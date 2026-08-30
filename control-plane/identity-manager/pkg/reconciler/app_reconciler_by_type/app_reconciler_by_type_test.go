package app_reconciler_by_type

import (
	"context"
	"errors"
	"testing"

	iamMock "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	dal_mocks "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/mocks"
	pkgModels "github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	realm_client "github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/client"
	reconciler_mocks "github.com/run-ai/runai/control-plane/identity-manager/pkg/reconciler/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

const (
	testTenantId = "tenant-1001"
	testVersion  = "v1"
	testRealm    = "test-realm"
)

func initTest(t *testing.T) (*iamMock.MockServicesContainer, *reconciler_mocks.TenantResolver, *dal_mocks.ReconcilerVersionsDal, *AppsReconciler) {
	mockContainer := iamMock.NewMockServicesContainer()
	tenantResolverMock := reconciler_mocks.NewTenantResolver(t)
	versionsDalMock := dal_mocks.NewReconcilerVersionsDal(t)

	subject := &AppsReconciler{
		iamServices:           *mockContainer.ToServicesContainer(),
		tenantResolver:        tenantResolverMock,
		reconcilerVersionsDal: versionsDalMock,
		version:               testVersion,
		config: AppsReconcilerConfig{
			hostTemplate: "example-host.com",
			appHost:      "example-app-host.com",
		},
	}
	return mockContainer, tenantResolverMock, versionsDalMock, subject
}

func clientTypeFilter() *iamModels.ClientFilter {
	clientType := realm_client.ClientTypeUserApp
	return &iamModels.ClientFilter{ClientType: &clientType}
}

func TestReconcile_EmptyTenantId_ReturnsError(t *testing.T) {
	_, tenantResolverMock, _, subject := initTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return("")

	err := subject.Reconcile(context.Background(), testRealm, false)

	assert.ErrorContains(t, err, "failed to get tenant id for realm test-realm")
}

func TestReconcile_SameVersion_SkipsReconcile(t *testing.T) {
	mockContainer, tenantResolverMock, versionsDalMock, subject := initTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return(testTenantId)
	versionsDalMock.On("GetReconcilerVersionsByTenantId", mock.Anything, testTenantId).
		Return(&pkgModels.ReconcilerVersions{TenantId: testTenantId, AccessKeys: testVersion}, nil)

	err := subject.Reconcile(context.Background(), testRealm, false)

	assert.NoError(t, err)
	mockContainer.MockClientService.AssertNotCalled(t, "GetClients")
}

func TestReconcile_DifferentVersion_ReconcilesCalled(t *testing.T) {
	mockContainer, tenantResolverMock, versionsDalMock, subject := initTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return(testTenantId)
	versionsDalMock.On("GetReconcilerVersionsByTenantId", mock.Anything, testTenantId).
		Return(&pkgModels.ReconcilerVersions{TenantId: testTenantId, AccessKeys: "old-version"}, nil)
	versionsDalMock.On("UpdateAccessKeysReconcilerVersions", mock.Anything, testTenantId, testVersion).Return(nil)

	apps := []iamModels.Client{
		{
			ClientID:   ptr.To("access-key-1"),
			Attributes: &iamModels.ClientAttributes{CreatedBy: ptr.To("user@example.com")},
		},
	}
	mockContainer.MockClientService.On("GetClients", mock.Anything, testRealm, clientTypeFilter()).Return(apps, nil)
	mockContainer.MockClientService.On("GetClientByClientId", mock.Anything, testRealm, "access-key-1").Return(&apps[0], nil)
	mockContainer.MockClientService.On("UpdateClient", mock.Anything, testRealm, mock.Anything).Return(nil)

	err := subject.Reconcile(context.Background(), testRealm, false)

	assert.NoError(t, err)
}

func TestReconcile_VersionDalError_ReconcilesCalled(t *testing.T) {
	mockContainer, tenantResolverMock, versionsDalMock, subject := initTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return(testTenantId)
	versionsDalMock.On("GetReconcilerVersionsByTenantId", mock.Anything, testTenantId).
		Return(nil, errors.New("db error"))
	versionsDalMock.On("UpdateAccessKeysReconcilerVersions", mock.Anything, testTenantId, testVersion).Return(nil)

	apps := []iamModels.Client{
		{
			ClientID:   ptr.To("access-key-1"),
			Attributes: &iamModels.ClientAttributes{CreatedBy: ptr.To("user@example.com")},
		},
	}
	mockContainer.MockClientService.On("GetClients", mock.Anything, testRealm, clientTypeFilter()).Return(apps, nil)
	mockContainer.MockClientService.On("GetClientByClientId", mock.Anything, testRealm, "access-key-1").Return(&apps[0], nil)
	mockContainer.MockClientService.On("UpdateClient", mock.Anything, testRealm, mock.Anything).Return(nil)

	err := subject.Reconcile(context.Background(), testRealm, false)

	assert.NoError(t, err)
}

func TestReconcile_Force_AlwaysReconciles(t *testing.T) {
	mockContainer, tenantResolverMock, versionsDalMock, subject := initTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return(testTenantId)
	versionsDalMock.On("UpdateAccessKeysReconcilerVersions", mock.Anything, testTenantId, testVersion).Return(nil)

	apps := []iamModels.Client{
		{
			ClientID:   ptr.To("access-key-1"),
			Attributes: &iamModels.ClientAttributes{CreatedBy: ptr.To("user@example.com")},
		},
	}
	mockContainer.MockClientService.On("GetClients", mock.Anything, testRealm, clientTypeFilter()).Return(apps, nil)
	mockContainer.MockClientService.On("GetClientByClientId", mock.Anything, testRealm, "access-key-1").Return(&apps[0], nil)
	mockContainer.MockClientService.On("UpdateClient", mock.Anything, testRealm, mock.Anything).Return(nil)

	err := subject.Reconcile(context.Background(), testRealm, true)

	assert.NoError(t, err)
	versionsDalMock.AssertNotCalled(t, "GetReconcilerVersionsByTenantId")
}

func TestReconcile_GetClientsError(t *testing.T) {
	mockContainer, tenantResolverMock, versionsDalMock, subject := initTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return(testTenantId)
	versionsDalMock.On("GetReconcilerVersionsByTenantId", mock.Anything, testTenantId).
		Return(nil, errors.New("db error"))
	mockContainer.MockClientService.On(
		"GetClients", mock.Anything, testRealm, clientTypeFilter(),
	).Return(nil, errors.New("iam unavailable"))

	err := subject.Reconcile(context.Background(), testRealm, false)

	assert.ErrorContains(t, err, "iam unavailable")
}

func TestReconcile_MissingCreatedByAttribute_ReturnsError(t *testing.T) {
	mockContainer, tenantResolverMock, versionsDalMock, subject := initTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return(testTenantId)
	versionsDalMock.On("GetReconcilerVersionsByTenantId", mock.Anything, testTenantId).
		Return(nil, errors.New("db error"))

	apps := []iamModels.Client{
		{ClientID: ptr.To("access-key-no-attrs")},
	}
	mockContainer.MockClientService.On("GetClients", mock.Anything, testRealm, clientTypeFilter()).Return(apps, nil)

	err := subject.Reconcile(context.Background(), testRealm, false)

	assert.ErrorContains(t, err, "missing 'createdBy' attribute")
}
