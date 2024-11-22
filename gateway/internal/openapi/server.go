package openapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/auth"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/clients"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/generated/openapi"
	cars_service "github.com/polnaya-katuxa/ds-lab-02/gateway/internal/generated/openapi/clients/cars-service"
	payment_service "github.com/polnaya-katuxa/ds-lab-02/gateway/internal/generated/openapi/clients/payment-service"
	rental_service "github.com/polnaya-katuxa/ds-lab-02/gateway/internal/generated/openapi/clients/rental-service"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/models"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/repository/kafka/retryqueue"
	"github.com/samber/lo"
)

type Server struct {
	cars       *clients.CarsServiceClient
	payment    *clients.PaymentServiceClient
	rental     *clients.RentalServiceClient
	retryQueue *retryqueue.RetryQueueProducer
}

func New(
	cars *clients.CarsServiceClient,
	payment *clients.PaymentServiceClient,
	rental *clients.RentalServiceClient,
	retryQueue *retryqueue.RetryQueueProducer,
) *Server {
	return &Server{
		cars:       cars,
		payment:    payment,
		rental:     rental,
		retryQueue: retryQueue,
	}
}

func (s *Server) GetCars(c echo.Context, params openapi.GetCarsParams) error {
	cars, err := s.cars.List(c.Request().Context(), &cars_service.ListParams{
		Page:    lo.ToPtr(float32(lo.FromPtr(params.Page) - 1)),
		Size:    lo.ToPtr(float32(lo.FromPtr(params.Size))),
		ShowAll: params.ShowAll,
	})
	if err != nil {
		return processError(c, err, "list cars")
	}

	return c.JSON(http.StatusOK, cars)
}

func (s *Server) GetUserRentals(c echo.Context) error {
	rentals, err := s.rental.List(c.Request().Context(), auth.GetToken(c.Request().Context()))
	if err != nil {
		return processError(c, err, "list user rentals")
	}

	result := make([]openapi.RentalResponse, len(rentals))
	for i, rental := range rentals {
		car, err := s.cars.Get(c.Request().Context(), rental.CarUid)
		if err != nil {
			if isLogicError(c, err) {
				return processError(c, err, "get car info")
			}

			car = &cars_service.CarResponse{
				CarUid: rental.CarUid,
			}
		}

		payment, err := s.payment.Get(c.Request().Context(), rental.PaymentUid)
		if err != nil {
			if isLogicError(c, err) {
				return processError(c, err, "get payment info")
			}

			payment = &payment_service.PaymentInfo{
				PaymentUid: rental.PaymentUid,
			}
		}

		result[i] = openapi.RentalResponse{
			Car: openapi.CarInfo{
				Brand:              car.Brand,
				CarUid:             car.CarUid,
				Model:              car.Model,
				RegistrationNumber: car.RegistrationNumber,
			},
			DateFrom: rental.DateFrom,
			DateTo:   rental.DateTo,
			Payment: openapi.PaymentInfo{
				PaymentUid: payment.PaymentUid,
				Price:      payment.Price,
				Status:     openapi.PaymentInfoStatus(payment.Status),
			},
			RentalUid: rental.RentalUid,
			Status:    openapi.RentalResponseStatus(rental.Status),
		}
	}

	return c.JSON(http.StatusOK, result)
}

func (s *Server) revertBook(c echo.Context, carUid uuid.UUID) error {
	err := s.cars.Unbook(c.Request().Context(), carUid)
	if err != nil {
		return processError(c, err, "unbook car")
	}

	return nil
}

func (s *Server) revertPayment(c echo.Context, paymentUid uuid.UUID) error {
	err := s.payment.Cancel(c.Request().Context(), paymentUid)
	if err != nil {
		return processError(c, err, "cancel payment")
	}

	return nil
}

