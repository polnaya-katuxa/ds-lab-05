package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

const (
	bearerKey   = "bearer"
	usernameKey = "username"
)

func CreateMiddleware(jwksURL, servicePassword string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Path() == "/manage/health" || c.Request().Header.Get("Service-Password") == servicePassword {
				return next(c)
			}

			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return c.NoContent(http.StatusUnauthorized)
			}

			prefix := "Bearer "
			if !strings.HasPrefix(header, prefix) {
				return c.NoContent(http.StatusUnauthorized)
			}

			token := strings.TrimPrefix(header, prefix)

			username, err := parseToken(token, jwksURL)
			fmt.Println(username, err)
			if err != nil {
				return c.NoContent(http.StatusUnauthorized)
			}

			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, bearerKey, token)
			ctx = context.WithValue(ctx, usernameKey, username)

			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func parseToken(token, jwksURL string) (string, error) {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
	if err != nil {
		return "", fmt.Errorf("get keyfunc: %w", err)
	}

	parsedToken, err := jwt.Parse(token, jwks.Keyfunc)
	if err != nil {
		return "", fmt.Errorf("parse jwt: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims type")
	}

	username, ok := claims["preferred_username"].(string)
	if !ok {
		return "", fmt.Errorf("missing username in claims")
	}

	return username, nil
}
