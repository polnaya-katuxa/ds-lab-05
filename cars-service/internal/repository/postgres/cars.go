package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/polnaya-katuxa/ds-lab-02/cars-service/internal/models"
	"gorm.io/gorm"
)

type Cars struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Cars {
	return &Cars{db: db}
}

func (c *Cars) List(ctx context.Context, paginator models.CarPaginator) (*models.CarList, error) {
	var cars []models.Car
	var total int64

	query := c.db.Table("cars").WithContext(ctx).Offset(paginator.Page * paginator.PageSize).Limit(paginator.PageSize)
	if !paginator.ShowAll {
		query = query.Where("availability = true")
	}

	err := query.Count(&total).Find(&cars).Error
	if err != nil {
		return nil, fmt.Errorf("find cars in db: %w", err)
	}

	return &models.CarList{
		Items:    cars,
		Total:    int(total),
		Page:     paginator.Page,
		PageSize: paginator.PageSize,
	}, nil
}

func (c *Cars) Get(ctx context.Context, uid uuid.UUID) (*models.Car, error) {
	var car models.Car

	err := c.db.Table("cars").WithContext(ctx).First(&car, "car_uid = ?", uid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get car from db: %w", models.ErrCarNotFound)
		}

		return nil, fmt.Errorf("get car from db: %w", err)
	}

	return &car, nil
}

func (c *Cars) Update(ctx context.Context, car *models.Car) error {
	res := c.db.Table("cars").WithContext(ctx).Save(car)
	if res.Error != nil {
		return fmt.Errorf("update car in db: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("update car in db: %w", models.ErrCarNotFound)
	}

	return nil
}
