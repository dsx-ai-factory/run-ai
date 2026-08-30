package main

import (
	"context"
	"flag"
	"os"
	"time"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	"github.com/run-ai/runai/control-plane/workloads/pkg/metrics_exporter"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_crud"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
	"github.com/run-ai/runai/runai-common-packages/setting"

	"github.com/run-ai/runai/control-plane/workloads/pkg/metrics_writer"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	settings_client "github.com/run-ai/runai/control-plane/workloads/pkg/clients/settings"
	workload_settings "github.com/run-ai/runai/control-plane/workloads/pkg/settings"
	"github.com/run-ai/runai/control-plane/workloads/pkg/watchdog"

	"github.com/run-ai/runai/control-plane/workloads/pkg/database"
	helper_job "github.com/run-ai/runai/control-plane/workloads/pkg/helper-job"
	"github.com/run-ai/runai/control-plane/workloads/pkg/notifiy_client"
	"github.com/run-ai/runai/runai-common-packages/runaihttp"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/go_migrations"
	pubsub "github.com/run-ai/runai/control-plane/workloads/pkg/pubsub"
	"github.com/run-ai/runai/control-plane/workloads/pkg/router"
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
	authorization_client.Init(nil)
	config.Init()
	dao_utils.Init(config.Database())
	err := go_migrations.Migrate()
	if err != nil {
		logging.Fatalf("Failed to run go DB migrations: %v", err)
	}

	err = cache.Init(nil)
	if err != nil {
		logging.Errorf("Failed to initialize cache: %v", err)
	}
	// Get configs
	servicesConf := runai_config.Get[*config.ServicesConfig]()

	// init clients
	policies.InitConnection(runai_config.Get[*config.ServicesConfig]().PolicyServiceUrl)
	orgunit.InitConnection(servicesConf.OrgUnitUrl)

	if !helper_job.ShouldRunAsJob() {
		runaipubsub.SetLogger(logging.Actual())
		_, err := runaipubsub.InitAsync(nil)
		if err != nil {
			logging.Fatalf("Failed to initialize pubsub: %v", err)
		}
	}

	checker.InitClusterCapabilitiesChecker(
		servicesConf.ClusterServiceUrl,
		&servicesConf.ClustersCache,
		servicesConf.LogLevel,
	)

	services_crud.InitTemplatePresets()

	backendSettingsClient := settings_client.NewBackendClient(runai_config.Get[*config.ServicesConfig]().TempBackendSettingsUrl)
	setting.Init(&setting.SettingConfig{
		BackendClient: backendSettingsClient,
	}, nil)

	if err := setting.Register[*workload_settings.EnablePolicies](); err != nil {
		logging.Errorf("Failed to register EnablePolicies setting: %v", err)
	}
	if err := setting.Register[*workload_settings.WorkloadSubmissionV2Enabled](); err != nil {
		logging.Errorf("Failed to register WorkloadSubmissionV2Enabled setting: %v", err)
	}
	if err := setting.Register[*workload_settings.EnableTemplatesV2](); err != nil {
		logging.Errorf("Failed to register EnableTemplatesV2 setting: %v", err)
	}
	if err := setting.Register[*workload_settings.EnableHybridSubmission](); err != nil {
		logging.Errorf("Failed to register EnableHybridSubmission setting: %v", err)
	}
	if err := setting.Register[*workload_settings.GpuOverProvisioning](); err != nil {
		logging.Errorf("Failed to register GpuOverProvisioning setting: %v", err)
	}
	if err := setting.Register[*workload_settings.ApplicationSubmission](); err != nil {
		logging.Errorf("Failed to register ApplicationSubmission setting: %v", err)
	}
	if err := setting.Register[*workload_settings.AnyWorkloadSubmissionEnabled](); err != nil {
		logging.Errorf("Failed to register AnyWorkloadSubmissionEnabled setting: %v", err)
	}
}

func main() {
	if helper_job.ShouldRunAsJob() {
		startHelperJob()
	}

	runSettingsMigration()

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	pubsub.Register()
	defer runaipubsub.Close()

	runaihttp.ServeHTTP(router.CreateRouter(), nil)

	logging.Infof("shutting down")
	os.Exit(0)
}

func startHelperJob() {
	stopper := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		close(stopper)
		cancel()
	}()

	database.InitGoRoutines(stopper)
	notifiy_client.InitGoRoutines(ctx, stopper)
	watchdog.InitGoRoutines(stopper)
	metrics_writer.InitGoRoutines(ctx, stopper)
	database.InitCategoryMigrator(stopper)
	metrics_exporter.InitGoRoutines(ctx)

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()
	runaihttp.ServeHTTP(router.CreateHelperRouter(), nil)
	logging.Infof("shutting down")
	<-stopper
	os.Exit(0)
}

func runSettingsMigration() {
	if err := setting.Migrate[*workload_settings.EnablePolicies](); err != nil {
		logging.Fatalf("failed to migrate EnablePolicies: %v", err)
	}
	if err := setting.Migrate[*workload_settings.WorkloadSubmissionV2Enabled](); err != nil {
		logging.Fatalf("failed to migrate WorkloadSubmissionV2Enabled: %v", err)
	}
	if err := setting.Migrate[*workload_settings.EnableTemplatesV2](); err != nil {
		logging.Fatalf("failed to migrate EnableTemplatesV2: %v", err)
	}
	if err := setting.Migrate[*workload_settings.EnableHybridSubmission](); err != nil {
		logging.Fatalf("failed to migrate EnableHybridSubmission: %v", err)
	}
	if err := setting.Migrate[*workload_settings.GpuOverProvisioning](); err != nil {
		logging.Fatalf("failed to migrate GpuOverProvisioning: %v", err)
	}
	logging.Infof("settings migration completed successfully")
}
