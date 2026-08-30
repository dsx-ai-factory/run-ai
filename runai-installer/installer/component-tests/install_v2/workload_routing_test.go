package install_v2_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorkloadRouting_SubdomainSupport verifies the cluster chart's
// clusterConfig.global.subdomainSupport value that the installer emits for each
// workload-routing selection. Host is the default; an explicit setting wins;
// and a subdomainSupport custom helm value drives routing when no first-class
// setting is given (and agrees with it, so no contradiction is flagged).
func TestWorkloadRouting_SubdomainSupport(t *testing.T) {
	const key = "clusterConfig.global.subdomainSupport"

	cases := []struct {
		name  string
		setup func(*installconfig.Config)
		want  string
	}{
		{
			name:  "default is host",
			setup: func(*installconfig.Config) {},
			want:  key + "=true",
		},
		{
			name:  "explicit host",
			setup: func(c *installconfig.Config) { c.Cluster.WorkloadRouting = "host" },
			want:  key + "=true",
		},
		{
			name:  "explicit path",
			setup: func(c *installconfig.Config) { c.Cluster.WorkloadRouting = "path" },
			want:  key + "=false",
		},
		{
			name: "path derived from subdomainSupport override",
			setup: func(c *installconfig.Config) {
				c.Cluster.CustomHelmValues = map[string]any{key: false}
			},
			want: key + "=false",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			helm := harness.NewFakeHelmWithReleases()
			k8s := &harness.FakeK8sExecutor{}
			_, provider := harness.NewFakeRunAIClientProvider(helm)
			deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

			eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
			cfg, opts := connectedConfig(t, eventsPath, false)
			tc.setup(cfg)
			term, _, _ := newTestTerminal()

			require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

			calls := helm.CallsByRelease("runai-cluster")
			require.Len(t, calls, 1)
			assert.Contains(t, calls[0].Values, tc.want)
			assert.Equal(t, 1, countWithPrefix(calls[0].Values, key+"="), "subdomainSupport must be emitted once")
		})
	}
}
