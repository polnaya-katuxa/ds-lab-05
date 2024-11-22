package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	Minivan  CarResponseType = "MINIVAN"
	Roadster CarResponseType = "ROADSTER"
	Sedan    CarResponseType = "SEDAN"
	SUV      CarResponseType = "SUV"
)

type PaginationResponse struct {
	Items         []CarResponse `json:"items"`
	Page          int           `json:"page"`
	PageSize      int           `json:"pageSize"`
	TotalElements int           `json:"totalElements"`
}

type CarResponse struct {
	Available          bool            `json:"available"`
	Brand              string          `json:"brand"`
	CarUid             uuid.UUID       `json:"carUid"`
	Model              string          `json:"model"`
	Power              *int            `json:"power,omitempty"`
	Price              int             `json:"price"`
	RegistrationNumber string          `json:"registrationNumber"`
	Type               CarResponseType `json:"type"`
}

type CarResponseType string

type CarUnbookRetryMsg struct {
	CarUid        uuid.UUID
	LastProcessed time.Time
}

type PaymentCancelRetryMsg struct {
	PaymentUid    uuid.UUID
	LastProcessed time.Time
}
