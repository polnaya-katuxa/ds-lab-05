package openapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/auth"
	"github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/generated/openapi"
	"github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/models"
	"github.com/samber/lo"
)

type Server struct {
	rentalLogic rentalLogic
}

func New(rentalLogic rentalLogic) *Server {
	return &Server{
		rentalLogic: rentalLogic,
	}
}

func (s *Server) GetUserRentals(c echo.Context) error {
	rents, err := s.rentalLogic.GetUserRentals(c.Request().Context(), auth.GetUsername(c.Request().Context()))
	if err != nil {
		return processError(c, err, "get user rentals")
	}

	return c.JSON(http.StatusOK, lo.Map(rents, func(r models.Rent, _ int) openapi.RentalResponse {
		return fromRent(r)
	}))
}

func (s *Server) Create(c echo.Context) error {
	var req openapi.CreateRentalRequest
	err := json.NewDecoder(c.Request().Body).Decode(&req)
	if err != nil {
		return processError(c, err, "cannot unmarshal request body")
	}

	logicReq, err := toRentCreateRequest(req, auth.GetUsername(c.Request().Context()))
	if err != nil {
		return processError(c, err, "validate request data")
	}

	rent, err := s.rentalLogic.Create(c.Request().Context(), *logicReq)
	if err != nil {
		return processError(c, err, "create rent")
	}

	return c.JSON(http.StatusCreated, fromRent(*rent))
}

func (s *Server) Cancel(c echo.Context, rentalUid openapi_types.UUID) error {
	err := s.rentalLogic.Cancel(c.Request().Context(), rentalUid, auth.GetUsername(c.Request().Context()))
	if err != nil {
		return processError(c, err, "cancel rent")
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) Get(c echo.Context, rentalUid openapi_types.UUID) error {
	rent, err := s.rentalLogic.Get(c.Request().Context(), rentalUid, auth.GetUsername(c.Request().Context()))
	if err != nil {
		return processError(c, err, "get rent")
	}

	return c.JSON(http.StatusOK, fromRent(*rent))
}

func (s *Server) Finish(c echo.Context, rentalUid openapi_types.UUID) error {
	err := s.rentalLogic.Finish(c.Request().Context(), rentalUid, auth.GetUsername(c.Request().Context()))
	if err != nil {
		return processError(c, err, "finish rent")
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) Live(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

type rentalLogic interface {
	GetUserRentals(ctx context.Context, username string) ([]models.Rent, error)
	Create(ctx context.Context, req models.CreateRentRequest) (*models.Rent, error)
	Cancel(ctx context.Context, uid uuid.UUID, username string) error
	Finish(ctx context.Context, uid uuid.UUID, username string) error
	Get(ctx context.Context, uid uuid.UUID, username string) (*models.Rent, error)
}
