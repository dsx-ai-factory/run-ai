package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Nerzal/gocloak/v13"
	kcMocks "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/gocloak/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

const (
	testHost                = "http://localhost:8080"
	testUsername            = "admin-user"
	testPassword            = "password"
	adminToken              = "admin-token"
	clientSecret            = "client-secret"
	newClientSecret         = "new-client-secret"
	testClientID            = "test-client-id"
	serviceAccountID        = "service-account-id"
	testServiceAccountToken = "app-token"
)

func initTest(t *testing.T) (*KeycloakAuthenticator, *kcMocks.GoCloak) {
	config := Configuration{
		Host:     testHost,
		Username: testUsername,
		Password: testPassword,
	}

	authenticator := NewKeycloakAuthenticator(config)
	mockKcClient := kcMocks.NewGoCloak(t)
	authenticator.KcClient = mockKcClient

	return authenticator, mockKcClient
}

func TestNewKeycloakAuthenticator(t *testing.T) {
	config := Configuration{
		Host:     testHost,
		Username: testUsername,
		Password: testPassword,
	}

	authenticator := NewKeycloakAuthenticator(config)

	assert.NotNil(t, authenticator)
	assert.Equal(t, config.Host, authenticator.host)
	assert.Equal(t, config.Username, authenticator.userAuth.username)
	assert.Equal(t, config.Password, authenticator.userAuth.password)
	assert.NotNil(t, authenticator.KcClient)
}

func TestGetHost(t *testing.T) {
	authenticator, _ := initTest(t)
	host := authenticator.GetHost()

	assert.Equal(t, testHost, host)
}

func TestGetAdminToken_AppLoginType_ValidTokenExists(t *testing.T) {
	authenticator, _ := initTest(t)
	authenticator.loginType = LoginTypeServiceAccount
	authenticator.appAuth.appToken = getMockJWT("valid-token", 3600)
	authenticator.appAuth.appTokenExpiresAt = time.Now().Add(1 * time.Hour)

	token, err := authenticator.GetAdminToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "valid-token", *token)
}

func TestGetAdminToken_AppLoginType_ExpiredToken(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.loginType = LoginTypeServiceAccount
	authenticator.appAuth.appToken = getMockJWT("expired-token", 3600)
	authenticator.appAuth.appTokenExpiresAt = time.Now().Add(-1 * time.Hour)

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT(adminToken, 3600), nil)
	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return([]*gocloak.Client{getMockClient(ServiceClientID, clientSecret)}, nil)
	mockKcClient.On("GetClientSecret", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return(getMockCredential(clientSecret), nil)
	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, clientSecret, MasterRealmName).Return(getMockJWT("new-app-token", 3600), nil)

	token, err := authenticator.GetAdminToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "new-app-token", *token)
	mockKcClient.AssertExpectations(t)
}

func TestGetAdminToken_AppLoginType_NoTokenExists(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.loginType = LoginTypeServiceAccount
	authenticator.appAuth.appToken = nil

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT(adminToken, 3600), nil)
	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return([]*gocloak.Client{getMockClient(ServiceClientID, clientSecret)}, nil)
	mockKcClient.On("GetClientSecret", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return(getMockCredential(clientSecret), nil)
	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, clientSecret, MasterRealmName).Return(getMockJWT("new-app-token", 3600), nil)

	token, err := authenticator.GetAdminToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "new-app-token", *token)
	mockKcClient.AssertExpectations(t)
}

func TestGetAdminToken_UserLoginType_ValidTokenExists(t *testing.T) {
	authenticator, _ := initTest(t)
	authenticator.loginType = LoginTypeUser
	authenticator.userAuth.userToken = getMockJWT("valid-token", 3600)
	authenticator.userAuth.userTokenExpiresAt = time.Now().Add(1 * time.Hour)

	token, err := authenticator.GetAdminToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "valid-token", *token)
}

func TestGetAdminToken_UserLoginType_ExpiredToken_RefreshSuccess(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.loginType = LoginTypeUser
	authenticator.userAuth.userToken = getMockJWT("expired-token", 3600)
	authenticator.userAuth.userToken.RefreshToken = "refresh-token"
	authenticator.userAuth.userTokenExpiresAt = time.Now().Add(-1 * time.Hour)

	mockKcClient.On("RefreshToken", mock.Anything, "refresh-token", MasterRealmAdminClient, "", MasterRealmName).Return(getMockJWT("new-token", 3600), nil)

	token, err := authenticator.GetAdminToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "new-token", *token)
	mockKcClient.AssertExpectations(t)
}

