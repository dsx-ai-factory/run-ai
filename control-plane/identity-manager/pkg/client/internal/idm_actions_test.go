package internal

import (
	"context"
	"net/http"
	"testing"

	openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/client/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

func initMocks(t *testing.T) (mock *mocks.TokensAPI, apiClient *openapi.APIClient) {
	mock = mocks.NewTokensAPI(t)
	config := openapi.NewConfiguration()
	apiClient = openapi.NewAPIClient(config)
	apiClient.TokensAPI = mock
	return
}

func TestAppLoginHappyFlow(t *testing.T) {
	tokenApiMock, apiClient := initMocks(t)
	response := &openapi.TokenResponse{
		AccessToken: ptr.To("accessToken"),
	}
	tokenApiMock.On("GrantToken", context.Background()).Return(openapi.ApiGrantTokenRequest{}, nil, nil)
	tokenApiMock.On("GrantTokenExecute", mock.Anything).Return(response, &http.Response{
		StatusCode: http.StatusOK,
	}, nil)
	appId := "appId"
	appSecret := "appSecret"
	clientResponse, err := clientCredentialsLogin(context.Background(), apiClient, appId, appSecret)
	assert.Equal(t, err, nil)
	assert.Equal(t, *clientResponse.AccessToken, *response.AccessToken)
}

func TestAppLoginBadParams(t *testing.T) {
	_, apiClient := initMocks(t)
	appSecret := "appSecret"
	clientResponse, err := clientCredentialsLogin(context.Background(), apiClient, "", appSecret)
	assert.NotNil(t, err)
	assert.Nil(t, clientResponse)
}

func TestAppLoginBadHttpResponse(t *testing.T) {
	tokenApiMock, apiClient := initMocks(t)
	tokenApiMock.On("GrantToken", context.Background()).Return(openapi.ApiGrantTokenRequest{}, nil, nil)
	tokenApiMock.On("GrantTokenExecute", mock.Anything).Return(nil, &http.Response{
		StatusCode: http.StatusBadGateway,
	}, nil)
	appId := "appId"
	appSecret := "appSecret"
	clientResponse, err := clientCredentialsLogin(context.Background(), apiClient, appId, appSecret)
	assert.NotNil(t, err)
	assert.Nil(t, clientResponse)
}
