package models

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidPayment  = errors.New("invalid payment")
)

type PaymentStatus string

const (
	Paid     PaymentStatus = "PAID"
	Canceled PaymentStatus = "CANCELED"
)

type Payment struct {
	ID     int           `gorm:"column:id;primaryKey"`
	UUID   uuid.UUID     `gorm:"column:payment_uid;type:uuid"`
	Price  int           `gorm:"column:price"`
	Status PaymentStatus `gorm:"column:status"`
}

type CreatePaymentRequest struct {
	Price int `gorm:"column:price" validate:"omitempty,gte=0"`
}

func (r *CreatePaymentRequest) Validate() error {
	err := validator.New().Struct(r)
	if err != nil {
		return fmt.Errorf("validate create payment: %w (%w)", err, ErrInvalidPayment)
	}

	return nil
}