func TestGetAdminToken_UserLoginType_ExpiredToken_RefreshFails_LoginSuccess(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.loginType = LoginTypeUser
	authenticator.userAuth.userToken = getMockJWT("expired-token", 3600)
	authenticator.userAuth.userToken.RefreshToken = "refresh-token"
	authenticator.userAuth.userTokenExpiresAt = time.Now().Add(-1 * time.Hour)

	mockKcClient.On("RefreshToken", mock.Anything, "refresh-token", MasterRealmAdminClient, "", MasterRealmName).Return(nil, errors.New("refresh failed"))
	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT("new-token", 3600), nil)

	token, err := authenticator.GetAdminToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "new-token", *token)
	mockKcClient.AssertExpectations(t)
}

func TestGetAdminToken_UserLoginType_ExpiredToken_RefreshFails_LoginFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.loginType = LoginTypeUser
	authenticator.userAuth.userToken = getMockJWT("expired-token", 3600)
	authenticator.userAuth.userToken.RefreshToken = "refresh-token"
	authenticator.userAuth.userTokenExpiresAt = time.Now().Add(-1 * time.Hour)

	mockKcClient.On("RefreshToken", mock.Anything, "refresh-token", MasterRealmAdminClient, "", MasterRealmName).Return(nil, errors.New("refresh failed"))
	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(nil, errors.New("login failed"))

	token, err := authenticator.GetAdminToken(context.Background())

	assert.Error(t, err)
	assert.Nil(t, token)
	mockKcClient.AssertExpectations(t)
}

func TestSetLoginType_AppLoginType_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT(adminToken, 3600), nil)
	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return([]*gocloak.Client{getMockClient(ServiceClientID, clientSecret)}, nil)
	mockKcClient.On("GetClientSecret", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return(getMockCredential(clientSecret), nil)
	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, clientSecret, MasterRealmName).Return(getMockJWT(testServiceAccountToken, 3600), nil)

	err := authenticator.SetLoginType(context.Background(), LoginTypeServiceAccount)

	assert.NoError(t, err)
	assert.Equal(t, LoginTypeServiceAccount, authenticator.loginType)
	mockKcClient.AssertExpectations(t)
}

func TestSetLoginType_AppLoginType_AdminLoginFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(nil, errors.New("invalid credentials"))

	err := authenticator.SetLoginType(context.Background(), LoginTypeServiceAccount)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrAdminLoginFailed)
	mockKcClient.AssertExpectations(t)
}

func TestSetLoginType_UserLoginType_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT(adminToken, 3600), nil)

	err := authenticator.SetLoginType(context.Background(), LoginTypeUser)

	assert.NoError(t, err)
	assert.Equal(t, LoginTypeUser, authenticator.loginType)
	mockKcClient.AssertExpectations(t)
}

func TestSetLoginType_UserLoginType_AdminLoginFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(nil, errors.New("invalid credentials"))

	err := authenticator.SetLoginType(context.Background(), LoginTypeUser)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrAdminLoginFailed)
	mockKcClient.AssertExpectations(t)
}

func TestSetAppLogin_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return([]*gocloak.Client{getMockClient(ServiceClientID, clientSecret)}, nil)
	mockKcClient.On("GetClientSecret", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return(getMockCredential(clientSecret), nil)
	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, clientSecret, MasterRealmName).Return(getMockJWT(testServiceAccountToken, 3600), nil)

	err := authenticator.setAppLogin(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, authenticator.appAuth.appToken)
	assert.Equal(t, testServiceAccountToken, authenticator.appAuth.appToken.AccessToken)
	mockKcClient.AssertExpectations(t)
}

