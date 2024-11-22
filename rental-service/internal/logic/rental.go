package logic

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/models"
)

type Rental struct {
	repo rentalRepo
}

func New(repo rentalRepo) *Rental {
	return &Rental{
		repo: repo,
	}
}

func (r *Rental) GetUserRentals(ctx context.Context, username string) ([]models.Rent, error) {
	rents, err := r.repo.GetUserRentals(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user rentals: %w", err)
	}

	return rents, nil
}

func (r *Rental) Create(ctx context.Context, req models.CreateRentRequest) (*models.Rent, error) {
	err := req.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate request: %w", err)
	}

	rentToCreate := models.Rent{
		UUID:        uuid.New(),
		Username:    req.Username,
		PaymentUUID: req.PaymentUUID,
		CarUUID:     req.CarUUID,
		DateFrom:    req.DateFrom,
		DateTo:      req.DateTo,
		Status:      models.InProgress,
	}

	rent, err := r.repo.Create(ctx, rentToCreate)
	if err != nil {
		return nil, fmt.Errorf("create rent: %w", err)
	}

	return rent, nil
}

func (r *Rental) Cancel(ctx context.Context, uid uuid.UUID, username string) error {
	rent, err := r.repo.Get(ctx, uid)
	if err != nil {
		return fmt.Errorf("get rent: %w", err)
	}

	if rent.Username != username {
		return fmt.Errorf("check user: %w", models.ErrForbidden)
	}

	err = r.repo.ChangeStatus(ctx, uid, models.Canceled)
	if err != nil {
		return fmt.Errorf("change rent status: %w", err)
	}

	return nil
}

func (r *Rental) Finish(ctx context.Context, uid uuid.UUID, username string) error {
	rent, err := r.repo.Get(ctx, uid)
	if err != nil {
		return fmt.Errorf("get rent: %w", err)
	}

	if rent.Username != username {
		return fmt.Errorf("check user: %w", models.ErrForbidden)
	}

	err = r.repo.ChangeStatus(ctx, uid, models.Finished)
	if err != nil {
		return fmt.Errorf("change rent status: %w", err)
	}

	return nil
}

func (r *Rental) Get(ctx context.Context, uid uuid.UUID, username string) (*models.Rent, error) {
	rent, err := r.repo.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get rent: %w", err)
	}

	if rent.Username != username {
		return nil, fmt.Errorf("check user: %w", models.ErrForbidden)
	}

	return rent, nil
}

//go:generate mockery --all --with-expecter --exported --output mocks/

type rentalRepo interface {
	Get(ctx context.Context, uid uuid.UUID) (*models.Rent, error)
	GetUserRentals(ctx context.Context, username string) ([]models.Rent, error)
	Create(ctx context.Context, rent models.Rent) (*models.Rent, error)
	ChangeStatus(ctx context.Context, uid uuid.UUID, status models.RentStatus) error
}
