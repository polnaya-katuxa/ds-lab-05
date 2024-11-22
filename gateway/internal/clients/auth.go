package clients

import (
	"context"
	"net/http"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/polnaya-katuxa/ds-lab-02/gateway/internal/auth"
)

func withToken(ctx context.Context) func(ctx context.Context, req *http.Request) error {
	provider, _ := securityprovider.NewSecurityProviderBearerToken(auth.GetToken(ctx))
	return provider.Intercept
}
