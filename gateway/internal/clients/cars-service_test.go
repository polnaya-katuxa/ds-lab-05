package clients

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	cars_service "github.com/polnaya-katuxa/ds-lab-02/gateway/internal/generated/openapi/clients/cars-service"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/generated/openapi/clients/cars-service/mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

func TestPaymentsLogic_Get(t *testing.T) {
	t.Run("got car", func(t *testing.T) {
		ctx := context.Background()

		uuid := uuid.MustParse("1bda4472-e536-4d74-b1e0-8f027aebf972")
		want := &cars_service.CarResponse{
			CarUid:             uuid,
			Available:          false,
			Brand:              "Mercedes Benz",
			Model:              "GLA 250",
			Power:              lo.ToPtr(450),
			Price:              1000,
			RegistrationNumber: "ЛО777Х799",
			Type:               "SEDAN",
		}

		client := mocks.NewClientInterface(t)
		data := `
		{
			"carUid": "1bda4472-e536-4d74-b1e0-8f027aebf972",
			"available": false,
			"brand": "Mercedes Benz",
			"model": "GLA 250",
			"power": 450,
			"price": 1000,
			"registrationNumber": "ЛО777Х799",
			"type": "SEDAN"
		}
		`
		buf := bytes.NewBufferString(data)
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(buf),
		}

		client.EXPECT().Get(ctx, uuid).Return(resp, nil)

		c := NewCarsServiceClient(client)
		got, err := c.Get(ctx, uuid)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := context.Background()

		client := mocks.NewClientInterface(t)
		data := `
		{
			"message": "service error"
		}
		`
		buf := bytes.NewBufferString(data)
		resp := &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(buf),
		}

		client.EXPECT().Get(ctx, uuid.UUID{}).Return(resp, nil)

		c := NewCarsServiceClient(client)
		_, err := c.Get(ctx, uuid.UUID{})
		require.Error(t, err)
	})

	t.Run("network error", func(t *testing.T) {
		ctx := context.Background()

		client := mocks.NewClientInterface(t)

		client.EXPECT().Get(ctx, uuid.UUID{}).Return(nil, errors.New("some error"))

		c := NewCarsServiceClient(client)
		_, err := c.Get(ctx, uuid.UUID{})
		require.Error(t, err)
	})
}
