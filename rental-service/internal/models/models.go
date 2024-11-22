package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	ErrRentNotFound = errors.New("payment not found")
	ErrInvalidRent  = errors.New("invalid rent")
	ErrForbidden    = errors.New("forbidden")
)

type RentStatus string

const (
	InProgress RentStatus = "IN_PROGRESS"
	Finished   RentStatus = "FINISHED"
	Canceled   RentStatus = "CANCELED"
)

type Rent struct {
	ID          int        `gorm:"column:id;primaryKey"`
	UUID        uuid.UUID  `gorm:"column:rental_uid;type:uuid"`
	Username    string     `gorm:"username:price"`
	PaymentUUID uuid.UUID  `gorm:"column:payment_uid;type:uuid"`
	CarUUID     uuid.UUID  `gorm:"column:car_uid;type:uuid"`
	DateFrom    time.Time  `gorm:"column:date_from;type:timestamptz"`
	DateTo      time.Time  `gorm:"column:date_to;type:timestamptz"`
	Status      RentStatus `gorm:"column:status"`
}

type CreateRentRequest struct {
	Username    string    `validate:"required"`
	PaymentUUID uuid.UUID `validate:"required"`
	CarUUID     uuid.UUID `validate:"required"`
	DateFrom    time.Time `validate:"required"`
	DateTo      time.Time `validate:"required"`
}

func (r *CreateRentRequest) Validate() error {
	err := validator.New().Struct(r)
	if err != nil {
		return fmt.Errorf("validate create rent: %w (%w)", err, ErrInvalidRent)
	}

	return nil
}
