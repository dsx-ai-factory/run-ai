package internal

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/client/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

const (
	CLIENT_ID      = "client_id"
	CLUSTER_DOMAIN = "cluster_domain"
	CLUSTER_ID     = "cluster_1"
)

func GetApiClient(redirectUris []string) []openapi.Client {
	attributeMap := make(map[string]interface{})

	attributeMap["runai.client_type"] = "workloads_auth"
	attributeMap["runai.cluster_id"] = CLUSTER_ID
	clientArray := []openapi.Client{
		{
			Id:           CLIENT_ID,
			Attributes:   attributeMap,
			RedirectUris: redirectUris,
		},
	}
	return clientArray
}

func initIdmService(t *testing.T) (mock *mocks.ClientsAPI, idmClient *IdmClient) {
	mock = mocks.NewClientsAPI(t)
	config := openapi.NewConfiguration()
	apiClient := openapi.NewAPIClient(config)
	apiClient.ClientsAPI = mock
	idmClient = &IdmClient{
		ApiClient: apiClient,
	}
	return
}

func initIdmServiceWithRealmAPI(t *testing.T) (realmMock *mocks.RealmAPI, idmClient *IdmClient) {
	realmMock = mocks.NewRealmAPI(t)
	config := openapi.NewConfiguration()
	apiClient := openapi.NewAPIClient(config)
	apiClient.RealmAPI = realmMock
	idmClient = &IdmClient{
		ApiClient: apiClient,
	}
	return
}

func initIdmServiceWithUsersAPI(t *testing.T) (usersMock *mocks.UsersAPI, idmClient *IdmClient) {
	usersMock = mocks.NewUsersAPI(t)
	config := openapi.NewConfiguration()
	apiClient := openapi.NewAPIClient(config)
	apiClient.UsersAPI = usersMock
	idmClient = &IdmClient{
		ApiClient: apiClient,
	}
	return
}

func TestGetClientsHappyFlow(t *testing.T) {
	clientServiceMock, testSubject := initIdmService(t)
	clients := GetApiClient(nil)
	clientServiceMock.On("GetAllClients", context.Background()).Return(openapi.ApiGetAllClientsRequest{})
	clientServiceMock.On("GetAllClientsExecute", mock.Anything).Return(clients, nil, nil)
	filter := fmt.Sprintf("%s:%s %s:%s", "runai.client_type", "workloads_auth", "runai.cluster_id", CLUSTER_ID)
	clientsRes, err := testSubject.GetInternalClientByFilter(context.Background(), filter)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(clientsRes), 1)
	assert.Equal(t, clientsRes, clients)
}

func TestGetClientsWithErr(t *testing.T) {
	clientServiceMock, testSubject := initIdmService(t)
	err := errors.Errorf("Error getting clients")
	response := http.Response{
		Status: "500",
	}
	clientServiceMock.On("GetAllClients", context.Background()).Return(openapi.ApiGetAllClientsRequest{})
	clientServiceMock.On("GetAllClientsExecute", mock.Anything).Return(nil, &response, err)
	filter := fmt.Sprintf("%s:%s %s:%s", "runai.client_type", "workloads_auth", "runai.cluster_id", CLUSTER_ID)
	clientsRes, errResponse := testSubject.GetInternalClientByFilter(context.Background(), filter)
	var expectedClientRes []openapi.Client = nil
	assert.True(t, errResponse != nil)
	assert.Equal(t, expectedClientRes, clientsRes)
}

func TestUpdateClientUriUriExistsOnClient(t *testing.T) {
	clientServiceMock, testSubject := initIdmService(t)
	redirectUris := []string{"https://test1", "https://test2"}
	clients := GetApiClient(redirectUris)
	clientServiceMock.On("UpdateRedirectUris", mock.Anything, mock.Anything).Return(openapi.ApiUpdateRedirectUrisRequest{})
	clientServiceMock.On("UpdateRedirectUrisExecute", mock.Anything).Return(nil, nil)
	expectedRedirectUris := []string{"https://test1", "https://test2", CLUSTER_DOMAIN + "/*"}
	expectedReq := getExpectedUpdateClientRedirectUrisReq(clientServiceMock, clients[0], expectedRedirectUris)
	err := testSubject.AppendRedirectUriToInternalClient(context.Background(), clients[0], CLUSTER_DOMAIN)
	assert.True(t, err == nil)
	clientServiceMock.AssertCalled(t, "UpdateRedirectUrisExecute", expectedReq)
}

