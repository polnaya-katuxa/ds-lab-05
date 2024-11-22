package openapi

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/generated/openapi"
	"github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/models"
)

func fromRent(r models.Rent) openapi.RentalResponse {
	return openapi.RentalResponse{
		CarUid:     r.CarUUID,
		DateFrom:   r.DateFrom.Format(time.DateOnly),
		DateTo:     r.DateTo.Format(time.DateOnly),
		PaymentUid: r.PaymentUUID,
		RentalUid:  r.UUID,
		Status:     openapi.RentalResponseStatus(r.Status),
	}
}

func toRentCreateRequest(r openapi.CreateRentalRequest, username string) (*models.CreateRentRequest, error) {
	dateFrom, err := time.Parse(time.DateOnly, r.DateFrom)
	if err != nil {
		return nil, fmt.Errorf("invalid date from (%w): %w", models.ErrInvalidRent, err)
	}

	dateTo, err := time.Parse(time.DateOnly, r.DateTo)
	if err != nil {
		return nil, fmt.Errorf("invalid date to (%w): %w", models.ErrInvalidRent, err)
	}

	return &models.CreateRentRequest{
		Username:    username,
		PaymentUUID: r.PaymentUid,
		CarUUID:     r.CarUid,
		DateFrom:    dateFrom,
		DateTo:      dateTo,
	}, nil
}

func processError(c echo.Context, err error, comment string) error {
	err = fmt.Errorf("%s: %w", comment, err)

	switch {
	case errors.Is(err, models.ErrInvalidRent):
		var valErrors validator.ValidationErrors
		if errors.As(err, &valErrors) {
			errorSlice := make([]openapi.ErrorDescription, 0, len(valErrors))
			for _, v := range valErrors {
				errorSlice = append(errorSlice, openapi.ErrorDescription{
					Error: v.Error(),
					Field: v.Field(),
				})
			}

			return c.JSON(http.StatusBadRequest, openapi.ValidationErrorResponse{
				Message: err.Error(),
				Errors:  errorSlice,
			})
		}
		return c.JSON(http.StatusBadRequest, openapi.ValidationErrorResponse{
			Message: err.Error(),
		})
	case errors.Is(err, models.ErrRentNotFound):
		return c.JSON(http.StatusNotFound, openapi.ErrorResponse{
			Message: err.Error(),
		})
	case errors.Is(err, models.ErrForbidden):
		return c.JSON(http.StatusForbidden, openapi.ErrorResponse{
			Message: err.Error(),
		})
	default:
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Message: err.Error(),
		})
	}
}
