package auth

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	auth_mocks "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"

	"k8s.io/utils/ptr"
)

type MockService struct {
	*auth_mocks.IAuthorizationClient
}

func (m MockService) AuthorizeMiddleware(next http.Handler, _ string, _ string, _ bool, _ string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader, err := runaictx.ParseSubjectId(r.Context())
		if err != nil || authHeader == "" {
			http.Error(w, "{ \"message\" : \"forbidden\"", http.StatusForbidden)
		} else {
			next.ServeHTTP(w, r.WithContext(r.Context()))
		}
	}
}

func (m MockService) GetPermittedScopes(_ context.Context, _ authorization_api.ResourceType, _ *authorization_api.Action) (*authorization_api.PermittedScopesActions, int, error) {
	permittedScopes := &authorization_api.PermittedScopesActions{
		Read: authorization_api.PermittedScopes{
			Tenants: []string{"1"},
			System:  ptr.To(true),
		},
	}
	return permittedScopes, 0, nil
}