func TestSetAppLogin_CreateNewClient(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return([]*gocloak.Client{}, nil)
	mockKcClient.On("CreateClient", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return(testClientID, nil)
	mockKcClient.On("GetClientServiceAccount", mock.Anything, adminToken, MasterRealmName, testClientID).Return(getMockUser(serviceAccountID), nil)
	mockKcClient.On("GetRealmRole", mock.Anything, adminToken, MasterRealmName, AdminRole).Return(getMockRole(AdminRole), nil)
	mockKcClient.On("GetRealmRole", mock.Anything, adminToken, MasterRealmName, DefaultMasterRole).Return(getMockRole(DefaultMasterRole), nil)
	mockKcClient.On("AddRealmRoleToUser", mock.Anything, adminToken, MasterRealmName, serviceAccountID, mock.Anything).Return(nil)
	mockKcClient.On("GetClientSecret", mock.Anything, adminToken, MasterRealmName, testClientID).Return(getMockCredential(newClientSecret), nil)
	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, newClientSecret, MasterRealmName).Return(getMockJWT(testServiceAccountToken, 3600), nil)

	err := authenticator.setAppLogin(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, authenticator.appAuth.appToken)
	assert.Equal(t, testServiceAccountToken, authenticator.appAuth.appToken.AccessToken)
	mockKcClient.AssertExpectations(t)
}

func TestSetAppLogin_CreateClientFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return([]*gocloak.Client{}, nil)
	mockKcClient.On("CreateClient", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return("", errors.New("failed to create client"))

	err := authenticator.setAppLogin(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create admin application")
	mockKcClient.AssertExpectations(t)
}

func TestRefreshToken_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT("old-token", 3600)
	authenticator.userAuth.userToken.RefreshToken = "refresh-token"

	mockKcClient.On("RefreshToken", mock.Anything, "refresh-token", MasterRealmAdminClient, "", MasterRealmName).Return(getMockJWT("new-token", 3600), nil)

	err := authenticator.refreshToken(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "new-token", authenticator.userAuth.userToken.AccessToken)
	mockKcClient.AssertExpectations(t)
}

func TestRefreshToken_NoRefreshToken(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT("old-token", 3600)
	authenticator.userAuth.userToken.RefreshToken = ""

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT("new-token", 3600), nil)

	err := authenticator.refreshToken(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "new-token", authenticator.userAuth.userToken.AccessToken)
	mockKcClient.AssertExpectations(t)
}

func TestRefreshToken_RefreshFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT("old-token", 3600)
	authenticator.userAuth.userToken.RefreshToken = "refresh-token"

	mockKcClient.On("RefreshToken", mock.Anything, "refresh-token", MasterRealmAdminClient, "", MasterRealmName).Return(nil, errors.New("refresh failed"))

	err := authenticator.refreshToken(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to refresh token")
	mockKcClient.AssertExpectations(t)
}

func TestSetAppAdminToken_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)
	authenticator.appAuth.appSecret = getMockCredential(clientSecret)

	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, clientSecret, MasterRealmName).Return(getMockJWT(testServiceAccountToken, 3600), nil)

	err := authenticator.setAppAdminToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, authenticator.appAuth.appToken)
	assert.Equal(t, testServiceAccountToken, authenticator.appAuth.appToken.AccessToken)
	mockKcClient.AssertExpectations(t)
}

func TestSetAppAdminToken_LoginFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)
	authenticator.appAuth.appSecret = getMockCredential(clientSecret)

	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, clientSecret, MasterRealmName).Return(nil, errors.New("login failed"))

	err := authenticator.setAppAdminToken(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to login admin app")
	mockKcClient.AssertExpectations(t)
}

func TestSetAppSecret_SecretExists(t *testing.T) {
	authenticator, _ := initTest(t)
	authenticator.appAuth.appSecret = getMockCredential(clientSecret)

	err := authenticator.setAppSecret(context.Background())

	assert.NoError(t, err)
}

func TestSetAppSecret_NoSecret_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.appAuth.appSecret = nil

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT(adminToken, 3600), nil)
	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return([]*gocloak.Client{getMockClient(ServiceClientID, clientSecret)}, nil)
	mockKcClient.On("GetClientSecret", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return(getMockCredential(clientSecret), nil)

	err := authenticator.setAppSecret(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, authenticator.appAuth.appSecret)
	mockKcClient.AssertExpectations(t)
}

func TestSetAppSecret_NoSecret_AdminLoginFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.appAuth.appSecret = nil

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(nil, errors.New("login failed"))

	err := authenticator.setAppSecret(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrAdminLoginFailed)
	mockKcClient.AssertExpectations(t)
}

