package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/polnaya-katuxa/ds-lab-02/payment-service/internal/models"
	"gorm.io/gorm"
)

type Payment struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Payment {
	return &Payment{
		db: db,
	}
}

func (p *Payment) Get(ctx context.Context, uid uuid.UUID) (*models.Payment, error) {
	var payment models.Payment

	err := p.db.Table("payment").WithContext(ctx).First(&payment, "payment_uid = ?", uid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get payment from db: %w", models.ErrPaymentNotFound)
		}

		return nil, fmt.Errorf("get payment from db: %w", err)
	}

	return &payment, nil
}

func (p *Payment) Create(ctx context.Context, payment models.Payment) (*models.Payment, error) {
	err := p.db.Table("payment").WithContext(ctx).Create(&payment).Error
	if err != nil {
		return nil, fmt.Errorf("create payment in db: %w", err)
	}

	return &payment, nil
}

func (p *Payment) ChangeStatus(ctx context.Context, uid uuid.UUID, status models.PaymentStatus) error {
	res := p.db.Table("payment").WithContext(ctx).Where("payment_uid = ?", uid).Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("update payment in db: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("update payment in db: %w", models.ErrPaymentNotFound)
	}

	return nil
}
