package main

import (
	"flag"
	"time"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"

	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"

	auth_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	service_conf "github.com/run-ai/runai/control-plane/workloads-manager/pkg/config"
	pubsub "github.com/run-ai/runai/control-plane/workloads-manager/pkg/pubsub"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/router"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/services"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/settings"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/config"
	daoutils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaihttp"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
	"github.com/run-ai/runai/runai-common-packages/setting"
	"github.com/run-ai/runai/runai-common-packages/tracing"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	if _, err := metrics.Init(nil); err != nil {
		logging.Errorf("Failed to initialize metrics: %v", err)
	}
	if _, err := tracing.Init(nil); err != nil {
		logging.Errorf("Failed to initialize tracing: %v", err)
	}
	service_conf.Init()

	servicesConfig := config.Get[*service_conf.Services]()

	auth_client.Init(&auth_client.Config{
		Url: servicesConfig.AuthorizationUrl,
	})

	if err := cache.Init(nil); err != nil {
		logging.Errorf("Failed to initialize cache: %v", err)
	}

	services.RegisterSuspendDefCacheStore()

	policies.InitConnection(servicesConfig.PolicyServiceUrl)

	checker.InitClusterCapabilitiesChecker(servicesConfig.ClusterServiceUrl, &checker.ClustersCacheConfig{
		ExpirationSec: servicesConfig.ClusterCacheExpirationSec,
		CleanupSec:    servicesConfig.ClusterCacheCleanupSec,
	}, config.Get[*logging.Config]().LogLevel)

	daoutils.Init(config.Get[*daoutils.DataBaseConfig]())

	setting.Init(nil, nil)
	if err := setting.Register[*settings.YamlSubmission](); err != nil {
		logging.Errorf("Failed to register YamlSubmission setting: %v", err)
	}

	runaipubsub.SetLogger(logging.Actual())
	_, err := runaipubsub.InitAsync(nil)
	if err != nil {
		logging.Fatalf("Failed to initialize pubsub: %v", err)
	}
}

func main() {
	logging.Infof("Starting Run:AI workloads-manager")

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully waits for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	pubsub.Register()
	defer runaipubsub.Close()

	runaihttp.ServeHTTP(
		router.CreateRouter(),
		nil,
	)
}
