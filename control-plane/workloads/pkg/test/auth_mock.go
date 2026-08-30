package testing

import (
	"context"
	"net/http"

	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	auth_mocks "github.com/run-ai/runai/control-plane/authorization/pkg/client/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	UnpermittedUser     = "unpermitted-subject"
	UserWithEmptyScopes = "empty-scopes-subject"

	UserWithCluster1And2Permissions = "user-with-cluster-1-and-2-permissions"
)

type SimpleMockAuthService struct {
	*auth_mocks.IAuthorizationClient

	currentCtx context.Context
}

func (m *SimpleMockAuthService) EvaluatePermission(ctx context.Context) authorizationClient.ApiEvaluatePermissionRequest {
	m.currentCtx = ctx
	return authorizationClient.ApiEvaluatePermissionRequest{}
}

func (m *SimpleMockAuthService) EvaluatePermissionExecute(r authorizationClient.ApiEvaluatePermissionRequest) (*authorizationClient.PermittedScopes, *http.Response, error) {
	subjectId := m.currentCtx.Value(runaictx.CfSubjectID)
	switch subjectId.(string) {
	case UnpermittedUser:
		return &authorizationClient.PermittedScopes{}, &http.Response{StatusCode: http.StatusForbidden}, nil
	case UserWithEmptyScopes:
		return nil, &http.Response{StatusCode: http.StatusNoContent}, nil
	case UserWithCluster1And2Permissions:
		return &authorizationClient.PermittedScopes{Clusters: []string{TestUuid1, TestUuid2}}, &http.Response{StatusCode: http.StatusOK}, nil
	}

	return &authorizationClient.PermittedScopes{
		Tenants: []string{"1"},
	}, &http.Response{StatusCode: http.StatusOK}, nil
}
