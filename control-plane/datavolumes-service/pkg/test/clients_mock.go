package test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	auth_mocks "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/config"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

const (
	TestTenant     = "1"
	TestTenantUint = 1
	TestTenant2    = "2"
	TestUser       = "user@run.ai"
	TestUser2      = "user2@run.ai"
)

type SimpleMockAuthService struct {
	auth_mocks.IAuthorizationClient

	currentCtx context.Context
}

func (m *SimpleMockAuthService) EvaluatePermission(ctx context.Context) authorizationClient.ApiEvaluatePermissionRequest {
	m.currentCtx = ctx
	return authorizationClient.ApiEvaluatePermissionRequest{}
}

func (m *SimpleMockAuthService) EvaluatePermissionExecute(r authorizationClient.ApiEvaluatePermissionRequest) (*authorizationClient.PermittedScopes, *http.Response, error) {
	subjectId := m.currentCtx.Value(runaictx.CfSubjectID)
	switch subjectId.(string) {
	case TestUser:
		return &authorizationClient.PermittedScopes{
			Tenants: []string{TestTenant},
		}, &http.Response{StatusCode: http.StatusOK}, nil
	case TestUser2:
		return &authorizationClient.PermittedScopes{
			Tenants: []string{TestTenant2},
		}, &http.Response{StatusCode: http.StatusOK}, nil
	}

	return &authorizationClient.PermittedScopes{
		System: ptr.To(true),
	}, &http.Response{StatusCode: http.StatusOK}, nil
}

func SetMockBackendServer() {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if got here - it means the request was not cached; return NotFound.
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Not found")
	}))

	runai_config.Set(&config.OrgHierarchyManagerConfig{
		OrgUnitServiceUrl: mockServer.URL,
		BackendUrl:        mockServer.URL,
	})
}