func TestSetNewAppSecret_ClientNotFound(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT(adminToken, 3600), nil)
	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return([]*gocloak.Client{}, nil)

	err := authenticator.setNewAppSecret(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check existing clients")
	mockKcClient.AssertExpectations(t)
}

func TestSetNewAppSecret_GetClientFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT(adminToken, 3600), nil)
	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return(nil, errors.New("get clients failed"))

	err := authenticator.setNewAppSecret(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get admin client")
	mockKcClient.AssertExpectations(t)
}

func TestSetNewAppSecret_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT(adminToken, 3600), nil)
	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return([]*gocloak.Client{getMockClient(ServiceClientID, clientSecret)}, nil)
	mockKcClient.On("GetClientSecret", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return(getMockCredential(clientSecret), nil)

	err := authenticator.setNewAppSecret(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, authenticator.appAuth.appSecret)
	mockKcClient.AssertExpectations(t)
}

func TestSetAppAdminSecret_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("GetClientSecret", mock.Anything, adminToken, MasterRealmName, testClientID).Return(getMockCredential(clientSecret), nil)

	err := authenticator.setAppAdminSecret(context.Background(), testClientID)

	assert.NoError(t, err)
	assert.NotNil(t, authenticator.appAuth.appSecret)
	assert.Equal(t, clientSecret, *authenticator.appAuth.appSecret.Value)
	mockKcClient.AssertExpectations(t)
}

func TestSetAppAdminSecret_GetSecretFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("GetClientSecret", mock.Anything, adminToken, MasterRealmName, testClientID).Return(nil, errors.New("get secret failed"))

	err := authenticator.setAppAdminSecret(context.Background(), testClientID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get admin application secret")
	mockKcClient.AssertExpectations(t)
}

func TestSetAppAdminSecret_NilSecret(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	nilSecret := &gocloak.CredentialRepresentation{Value: nil}
	mockKcClient.On("GetClientSecret", mock.Anything, adminToken, MasterRealmName, testClientID).Return(nilSecret, nil)

	err := authenticator.setAppAdminSecret(context.Background(), testClientID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get admin application secret")
	mockKcClient.AssertExpectations(t)
}

func TestUserAdminLogin_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT(adminToken, 3600), nil)

	err := authenticator.userAdminLogin(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, authenticator.userAuth.userToken)
	assert.Equal(t, adminToken, authenticator.userAuth.userToken.AccessToken)
	mockKcClient.AssertExpectations(t)
}

func TestUserAdminLogin_Failure(t *testing.T) {
	authenticator, mockKcClient := initTest(t)

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(nil, errors.New("invalid credentials"))

	err := authenticator.userAdminLogin(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to login admin user")
	mockKcClient.AssertExpectations(t)
}

func TestAppAdminLogin_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.appAuth.appSecret = getMockCredential(clientSecret)

	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, clientSecret, MasterRealmName).Return(getMockJWT(testServiceAccountToken, 3600), nil)

	err := authenticator.appAdminLogin(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, authenticator.appAuth.appToken)
	assert.Equal(t, testServiceAccountToken, authenticator.appAuth.appToken.AccessToken)
	mockKcClient.AssertExpectations(t)
}

func TestAppAdminLogin_Failure(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.appAuth.appSecret = getMockCredential(clientSecret)

	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, clientSecret, MasterRealmName).Return(nil, errors.New("invalid client credentials"))

	err := authenticator.appAdminLogin(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to login admin app")
	mockKcClient.AssertExpectations(t)
}

func TestCreateNewAppAdminClient_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("CreateClient", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return(testClientID, nil)
	mockKcClient.On("GetClientServiceAccount", mock.Anything, adminToken, MasterRealmName, testClientID).Return(getMockUser(serviceAccountID), nil)
	mockKcClient.On("GetRealmRole", mock.Anything, adminToken, MasterRealmName, AdminRole).Return(getMockRole(AdminRole), nil)
	mockKcClient.On("GetRealmRole", mock.Anything, adminToken, MasterRealmName, DefaultMasterRole).Return(getMockRole(DefaultMasterRole), nil)
	mockKcClient.On("AddRealmRoleToUser", mock.Anything, adminToken, MasterRealmName, serviceAccountID, mock.Anything).Return(nil)

	clientID, err := authenticator.createNewAppAdminClient(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, testClientID, clientID)
	mockKcClient.AssertExpectations(t)
}