func (s *Server) BookCar(c echo.Context) error {
	var req openapi.BookCarJSONRequestBody
	err := json.NewDecoder(c.Request().Body).Decode(&req)
	if err != nil {
		return processError(c, err, "cannot unmarshal request body")
	}

	dateFrom, err := time.Parse(time.DateOnly, req.DateFrom)
	if err != nil {
		return processError(c, models.ValidationError{Message: err.Error()}, "parse date from")
	}

	dateTo, err := time.Parse(time.DateOnly, req.DateTo)
	if err != nil {
		return processError(c, models.ValidationError{Message: err.Error()}, "parse date to")
	}

	numDays := int(dateTo.Sub(dateFrom).Hours()) / 24
	if numDays < 1 {
		return processError(c, models.ValidationError{Message: "should rent min to 1 day"}, "check rent dates")
	}

	car, err := s.cars.Get(c.Request().Context(), req.CarUid)
	if err != nil {
		return processError(c, err, "get car")
	}

	_, err = s.cars.Book(c.Request().Context(), car.CarUid)
	if err != nil {
		return processError(c, err, "book car")
	}

	totalPrice := car.Price * numDays
	payment, err := s.payment.Create(c.Request().Context(), payment_service.CreatePaymentRequest{
		Price: totalPrice,
	})
	if err != nil {
		if !isLogicError(c, err) {
			revertErr := s.revertBook(c, car.CarUid)
			if revertErr != nil {
				return processError(c, revertErr, "revert book")
			}
		}
		return processAndHideError(c, err, "Payment Service unavailable")
	}

	rental, err := s.rental.Create(c.Request().Context(), auth.GetToken(c.Request().Context()), rental_service.CreateRentalRequest{
		CarUid:     car.CarUid,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
		PaymentUid: payment.PaymentUid,
	})
	if err != nil {
		if !isLogicError(c, err) {
			revertErr := s.revertBook(c, car.CarUid)
			if revertErr != nil {
				return processError(c, revertErr, "revert book")
			}

			revertErr = s.revertPayment(c, payment.PaymentUid)
			if revertErr != nil {
				return processError(c, revertErr, "revert payment")
			}
		}
		return processError(c, err, "create rental")
	}

	result := openapi.CreateRentalResponse{
		CarUid:   car.CarUid,
		DateFrom: rental.DateFrom,
		DateTo:   rental.DateTo,
		Payment: openapi.PaymentInfo{
			PaymentUid: payment.PaymentUid,
			Price:      payment.Price,
			Status:     openapi.PaymentInfoStatus(payment.Status),
		},
		RentalUid: rental.RentalUid,
		Status:    openapi.CreateRentalResponseStatus(rental.Status),
	}

	return c.JSON(http.StatusOK, result)
}

func (s *Server) CancelRental(c echo.Context, rentalUid openapi_types.UUID) error {
	rental, err := s.rental.Get(c.Request().Context(), auth.GetToken(c.Request().Context()), rentalUid)
	if err != nil {
		return processError(c, err, "get user rental")
	}

	err = s.cars.Unbook(c.Request().Context(), rental.CarUid)
	if err != nil {
		if isUnavailableError(c, err) {
			s.retryQueue.RetryCarUnbook(rental.CarUid)
		} else {
			return processError(c, err, "make car available")
		}
	}

	err = s.rental.Cancel(c.Request().Context(), auth.GetToken(c.Request().Context()), rentalUid)
	if err != nil {
		return processError(c, err, "cancel rental")
	}

	err = s.payment.Cancel(c.Request().Context(), rental.PaymentUid)
	if err != nil {
		if isUnavailableError(c, err) {
			s.retryQueue.RetryPaymentCancel(rental.PaymentUid)
		} else {
			return processError(c, err, "cancel payment")
		}
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) GetUserRental(c echo.Context, rentalUid openapi_types.UUID) error {
	rental, err := s.rental.Get(c.Request().Context(), auth.GetToken(c.Request().Context()), rentalUid)
	if err != nil {
		return processError(c, err, "get user rental")
	}

	car, err := s.cars.Get(c.Request().Context(), rental.CarUid)
	if err != nil {
		if isLogicError(c, err) {
			return processError(c, err, "get car info")
		}

		car = &cars_service.CarResponse{
			CarUid: rental.CarUid,
		}
	}

	payment, err := s.payment.Get(c.Request().Context(), rental.PaymentUid)
	if err != nil {
		if isLogicError(c, err) {
			return processError(c, err, "get payment info")
		}

		payment = &payment_service.PaymentInfo{
			PaymentUid: rental.PaymentUid,
		}
	}

	result := openapi.RentalResponse{
		Car: openapi.CarInfo{
			Brand:              car.Brand,
			CarUid:             car.CarUid,
			Model:              car.Model,
			RegistrationNumber: car.RegistrationNumber,
		},
		DateFrom: rental.DateFrom,
		DateTo:   rental.DateTo,
		Payment: openapi.PaymentInfo{
			PaymentUid: payment.PaymentUid,
			Price:      payment.Price,
			Status:     openapi.PaymentInfoStatus(payment.Status),
		},
		RentalUid: rental.RentalUid,
		Status:    openapi.RentalResponseStatus(rental.Status),
	}

	return c.JSON(http.StatusOK, result)
}

func (s *Server) FinishRental(c echo.Context, rentalUid openapi_types.UUID) error {
	rental, err := s.rental.Get(c.Request().Context(), auth.GetToken(c.Request().Context()), rentalUid)
	if err != nil {
		return processError(c, err, "get user rental")
	}

	err = s.cars.Unbook(c.Request().Context(), rental.CarUid)
	if err != nil {
		if isUnavailableError(c, err) {
			s.retryQueue.RetryCarUnbook(rental.CarUid)
		} else {
			return processError(c, err, "make car available")
		}
	}

	err = s.rental.Finish(c.Request().Context(), auth.GetToken(c.Request().Context()), rentalUid)
	if err != nil {
		return processError(c, err, "finish rental")
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) Live(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
