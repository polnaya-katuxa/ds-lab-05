// Package openapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.3.0 DO NOT EDIT.
package openapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Defines values for RentalResponseStatus.
const (
	CANCELED   RentalResponseStatus = "CANCELED"
	FINISHED   RentalResponseStatus = "FINISHED"
	INPROGRESS RentalResponseStatus = "IN_PROGRESS"
)

// CreateRentalRequest defines model for CreateRentalRequest.
type CreateRentalRequest struct {
	// CarUid UUID автомобиля
	CarUid openapi_types.UUID `json:"carUid"`

	// DateFrom Дата начала аренды
	DateFrom string `json:"dateFrom"`

	// DateTo Дата окончания аренды
	DateTo string `json:"dateTo"`

	// PaymentUid UUID платежа
	PaymentUid openapi_types.UUID `json:"paymentUid"`
}

// ErrorDescription defines model for ErrorDescription.
type ErrorDescription struct {
	Error string `json:"error"`
	Field string `json:"field"`
}

// ErrorResponse defines model for ErrorResponse.
type ErrorResponse struct {
	// Message Информация об ошибке
	Message string `json:"message"`
}

// RentalResponse defines model for RentalResponse.
type RentalResponse struct {
	// CarUid UUID автомобиля
	CarUid openapi_types.UUID `json:"carUid"`

	// DateFrom Дата начала аренды
	DateFrom string `json:"dateFrom"`

	// DateTo Дата окончания аренды
	DateTo string `json:"dateTo"`

	// PaymentUid UUID платежа
	PaymentUid openapi_types.UUID `json:"paymentUid"`

	// RentalUid UUID аренды
	RentalUid openapi_types.UUID `json:"rentalUid"`

	// Status Статус аренды
	Status RentalResponseStatus `json:"status"`
}

// RentalResponseStatus Статус аренды
type RentalResponseStatus string

// ValidationErrorResponse defines model for ValidationErrorResponse.
type ValidationErrorResponse struct {
	// Errors Массив полей с описанием ошибки
	Errors []ErrorDescription `json:"errors"`

	// Message Информация об ошибке
	Message string `json:"message"`
}

// CreateJSONRequestBody defines body for Create for application/json ContentType.
type CreateJSONRequestBody = CreateRentalRequest

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Получить информацию о всех арендах пользователя
	// (GET /api/v1/rental)
	GetUserRentals(ctx echo.Context) error
	// Оформить аренду
	// (POST /api/v1/rental)
	Create(ctx echo.Context) error
	// Отмена аренды
	// (DELETE /api/v1/rental/{rentalUid})
	Cancel(ctx echo.Context, rentalUid openapi_types.UUID) error
	// Информация по конкретной аренде пользователя
	// (GET /api/v1/rental/{rentalUid})
	Get(ctx echo.Context, rentalUid openapi_types.UUID) error
	// Завершение аренды
	// (POST /api/v1/rental/{rentalUid}/finish)
	Finish(ctx echo.Context, rentalUid openapi_types.UUID) error
	// Liveness probe
	// (GET /manage/health)
	Live(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetUserRentals converts echo context to params.
func (w *ServerInterfaceWrapper) GetUserRentals(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetUserRentals(ctx)
	return err
}

// Create converts echo context to params.
func (w *ServerInterfaceWrapper) Create(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.Create(ctx)
	return err
}

// Cancel converts echo context to params.
func (w *ServerInterfaceWrapper) Cancel(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "rentalUid" -------------
	var rentalUid openapi_types.UUID

	err = runtime.BindStyledParameterWithOptions("simple", "rentalUid", ctx.Param("rentalUid"), &rentalUid, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter rentalUid: %s", err))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.Cancel(ctx, rentalUid)
	return err
}

// Get converts echo context to params.
func (w *ServerInterfaceWrapper) Get(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "rentalUid" -------------
	var rentalUid openapi_types.UUID

	err = runtime.BindStyledParameterWithOptions("simple", "rentalUid", ctx.Param("rentalUid"), &rentalUid, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter rentalUid: %s", err))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.Get(ctx, rentalUid)
	return err
}

// Finish converts echo context to params.
func (w *ServerInterfaceWrapper) Finish(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "rentalUid" -------------
	var rentalUid openapi_types.UUID

	err = runtime.BindStyledParameterWithOptions("simple", "rentalUid", ctx.Param("rentalUid"), &rentalUid, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter rentalUid: %s", err))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.Finish(ctx, rentalUid)
	return err
}

// Live converts echo context to params.
func (w *ServerInterfaceWrapper) Live(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.Live(ctx)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/api/v1/rental", wrapper.GetUserRentals)
	router.POST(baseURL+"/api/v1/rental", wrapper.Create)
	router.DELETE(baseURL+"/api/v1/rental/:rentalUid", wrapper.Cancel)
	router.GET(baseURL+"/api/v1/rental/:rentalUid", wrapper.Get)
	router.POST(baseURL+"/api/v1/rental/:rentalUid/finish", wrapper.Finish)
	router.GET(baseURL+"/manage/health", wrapper.Live)

}