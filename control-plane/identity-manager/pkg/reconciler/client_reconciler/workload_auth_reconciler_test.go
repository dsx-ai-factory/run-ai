package client_reconciler

import (
	"context"
	"errors"
	"testing"

	iamMock "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	reconciler_mocks "github.com/run-ai/runai/control-plane/identity-manager/pkg/reconciler/client_reconciler/mocks"
	tenant_mocks "github.com/run-ai/runai/control-plane/identity-manager/pkg/reconciler/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

func initWorkloadAuthReconcilerTest(t *testing.T) (*iamMock.MockServicesContainer, *tenant_mocks.TenantResolver, *reconciler_mocks.ClientReconcilerHelperInterface, *WorkloadAuthReconciler) {
	mockContainer := iamMock.NewMockServicesContainer()
	tenantResolverMock := tenant_mocks.NewTenantResolver(t)
	helperMock := reconciler_mocks.NewClientReconcilerHelperInterface(t)

	subject := &WorkloadAuthReconciler{
		iamServices:            *mockContainer.ToServicesContainer(),
		tenantResolver:         tenantResolverMock,
		clientReconcilerHelper: helperMock,
		config: WorkloadAuthReconcilerConfig{
			hostTemplate: "example-host.com",
			appHost:      "example-app-host.com",
		},
	}
	return mockContainer, tenantResolverMock, helperMock, subject
}

func TestWorkloadAuthReconcile_EmptyTenantIdReturnsError(t *testing.T) {
	_, tenantResolverMock, _, subject := initWorkloadAuthReconcilerTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return("")

	err := subject.Reconcile(context.Background(), "test-realm", false)

	assert.ErrorContains(t, err, "failed to get tenant id for realm test-realm")
}

func TestWorkloadAuthReconcile_GetClientsError(t *testing.T) {
	mockContainer, tenantResolverMock, _, subject := initWorkloadAuthReconcilerTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return("tenant-1001")

	clientType := utils.ClientTypeWorkloadsAuth
	mockContainer.MockClientService.On(
		"GetClients", mock.Anything, "test-realm", &iamModels.ClientFilter{ClientType: &clientType},
	).Return(nil, errors.New("iam unavailable"))

	err := subject.Reconcile(context.Background(), "test-realm", false)

	assert.ErrorContains(t, err, "iam unavailable")
}

func TestWorkloadAuthReconcile_NoClients_Success(t *testing.T) {
	mockContainer, tenantResolverMock, _, subject := initWorkloadAuthReconcilerTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return("tenant-1001")

	clientType := utils.ClientTypeWorkloadsAuth
	mockContainer.MockClientService.On(
		"GetClients", mock.Anything, "test-realm", &iamModels.ClientFilter{ClientType: &clientType},
	).Return([]iamModels.Client{}, nil)

	err := subject.Reconcile(context.Background(), "test-realm", false)

	assert.NoError(t, err)
}

func TestWorkloadAuthReconcile_WithClients_NoClusterId(t *testing.T) {
	mockContainer, tenantResolverMock, helperMock, subject := initWorkloadAuthReconcilerTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return("tenant-1001")

	clients := []iamModels.Client{
		{ClientID: ptr.To("workload-client-1")},
	}
	clientType := utils.ClientTypeWorkloadsAuth
	mockContainer.MockClientService.On(
		"GetClients", mock.Anything, "test-realm", &iamModels.ClientFilter{ClientType: &clientType},
	).Return(clients, nil)
	helperMock.On("ReconcileClient", mock.Anything, "test-realm", mock.Anything).Return(nil)

	err := subject.Reconcile(context.Background(), "test-realm", false)

	assert.NoError(t, err)
	helperMock.AssertNumberOfCalls(t, "ReconcileClient", 1)
}

func TestWorkloadAuthReconcile_WithClients_WithClusterId(t *testing.T) {
	mockContainer, tenantResolverMock, helperMock, subject := initWorkloadAuthReconcilerTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return("tenant-1001")

	clients := []iamModels.Client{
		{
			ClientID:   ptr.To("workload-client-1"),
			Attributes: &iamModels.ClientAttributes{ClusterId: ptr.To("cluster-abc")},
		},
	}
	clientType := utils.ClientTypeWorkloadsAuth
	mockContainer.MockClientService.On(
		"GetClients", mock.Anything, "test-realm", &iamModels.ClientFilter{ClientType: &clientType},
	).Return(clients, nil)
	helperMock.On("ReconcileClient", mock.Anything, "test-realm", mock.Anything).Return(nil)

	err := subject.Reconcile(context.Background(), "test-realm", false)

	assert.NoError(t, err)
	helperMock.AssertNumberOfCalls(t, "ReconcileClient", 1)
}

func TestWorkloadAuthReconcile_ReconcileClientError(t *testing.T) {
	mockContainer, tenantResolverMock, helperMock, subject := initWorkloadAuthReconcilerTest(t)
	tenantResolverMock.On("GetRealmTenantId", mock.Anything, mock.Anything).Return("tenant-1001")

	clients := []iamModels.Client{
		{ClientID: ptr.To("workload-client-1")},
	}
	clientType := utils.ClientTypeWorkloadsAuth
	mockContainer.MockClientService.On(
		"GetClients", mock.Anything, "test-realm", &iamModels.ClientFilter{ClientType: &clientType},
	).Return(clients, nil)
	helperMock.On("ReconcileClient", mock.Anything, "test-realm", mock.Anything).Return(errors.New("reconcile failed"))

	err := subject.Reconcile(context.Background(), "test-realm", false)

	assert.ErrorContains(t, err, "reconcile failed")
}
