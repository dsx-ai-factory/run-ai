package install_v2_test

import (
	"context"
	"encoding/base64"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
)

func depsWithProbe(probe *harness.FakeClusterProbe) (orchestrator2.Deps, *harness.FakeHelmInstaller, *harness.FakeK8sExecutor, *harness.FakeRunAIAPIClient) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: probe}
	return deps, helm, k8s, apiClient
}

func findPreflightCompleted(evs []map[string]any) map[string]any {
	for _, e := range evs {
		if e["kind"] == "preflight_completed" {
			return e
		}
	}
	return nil
}

func preflightOutcome(t *testing.T, evs []map[string]any) string {
	t.Helper()
	pc := findPreflightCompleted(evs)
	require.NotNil(t, pc, "a preflight_completed event must be emitted")
	summary, ok := pc["preflight"].(map[string]any)
	require.True(t, ok, "preflight_completed must carry a summary")
	return summary["outcome"].(string)
}

func TestPreflight_Pass_InstallProceeds(t *testing.T) {
	deps, helm, _, _ := depsWithProbe(&harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "passed", preflightOutcome(t, evs))

	reachable := findEvent(evs, "step_succeeded", "preflight-cluster-reachable")
	require.NotNil(t, reachable)
	reachableResults := reachable["results"].([]any)
	require.NotEmpty(t, reachableResults, "cluster-reachable check must emit its detected k8s version")
	data := reachableResults[0].(map[string]any)["data"].(map[string]any)
	assert.Equal(t, "v1.30.0", data["kubernetesVersion"], "the probed Kubernetes version is reported in the event")

	require.NotNil(t, findEvent(evs, "step_succeeded", "preflight-rbac"))
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"), "install must run after a passing preflight")
	assert.NotEmpty(t, helm.Calls, "install must perform helm releases")
}

func TestPreflight_UnsupportedKubernetesVersionBlocksInstall(t *testing.T) {
	deps, helm, _, _ := depsWithProbe(&harness.FakeClusterProbe{Version: "v1.29.9", AllowAll: true})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	err := specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts())
	require.ErrorIs(t, err, specrun.ErrPreflightFailed)

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "failed", preflightOutcome(t, evs))
	reachable := findEvent(evs, "step_succeeded", "preflight-cluster-reachable")
	require.NotNil(t, reachable)
	result := reachable["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "fail", result["status"])
	assert.Contains(t, result["reason"], "minimum required version is 1.30.0")
	assert.False(t, hasPipelineEvent(evs, "pipeline_started", "runai"))
	assert.Empty(t, helm.Calls)
}

// TestPreflight_StorageClass_WarnsWhenNoDefault verifies the default-StorageClass
// check warns (never fails) when the cluster has no default class, and that the
// install still proceeds.
func TestPreflight_StorageClass_WarnsWhenNoDefault(t *testing.T) {
	deps, helm, _, _ := depsWithProbe(&harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true, NoDefaultStorageClass: true})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs), "a missing default StorageClass warns, it does not fail")

	step := findEvent(evs, "step_succeeded", "preflight-storage-class")
	require.NotNil(t, step, "the storage-class check must run")
	result := step["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "no default StorageClass")

	assert.NotEmpty(t, helm.Calls, "a warning must not block the install")
}

// TestPreflight_StorageClass_PassesWithDefault verifies the check is silent when
// a default StorageClass exists.
func TestPreflight_StorageClass_PassesWithDefault(t *testing.T) {
	deps, _, _, _ := depsWithProbe(&harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	step := findEvent(evs, "step_succeeded", "preflight-storage-class")
	require.NotNil(t, step, "the storage-class check must run")
	assert.Empty(t, step["results"], "no finding when a default StorageClass exists")
}

// TestPreflight_OpenShiftMonitoring_FailsWhenMissing verifies that on OpenShift
// the monitoring check fails (blocking the install) when the openshift-monitoring
// namespace is absent.
func TestPreflight_OpenShiftMonitoring_FailsWhenMissing(t *testing.T) {
	probe := &harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true, AbsentNamespaces: map[string]bool{"openshift-monitoring": true}}
	deps, helm, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil // OpenShift uses Routes
	cfg.Prerequisites = &installconfig.Prerequisites{Skip: true}
	opts.Output = spec.OutputFormatText
	term, _, _ := newTestTerminal()

	err := specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts())
	require.ErrorIs(t, err, specrun.ErrPreflightFailed)

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "failed", preflightOutcome(t, evs))

	step := findEvent(evs, "step_succeeded", "preflight-openshift-monitoring")
	require.NotNil(t, step, "the openshift-monitoring check must run on OpenShift")
	result := step["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "fail", result["status"])
	assert.Contains(t, result["reason"], "openshift-monitoring")

	assert.False(t, hasPipelineEvent(evs, "pipeline_started", "runai"), "install must not start after a failed preflight")
	assert.Empty(t, helm.Calls)
}

