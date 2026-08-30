package auth

import (
	"context"
	"net/http"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	auth_mocks "github.com/run-ai/runai/control-plane/authorization/pkg/client/mocks"
	"github.com/run-ai/runai/control-plane/authorization/pkg/client/structs"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
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
			isSystem := runaictx.IsManagementTenant(r.Context())
			ctx := context.WithValue(r.Context(), structs.PermittedScopesKey, &authorization_api.PermittedScopes{
				System: ptr.To(isSystem),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
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

func (m MockService) GetGroupsForTenant(_ context.Context, _ string) ([]string, error) {
	return []string{"group1", "group2"}, nil
}

func (m MockService) DeleteSubjectAccessRules(_ context.Context) authorization_api.ApiDeleteSubjectAccessRulesRequest {
	return authorization_api.ApiDeleteSubjectAccessRulesRequest{}
}

func (m MockService) DeleteSubjectAccessRulesExecute(_ authorization_api.ApiDeleteSubjectAccessRulesRequest) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK}, nil
}

func (m MockService) SetAppPermissions(_ context.Context, _ string, _ string, _ authorization_api.InternalCreateAccessRuleRequestScope, _ authorization_api.AllowedInternalCreateRoles) error {
	return nil
}

func (m MockService) PrivilegeEscalationEvaluate(_ context.Context) authorization_api.ApiPrivilegeEscalationEvaluateRequest {
	return authorization_api.ApiPrivilegeEscalationEvaluateRequest{
		ApiService: m,
	}
}

func (m MockService) PrivilegeEscalationEvaluateExecute(_ authorization_api.ApiPrivilegeEscalationEvaluateRequest) (*http.Response, error) {
	return &http.Response{StatusCode: 204, Body: http.NoBody}, nil
}
