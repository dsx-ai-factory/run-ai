package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	pending_deletion_cache "github.com/run-ai/runai/control-plane/authorization/pkg/cache"
	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/authorization/pkg/clients"
	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/authorization/pkg/pubsub"
	"github.com/run-ai/runai/control-plane/authorization/pkg/router"
	"github.com/run-ai/runai/control-plane/authorization/pkg/scopes_sync"
	"github.com/run-ai/runai/control-plane/authorization/pkg/settings"
	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	tenants_api "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
	org_unit_client "github.com/run-ai/runai/runai-clients/org-unit-service-client"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaihttp"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
	"github.com/run-ai/runai/runai-common-packages/setting"
	"github.com/run-ai/runai/runai-common-packages/tracing"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	metrics.Init(nil)
	tracing.Init(nil)
	audit.Init(nil)
	if err := authorization_config.Init(); err != nil {
		logging.Fatalf("Failed to initialize authorization config: %v", err)
	}
	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("Failed to initialize database: %v", err)
	}
	dao.InitStaticPermissionSets()
	dao.InitStaticLegacyRoles()
	dao.InitStaticRoles()
	if err := cache.Init(nil); err != nil {
		logging.Errorf("Failed to initialize cache: %v", err)
	}
	authorization_client.Init(&authorization_client.Config{
		Url: fmt.Sprintf("http://localhost:%d", config.Get[*authorization_config.Server]().ServicePort),
	})
	pubsub.Init()

	backendClient := clients.NewBackendClient(config.Get[*authorization_config.Server]().TempBackendSettingsUrl)
	setting.Init(&setting.SettingConfig{
		BackendClient: backendClient,
	}, nil)

	if err := setting.Register[*settings.EnableScopedAuthorization](); err != nil {
		logging.Errorf("Failed to register EnableScopedAuthorization setting: %v", err)
	}
	if err := setting.Register[*settings.TenantIsCspManaged](); err != nil {
		logging.Errorf("Failed to register TenantIsCspManaged setting: %v", err)
	}

	logging.Infof("Initializing cluster capabilities checker")
	checker.InitClusterCapabilitiesChecker(config.Get[*authorization_config.Server]().ClusterServiceUrl, nil, "info")
}

func main() {
	logging.Infof("Starting authorization service. <<<<<<<<<<<<<<< trigger build")

	stopper := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		close(stopper)
		cancel()
		if err := runaidb.Close(); err != nil {
			logging.Errorf("Failed to close database: %v", err)
		}
	}()

	runSettingsMigration()

	initAccessRulesCleaner(ctx, stopper)
	initPendingDeletionCache(ctx)

	runaihttp.ServeHTTP(
		router.CreateRouter(),
		nil,
	)
	runaipubsub.Close()

	<-stopper
	os.Exit(0)
}

func initPendingDeletionCache(ctx context.Context) {
	clustersClient := cluster_service_client.NewClusterClient(config.Get[*authorization_config.Server]().ClusterServiceUrl)
	if err := pending_deletion_cache.InitPendingDeletionCache(ctx, clustersClient); err != nil {
		logging.Errorf("Failed to initialize pending deletion cache: %v", err)
	}
}

func runSettingsMigration() {
	logging.Infof("running settings migration...")
	var errs []error
	if err := setting.Migrate[*settings.EnableScopedAuthorization](); err != nil {
		errs = append(errs, fmt.Errorf("EnableScopedAuthorization: %s", err.Error()))
	}
	if err := setting.Migrate[*settings.TenantIsCspManaged](); err != nil {
		errs = append(errs, fmt.Errorf("TenantIsCspManaged: %s", err.Error()))
	}
	if len(errs) > 0 {
		logging.Fatalf("failed to migrate settings: %v", errors.Join(errs...))
	}
	logging.Infof("settings migration completed successfully")
}

func initAccessRulesCleaner(ctx context.Context, stopCh <-chan struct{}) {
	tenantsClient := tenants_api.NewTenantsManagerClient(config.Get[*authorization_config.Server]().TenantServiceUrl)
	orgUnitClient := org_unit_client.NewOrgUnitClient(config.Get[*authorization_config.Server]().OrgUnitServiceUrl)

	scopes_sync.InitGoRoutines(ctx, stopCh, tenantsClient, orgUnitClient)
}