func TestUpdateClientUriEmptyUrisOnClient(t *testing.T) {
	clientServiceMock, testSubject := initIdmService(t)
	clients := GetApiClient(nil)
	clientServiceMock.On("UpdateRedirectUris", mock.Anything, mock.Anything).Return(openapi.ApiUpdateRedirectUrisRequest{})
	clientServiceMock.On("UpdateRedirectUrisExecute", mock.Anything).Return(nil, nil)
	expectedRedirectUris := []string{CLUSTER_DOMAIN + "/*"}
	expectedReq := getExpectedUpdateClientRedirectUrisReq(clientServiceMock, clients[0], expectedRedirectUris)
	err := testSubject.AppendRedirectUriToInternalClient(context.Background(), clients[0], CLUSTER_DOMAIN)
	assert.True(t, err == nil)
	clientServiceMock.AssertCalled(t, "UpdateRedirectUrisExecute", expectedReq)
}

func TestUpdateClientUriWhenAlreadyExist(t *testing.T) {
	clientServiceMock, testSubject := initIdmService(t)
	redirectUris := []string{"https://test1", "https://test2", CLUSTER_DOMAIN + "/*"}
	clients := GetApiClient(redirectUris)
	err := testSubject.AppendRedirectUriToInternalClient(context.Background(), clients[0], CLUSTER_DOMAIN)
	assert.True(t, err == nil)
	clientServiceMock.AssertNotCalled(t, "UpdateRedirectUrisExecute", mock.Anything)
}

func TestCreateClient_HappyFlow(t *testing.T) {
	clientServiceMock, testSubject := initIdmService(t)
	resp := &openapi.CreateInternalClient201Response{
		Id: ptr.To("ID"),
	}
	httpResp := &http.Response{
		StatusCode: http.StatusOK,
	}
	clientServiceMock.On("CreateInternalClient", mock.Anything).Return(openapi.ApiCreateInternalClientRequest{})
	clientServiceMock.On("CreateInternalClientExecute", mock.Anything).Return(resp, httpResp, nil)
	clientId, err := testSubject.CreateClient(context.Background(), "name", "Application", ptr.To("ID"))
	assert.Equal(t, "ID", *clientId)
	assert.Nil(t, err)
}

func TestCreateClient_Error(t *testing.T) {
	clientServiceMock, testSubject := initIdmService(t)
	resp := &openapi.CreateInternalClient201Response{
		Id: ptr.To("ID"),
	}
	clientServiceMock.On("CreateInternalClient", mock.Anything).Return(openapi.ApiCreateInternalClientRequest{})
	clientServiceMock.On("CreateInternalClientExecute", mock.Anything).Return(resp, nil, fmt.Errorf("error"))
	clientId, err := testSubject.CreateClient(context.Background(), "name", "Application", ptr.To("ID"))
	assert.Nil(t, clientId)
	assert.NotNil(t, err)
}

func TestCreateClient_HttpError(t *testing.T) {
	clientServiceMock, testSubject := initIdmService(t)
	resp := &openapi.CreateInternalClient201Response{
		Id: ptr.To("ID"),
	}
	httpResp := &http.Response{
		StatusCode: http.StatusInternalServerError,
	}
	clientServiceMock.On("CreateInternalClient", mock.Anything).Return(openapi.ApiCreateInternalClientRequest{})
	clientServiceMock.On("CreateInternalClientExecute", mock.Anything).Return(resp, httpResp, fmt.Errorf("error"))
	clientId, err := testSubject.CreateClient(context.Background(), "name", "Application", ptr.To("ID"))
	assert.Nil(t, clientId)
	assert.NotNil(t, err)
}

func getExpectedUpdateClientRedirectUrisReq(mock *mocks.ClientsAPI, client openapi.Client, expectedRedirectUris []string) openapi.ApiUpdateRedirectUrisRequest {
	expectedInnerReq := openapi.NewClientRedirectUrisUpdateRequest(expectedRedirectUris)
	expectedUpdateUriReq := mock.UpdateRedirectUris(context.Background(), client.ClientId)
	expectedUpdateUriReq = expectedUpdateUriReq.ClientRedirectUrisUpdateRequest(*expectedInnerReq)
	return expectedUpdateUriReq
}