// TestPreflight_OpenShiftMonitoring_PassesWhenPresent verifies the check runs and
// is silent when the openshift-monitoring namespace exists, letting the install
// proceed.
func TestPreflight_OpenShiftMonitoring_PassesWhenPresent(t *testing.T) {
	// AbsentNamespaces is empty, so every namespace (including
	// openshift-monitoring) is reported present.
	deps, helm, _, _ := depsWithProbe(&harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil // OpenShift uses Routes
	cfg.Prerequisites = &installconfig.Prerequisites{Skip: true}
	opts.Output = spec.OutputFormatText
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	step := findEvent(evs, "step_succeeded", "preflight-openshift-monitoring")
	require.NotNil(t, step, "the openshift-monitoring check must run on OpenShift")
	assert.Empty(t, step["results"], "no finding when the monitoring stack is present")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.NotEmpty(t, helm.Calls)
}

// TestPreflight_OpenShiftMonitoring_SkippedOnNonOpenShift verifies the check does
// not run on a non-OpenShift install (even if monitoring is absent).
func TestPreflight_OpenShiftMonitoring_SkippedOnNonOpenShift(t *testing.T) {
	probe := &harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true, AbsentNamespaces: map[string]bool{"openshift-monitoring": true}}
	deps, _, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Nil(t, findEvent(evs, "step_started", "preflight-openshift-monitoring"), "the check must not run off OpenShift")
	assert.Equal(t, "passed", preflightOutcome(t, evs))
}

func TestPreflight_ClusterUnreachable_Exit2(t *testing.T) {
	deps, helm, k8s, apiClient := depsWithProbe(&harness.FakeClusterProbe{VersionErr: errors.New("connection refused"), AllowAll: true})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	err := specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts())
	require.ErrorIs(t, err, specrun.ErrPreflightFailed)

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "failed", preflightOutcome(t, evs))
	assert.False(t, hasPipelineEvent(evs, "pipeline_started", "runai"), "install must not start after a failed preflight")
	assert.Empty(t, helm.Calls)
	assert.Empty(t, k8s.CallsByVerb("apply"))
	assert.False(t, apiClient.ClusterCreated)
}

func TestPreflight_RBACDenied_Warns(t *testing.T) {
	// Reachable, but no permissions granted (AllowAll false, empty Allowed).
	// Missing permissions are a warning, not a hard failure: the install still
	// proceeds (the SelfSubjectAccessReview is an approximation).
	// nginx ingress class present so the external-connection check passes and
	// RBAC is the only variable under test.
	deps, helm, _, _ := depsWithProbe(&harness.FakeClusterProbe{Version: "v1.30.0", IngressClasses: map[string]bool{"nginx": true}})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs))
	assert.NotEmpty(t, helm.Calls, "a warning must not block the install")

	rbac := findEvent(evs, "step_succeeded", "preflight-rbac")
	require.NotNil(t, rbac)
	result := rbac["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "missing required permissions")
}

