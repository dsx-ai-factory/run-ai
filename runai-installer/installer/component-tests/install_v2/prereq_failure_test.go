package install_v2_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConnectedInstall_PrereqFailureAborts verifies that a failure in any single
// prerequisite aborts the install before Run:ai: prereqs install in order until
// the failing one, the prereq_completed summary reports the failure, and the
// Run:ai control plane / cluster are never touched. One case per prerequisite
// confirms each release name is wired correctly and each one triggers the abort.
//
// Prereqs run in this order: prometheus -> gpu-operator -> network-operator. The
// summary outcome is "failed" only when nothing installed before the failure,
// otherwise "partial". (Feature failures, which happen after Run:ai, are covered
// by TestConnectedInstall_NIMOperator_FailureFailsRun.)
func TestConnectedInstall_PrereqFailureAborts(t *testing.T) {
	cases := []struct {
		name          string
		failRelease   string
		wantOutcome   string
		wantInstalled int
		ranBefore     []string
		notRun        []string
	}{
		{
			name:        "prometheus",
			failRelease: "kube-prometheus-stack",
			wantOutcome: "failed", wantInstalled: 0,
			notRun: []string{"gpu-operator", "network-operator"},
		},
		{
			name:        "gpu-operator",
			failRelease: "gpu-operator",
			wantOutcome: "partial", wantInstalled: 1,
			ranBefore: []string{"kube-prometheus-stack"},
			notRun:    []string{"network-operator"},
		},
		{
			name:        "network-operator",
			failRelease: "network-operator",
			wantOutcome: "partial", wantInstalled: 2,
			ranBefore: []string{"kube-prometheus-stack", "gpu-operator"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			helm := harness.NewFakeHelmWithReleases()
			helm.InstallErrors = map[string]error{tc.failRelease: errors.New("helm boom")}
			k8s := &harness.FakeK8sExecutor{}
			apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
			deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

			eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
			cfg, opts := connectedConfig(t, eventsPath, false)
			term, _, _ := newTestTerminal()

			err := specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts())
			require.Error(t, err, "a failed prerequisite must abort the install")

			evs := readEvents(t, eventsPath)
			summary := prereqSummaryEvent(t, evs)
			assert.Equal(t, tc.wantOutcome, summary["outcome"])
			assert.EqualValues(t, tc.wantInstalled, summary["installed"])
			assert.EqualValues(t, 1, summary["failed"])

			assert.Len(t, helm.CallsByRelease(tc.failRelease), 1, "the failing prerequisite was attempted")
			for _, r := range tc.ranBefore {
				assert.Len(t, helm.CallsByRelease(r), 1, "%s installs before the failure", r)
			}
			for _, r := range tc.notRun {
				assert.Empty(t, helm.CallsByRelease(r), "%s must not run after the abort", r)
			}

			// The Run:ai install never starts once a prerequisite fails.
			assert.True(t, hasPipelineEvent(evs, "pipeline_failed", "prereq"))
			assert.False(t, hasPipelineEvent(evs, "pipeline_started", "runai"), "Run:ai install must not start after a prereq failure")
			assert.Empty(t, helm.CallsByRelease("runai-backend"), "control plane must not install")
			assert.Empty(t, helm.CallsByRelease("runai-cluster"))
			assert.False(t, apiClient.ClusterCreated)
		})
	}
}
