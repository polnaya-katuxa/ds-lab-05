package openapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/polnaya-katuxa/ds-lab-02/cars-service/internal/generated/openapi"
	"github.com/polnaya-katuxa/ds-lab-02/cars-service/internal/models"
	"github.com/samber/lo"
)

type Server struct {
	carsLogic carsLogic
}

func New(carsLogic carsLogic) *Server {
	return &Server{
		carsLogic: carsLogic,
	}
}

func (s *Server) List(c echo.Context, params openapi.ListParams) error {
	list, err := s.carsLogic.List(c.Request().Context(), models.CarPaginator{
		Page:     int(lo.FromPtr(params.Page)),
		PageSize: int(lo.FromPtr(params.Size)),
		ShowAll:  lo.FromPtr(params.ShowAll),
	})
	if err != nil {
		return processError(c, err, "list cars")
	}

	return c.JSON(http.StatusOK, fromCarList(*list))
}

func (s *Server) Get(c echo.Context, carUid openapi_types.UUID) error {
	car, err := s.carsLogic.Get(c.Request().Context(), carUid)
	if err != nil {
		return processError(c, err, "get car")
	}

	return c.JSON(http.StatusOK, fromCar(*car))
}

func (s *Server) Book(c echo.Context, carUid openapi_types.UUID) error {
	car, err := s.carsLogic.Book(c.Request().Context(), carUid)
	if err != nil {
		return processError(c, err, "book car")
	}

	return c.JSON(http.StatusOK, fromCar(*car))
}

func (s *Server) Unbook(c echo.Context, carUid openapi_types.UUID) error {
	_, err := s.carsLogic.Unbook(c.Request().Context(), carUid)
	if err != nil {
		return processError(c, err, "unbook car")
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) Live(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

type carsLogic interface {
	List(ctx context.Context, paginator models.CarPaginator) (*models.CarList, error)
	Get(ctx context.Context, uid uuid.UUID) (*models.Car, error)
	Book(ctx context.Context, uid uuid.UUID) (*models.Car, error)
	Unbook(ctx context.Context, uid uuid.UUID) (*models.Car, error)
}
