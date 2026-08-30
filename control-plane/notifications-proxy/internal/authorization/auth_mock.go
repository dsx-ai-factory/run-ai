package authorization

import (
	"context"
	"net/http"

	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	auth_mocks "github.com/run-ai/runai/control-plane/authorization/pkg/client/mocks"
	"github.com/run-ai/runai/control-plane/authorization/pkg/client/structs"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

type AuthMock struct {
	*auth_mocks.IAuthorizationClient
}

func (m *AuthMock) AuthorizeMiddleware(next http.Handler, _ string, _ string,
	_ bool, _ string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tenantId, _ := runaictx.ParseTenantId(ctx)
		permittedScopes := &openapi.PermittedScopes{}
		if tenantId == "3" {
			permittedScopes = &openapi.PermittedScopes{
				Tenants: []string{"3"},
			}
		}
		if tenantId == "1" {
			ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
			permittedScopes = &openapi.PermittedScopes{
				System:  ptr.To(true),
				Tenants: []string{"1"},
			}
		}
		ctx = context.WithValue(ctx, structs.PermittedScopesKey, permittedScopes)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
