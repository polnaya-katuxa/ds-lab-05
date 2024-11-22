package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	cars_service "github.com/polnaya-katuxa/ds-lab-02/gateway/internal/generated/openapi/clients/cars-service"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/models"
)

type CarsServiceClient struct {
	c               cars_service.ClientInterface
	servicePassword string
}

func NewCarsServiceClient(c cars_service.ClientInterface, servicePassword string) *CarsServiceClient {
	return &CarsServiceClient{
		c:               c,
		servicePassword: servicePassword,
	}
}

func (c *CarsServiceClient) List(ctx context.Context, params *cars_service.ListParams) (*cars_service.PaginationResponse, error) {
	resp, err := c.c.List(ctx, params, withToken(ctx))
	if err != nil {
		return nil, fmt.Errorf("get cars list: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusBadRequest:
		var validationError models.ValidationError
		err := json.Unmarshal(body, &validationError)
		if err != nil {
			return nil, fmt.Errorf("parse service error: %w", err)
		}

		return nil, validationError
	case http.StatusInternalServerError:
		var internalError models.InternalError
		err := json.Unmarshal(body, &internalError)
		if err != nil {
			return nil, fmt.Errorf("parse service error: %w", err)
		}

		internalError.StatusCode = resp.StatusCode

		return nil, internalError
	case http.StatusOK:
		var carsList cars_service.PaginationResponse
		err := json.Unmarshal(body, &carsList)
		if err != nil {
			return nil, fmt.Errorf("parse cars list: %w", err)
		}

		return &carsList, nil
	default:
		return nil, fmt.Errorf("unknown response %d: %w", resp.StatusCode, models.ErrUnknownResponseStatus)
	}
}

func (c *CarsServiceClient) Get(ctx context.Context, carUid uuid.UUID) (*cars_service.CarResponse, error) {
	resp, err := c.c.Get(ctx, carUid, withToken(ctx))
	if err != nil {
		return nil, fmt.Errorf("get car: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound, http.StatusInternalServerError:
		var internalError models.InternalError
		err := json.Unmarshal(body, &internalError)
		if err != nil {
			return nil, fmt.Errorf("parse service error: %w", err)
		}

		internalError.StatusCode = resp.StatusCode

		return nil, internalError
	case http.StatusOK:
		var carResponse cars_service.CarResponse
		err := json.Unmarshal(body, &carResponse)
		if err != nil {
			return nil, fmt.Errorf("parse car response: %w", err)
		}

		return &carResponse, nil
	default:
		return nil, fmt.Errorf("unknown response %d: %w", resp.StatusCode, models.ErrUnknownResponseStatus)
	}
}

func (c *CarsServiceClient) Book(ctx context.Context, carUid uuid.UUID) (*cars_service.CarResponse, error) {
	resp, err := c.c.Book(ctx, carUid, withToken(ctx))
	if err != nil {
		return nil, fmt.Errorf("book car: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		var internalError models.InternalError
		err := json.Unmarshal(body, &internalError)
		if err != nil {
			return nil, fmt.Errorf("parse service error: %w", err)
		}

		return nil, internalError
	case http.StatusInternalServerError, http.StatusConflict:
		var internalError models.InternalError
		err := json.Unmarshal(body, &internalError)
		if err != nil {
			return nil, fmt.Errorf("parse service error: %w", err)
		}

		internalError.StatusCode = resp.StatusCode

		return nil, internalError
	case http.StatusOK:
		var carResponse cars_service.CarResponse
		err := json.Unmarshal(body, &carResponse)
		if err != nil {
			return nil, fmt.Errorf("parse car response: %w", err)
		}

		return &carResponse, nil
	default:
		return nil, fmt.Errorf("unknown response %d: %w", resp.StatusCode, models.ErrUnknownResponseStatus)
	}
}

func (c *CarsServiceClient) Unbook(ctx context.Context, carUid uuid.UUID) error {
	resp, err := c.c.Unbook(ctx, carUid, withToken(ctx))
	if err != nil {
		return fmt.Errorf("unbook car: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound, http.StatusInternalServerError, http.StatusConflict:
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

func (c *CarsServiceClient) RetryUnbook(ctx context.Context, carUid uuid.UUID) error {
	resp, err := c.c.Unbook(ctx, carUid, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Service-Password", c.servicePassword)
		return nil
	})
	if err != nil {
		return fmt.Errorf("unbook car: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound, http.StatusInternalServerError, http.StatusConflict:
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
