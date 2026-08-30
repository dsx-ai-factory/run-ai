package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	auth_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	clusters_api "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	clients "github.com/run-ai/runai/control-plane/org-unit-service/pkg/clients/backend"
	org_unit_config "github.com/run-ai/runai/control-plane/org-unit-service/pkg/config"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/helper_job"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/overtime_metrics_writer"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/pubsub"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/resources_sync"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/router"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/settings"
	tenants_api "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
	datavolumes_service_client "github.com/run-ai/runai/runai-clients/datavolumes-service-client"
	workloads_client "github.com/run-ai/runai/runai-clients/workloads-client"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/run-ai/runai/runai-common-packages/runaihttp"
	"github.com/run-ai/runai/runai-common-packages/s3"
	setting "github.com/run-ai/runai/runai-common-packages/setting"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/tracing"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	metrics.Init(nil)
	tracing.Init(nil)
	audit.Init(nil)
	if err := org_unit_config.Init(); err != nil {
		logging.Fatalf("failed to init configuration: %v", err)
	}
	dao_utils.Init(dao_utils.ConstructConfig(org_unit_config.Schema))

	backendClient := clients.NewBackendClient(config.Get[*org_unit_config.Server]().TempBackendSettingsUrl)
	setting.Init(&setting.SettingConfig{
		BackendClient: backendClient,
	}, nil)
	if err := setting.Register[*settings.LimitQuotaOverSubscription](); err != nil {
		logging.Errorf("Failed to register LimitQuotaOverSubscription setting: %v", err)
	}

	err := cache.Init(nil)
	if err != nil {
		logging.Errorf("Failed to initialize cache: %v", err)
	}
	database.ClearCache()
	auth_client.Init(nil)
	if !helper_job.ShouldRunAsJob() {
		pubsub.Init()
	}
	s3.Init(nil)
	logging.Infof("Initializing cluster capabilities checker")
	serverConf := config.Get[*org_unit_config.Server]()
	generalConf := config.Get[*org_unit_config.General]()
	checker.InitClusterCapabilitiesChecker(serverConf.ClustersServiceUrl, nil, generalConf.LogLevel)
}

func main() {
	serverConf := config.Get[*org_unit_config.Server]()

	tmClient := tenants_api.NewTenantsManagerClient(serverConf.TenantsManagerUrl)
	clusterClient := clusters_api.NewClusterClient(serverConf.ClustersServiceUrl)

	if helper_job.ShouldRunAsJob() {
		startHelperJob(tmClient, clusterClient)
	}

	runSettingsMigration()

	logging.Infof("Starting Run:AI org-unit-service")

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	clustersClient := clusters_api.NewCachedClusterClient(serverConf.ClustersServiceUrl)
	workloadsClient := workloads_client.NewWorkloadsClient(serverConf.WorkloadsServiceUrl)
	dvsClient := datavolumes_service_client.NewDatavolumesClient(serverConf.DatavolumesServiceUrl)

	authorizationClient := auth_client.NewAuthorizationClient()

	runaihttp.ServeHTTP(
		router.CreateRouter(authorizationClient, clustersClient, tmClient, workloadsClient, dvsClient),
		nil,
	)

	logging.Infof("Shutting down org-unit-service")
	runaipubsub.Close()
	os.Exit(0)
}

func runSettingsMigration() {
	logging.Infof("running settings migration...")
	var errs []error
	if err := setting.Migrate[*settings.LimitQuotaOverSubscription](); err != nil {
		errs = append(errs, fmt.Errorf("LimitQuotaOverSubscription: %s", err.Error()))
	}
	if len(errs) > 0 {
		logging.Fatalf("failed to migrate settings: %v", errors.Join(errs...))
	}
	logging.Infof("settings migration completed successfully")
}

func startHelperJob(tenantsClient tenants_api.ITenantsManager, clusterClient clusters_api.IClusterClient) {
	logging.Infof("Starting Run:AI org-unit-helper")

	stopper := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		close(stopper)
		cancel()
	}()

	overtime_metrics_writer.InitGoRoutines(ctx, stopper, tenantsClient)
	resources_sync.InitGoRoutines(ctx, stopper, tenantsClient, clusterClient)
	database.InitCleaner(stopper)

	<-stopper
	os.Exit(0)
}
