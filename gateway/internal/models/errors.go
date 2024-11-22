package models

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownResponseStatus = errors.New("unknown response status")
)

type ValidationError struct {
	Errors  []ErrorDescription `json:"errors"`
	Message string             `json:"message"`
}

type ErrorDescription struct {
	Error string `json:"error"`
	Field string `json:"field"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %+v", e.Message, e.Errors)
}

type InternalError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e InternalError) Error() string {
	return e.Message
}