func TestCreateNewAppAdminClient_CreateClientFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("CreateClient", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return("", errors.New("failed to create client"))

	clientID, err := authenticator.createNewAppAdminClient(context.Background())

	assert.Error(t, err)
	assert.Empty(t, clientID)
	assert.Contains(t, err.Error(), "failed to create admin app client")
	mockKcClient.AssertExpectations(t)
}

func TestCreateNewAppAdminClient_GetServiceAccountFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("CreateClient", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return(testClientID, nil)
	mockKcClient.On("GetClientServiceAccount", mock.Anything, adminToken, MasterRealmName, testClientID).Return(nil, errors.New("failed to get service account"))

	clientID, err := authenticator.createNewAppAdminClient(context.Background())

	assert.Error(t, err)
	assert.Empty(t, clientID)
	assert.Contains(t, err.Error(), "failed to get service account user")
	mockKcClient.AssertExpectations(t)
}

func TestGetAppAdminClient_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	expectedClient := getMockClient(ServiceClientID, clientSecret)

	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return([]*gocloak.Client{expectedClient}, nil)

	client, err := authenticator.getAppAdminClient(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedClient, client)
	mockKcClient.AssertExpectations(t)
}

func TestGetAppAdminClient_NoClientsFound(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return([]*gocloak.Client{}, nil)

	client, err := authenticator.getAppAdminClient(context.Background())

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to check existing clients")
	mockKcClient.AssertExpectations(t)
}

func TestGetAppAdminClient_GetClientsFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return(nil, errors.New("failed to get clients"))

	client, err := authenticator.getAppAdminClient(context.Background())

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to check existing clients")
	mockKcClient.AssertExpectations(t)
}

func TestAssignAdminRolesToAdminApp_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	serviceAccountUser := getMockUser(serviceAccountID)

	mockKcClient.On("GetRealmRole", mock.Anything, adminToken, MasterRealmName, AdminRole).Return(getMockRole(AdminRole), nil)
	mockKcClient.On("GetRealmRole", mock.Anything, adminToken, MasterRealmName, DefaultMasterRole).Return(getMockRole(DefaultMasterRole), nil)
	mockKcClient.On("AddRealmRoleToUser", mock.Anything, adminToken, MasterRealmName, serviceAccountID, mock.Anything).Return(nil)

	err := authenticator.assignAdminRolesToAdminApp(context.Background(), adminToken, serviceAccountUser)

	assert.NoError(t, err)
	mockKcClient.AssertExpectations(t)
}

func TestAssignAdminRolesToAdminApp_NilUserID(t *testing.T) {
	authenticator, _ := initTest(t)
	serviceAccountUser := &gocloak.User{}

	err := authenticator.assignAdminRolesToAdminApp(context.Background(), adminToken, serviceAccountUser)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service account user ID is nil")
}

func TestAssignAdminRolesToAdminApp_NoRolesFound(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	serviceAccountUser := getMockUser(serviceAccountID)

	mockKcClient.On("GetRealmRole", mock.Anything, adminToken, MasterRealmName, AdminRole).Return(nil, errors.New("role not found"))
	mockKcClient.On("GetRealmRole", mock.Anything, adminToken, MasterRealmName, DefaultMasterRole).Return(nil, errors.New("role not found"))

	err := authenticator.assignAdminRolesToAdminApp(context.Background(), adminToken, serviceAccountUser)

	assert.NoError(t, err)
	mockKcClient.AssertExpectations(t)
}

func TestAssignAdminRolesToAdminApp_AddRoleFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	serviceAccountUser := getMockUser(serviceAccountID)

	mockKcClient.On("GetRealmRole", mock.Anything, adminToken, MasterRealmName, AdminRole).Return(getMockRole(AdminRole), nil)
	mockKcClient.On("GetRealmRole", mock.Anything, adminToken, MasterRealmName, DefaultMasterRole).Return(getMockRole(DefaultMasterRole), nil)
	mockKcClient.On("AddRealmRoleToUser", mock.Anything, adminToken, MasterRealmName, serviceAccountID, mock.Anything).Return(errors.New("failed to add role"))

	err := authenticator.assignAdminRolesToAdminApp(context.Background(), adminToken, serviceAccountUser)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign admin roles to admin app")
	mockKcClient.AssertExpectations(t)
}

