package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	rental_service "github.com/polnaya-katuxa/ds-lab-02/gateway/internal/generated/openapi/clients/rental-service"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/models"
)

type RentalServiceClient struct {
	c *rental_service.Client
}

func NewRentalServiceClient(c *rental_service.Client) *RentalServiceClient {
	return &RentalServiceClient{
		c: c,
	}
}

func (c *RentalServiceClient) List(ctx context.Context, userName string) ([]rental_service.RentalResponse, error) {
	resp, err := c.c.GetUserRentals(ctx, withToken(ctx))
	if err != nil {
		return nil, fmt.Errorf("list user rentals: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		var internalError models.InternalError
		err := json.Unmarshal(body, &internalError)
		if err != nil {
			return nil, fmt.Errorf("parse service error: %w", err)
		}

		return nil, internalError
	case http.StatusOK:
		var rentals []rental_service.RentalResponse
		err := json.Unmarshal(body, &rentals)
		if err != nil {
			return nil, fmt.Errorf("parse rentals info: %w", err)
		}

		return rentals, nil
	default:
		return nil, fmt.Errorf("unknown response %d: %w", resp.StatusCode, models.ErrUnknownResponseStatus)
	}
}

func (c *RentalServiceClient) Create(ctx context.Context, userName string, req rental_service.CreateRentalRequest) (*rental_service.RentalResponse, error) {
	resp, err := c.c.Create(ctx, req, withToken(ctx))
	if err != nil {
		return nil, fmt.Errorf("create user rental: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		var internalError models.InternalError
		err := json.Unmarshal(body, &internalError)
		if err != nil {
			return nil, fmt.Errorf("parse service error: %w", err)
		}

		internalError.StatusCode = resp.StatusCode

		return nil, internalError
	case http.StatusBadRequest:
		var serviceError models.ValidationError
		err := json.Unmarshal(body, &serviceError)
		if err != nil {
			return nil, fmt.Errorf("parse service error: %w", err)
		}

		return nil, serviceError
	case http.StatusCreated:
		var rental rental_service.RentalResponse
		err := json.Unmarshal(body, &rental)
		if err != nil {
			return nil, fmt.Errorf("parse rentals info: %w", err)
		}

		return &rental, nil
	default:
		return nil, fmt.Errorf("unknown response %d: %w", resp.StatusCode, models.ErrUnknownResponseStatus)
	}
}

func (c *RentalServiceClient) Get(ctx context.Context, userName string, rentalUid uuid.UUID) (*rental_service.RentalResponse, error) {
	resp, err := c.c.Get(ctx, rentalUid, withToken(ctx))
	if err != nil {
		return nil, fmt.Errorf("get user rental: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusInternalServerError, http.StatusForbidden, http.StatusNotFound:
		var internalError models.InternalError
		err := json.Unmarshal(body, &internalError)
		if err != nil {
			return nil, fmt.Errorf("parse service error: %w", err)
		}

		internalError.StatusCode = resp.StatusCode

		return nil, internalError
	case http.StatusOK:
		var rental rental_service.RentalResponse
		err := json.Unmarshal(body, &rental)
		if err != nil {
			return nil, fmt.Errorf("parse rentals info: %w", err)
		}

		return &rental, nil
	default:
		return nil, fmt.Errorf("unknown response %d: %w", resp.StatusCode, models.ErrUnknownResponseStatus)
	}
}

func (c *RentalServiceClient) Cancel(ctx context.Context, userName string, rentalUid uuid.UUID) error {
	resp, err := c.c.Cancel(ctx, rentalUid, withToken(ctx))
	if err != nil {
		return fmt.Errorf("cancel user rental: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusInternalServerError, http.StatusForbidden, http.StatusNotFound:
		var internalError models.InternalError
		err := json.Unmarshal(body, &internalError)
		if err != nil {
			return fmt.Errorf("parse service error: %w", err)
		}

		internalError.StatusCode = resp.StatusCode

		return internalError
	case http.StatusNoContent:
		return nil
	default:
		return fmt.Errorf("unknown response %d: %w", resp.StatusCode, models.ErrUnknownResponseStatus)
	}
}

func (c *RentalServiceClient) Finish(ctx context.Context, userName string, rentalUid uuid.UUID) error {
	resp, err := c.c.Finish(ctx, rentalUid, withToken(ctx))
	if err != nil {
		return fmt.Errorf("finish user rental: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusInternalServerError, http.StatusForbidden, http.StatusNotFound:
		var internalError models.InternalError
		err := json.Unmarshal(body, &internalError)
		if err != nil {
			return fmt.Errorf("parse service error: %w", err)
		}

		internalError.StatusCode = resp.StatusCode

		return internalError
	case http.StatusNoContent:
		return nil
	default:
		return fmt.Errorf("unknown response %d: %w", resp.StatusCode, models.ErrUnknownResponseStatus)
	}
}
