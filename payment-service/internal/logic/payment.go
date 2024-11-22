package logic

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/polnaya-katuxa/ds-lab-02/payment-service/internal/models"
)

type Payment struct {
	repo paymentRepo
}

func New(repo paymentRepo) *Payment {
	return &Payment{
		repo: repo,
	}
}

func (p *Payment) Create(ctx context.Context, req models.CreatePaymentRequest) (*models.Payment, error) {
	err := req.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate request: %w", err)
	}

	paymentToCreate := models.Payment{
		UUID:   uuid.New(),
		Price:  req.Price,
		Status: models.Paid,
	}

	payment, err := p.repo.Create(ctx, paymentToCreate)
	if err != nil {
		return nil, fmt.Errorf("create payment in repo: %w", err)
	}

	return payment, nil
}

func (p *Payment) Cancel(ctx context.Context, uid uuid.UUID) error {
	err := p.repo.ChangeStatus(ctx, uid, models.Canceled)
	if err != nil {
		return fmt.Errorf("change payment status: %w", err)
	}

	return nil
}

func (p *Payment) Get(ctx context.Context, uid uuid.UUID) (*models.Payment, error) {
	payment, err := p.repo.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get payment from repo: %w", err)
	}

	return payment, nil
}

//go:generate mockery --all --with-expecter --exported --output mocks/

type paymentRepo interface {
	Get(ctx context.Context, uid uuid.UUID) (*models.Payment, error)
	Create(ctx context.Context, payment models.Payment) (*models.Payment, error)
	ChangeStatus(ctx context.Context, uid uuid.UUID, status models.PaymentStatus) error
}