func TestPreflight_RBACPartialDenial_ReportsDeniedSubset(t *testing.T) {
	// Some permissions granted, some denied: only the denied ones are reported.
	probe := &harness.FakeClusterProbe{
		Version:        "v1.30.0",
		IngressClasses: map[string]bool{"nginx": true},
		Allowed: map[string]bool{
			"create /namespaces": true,
			"create /secrets":    true,
		},
	}
	deps, helm, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs))
	assert.NotEmpty(t, helm.Calls, "a warning must not block the install")

	rbac := findEvent(evs, "step_succeeded", "preflight-rbac")
	require.NotNil(t, rbac)
	result := rbac["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	reason := result["reason"].(string)
	assert.Contains(t, reason, "missing required permissions")
	assert.Contains(t, reason, "create customresourcedefinitions.apiextensions.k8s.io")
	assert.Contains(t, reason, "create clusterroles.rbac.authorization.k8s.io")
	assert.NotContains(t, reason, "namespaces", "granted permissions must not be reported as denied")
	assert.NotContains(t, reason, "secrets", "granted permissions must not be reported as denied")
}

func TestPreflight_RBACEvalError_Warns(t *testing.T) {
	// CanI itself errors (API problem), distinct from a clean denial. Reported
	// as a warning, so the install still proceeds.
	deps, helm, _, _ := depsWithProbe(&harness.FakeClusterProbe{Version: "v1.30.0", IngressClasses: map[string]bool{"nginx": true}, CanIErr: errors.New("apiserver unavailable")})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs))
	assert.NotEmpty(t, helm.Calls, "a warning must not block the install")

	rbac := findEvent(evs, "step_succeeded", "preflight-rbac")
	require.NotNil(t, rbac)
	result := rbac["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "could not evaluate permissions")
}

func TestPreflight_SkipPreflight_BypassesChecks(t *testing.T) {
	// A probe that would fail preflight (unreachable); --skip-preflight must
	// bypass the checks entirely and let the install run anyway.
	deps, helm, _, _ := depsWithProbe(&harness.FakeClusterProbe{VersionErr: errors.New("connection refused")})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	exec := execOpts()
	exec.SkipPreflight = true
	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, exec))

	evs := readEvents(t, eventsPath)
	assert.Nil(t, findPreflightCompleted(evs), "no preflight summary when preflight is skipped")
	assert.False(t, hasPipelineEvent(evs, "pipeline_started", "preflight"), "preflight pipeline must not run with --skip-preflight")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"), "install runs despite a probe that would fail preflight")
	assert.NotEmpty(t, helm.Calls)
}

func TestPreflight_PreflightOnly_SkipsInstall(t *testing.T) {
	deps, helm, k8s, apiClient := depsWithProbe(&harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true})

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	exec := execOpts()
	exec.PreflightOnly = true
	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, exec))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "passed", preflightOutcome(t, evs))
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "preflight"), "preflight pipeline must run")
	assert.False(t, hasPipelineEvent(evs, "pipeline_started", "runai"), "install must not start with --preflight-only")
	assert.Empty(t, helm.Calls, "no install work with --preflight-only")
	assert.Empty(t, k8s.CallsByVerb("apply"))
	assert.False(t, apiClient.ClusterCreated)
}

var nodeResourceStepNames = []string{"preflight-node-cpu", "preflight-node-memory", "preflight-node-disk"}

func workerNode(name, cpu, memory, disk string) orchestrator2.NodeInfo {
	return orchestrator2.NodeInfo{
		Name:   name,
		Roles:  []string{"worker"},
		CPU:    resource.MustParse(cpu),
		Memory: resource.MustParse(memory),
		Disk:   resource.MustParse(disk),
	}
}

// TestPreflight_NodeResources_RunAndPass verifies the CPU/memory/disk node-sizing
// checks actually run as preflight steps and pass when nodes meet the minimums.
func TestPreflight_NodeResources_RunAndPass(t *testing.T) {
	probe := &harness.FakeClusterProbe{
		Version:  "v1.30.0",
		AllowAll: true,
		Nodes:    []orchestrator2.NodeInfo{workerNode("node-1", "16", "64Gi", "200Gi")},
	}
	deps, helm, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "passed", preflightOutcome(t, evs))
	for _, name := range nodeResourceStepNames {
		step := findEvent(evs, "step_succeeded", name)
		require.NotNil(t, step, "%s must run as a preflight step", name)
		assert.Empty(t, step["results"], "%s must report no findings when the node meets the minimum", name)
	}
	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	assert.Empty(t, summary["resourceIssues"], "no resource issues when nodes meet the minimums")
	assert.NotEmpty(t, helm.Calls, "install proceeds after a passing preflight")
}

