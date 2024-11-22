package logic

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/polnaya-katuxa/ds-lab-02/cars-service/internal/models"
)

type Cars struct {
	repo carsRepo
}

func New(repo carsRepo) *Cars {
	return &Cars{
		repo: repo,
	}
}

func (c *Cars) List(ctx context.Context, paginator models.CarPaginator) (*models.CarList, error) {
	err := paginator.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate paginator: %w", err)
	}

	list, err := c.repo.List(ctx, paginator)
	if err != nil {
		return nil, fmt.Errorf("get cars list from repo: %w", err)
	}

	return list, nil
}

func (c *Cars) Get(ctx context.Context, uid uuid.UUID) (*models.Car, error) {
	car, err := c.repo.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get car from repo: %w", err)
	}

	return car, nil
}

func (c *Cars) Book(ctx context.Context, uid uuid.UUID) (*models.Car, error) {
	car, err := c.repo.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get car from repo: %w", err)
	}

	if !car.Available {
		return nil, fmt.Errorf("check car availability: %w", models.ErrCarCantBeBooked)
	}

	car.Available = false
	err = c.repo.Update(ctx, car)
	if err != nil {
		return nil, fmt.Errorf("update car: %w", err)
	}

	return car, nil
}

func (c *Cars) Unbook(ctx context.Context, uid uuid.UUID) (*models.Car, error) {
	car, err := c.repo.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get car from repo: %w", err)
	}

	if car.Available {
		return nil, fmt.Errorf("check car availability: %w", models.ErrCarIsNotBooked)
	}

	car.Available = true
	err = c.repo.Update(ctx, car)
	if err != nil {
		return nil, fmt.Errorf("update car: %w", err)
	}

	return car, nil
}

//go:generate mockery --all --with-expecter --exported --output mocks/

type carsRepo interface {
	List(ctx context.Context, paginator models.CarPaginator) (*models.CarList, error)
	Get(ctx context.Context, uid uuid.UUID) (*models.Car, error)
	Update(ctx context.Context, car *models.Car) error
}
