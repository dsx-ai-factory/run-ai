package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config"
	rcpconfig "github.com/run-ai/runai/runai-common-packages/config"
)

const (
	TestClusterId           = "7daba6ee-2aa9-4ab3-b735-a3d4bcdff72f"
	TestClientName          = "clientName"
	TestClientSecret        = "clientSecret"
	agentSyncIntervalMillis = 150

	DefaultTimeout      = 10 * time.Second
	ConsistentlyTimeout = 5 * time.Second
)

// MockEnv Notice that after test is done env variables are cleaned up
func MockEnv(t *testing.T) {
	t.Setenv("ENDPOINT", "127.0.0.1")
	t.Setenv("TLS_REJECT_UNAUTHORIZED", "false")
	t.Setenv("CLUSTER_VERSION", "9.9.9")
	t.Setenv("NAMESPACE", "runai")
	t.Setenv("AGENT_SYNC_INTERVAL_MILLIS", fmt.Sprintf("%d", agentSyncIntervalMillis))
	t.Setenv("DEFAULT_OBJECTS_LIMIT", "2")
	t.Setenv("PROJECTS_LIMIT", "2")
	t.Setenv("DEPARTMENTS_LIMIT", "2")
	t.Setenv("NODE_POOLS_LIMIT", "2")
	t.Setenv("ACCESS_RULES_LIMIT", "2")
	t.Setenv("POLICY_SYNC_ENABLED", "true")

	ag, _ := rcpconfig.GetR[*config.AgentConfig]()
	ag.ClientName = TestClientName
	ag.ClusterUuid = TestClusterId
	ag.ClientSecret = TestClientSecret
	_ = rcpconfig.Set(ag)
}