func TestUserAuthentication_GetAccessToken_ValidToken(t *testing.T) {
	ua := &UserAuthentication{
		userToken: getMockJWT("valid-token", 3600),
	}
	token := ua.getAccessToken()
	assert.NotNil(t, token)
	assert.Equal(t, "valid-token", *token)
}

func TestUserAuthentication_GetAccessToken_NilToken(t *testing.T) {
	ua := &UserAuthentication{
		userToken: nil,
	}
	token := ua.getAccessToken()
	assert.Nil(t, token)
}

func TestUserAuthentication_GetAccessToken_EmptyToken(t *testing.T) {
	ua := &UserAuthentication{
		userToken: getMockJWT("", 3600),
	}
	token := ua.getAccessToken()
	assert.Nil(t, token)
}

func TestUserAuthentication_IsTokenValid_ValidToken(t *testing.T) {
	ua := &UserAuthentication{
		userToken:          getMockJWT("valid-token", 3600),
		userTokenExpiresAt: time.Now().Add(1 * time.Minute),
	}
	isValid := ua.isTokenValid()
	assert.True(t, isValid)
}

func TestUserAuthentication_IsTokenValid_ExpiredToken(t *testing.T) {
	ua := &UserAuthentication{
		userToken:          getMockJWT("expired-token", 3600),
		userTokenExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	isValid := ua.isTokenValid()
	assert.False(t, isValid)
}

func TestUserAuthentication_IsTokenValid_NilToken(t *testing.T) {
	ua := &UserAuthentication{
		userToken: nil,
	}
	isValid := ua.isTokenValid()
	assert.False(t, isValid)
}

func TestUserAuthentication_IsTokenValid_EmptyToken(t *testing.T) {
	ua := &UserAuthentication{
		userToken:          getMockJWT("", 3600),
		userTokenExpiresAt: time.Now().Add(1 * time.Minute),
	}
	isValid := ua.isTokenValid()
	assert.False(t, isValid)
}

func TestUserAuthentication_SetToken(t *testing.T) {
	ua := &UserAuthentication{}
	token := getMockJWT("new-token", 3600)
	ua.setToken(token)
	assert.Equal(t, token, ua.userToken)
	assert.True(t, ua.userTokenExpiresAt.After(time.Now()))
}

func TestServiceAccountAuthentication_GetAccessToken_ValidToken(t *testing.T) {
	aa := &ServiceAccountAuthentication{
		appToken: getMockJWT("valid-token", 3600),
	}
	token := aa.getAccessToken()
	assert.NotNil(t, token)
	assert.Equal(t, "valid-token", *token)
}

func TestServiceAccountAuthentication_GetAccessToken_NilToken(t *testing.T) {
	aa := &ServiceAccountAuthentication{
		appToken: nil,
	}
	token := aa.getAccessToken()
	assert.Nil(t, token)
}

func TestServiceAccountAuthentication_GetAccessToken_EmptyToken(t *testing.T) {
	aa := &ServiceAccountAuthentication{
		appToken: getMockJWT("", 3600),
	}
	token := aa.getAccessToken()
	assert.Nil(t, token)
}

func TestServiceAccountAuthentication_IsTokenValid_ValidToken(t *testing.T) {
	aa := &ServiceAccountAuthentication{
		appToken:          getMockJWT("valid-token", 3600),
		appTokenExpiresAt: time.Now().Add(1 * time.Minute),
	}
	isValid := aa.isTokenValid()
	assert.True(t, isValid)
}

func TestServiceAccountAuthentication_IsTokenValid_ExpiredToken(t *testing.T) {
	aa := &ServiceAccountAuthentication{
		appToken:          getMockJWT("expired-token", 3600),
		appTokenExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	isValid := aa.isTokenValid()
	assert.False(t, isValid)
}

func TestServiceAccountAuthentication_IsTokenValid_NilToken(t *testing.T) {
	aa := &ServiceAccountAuthentication{
		appToken: nil,
	}
	isValid := aa.isTokenValid()
	assert.False(t, isValid)
}

func TestServiceAccountAuthentication_IsTokenValid_EmptyToken(t *testing.T) {
	aa := &ServiceAccountAuthentication{
		appToken:          getMockJWT("", 3600),
		appTokenExpiresAt: time.Now().Add(1 * time.Minute),
	}
	isValid := aa.isTokenValid()
	assert.False(t, isValid)
}

func TestServiceAccountAuthentication_SetToken(t *testing.T) {
	aa := &ServiceAccountAuthentication{}
	token := getMockJWT("new-token", 3600)
	aa.setToken(token)
	assert.Equal(t, token, aa.appToken)
	assert.True(t, aa.appTokenExpiresAt.After(time.Now()))
}

func TestServiceAccountAuthentication_GetClientSecret_ValidSecret(t *testing.T) {
	aa := &ServiceAccountAuthentication{
		appSecret: getMockCredential(clientSecret),
	}
	secret := aa.getClientSecret()
	assert.NotNil(t, secret)
	assert.Equal(t, clientSecret, *secret)
}

func TestServiceAccountAuthentication_GetClientSecret_NilSecret(t *testing.T) {
	aa := &ServiceAccountAuthentication{
		appSecret: nil,
	}
	secret := aa.getClientSecret()
	assert.Nil(t, secret)
}

func TestServiceAccountAuthentication_GetClientSecret_NilValue(t *testing.T) {
	aa := &ServiceAccountAuthentication{
		appSecret: &gocloak.CredentialRepresentation{Value: nil},
	}
	secret := aa.getClientSecret()
	assert.Nil(t, secret)
}

func TestGetAppAdminAccessToken_ValidToken(t *testing.T) {
	authenticator, _ := initTest(t)
	authenticator.appAuth.appToken = getMockJWT("valid-token", 3600)
	authenticator.appAuth.appTokenExpiresAt = time.Now().Add(1 * time.Hour)

	token, err := authenticator.getAppAdminAccessToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "valid-token", *token)
}

func TestGetAppAdminAccessToken_ExpiredToken_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.appAuth.appToken = getMockJWT("expired-token", 3600)
	authenticator.appAuth.appTokenExpiresAt = time.Now().Add(-1 * time.Hour)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)
	authenticator.appAuth.appSecret = getMockCredential(clientSecret)

	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, clientSecret, MasterRealmName).Return(getMockJWT("new-token", 3600), nil)

	token, err := authenticator.getAppAdminAccessToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "new-token", *token)
	mockKcClient.AssertExpectations(t)
}

