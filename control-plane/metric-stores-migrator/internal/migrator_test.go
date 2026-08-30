package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/run-ai/runai/control-plane/metric-stores-migrator/internal/mocks"
	metrics_api "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	pkgapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	api "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
	"github.com/stretchr/testify/require"

	metricsmocks "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/mocks"
)

var ctx = context.Background()

func TestMigrate(t *testing.T) {
	stores := []*api.ExtendedMetricStore{{
		Type: "mimir", DatasourceId: "some-id",
	}}
	mc := metricsmocks.NewMetricStoresClienter(t)
	tc := mocks.NewTenantsClienter(t)
	tc.On("GetTenants", ctx).Return([]api.Tenant{{Id: int32(1)}}, nil)
	tc.On("GetMetricStore", ctx, int32(1)).Return(stores[0], nil)
	mc.On("CreateMetricStore", ctx, &pkgapi.MetricStore{
		TenantId:     int32(1),
		DatasourceId: pkgapi.PtrString("some-id"),
		MetricStoreDetails: *pkgapi.NewNullableMetricStoreDetails(&pkgapi.MetricStoreDetails{
			Type:  pkgapi.METRICSTORETYPE_MIMIR,
			Read:  *getMetricStoreEndpoint(stores[0].Read),
			Write: *getMetricStoreEndpoint(stores[0].Write),
		})}).Return(nil, nil)
	migrator := &Migrator{tenantsClient: tc, metricStoresClient: mc}
	migrator.Migrate(ctx)
	require.Equal(t, migrator.status.Succeed, 1)
}

func TestMigrate_FailedToGetTenants(t *testing.T) {
	// Migrate calls logging.Fatalf (os.Exit(1)) when it cannot fetch tenants, so
	// the path is exercised in a subprocess and we assert it exits non-zero.
	if os.Getenv("BE_MIGRATE_FATAL") == "1" {
		tc := mocks.NewTenantsClienter(t)
		tc.On("GetTenants", ctx).Return(nil, fmt.Errorf("failed to get tenants"))
		migrator := &Migrator{tenantsClient: tc}
		migrator.Migrate(ctx)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMigrate_FailedToGetTenants$")
	cmd.Env = append(os.Environ(), "BE_MIGRATE_FATAL=1")
	err := cmd.Run()

	var exitErr *exec.ExitError
	require.True(t, errors.As(err, &exitErr), "expected Migrate to terminate the process")
	require.False(t, exitErr.Success(), "expected a non-zero exit code")
}

func TestMigrateTenant_FailedToGetMetricStore(t *testing.T) {
	tc := mocks.NewTenantsClienter(t)
	tc.On("GetMetricStore", ctx, int32(1)).Return(nil, fmt.Errorf("failed"))
	migrator := &Migrator{tenantsClient: tc}
	require.Equal(t, migrator.migrateTenant(ctx, &api.Tenant{Id: int32(1)}).Error(), "failed to get metric store from tenants manager: failed")
}

func TestMigrateTenant_MetricStoreNotExist(t *testing.T) {
	mc := metricsmocks.NewMetricStoresClienter(t)
	tc := mocks.NewTenantsClienter(t)
	tc.On("GetMetricStore", ctx, int32(1)).Return(nil, nil)
	mc.On("CreateMetricStore", ctx, pkgapi.NewMetricStore(int32(1))).Return(nil, nil)
	migrator := &Migrator{tenantsClient: tc, metricStoresClient: mc}
	require.Nil(t, migrator.migrateTenant(ctx, &api.Tenant{Id: int32(1)}))
}

func TestMigrateTenant_FailedToCreateNewStore(t *testing.T) {
	stores := []*api.ExtendedMetricStore{{
		Type: "mimir", DatasourceId: "some-id",
	}}
	mc := metricsmocks.NewMetricStoresClienter(t)
	tc := mocks.NewTenantsClienter(t)
	tc.On("GetMetricStore", ctx, int32(1)).Return(stores[0], nil)
	mc.On("CreateMetricStore", ctx, &pkgapi.MetricStore{
		TenantId:     int32(1),
		DatasourceId: pkgapi.PtrString("some-id"),
		MetricStoreDetails: *pkgapi.NewNullableMetricStoreDetails(&pkgapi.MetricStoreDetails{
			Type:  pkgapi.METRICSTORETYPE_MIMIR,
			Read:  *getMetricStoreEndpoint(stores[0].Read),
			Write: *getMetricStoreEndpoint(stores[0].Write),
		})}).Return(nil, fmt.Errorf("failed"))
	migrator := &Migrator{tenantsClient: tc, metricStoresClient: mc}
	require.Equal(t, migrator.migrateTenant(ctx, &api.Tenant{Id: int32(1)}).Error(), "failed to create metric store: failed")
}

func TestMigrateTenant_FailedToCreateNewStoreAllreadyExist(t *testing.T) {
	stores := []*api.ExtendedMetricStore{{
		Type: "mimir", DatasourceId: "some-id",
	}}
	mc := metricsmocks.NewMetricStoresClienter(t)

	tc := mocks.NewTenantsClienter(t)
	tc.On("GetTenants", ctx).Return([]api.Tenant{{Id: int32(1)}}, nil)
	tc.On("GetMetricStore", ctx, int32(1)).Return(stores[0], nil)
	mc.On("CreateMetricStore", ctx, &pkgapi.MetricStore{
		TenantId:     int32(1),
		DatasourceId: pkgapi.PtrString("some-id"),
		MetricStoreDetails: *pkgapi.NewNullableMetricStoreDetails(&pkgapi.MetricStoreDetails{
			Type:  pkgapi.METRICSTORETYPE_MIMIR,
			Read:  *getMetricStoreEndpoint(stores[0].Read),
			Write: *getMetricStoreEndpoint(stores[0].Write),
		})}).Return(nil, &metrics_api.Error{Code: http.StatusConflict, Message: "already exists"})
	migrator := &Migrator{tenantsClient: tc, metricStoresClient: mc}
	migrator.Migrate(ctx)
	require.Equal(t, migrator.status.AlreadyMigrated, 1)
}

func TestMigrateTenant_MetricStoreNotExist_FailedToCreate(t *testing.T) {
	mc := metricsmocks.NewMetricStoresClienter(t)
	tc := mocks.NewTenantsClienter(t)
	tc.On("GetMetricStore", ctx, int32(1)).Return(nil, nil)
	mc.On("CreateMetricStore", ctx, pkgapi.NewMetricStore(int32(1))).Return(nil, fmt.Errorf("failed"))
	migrator := &Migrator{tenantsClient: tc, metricStoresClient: mc}
	require.EqualError(t, migrator.migrateTenant(ctx, &api.Tenant{Id: int32(1)}), "failed to create metric store: failed")
}

func TestMigrateTenant_BadType(t *testing.T) {
	stores := []*api.ExtendedMetricStore{{
		Type: "unknown-type", DatasourceId: "some-id",
	}}
	tc := mocks.NewTenantsClienter(t)
	tc.On("GetMetricStore", ctx, int32(1)).Return(stores[0], nil)
	migrator := &Migrator{tenantsClient: tc}
	require.Equal(t, migrator.migrateTenant(ctx, &api.Tenant{Id: int32(1)}).Error(), "failed to parse metric store type: unknown-type: invalid value 'unknown-type' for MetricStoreType: valid values are [grafanalabs mimir thanos]")
}
