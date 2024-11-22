package openapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/polnaya-katuxa/ds-lab-02/payment-service/internal/generated/openapi"
	"github.com/polnaya-katuxa/ds-lab-02/payment-service/internal/models"
)

func fromPayment(p models.Payment) openapi.PaymentInfo {
	return openapi.PaymentInfo{
		PaymentUid: p.UUID,
		Price:      p.Price,
		Status:     openapi.PaymentInfoStatus(p.Status),
	}
}

func processError(c echo.Context, err error, comment string) error {
	err = fmt.Errorf("%s: %w", comment, err)

	switch {
	case errors.Is(err, models.ErrInvalidPayment):
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
	case errors.Is(err, models.ErrPaymentNotFound):
		return c.JSON(http.StatusNotFound, openapi.ErrorResponse{
			Message: err.Error(),
		})
	default:
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Message: err.Error(),
		})
	}
}
