package connections

import (
	"context"
	"errors"
	"testing"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/connections/mocks"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestConnectionManager_GetConnection_NoMetricStore(t *testing.T) {
	mockGetter := mocks.NewMetricStoreGetter(t)
	ctx := context.Background()
	mockGetter.On("GetByTenantId", ctx, int32(1)).Return(nil, nil)
	manager := NewManager(mockGetter)
	connection, err := manager.GetConnection(ctx, int32(1))
	assert.EqualError(t, err, "metric store not found for tenant 1")
	assert.Nil(t, connection)
}

func TestConnectionManager_GetConnection_NoAuth(t *testing.T) {
	mockGetter := mocks.NewMetricStoreGetter(t)
	ctx := context.Background()
	mockGetter.On("GetByTenantId", ctx, int32(1)).Return(&api.MetricStore{
		TenantId: int32(1),
		MetricStoreDetails: &api.MetricStoreDetails{
			Read: api.MetricStoreEndpoint{
				Url: "test",
			},
		},
	}, nil)
	manager := NewManager(mockGetter)
	connection, err := manager.GetConnection(ctx, int32(1))
	assert.NoError(t, err)
	assert.NotNil(t, connection)
	assert.Equal(t, "test", connection.Url)
	assert.Equal(t, &EmptyAuthConfigurator{}, connection.AuthConfigurator)
}

func TestConnectionManager_GetConnection_BasicAuth(t *testing.T) {
	mockGetter := mocks.NewMetricStoreGetter(t)
	ctx := context.Background()
	mockGetter.On("GetByTenantId", ctx, int32(1)).Return(&api.MetricStore{
		TenantId: int32(1),
		MetricStoreDetails: &api.MetricStoreDetails{
			Read: api.MetricStoreEndpoint{
				Url: "test",
				Auth: &api.MetricStoreEndpointAuth{
					Basic: &api.MetricStoreEndpointAuthBasic{
						Username: "user",
						Password: "pass",
					},
					Type: api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
				},
			},
		},
	}, nil)
	manager := NewManager(mockGetter)
	connection, err := manager.GetConnection(ctx, int32(1))
	assert.NoError(t, err)
	assert.NotNil(t, connection)
	assert.Equal(t, "test", connection.Url)
	assert.Equal(t, &BasicAuthConfigurator{Username: "user", Password: "pass"}, connection.AuthConfigurator)
}

func TestConnectionManager_GetConnection_GetterError(t *testing.T) {
	mockGetter := mocks.NewMetricStoreGetter(t)
	ctx := context.Background()
	mockGetter.On("GetByTenantId", ctx, int32(1)).Return(nil, errors.New("error"))
	manager := NewManager(mockGetter)
	connection, err := manager.GetConnection(ctx, int32(1))
	assert.Error(t, err)
	assert.Nil(t, connection)
}
