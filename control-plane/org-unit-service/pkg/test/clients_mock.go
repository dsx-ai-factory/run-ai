package testing

import (
	"context"
	"net/http"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	auth_mocks "github.com/run-ai/runai/control-plane/authorization/pkg/client/mocks"
	"github.com/stretchr/testify/mock"
)

// depId to mock with access rules..
var toMockAccessrulesDepId *string = nil

type SimpleMockAuthService struct {
	auth_mocks.IAuthorizationClient

	currentCtx context.Context
}

func SetAuthMock() {
	authMock := &SimpleMockAuthService{}
	authMock.On("GetScopeAccessRules", mock.Anything).Return(authorizationClient.ApiGetScopeAccessRulesRequest{ApiService: authMock})
	authMock.On("DeleteScopeAccessRules", mock.Anything).Return(authorizationClient.ApiDeleteScopeAccessRulesRequest{ApiService: authMock})
	authMock.On("DeleteScopeAccessRulesExecute", mock.Anything).Return(&http.Response{StatusCode: http.StatusNoContent}, nil)
	authorization_client.SetMockAuthorizationClient(authMock)
}

func (m *SimpleMockAuthService) GetScopeAccessRulesExecute(r authorizationClient.ApiGetScopeAccessRulesRequest) (*authorizationClient.GetAccessRules200Response, *http.Response, error) {
	if toMockAccessrulesDepId == nil {
		return &authorizationClient.GetAccessRules200Response{
			AccessRules: []authorizationClient.AccessRule{},
		}, &http.Response{StatusCode: http.StatusOK}, nil
	}

	accessRules := []authorizationClient.AccessRule{
		{
			SubjectId:   "some-user",
			SubjectType: authorizationClient.SUBJECTTYPEEXTENDED_USER,
			ScopeId:     *toMockAccessrulesDepId,
			ScopeType:   authorizationClient.SCOPETYPE_DEPARTMENT,
		},
		{
			SubjectId:   "some-user2",
			SubjectType: authorizationClient.SUBJECTTYPEEXTENDED_USER,
			ScopeId:     *toMockAccessrulesDepId,
			ScopeType:   authorizationClient.SCOPETYPE_DEPARTMENT,
		},
		{
			SubjectId:   "some-group",
			SubjectType: authorizationClient.SUBJECTTYPEEXTENDED_GROUP,
			ScopeId:     *toMockAccessrulesDepId,
			ScopeType:   authorizationClient.SCOPETYPE_DEPARTMENT,
		},
	}
	return &authorizationClient.GetAccessRules200Response{
		AccessRules: accessRules,
	}, &http.Response{StatusCode: http.StatusOK}, nil
}