// TestPreflight_NodeResources_WarnBelowMinimum verifies an undersized node makes
// each node-sizing check emit a non-blocking warning (install still proceeds).
func TestPreflight_NodeResources_WarnBelowMinimum(t *testing.T) {
	probe := &harness.FakeClusterProbe{
		Version:  "v1.30.0",
		AllowAll: true,
		Nodes:    []orchestrator2.NodeInfo{workerNode("tiny-node", "2", "4Gi", "20Gi")},
	}
	deps, helm, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs), "undersized nodes warn, they do not fail")

	// The per-check line stays a single concise warning (no inline node list).
	for _, name := range nodeResourceStepNames {
		step := findEvent(evs, "step_succeeded", name)
		require.NotNil(t, step, "%s must run", name)
		results := step["results"].([]any)
		require.NotEmpty(t, results, "%s must report a finding for the undersized node", name)
		result := results[0].(map[string]any)
		assert.Equal(t, "warn", result["status"], "%s must be a warning, never a failure", name)
		assert.Contains(t, result["reason"], "below recommended")
		assert.NotContains(t, result["reason"], "tiny-node", "the per-check line must not inline the node list")
	}

	// The full per-node breakdown is aggregated into the preflight summary.
	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	issues := summary["resourceIssues"].([]any)
	require.Len(t, issues, 3, "one issue per resource for the single undersized node")
	resources := map[string]bool{}
	for _, raw := range issues {
		issue := raw.(map[string]any)
		assert.Equal(t, "tiny-node", issue["node"])
		resources[issue["resource"].(string)] = true
	}
	assert.Equal(t, map[string]bool{"CPU": true, "memory": true, "disk (ephemeral storage)": true}, resources)

	assert.NotEmpty(t, helm.Calls, "a warning must not block the install")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}

func nodeWithRoles(name string, roles []string, cpu, memory, disk string) orchestrator2.NodeInfo {
	return orchestrator2.NodeInfo{
		Name:   name,
		Roles:  roles,
		CPU:    resource.MustParse(cpu),
		Memory: resource.MustParse(memory),
		Disk:   resource.MustParse(disk),
	}
}

// TestPreflight_NodeResources_ListError verifies a node-list failure is surfaced
// as a (non-blocking) warning rather than crashing or silently passing.
func TestPreflight_NodeResources_ListError(t *testing.T) {
	probe := &harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true, ListNodesErr: errors.New("forbidden: cannot list nodes")}
	deps, helm, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs))
	for _, name := range nodeResourceStepNames {
		step := findEvent(evs, "step_succeeded", name)
		require.NotNil(t, step, "%s must run", name)
		result := step["results"].([]any)[0].(map[string]any)
		assert.Equal(t, "warn", result["status"])
		assert.Contains(t, result["reason"], "could not list nodes")
	}
	assert.NotEmpty(t, helm.Calls, "a warning must not block the install")
}

// TestPreflight_NodeResources_RoleSpecificMinimums verifies control-plane-only
// nodes are held to the (lower) control-plane minimums while combined
// control-plane+worker nodes are held to the worker minimums.
func TestPreflight_NodeResources_RoleSpecificMinimums(t *testing.T) {
	probe := &harness.FakeClusterProbe{
		Version:  "v1.30.0",
		AllowAll: true,
		Nodes: []orchestrator2.NodeInfo{
			// Meets the control-plane minimums (4/16Gi/50Gi) but is below the
			// worker minimums; must pass because it is control-plane only.
			nodeWithRoles("cp-only", []string{"control-plane"}, "4", "16Gi", "50Gi"),
			// Same sizing, but it also runs workloads -> held to worker minimums.
			nodeWithRoles("cp-worker", []string{"control-plane", "worker"}, "4", "16Gi", "50Gi"),
		},
	}
	deps, _, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	issues := summary["resourceIssues"].([]any)
	nodes := map[string]bool{}
	for _, raw := range issues {
		i := raw.(map[string]any)
		nodes[i["node"].(string)] = true
		assert.Equal(t, "worker", i["role"], "the combined node is held to the worker minimums")
	}
	assert.Equal(t, map[string]bool{"cp-worker": true}, nodes,
		"control-plane-only node passes on control-plane minimums; only the combined node is flagged")
}