func setRealmAPIRequestApiService(req *openapi.ApiCreateRealmRequest, api openapi.RealmAPI) {
	reflect.ValueOf(req).Elem().FieldByName("ApiService").Set(reflect.ValueOf(api))
}

func setUsersAPIRequestApiService(req *openapi.ApiCreateIntenralUserRequest, api openapi.UsersAPI) {
	reflect.ValueOf(req).Elem().FieldByName("ApiService").Set(reflect.ValueOf(api))
}

func TestCreateRealm_HappyFlow(t *testing.T) {
	realmMock, testSubject := initIdmServiceWithRealmAPI(t)
	realmName := "test-realm"
	tenantId := "123"
	realmResp := &openapi.RealmCreationResponse{
		RealmName: realmName,
		Domain:    "test-domain",
		Clients:   []openapi.Client{},
	}
	httpResp := &http.Response{StatusCode: http.StatusOK}

	createRealmReq := openapi.ApiCreateRealmRequest{}
	setRealmAPIRequestApiService(&createRealmReq, realmMock)
	realmMock.On("CreateRealm", mock.Anything).Return(createRealmReq)
	realmMock.On("CreateRealmExecute", mock.Anything).Return(realmResp, httpResp, nil)

	resp, respHTTP, err := testSubject.CreateRealm(context.Background(), realmName, tenantId)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, realmName, resp.RealmName)
	assert.Equal(t, "test-domain", resp.Domain)
	assert.Equal(t, http.StatusOK, respHTTP.StatusCode)
}

func TestCreateRealm_Error(t *testing.T) {
	realmMock, testSubject := initIdmServiceWithRealmAPI(t)
	createRealmReq := openapi.ApiCreateRealmRequest{}
	setRealmAPIRequestApiService(&createRealmReq, realmMock)
	realmMock.On("CreateRealm", mock.Anything).Return(createRealmReq)
	realmMock.On("CreateRealmExecute", mock.Anything).Return((*openapi.RealmCreationResponse)(nil), (*http.Response)(nil), fmt.Errorf("realm error"))

	resp, _, err := testSubject.CreateRealm(context.Background(), "realm", "456")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateInternalUser_HappyFlow(t *testing.T) {
	usersMock, testSubject := initIdmServiceWithUsersAPI(t)
	req := openapi.InternalUserCreationRequest{
		Email:         "user@example.com",
		Realm:         "test-realm",
		ResetPassword: true,
	}
	req.Notify = *openapi.NewNullableBool(ptr.To(true))
	userResp := &openapi.UserPostResponse{
		Id:           "user-id-123",
		Username:     "user@example.com",
		TempPassword: "temp-pass",
	}
	httpResp := &http.Response{StatusCode: http.StatusCreated}

	createUserReq := openapi.ApiCreateIntenralUserRequest{}
	setUsersAPIRequestApiService(&createUserReq, usersMock)
	usersMock.On("CreateIntenralUser", mock.Anything).Return(createUserReq)
	usersMock.On("CreateIntenralUserExecute", mock.Anything).Return(userResp, httpResp, nil)

	resp, respHTTP, err := testSubject.CreateInternalUser(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "user-id-123", resp.Id)
	assert.Equal(t, "temp-pass", resp.TempPassword)
	assert.Equal(t, http.StatusCreated, respHTTP.StatusCode)
}

func TestCreateInternalUser_Error(t *testing.T) {
	usersMock, testSubject := initIdmServiceWithUsersAPI(t)
	req := openapi.InternalUserCreationRequest{
		Email:         "user@example.com",
		Realm:         "test-realm",
		ResetPassword: false,
	}
	createUserReq := openapi.ApiCreateIntenralUserRequest{}
	setUsersAPIRequestApiService(&createUserReq, usersMock)
	usersMock.On("CreateIntenralUser", mock.Anything).Return(createUserReq)
	usersMock.On("CreateIntenralUserExecute", mock.Anything).Return((*openapi.UserPostResponse)(nil), (*http.Response)(nil), fmt.Errorf("user error"))

	resp, _, err := testSubject.CreateInternalUser(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
