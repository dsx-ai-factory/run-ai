package test_deprecated

import (
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	rcpconfig "github.com/run-ai/runai/runai-common-packages/config"
	"k8s.io/apimachinery/pkg/runtime"
	"sync"
	"testing"
)

const (
	_migValidNameOne = "nvidia.com/mig-1g.5gb"  //nolint:unused
	_migValidNameTwo = "nvidia.com/mig-2g.10gb" //nolint:unused
)

var (
	_scheme *runtime.Scheme //nolint:unused
	_once   sync.Once       //nolint:unused
)

const ClusterIdForTests = "7daba6ee-2ea9-4af3-b735-a364bcdff72f"
const ClusterDomainForTests = "some-url.run.ai"
const ClientNameForTests = "clientName"
const ClientSecretForTests = "clientSecret"
const syncEventsIntervalSeconds = 2
const numOfPodsPerRequest = 3
const numOfWorkloadsPerRequest = 4
const numOfEventsPerRequest = 5

const tokenResponse = "token"

// MockEnvDeprecated sets up environment variables for tests. Notice that after test is done env variables are cleaned up.
func MockEnvDeprecated(t *testing.T) {
	t.Setenv("CLUSTER_DOMAIN", ClusterDomainForTests)
	t.Setenv("ENDPOINT", "127.0.0.1")
	t.Setenv("LIVENESS_INTERVAL_SECONDS", "1")
	t.Setenv("MAX_NUM_OF_EVENTS_PER_REQUEST", fmt.Sprintf("%d", numOfEventsPerRequest))
	t.Setenv("SYNC_EVENTS_INTERVAL_SECONDS", fmt.Sprintf("%d", syncEventsIntervalSeconds))
	t.Setenv("SSL_CERTIFICATE_VERIFICATION", "false")
	t.Setenv("MAX_NUM_OF_PODS_PER_REQUEST", fmt.Sprintf("%d", numOfPodsPerRequest))
	t.Setenv("MAX_NUM_OF_WORKLOADS_PER_REQUEST", fmt.Sprintf("%d", numOfWorkloadsPerRequest))
	t.Setenv("ENABLE_NODE_NVIDIA_COMPONENTS_READINESS", "true")

	t.Setenv("RATE_LIMITER_WORKLOADS_DELAY_SECONDS", "0")
	t.Setenv("RATE_LIMITER_WORKLOADS_NUM_OF_WORKERS", "1")
	t.Setenv("RATE_LIMITER_REGULAR_DELAY_SECONDS", "0")
	t.Setenv("RATE_LIMITER_REGULAR_NUM_OF_WORKERS", "1")
	// deprecated flags:
	t.Setenv("SYNC_PODS_WITH_WORKLOADS", "false")

	c, err := rcpconfig.GetR[*config.ClusterSyncConfig]()
	if err != nil {
		t.Fatalf("Failed to get agent config: %v", err)
	}

	c.ClusterUuid = ClusterIdForTests
	c.ClientName = ClientNameForTests
	c.ClientSecret = ClientSecretForTests

	err = rcpconfig.Set(c)
	if err != nil {
		t.Fatalf("Failed to set agent config: %v", err)
	}

}
