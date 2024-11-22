package models

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	ErrCarCantBeBooked = errors.New("car is not available")
	ErrCarIsNotBooked  = errors.New("car was not booked")
	ErrInvalidData     = errors.New("invalid data")
	ErrCarNotFound     = errors.New("car not found")
)

type CarType string

const (
	Minivan  CarType = "MINIVAN"
	Roadster CarType = "ROADSTER"
	Sedan    CarType = "SEDAN"
	SUV      CarType = "SUV"
)

type Car struct {
	ID   int       `gorm:"column:id;primaryKey"`
	UUID uuid.UUID `gorm:"column:car_uid;type:uuid" json:"car_uid"`

	Available          bool    `gorm:"column:availability"`
	Brand              string  `gorm:"column:brand"`
	Model              string  `gorm:"column:model"`
	Power              *int    `gorm:"column:power"`
	Price              int     `gorm:"column:price"`
	RegistrationNumber string  `gorm:"column:registration_number"`
	Type               CarType `gorm:"column:type"`
}

type CarList struct {
	Items    []Car
	Total    int
	Page     int
	PageSize int
}

type CarPaginator struct {
	Page     int `validate:"omitempty,gte=0"`
	PageSize int `validate:"omitempty,gte=0"`
	ShowAll  bool
}

func (p *CarPaginator) Validate() error {
	err := validator.New().Struct(p)
	if err != nil {
		return fmt.Errorf("validate paginator: %w (%w)", err, ErrInvalidData)
	}

	return nil
}