func TestGetAppAdminAccessToken_ExpiredToken_LoginFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.appAuth.appToken = getMockJWT("expired-token", 3600)
	authenticator.appAuth.appTokenExpiresAt = time.Now().Add(-1 * time.Hour)
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)
	authenticator.appAuth.appSecret = getMockCredential(clientSecret)

	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, clientSecret, MasterRealmName).Return(nil, errors.New("login failed"))

	token, err := authenticator.getAppAdminAccessToken(context.Background())

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "failed to login admin app")
	mockKcClient.AssertExpectations(t)
}

func TestGetAppAdminAccessToken_NoToken_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.appAuth.appToken = nil
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)
	authenticator.appAuth.appSecret = getMockCredential(clientSecret)

	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, clientSecret, MasterRealmName).Return(getMockJWT("new-token", 3600), nil)

	token, err := authenticator.getAppAdminAccessToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "new-token", *token)
	mockKcClient.AssertExpectations(t)
}

func TestGetAppAdminAccessToken_NoToken_NoSecret_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.appAuth.appToken = nil
	authenticator.appAuth.appSecret = nil
	authenticator.userAuth.userToken = getMockJWT(adminToken, 3600)

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT(adminToken, 3600), nil)
	mockKcClient.On("GetClients", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return([]*gocloak.Client{getMockClient(ServiceClientID, clientSecret)}, nil)
	mockKcClient.On("GetClientSecret", mock.Anything, adminToken, MasterRealmName, mock.Anything).Return(getMockCredential(clientSecret), nil)
	mockKcClient.On("LoginClient", mock.Anything, ServiceClientID, clientSecret, MasterRealmName).Return(getMockJWT("new-token", 3600), nil)

	token, err := authenticator.getAppAdminAccessToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "new-token", *token)
	mockKcClient.AssertExpectations(t)
}

