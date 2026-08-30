package flows

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/smithy-go/ptr"
	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authMock "github.com/run-ai/runai/control-plane/authorization/pkg/client/mocks"
	tm_config "github.com/run-ai/runai/control-plane/tenants-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	mngTenantName = "mng-tenant"
	mngTenantId   = "42"
	mngUsername   = "admin@run.ai"
)

func initCreateMngTenantFlowTest(t *testing.T) (authClientMock *authMock.IAuthorizationClient, testSubject *CreateMngTenantFlow) {
	authClientMock = authMock.NewIAuthorizationClient(t)
	testSubject = &CreateMngTenantFlow{
		authClient: authClientMock,
		tenantConf: &tm_config.Tenant{
			MngName:     mngTenantName,
			MngUsername: mngUsername,
		},
	}
	return
}

func TestCreateApplicationAccessRules_CallsBothRoles(t *testing.T) {
	authClientMock, testSubject := initCreateMngTenantFlowTest(t)

	systemScope := authorization_api.InternalCreateAccessRuleRequestScope{
		Id:   mngTenantId,
		Type: authorization_api.SCOPETYPE_SYSTEM,
		Name: ptr.String(mngTenantName),
	}
	tenantScope := authorization_api.InternalCreateAccessRuleRequestScope{
		Id:   mngTenantId,
		Type: authorization_api.SCOPETYPE_TENANT,
		Name: ptr.String(mngTenantName),
	}

	authClientMock.On("SetAppPermissions", mock.Anything, mngTenantId, MNG_APPLICATION_NAME, systemScope, authorization_api.ALLOWEDINTERNALCREATEROLES_CLOUD_OPERATOR).Return(nil)
	authClientMock.On("SetAppPermissions", mock.Anything, mngTenantId, MNG_APPLICATION_NAME, tenantScope, authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR).Return(nil)

	err := testSubject.createApplicationAccessRules(context.Background(), mngTenantId, mngTenantName)

	assert.NoError(t, err)
	authClientMock.AssertNumberOfCalls(t, "SetAppPermissions", 2)
}

func TestCreateApplicationAccessRules_StopsOnCloudOperatorError(t *testing.T) {
	authClientMock, testSubject := initCreateMngTenantFlowTest(t)

	systemScope := authorization_api.InternalCreateAccessRuleRequestScope{
		Id:   mngTenantId,
		Type: authorization_api.SCOPETYPE_SYSTEM,
		Name: ptr.String(mngTenantName),
	}

	authClientMock.On("SetAppPermissions", mock.Anything, mngTenantId, MNG_APPLICATION_NAME, systemScope, authorization_api.ALLOWEDINTERNALCREATEROLES_CLOUD_OPERATOR).Return(assert.AnError)

	err := testSubject.createApplicationAccessRules(context.Background(), mngTenantId, mngTenantName)

	assert.Error(t, err)
	authClientMock.AssertNumberOfCalls(t, "SetAppPermissions", 1)
}

func TestAddConsoleAdminSystemAdminRole_HappyFlow(t *testing.T) {
	authClientMock, testSubject := initCreateMngTenantFlowTest(t)
	httpOK := &http.Response{StatusCode: http.StatusOK}

	authClientMock.On("CreateInternalAccessRule", mock.Anything).Return(authorization_api.ApiCreateInternalAccessRuleRequest{})
	authClientMock.On("CreateInternalAccessRuleExecute", mock.Anything).Return(nil, httpOK, nil)

	err := testSubject.addConsoleAdminSystemAdminRole(context.Background(), mngTenantId, mngTenantName)

	assert.NoError(t, err)
	authClientMock.AssertCalled(t, "CreateInternalAccessRule", mock.Anything)
	authClientMock.AssertCalled(t, "CreateInternalAccessRuleExecute", mock.Anything)
}

func TestAddConsoleAdminSystemAdminRole_ConflictIsIgnored(t *testing.T) {
	authClientMock, testSubject := initCreateMngTenantFlowTest(t)
	httpConflict := &http.Response{StatusCode: http.StatusConflict}

	authClientMock.On("CreateInternalAccessRule", mock.Anything).Return(authorization_api.ApiCreateInternalAccessRuleRequest{})
	authClientMock.On("CreateInternalAccessRuleExecute", mock.Anything).Return(nil, httpConflict, assert.AnError)

	// Should not panic or log fatal — conflict is idempotent
	err := testSubject.addConsoleAdminSystemAdminRole(context.Background(), mngTenantId, mngTenantName)

	assert.NoError(t, err)
	authClientMock.AssertCalled(t, "CreateInternalAccessRuleExecute", mock.Anything)
}