// TestPreflight_NodeResources_SkipsNodeMissingCapacity verifies a node that does
// not report a resource (zero capacity) is skipped, not flagged as "0 below".
func TestPreflight_NodeResources_SkipsNodeMissingCapacity(t *testing.T) {
	probe := &harness.FakeClusterProbe{
		Version:  "v1.30.0",
		AllowAll: true,
		// Reports CPU/memory but no ephemeral-storage (disk == 0).
		Nodes: []orchestrator2.NodeInfo{nodeWithRoles("no-disk", []string{"worker"}, "16", "64Gi", "0")},
	}
	deps, _, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "passed", preflightOutcome(t, evs), "a node not reporting a resource must not be flagged as 0")
	diskStep := findEvent(evs, "step_succeeded", "preflight-node-disk")
	require.NotNil(t, diskStep)
	assert.Empty(t, diskStep["results"], "node with no reported disk capacity is skipped, not flagged")
	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	assert.Empty(t, summary["resourceIssues"])
}

// TestPreflight_Clock_HeartbeatError verifies that a failure to read node leases
// is surfaced as a warning by both clock checks.
func TestPreflight_Clock_HeartbeatError(t *testing.T) {
	probe := &harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true, NodeHeartbeatsErr: errors.New("forbidden: cannot list leases")}
	deps, _, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs))
	for _, name := range []string{"preflight-clock-skew", "preflight-clock-drift"} {
		step := findEvent(evs, "step_succeeded", name)
		require.NotNil(t, step, "%s must run", name)
		result := step["results"].([]any)[0].(map[string]any)
		assert.Equal(t, "warn", result["status"])
		assert.Contains(t, result["reason"], "could not read node heartbeats")
	}
}

// TestPreflight_ClockSkew_SingleNodeSkipped verifies the skew check is a no-op
// when there is only one node (nothing to compare against).
func TestPreflight_ClockSkew_SingleNodeSkipped(t *testing.T) {
	probe := &harness.FakeClusterProbe{
		Version:    "v1.30.0",
		AllowAll:   true,
		Heartbeats: []orchestrator2.NodeHeartbeat{{Name: "only-node", RenewTime: time.Now()}},
	}
	deps, _, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	step := findEvent(evs, "step_succeeded", "preflight-clock-skew")
	require.NotNil(t, step)
	assert.Empty(t, step["results"], "a single node has nothing to compare; no skew finding")
	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	assert.Empty(t, summary["clockSkew"])
}

// TestPreflight_Registry_AllGood verifies the registry check passes silently when
// the registry is reachable and the credentials authenticate.
func TestPreflight_Registry_AllGood(t *testing.T) {
	token := base64.StdEncoding.EncodeToString([]byte(`{"auths":{"runai.jfrog.io":{"username":"u","password":"p"}}}`))
	fake := &harness.FakeRegistryProber{Default: orchestrator2.RegistryCheckResult{Reachable: true, Authenticated: true, CanPull: true}}
	deps, helm, _, _ := depsWithProbe(&harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true})
	deps.RegistryProber = fake

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	opts.Registry = &spec.Registry{Type: spec.RegistryTypeJFrog, Token: token}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "passed", preflightOutcome(t, evs))
	step := findEvent(evs, "step_succeeded", "preflight-registry")
	require.NotNil(t, step, "the registry check must run")
	assert.Empty(t, step["results"], "no findings when the registry is reachable and authenticated")
	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	assert.Empty(t, summary["registryIssues"])
	assert.NotEmpty(t, helm.Calls)
}

// TestPreflight_ClockSkew_WarnsAndAggregates verifies the clock-skew check warns
// (never fails) on a node whose heartbeat lags far behind, keeps the per-check
// line to one line, and aggregates the per-node detail into the summary.
func TestPreflight_ClockSkew_WarnsAndAggregates(t *testing.T) {
	base := time.Now()
	probe := &harness.FakeClusterProbe{
		Version:  "v1.30.0",
		AllowAll: true,
		Heartbeats: []orchestrator2.NodeHeartbeat{
			{Name: "node-a", RenewTime: base},
			{Name: "node-b", RenewTime: base.Add(-2 * time.Second)},
			{Name: "skewed-node", RenewTime: base.Add(-5 * time.Minute)},
		},
	}
	deps, helm, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs), "clock skew warns, it does not fail")

	step := findEvent(evs, "step_succeeded", "preflight-clock-skew")
	require.NotNil(t, step)
	result := step["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "clock skew")
	assert.NotContains(t, result["reason"], "skewed-node", "the per-check line must not inline the node list")

	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	skew := summary["clockSkew"].([]any)
	require.Len(t, skew, 1, "only the lagging node is reported")
	assert.Equal(t, "skewed-node", skew[0].(map[string]any)["node"])

	assert.NotEmpty(t, helm.Calls, "a warning must not block the install")
}

