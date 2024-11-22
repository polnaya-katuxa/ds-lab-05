package openapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/polnaya-katuxa/ds-lab-02/payment-service/internal/generated/openapi"
	"github.com/polnaya-katuxa/ds-lab-02/payment-service/internal/models"
)

type Server struct {
	paymentLogic paymentLogic
}

func New(paymentLogic paymentLogic) *Server {
	return &Server{
		paymentLogic: paymentLogic,
	}
}

func (s *Server) Create(c echo.Context) error {
	var req openapi.CreatePaymentRequest
	err := json.NewDecoder(c.Request().Body).Decode(&req)
	if err != nil {
		return processError(c, err, "cannot unmarshal request body")
	}

	payment, err := s.paymentLogic.Create(c.Request().Context(), models.CreatePaymentRequest{
		Price: int(req.Price),
	})
	if err != nil {
		return processError(c, err, "create payment")
	}

	return c.JSON(http.StatusOK, fromPayment(*payment))
}

func (s *Server) Cancel(c echo.Context, paymentUid openapi_types.UUID) error {
	err := s.paymentLogic.Cancel(c.Request().Context(), paymentUid)
	if err != nil {
		return processError(c, err, "cancel payment")
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) Get(c echo.Context, paymentUid openapi_types.UUID) error {
	payment, err := s.paymentLogic.Get(c.Request().Context(), paymentUid)
	if err != nil {
		return processError(c, err, "get payment")
	}

	return c.JSON(http.StatusOK, fromPayment(*payment))
}

func (s *Server) Live(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

type paymentLogic interface {
	Create(ctx context.Context, req models.CreatePaymentRequest) (*models.Payment, error)
	Cancel(ctx context.Context, uid uuid.UUID) error
	Get(ctx context.Context, uid uuid.UUID) (*models.Payment, error)
}
