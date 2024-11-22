package logic

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/polnaya-katuxa/ds-lab-02/cars-service/internal/logic/mocks"
	"github.com/polnaya-katuxa/ds-lab-02/cars-service/internal/models"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

func TestCarsLogic_Get(t *testing.T) {
	t.Run("got car", func(t *testing.T) {
		ctx := context.Background()

		id := 1
		uuid := uuid.New()
		want := &models.Car{
			ID:                 id,
			UUID:               uuid,
			Available:          false,
			Brand:              "Mercedes Benz",
			Model:              "GLA 250",
			Power:              &id,
			Price:              1000,
			RegistrationNumber: "ЛО777Х799",
			Type:               models.Sedan,
		}

		repository := mocks.NewCarsRepo(t)
		repository.EXPECT().Get(ctx, uuid).Return(want, nil)

		p := New(repository)
		got, err := p.Get(ctx, uuid)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx := context.Background()

		uuid := uuid.New()
		repository := mocks.NewCarsRepo(t)
		repository.EXPECT().Get(ctx, uuid).Return(nil, errors.New("error"))

		p := New(repository)
		got, err := p.Get(ctx, uuid)
		require.Error(t, err)
		require.Nil(t, got)
	})
}
