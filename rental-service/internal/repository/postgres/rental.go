package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/models"
	"gorm.io/gorm"
)

type Rental struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Rental {
	return &Rental{
		db: db,
	}
}

func (r *Rental) Get(ctx context.Context, uid uuid.UUID) (*models.Rent, error) {
	var rent models.Rent

	err := r.db.Table("rental").WithContext(ctx).First(&rent, "rental_uid = ?", uid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get rental from db: %w", models.ErrRentNotFound)
		}

		return nil, fmt.Errorf("get rental from db: %w", err)
	}

	return &rent, nil
}

func (r *Rental) GetUserRentals(ctx context.Context, username string) ([]models.Rent, error) {
	var rents []models.Rent

	err := r.db.Table("rental").WithContext(ctx).Where("username = ?", username).Find(&rents).Error
	if err != nil {
		return nil, fmt.Errorf("find rentals in db: %w", err)
	}

	return rents, nil
}

func (r *Rental) Create(ctx context.Context, rent models.Rent) (*models.Rent, error) {
	err := r.db.Table("rental").WithContext(ctx).Create(&rent).Error
	if err != nil {
		return nil, fmt.Errorf("create rental in db: %w", err)
	}

	return &rent, nil
}

func (r *Rental) ChangeStatus(ctx context.Context, uid uuid.UUID, status models.RentStatus) error {
	res := r.db.Table("rental").WithContext(ctx).Where("rental_uid = ?", uid).Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("update rental in db: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("update rental in db: %w", models.ErrRentNotFound)
	}

	return nil
}