// TestPreflight_ClockSkew_PassWhenInSync verifies that nodes whose heartbeats
// are close together produce no clock-skew finding.
func TestPreflight_ClockSkew_PassWhenInSync(t *testing.T) {
	base := time.Now()
	probe := &harness.FakeClusterProbe{
		Version:  "v1.30.0",
		AllowAll: true,
		Heartbeats: []orchestrator2.NodeHeartbeat{
			{Name: "node-a", RenewTime: base},
			{Name: "node-b", RenewTime: base.Add(-3 * time.Second)},
			{Name: "node-c", RenewTime: base.Add(-5 * time.Second)},
		},
	}
	deps, _, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	step := findEvent(evs, "step_succeeded", "preflight-clock-skew")
	require.NotNil(t, step, "the clock-skew check must run")
	assert.Empty(t, step["results"], "in-sync nodes produce no finding")

	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	assert.Empty(t, summary["clockSkew"], "no clock-skew entries when nodes are in sync")
}

// TestPreflight_ClockDrift_WarnsAgainstLocalClock verifies the clock-drift check
// warns (never fails) when a node's clock differs from the installer machine's
// clock, makes the reference explicit, and aggregates per-node detail.
func TestPreflight_ClockDrift_WarnsAgainstLocalClock(t *testing.T) {
	base := time.Now()
	probe := &harness.FakeClusterProbe{
		Version:  "v1.30.0",
		AllowAll: true,
		Heartbeats: []orchestrator2.NodeHeartbeat{
			{Name: "in-sync-node", RenewTime: base.Add(-2 * time.Second)},
			{Name: "behind-node", RenewTime: base.Add(-5 * time.Minute)},
			{Name: "ahead-node", RenewTime: base.Add(5 * time.Minute)},
		},
	}
	deps, helm, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs), "clock drift warns, it does not fail")

	step := findEvent(evs, "step_succeeded", "preflight-clock-drift")
	require.NotNil(t, step)
	result := step["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "this machine's clock", "the message must make the reference clock explicit")
	assert.NotContains(t, result["reason"], "behind-node", "the per-check line must not inline the node list")

	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	assert.NotEmpty(t, summary["clockDriftReference"], "the machine clock reference must be recorded")
	drift := summary["clockDrift"].([]any)
	require.Len(t, drift, 2, "the in-sync node is not reported")
	offsets := map[string]string{}
	for _, raw := range drift {
		d := raw.(map[string]any)
		offsets[d["node"].(string)] = d["offset"].(string)
	}
	assert.Contains(t, offsets["behind-node"], "behind this machine")
	assert.Contains(t, offsets["ahead-node"], "ahead of this machine")

	assert.NotEmpty(t, helm.Calls, "a warning must not block the install")
}

// TestPreflight_ClockDrift_PassWhenInSync verifies no drift finding when all
// nodes are close to the installer machine's clock.
func TestPreflight_ClockDrift_PassWhenInSync(t *testing.T) {
	base := time.Now()
	probe := &harness.FakeClusterProbe{
		Version:  "v1.30.0",
		AllowAll: true,
		Heartbeats: []orchestrator2.NodeHeartbeat{
			{Name: "node-a", RenewTime: base.Add(-2 * time.Second)},
			{Name: "node-b", RenewTime: base.Add(-4 * time.Second)},
		},
	}
	deps, _, _, _ := depsWithProbe(probe)

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	step := findEvent(evs, "step_succeeded", "preflight-clock-drift")
	require.NotNil(t, step, "the clock-drift check must run")
	assert.Empty(t, step["results"], "in-sync nodes produce no finding")

	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	assert.Empty(t, summary["clockDrift"], "no clock-drift entries when nodes match this machine")
}