func TestGetUserAdminAccessToken_ValidToken(t *testing.T) {
	authenticator, _ := initTest(t)
	authenticator.userAuth.userToken = getMockJWT("valid-token", 3600)
	authenticator.userAuth.userTokenExpiresAt = time.Now().Add(1 * time.Hour)

	token, err := authenticator.getUserAdminAccessToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "valid-token", *token)
}

func TestGetUserAdminAccessToken_ExpiredToken_RefreshSuccess(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT("expired-token", 3600)
	authenticator.userAuth.userToken.RefreshToken = "refresh-token"
	authenticator.userAuth.userTokenExpiresAt = time.Now().Add(-1 * time.Hour)

	mockKcClient.On("RefreshToken", mock.Anything, "refresh-token", MasterRealmAdminClient, "", MasterRealmName).Return(getMockJWT("new-token", 3600), nil)

	token, err := authenticator.getUserAdminAccessToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "new-token", *token)
	mockKcClient.AssertExpectations(t)
}

func TestGetUserAdminAccessToken_ExpiredToken_RefreshFails_LoginSuccess(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT("expired-token", 3600)
	authenticator.userAuth.userToken.RefreshToken = "refresh-token"
	authenticator.userAuth.userTokenExpiresAt = time.Now().Add(-1 * time.Hour)

	mockKcClient.On("RefreshToken", mock.Anything, "refresh-token", MasterRealmAdminClient, "", MasterRealmName).Return(nil, errors.New("refresh failed"))
	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT("new-token", 3600), nil)

	token, err := authenticator.getUserAdminAccessToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "new-token", *token)
	mockKcClient.AssertExpectations(t)
}

func TestGetUserAdminAccessToken_ExpiredToken_RefreshFails_LoginFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT("expired-token", 3600)
	authenticator.userAuth.userToken.RefreshToken = "refresh-token"
	authenticator.userAuth.userTokenExpiresAt = time.Now().Add(-1 * time.Hour)

	mockKcClient.On("RefreshToken", mock.Anything, "refresh-token", MasterRealmAdminClient, "", MasterRealmName).Return(nil, errors.New("refresh failed"))
	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(nil, errors.New("login failed"))

	token, err := authenticator.getUserAdminAccessToken(context.Background())

	assert.Error(t, err)
	assert.Nil(t, token)
	mockKcClient.AssertExpectations(t)
}

func TestGetUserAdminAccessToken_NoToken_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = nil

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT("new-token", 3600), nil)

	token, err := authenticator.getUserAdminAccessToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "new-token", *token)
	mockKcClient.AssertExpectations(t)
}

func TestGetUserAdminAccessToken_NoToken_LoginFails(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = nil

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(nil, errors.New("login failed"))

	token, err := authenticator.getUserAdminAccessToken(context.Background())

	assert.Error(t, err)
	assert.Nil(t, token)
	mockKcClient.AssertExpectations(t)
}

func TestGetUserAdminAccessToken_NoRefreshToken_Success(t *testing.T) {
	authenticator, mockKcClient := initTest(t)
	authenticator.userAuth.userToken = getMockJWT("expired-token", 3600)
	authenticator.userAuth.userToken.RefreshToken = ""
	authenticator.userAuth.userTokenExpiresAt = time.Now().Add(-1 * time.Hour)

	mockKcClient.On("LoginAdmin", mock.Anything, testUsername, testPassword, MasterRealmName).Return(getMockJWT("new-token", 3600), nil)

	token, err := authenticator.getUserAdminAccessToken(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "new-token", *token)
	mockKcClient.AssertExpectations(t)
}

func getMockJWT(accessToken string, expiresIn int) *gocloak.JWT {
	return &gocloak.JWT{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	}
}

func getMockClient(clientID, secret string) *gocloak.Client {
	return &gocloak.Client{
		ID:       ptr.To("client-id"),
		ClientID: ptr.To(clientID),
		Secret:   ptr.To(secret),
	}
}

func getMockUser(userID string) *gocloak.User {
	return &gocloak.User{
		ID: ptr.To(userID),
	}
}

func getMockRole(roleName string) *gocloak.Role {
	return &gocloak.Role{
		Name: ptr.To(roleName),
	}
}

func getMockCredential(value string) *gocloak.CredentialRepresentation {
	return &gocloak.CredentialRepresentation{
		Value: ptr.To(value),
	}
}
