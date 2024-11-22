package logic

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/polnaya-katuxa/ds-lab-02/payment-service/internal/logic/mocks"
	"github.com/polnaya-katuxa/ds-lab-02/payment-service/internal/models"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

func TestPaymentsLogic_Get(t *testing.T) {
	t.Run("got payment", func(t *testing.T) {
		ctx := context.Background()

		id := 1
		uuid := uuid.New()
		want := &models.Payment{
			ID:     id,
			UUID:   uuid,
			Price:  1000,
			Status: "PAID",
		}

		repository := mocks.NewPaymentRepo(t)
		repository.EXPECT().Get(ctx, uuid).Return(want, nil)

		p := New(repository)
		got, err := p.Get(ctx, uuid)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx := context.Background()

		uuid := uuid.New()
		repository := mocks.NewPaymentRepo(t)
		repository.EXPECT().Get(ctx, uuid).Return(nil, errors.New("error"))

		p := New(repository)
		got, err := p.Get(ctx, uuid)
		require.Error(t, err)
		require.Nil(t, got)
	})
}