// TestPreflight_Registry_ConnectedWarnsOnAuthFailure verifies the registry check
// probes each image registry from the pull-secret token, warns (non-blocking) on
// an auth failure, and aggregates the finding into the summary.
func TestPreflight_Registry_ConnectedWarnsOnAuthFailure(t *testing.T) {
	token := base64.StdEncoding.EncodeToString([]byte(`{"auths":{"runai.jfrog.io":{"username":"u","password":"p"}}}`))
	fake := &harness.FakeRegistryProber{Results: map[string]orchestrator2.RegistryCheckResult{
		"runai.jfrog.io": {Reachable: true, Authenticated: false, Detail: "bad credentials"},
	}}
	deps, helm, _, _ := depsWithProbe(&harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true})
	deps.RegistryProber = fake

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	opts.Registry = &spec.Registry{Type: spec.RegistryTypeJFrog, Token: token}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs), "a registry problem warns, it does not fail")

	step := findEvent(evs, "step_succeeded", "preflight-registry")
	require.NotNil(t, step)
	result := step["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "registry issue")

	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	issues := summary["registryIssues"].([]any)
	require.Len(t, issues, 1)
	issue := issues[0].(map[string]any)
	assert.Contains(t, issue["registry"], "runai.jfrog.io")
	assert.Contains(t, issue["issue"], "Invalid registry credentials")

	require.Len(t, fake.Calls, 1)
	assert.Equal(t, "runai.jfrog.io", fake.Calls[0].Host)
	assert.Equal(t, "u", fake.Calls[0].Username)
	assert.False(t, fake.Calls[0].CheckPush, "connected install only needs pull")

	assert.NotEmpty(t, helm.Calls, "a warning must not block the install")
}

// TestPreflight_Registry_MalformedTokenWarns verifies that an invalid pull-secret
// token surfaces a config warning instead of silently skipping the check.
func TestPreflight_Registry_MalformedTokenWarns(t *testing.T) {
	fake := &harness.FakeRegistryProber{Default: orchestrator2.RegistryCheckResult{Reachable: true, Authenticated: true, CanPull: true}}
	deps, _, _, _ := depsWithProbe(&harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true})
	deps.RegistryProber = fake

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	// Valid base64 (so it passes input validation) but not a docker config: this
	// is the silent path that previously produced no check at all.
	opts.Registry = &spec.Registry{Type: spec.RegistryTypeJFrog, Token: base64.StdEncoding.EncodeToString([]byte("garbage"))}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs))

	step := findEvent(evs, "step_succeeded", "preflight-registry")
	require.NotNil(t, step, "the registry check must run and report the malformed token, not be skipped")

	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	issues := summary["registryIssues"].([]any)
	require.Len(t, issues, 1)
	assert.Contains(t, issues[0].(map[string]any)["issue"], "not a valid docker config")
	assert.Empty(t, fake.Calls, "a malformed token must not reach the network probe")
}

// TestPreflight_Registry_UnreachableWarns verifies the "no route to the registry"
// case (network unreachable) surfaces as a warning.
func TestPreflight_Registry_UnreachableWarns(t *testing.T) {
	token := base64.StdEncoding.EncodeToString([]byte(`{"auths":{"runai.jfrog.io":{"username":"u","password":"p"}}}`))
	fake := &harness.FakeRegistryProber{Results: map[string]orchestrator2.RegistryCheckResult{
		"runai.jfrog.io": {Reachable: false, Detail: "dial tcp: i/o timeout"},
	}}
	deps, helm, _, _ := depsWithProbe(&harness.FakeClusterProbe{Version: "v1.30.0", AllowAll: true})
	deps.RegistryProber = fake

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	opts.Registry = &spec.Registry{Type: spec.RegistryTypeJFrog, Token: token}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), newInstallRunConfig().ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Equal(t, "warning", preflightOutcome(t, evs))
	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	issues := summary["registryIssues"].([]any)
	require.Len(t, issues, 1)
	assert.Contains(t, issues[0].(map[string]any)["issue"], "not reachable")
	assert.Contains(t, issues[0].(map[string]any)["issue"], "i/o timeout")
	assert.NotEmpty(t, helm.Calls, "a warning must not block the install")
}

