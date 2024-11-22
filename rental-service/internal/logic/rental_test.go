package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/logic/mocks"
	"github.com/polnaya-katuxa/ds-lab-02/rental-service/internal/models"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

func TestRentalLogic_Get(t *testing.T) {
	t.Run("got rental", func(t *testing.T) {
		ctx := context.Background()

		id := 1
		uuid := uuid.New()
		want := &models.Rent{
			ID:          id,
			UUID:        uuid,
			Username:    "user",
			PaymentUUID: uuid,
			CarUUID:     uuid,
			DateFrom:    time.Now(),
			DateTo:      time.Now(),
			Status:      "PAID",
		}

		repository := mocks.NewRentalRepo(t)
		repository.EXPECT().Get(ctx, uuid).Return(want, nil)

		p := New(repository)
		got, err := p.Get(ctx, uuid, want.Username)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx := context.Background()

		uuid := uuid.New()
		repository := mocks.NewRentalRepo(t)
		repository.EXPECT().Get(ctx, uuid).Return(nil, errors.New("error"))

		p := New(repository)
		got, err := p.Get(ctx, uuid, "user")
		require.Error(t, err)
		require.Nil(t, got)
	})
}
