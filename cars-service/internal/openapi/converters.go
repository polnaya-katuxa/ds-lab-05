package openapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/polnaya-katuxa/ds-lab-02/cars-service/internal/generated/openapi"
	"github.com/polnaya-katuxa/ds-lab-02/cars-service/internal/models"
	"github.com/samber/lo"
)

func fromCarList(list models.CarList) openapi.PaginationResponse {
	items := lo.Map(list.Items, func(car models.Car, _ int) openapi.CarResponse {
		return fromCar(car)
	})

	return openapi.PaginationResponse{
		Items:         items,
		Page:          list.Page,
		PageSize:      list.PageSize,
		TotalElements: list.Total,
	}
}

func fromCar(car models.Car) openapi.CarResponse {
	return openapi.CarResponse{
		Available:          car.Available,
		Brand:              car.Brand,
		CarUid:             car.UUID,
		Model:              car.Model,
		Power:              car.Power,
		Price:              car.Price,
		RegistrationNumber: car.RegistrationNumber,
		Type:               openapi.CarResponseType(car.Type),
	}
}

func processError(c echo.Context, err error, comment string) error {
	err = fmt.Errorf("%s: %w", comment, err)

	switch {
	case errors.Is(err, models.ErrInvalidData):
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
	case errors.Is(err, models.ErrCarNotFound):
		return c.JSON(http.StatusNotFound, openapi.ErrorResponse{
			Message: err.Error(),
		})
	case errors.Is(err, models.ErrCarIsNotBooked), errors.Is(err, models.ErrCarCantBeBooked):
		return c.JSON(http.StatusConflict, openapi.ErrorResponse{
			Message: err.Error(),
		})
	default:
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Message: err.Error(),
		})
	}
}