// TestPreflight_Registry_AirgapUsernamePasswordAuthFailure verifies the username/
// password (air-gap mirror) path: bad credentials surface as an auth-failure
// warning and the credentials are passed through to the probe.
func TestPreflight_Registry_AirgapUsernamePasswordAuthFailure(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	fake := &harness.FakeRegistryProber{Results: map[string]orchestrator2.RegistryCheckResult{
		"registry.local:5000": {Reachable: true, Authenticated: false, Detail: "basic auth returned 401"},
	}}
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), RegistryProber: fake}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := airgapConfig(t, harness.CreateTestTarball(t), eventsPath)
	opts.Registry = &spec.Registry{Host: "registry.local:5000", Username: "admin", Password: "wrong-password"}
	opts.Output = spec.OutputFormatText
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	issues := summary["registryIssues"].([]any)
	require.Len(t, issues, 1)
	assert.Contains(t, issues[0].(map[string]any)["issue"], "Invalid registry credentials")

	require.Len(t, fake.Calls, 1)
	assert.Equal(t, "admin", fake.Calls[0].Username, "username/password must reach the probe")
	assert.Equal(t, "wrong-password", fake.Calls[0].Password)
}

// TestPreflight_Registry_AirgapPullDenied verifies that authenticated-but-no-pull
// permission surfaces as a warning.
func TestPreflight_Registry_AirgapPullDenied(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	fake := &harness.FakeRegistryProber{Results: map[string]orchestrator2.RegistryCheckResult{
		"registry.local:5000": {
			Reachable: true, Authenticated: true, CanPull: false,
			ProjectChecked: true, ProjectExists: true,
			PushChecked: true, CanPush: true,
		},
	}}
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), RegistryProber: fake}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := airgapConfig(t, harness.CreateTestTarball(t), eventsPath)
	opts.Registry = &spec.Registry{Host: "registry.local:5000/runai", Username: "admin", Password: "Harbor12345"}
	opts.Output = spec.OutputFormatText
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	issues := summary["registryIssues"].([]any)
	texts := []string{}
	for _, raw := range issues {
		texts = append(texts, raw.(map[string]any)["issue"].(string))
	}
	assert.Contains(t, texts, "Credentials cannot pull images from registry")
}

// TestPreflight_Registry_AirgapVerifiesPushAndProjectReadOnly verifies the
// air-gap mirror is probed for push permission (read-only, never pushing) and
// project existence, and that both failures surface as warnings.
func TestPreflight_Registry_AirgapVerifiesPushAndProjectReadOnly(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	fake := &harness.FakeRegistryProber{Results: map[string]orchestrator2.RegistryCheckResult{
		"registry.local:5000": {
			Reachable: true, Authenticated: true, CanPull: true,
			ProjectChecked: true, ProjectExists: false,
			PushChecked: true, CanPush: false,
		},
	}}
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), RegistryProber: fake}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := airgapConfig(t, harness.CreateTestTarball(t), eventsPath)
	opts.Registry = &spec.Registry{Host: "registry.local:5000/runai", Username: "admin", Password: "Harbor12345"}
	opts.Output = spec.OutputFormatText
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	step := findEvent(evs, "step_succeeded", "preflight-registry")
	require.NotNil(t, step)

	summary := findPreflightCompleted(evs)["preflight"].(map[string]any)
	issues := summary["registryIssues"].([]any)
	texts := []string{}
	for _, raw := range issues {
		texts = append(texts, raw.(map[string]any)["issue"].(string))
	}
	assert.Contains(t, texts, `Project "runai" does not exist in registry`)
	assert.Contains(t, texts, "Credentials cannot push images to registry")

	// Push was verified read-only: the probe requested push scope on a sentinel
	// repo under the project and never performed a push.
	require.Len(t, fake.Calls, 1)
	assert.True(t, fake.Calls[0].CheckPush)
	assert.Equal(t, "registry.local:5000", fake.Calls[0].Host)
	assert.Equal(t, "runai", fake.Calls[0].Project)
	assert.Equal(t, "runai/runai-preflight-probe", fake.Calls[0].Repo)
}
