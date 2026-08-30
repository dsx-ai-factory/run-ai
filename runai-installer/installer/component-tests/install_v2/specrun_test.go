package install_v2_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"os"
	"testing"

	"github.com/pterm/pterm"
	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/plan"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/run-ai/runai/runai-installer/installer/internal/terminal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	pterm.DisableColor()
	pterm.DisableStyling()
}

func newInstallRunConfig() *installconfig.Config {
	return &installconfig.Config{
		Install:      installconfig.Install{Mode: "full_installation", Connectivity: "connected"},
		ControlPlane: &installconfig.ControlPlane{Enabled: true, Domain: "test.runai.local"},
		Cluster:      &installconfig.Cluster{Enabled: true, Name: "test-cluster"},
		ExternalConnection: &installconfig.ExternalConnection{
			Type:    "ingress",
			Ingress: &installconfig.IngressConnection{ClassName: "nginx"},
		},
	}
}

func baseRunOptions() installconfig.RunOptions {
	return installconfig.RunOptions{
		Yes:      true,
		Registry: &spec.Registry{Type: spec.RegistryTypeJFrog, Token: base64.StdEncoding.EncodeToString([]byte(`{"auths":{}}`))},
	}
}

func execOpts() specrun.ExecOptions {
	return specrun.ExecOptions{PlanBuilder: plan.Resolve}
}

func newTestTerminal() (*terminal.Terminal, *bytes.Buffer, *bytes.Buffer) {
	var out, errOut bytes.Buffer
	term := terminal.New(
		terminal.WithWriter(&out),
		terminal.WithErrWriter(&errOut),
		terminal.WithYes(true),
	)
	return term, &out, &errOut
}

func TestSpecRun_TextOutput(t *testing.T) {
	term, out, _ := newTestTerminal()

	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	s := newInstallRunConfig().ToSpec(opts)

	err := specrun.Execute(context.Background(), s, newTestDeps(&harness.FakeHelmInstaller{}), term, execOpts())
	require.NoError(t, err)

	// The text UI renders human-readable titles, not the machine names. The
	// machine names remain the identity in JSON/NDJSON (asserted elsewhere).
	stdout := out.String()
	assert.Contains(t, stdout, "# Preflight checks", "pipeline heading uses the human title")
	assert.Contains(t, stdout, "Node CPU capacity", "preflight step uses a human title, not preflight-node-cpu")
	assert.Contains(t, stdout, "Preflight checks: 9 passed", "summary uses the human label")
	assert.Contains(t, stdout, "# Prerequisites")
	assert.Contains(t, stdout, "# NVIDIA Run:ai")
	assert.Contains(t, stdout, "Control plane", "flow uses a human title, not control-plane")
	assert.Contains(t, stdout, "Installing control plane", "step uses a human title, not install-control-plane-chart")
	assert.Contains(t, stdout, "Installing Run:ai cluster")
	assert.Contains(t, stdout, "NVIDIA Run:ai complete", "pipeline completion uses the human title")

	// Machine names (kebab identifiers) must NOT leak into the human text UI.
	assert.NotContains(t, stdout, "preflight-node-cpu")
	assert.NotContains(t, stdout, "install-control-plane-chart")
	assert.NotContains(t, stdout, "pipeline complete:")

	// The runai summary and the final run banner render at the end.
	assert.Contains(t, stdout, "Run:ai control plane is up", "runai summary renders")
	assert.Contains(t, stdout, "https://test.runai.local")
	assert.Contains(t, stdout, `Run:ai cluster "test-cluster" is registered and running`)
	assert.Contains(t, stdout, "NVIDIA Run:ai installation completed", "final run banner renders")
	assert.Contains(t, stdout, "Completed in")
}

func TestSpecRun_InstallsControlPlaneAndCluster(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	fakeK8s := &harness.FakeK8sExecutor{}
	term, _, _ := newTestTerminal()

	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	s := newInstallRunConfig().ToSpec(opts)

	err := specrun.Execute(context.Background(), s, newTestDepsWith(fakeHelm, fakeK8s), term, execOpts())
	require.NoError(t, err)

	require.Len(t, fakeHelm.Calls, 2)

	cp := fakeHelm.CallsByRelease("runai-backend")
	require.Len(t, cp, 1)
	assert.Equal(t, "control-plane", cp[0].ChartName)
	assert.Equal(t, "runai-backend", cp[0].Namespace)
	assert.True(t, harness.ContainsValue(cp[0].Values, "global.domain=test.runai.local"))

	cl := fakeHelm.CallsByRelease("runai-cluster")
	require.Len(t, cl, 1)
	assert.Equal(t, "runai-cluster", cl[0].ChartName)
	assert.Equal(t, "runai", cl[0].Namespace)
	assert.True(t, harness.ContainsValue(cl[0].Values, "cluster.uid=uuid-airgap-test"))

	var namespaces []string
	for _, c := range fakeK8s.CallsByVerb("apply") {
		if c.Kind == "Namespace" {
			namespaces = append(namespaces, c.Name)
		}
	}
	assert.ElementsMatch(t, []string{"runai-backend", "runai"}, namespaces)
}

func TestSpecRun_EventsFileNDJSON(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	term, _, _ := newTestTerminal()

	eventsPath := t.TempDir() + "/events.ndjson"
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	s := newInstallRunConfig().ToSpec(opts)

	err := specrun.Execute(context.Background(), s, newTestDeps(fakeHelm), term, execOpts())
	require.NoError(t, err)

	evs := readEvents(t, eventsPath)
	require.NotEmpty(t, evs)
	assert.Equal(t, "pipeline_started", evs[0]["kind"])
	assert.Equal(t, "run_completed", evs[len(evs)-1]["kind"], "run_completed is the terminal event")

	require.NotNil(t, findEvent(evs, "step_succeeded", "install-control-plane-chart"))
	require.NotNil(t, findEvent(evs, "step_succeeded", "install-cluster-chart"))

	// The runai pipeline emits a runai_completed summary of the control plane
	// and cluster.
	runaiEv := findEventByKind(evs, "runai_completed")
	require.NotNil(t, runaiEv, "a runai_completed event must be emitted")
	runai := runaiEv["runai"].(map[string]any)
	assert.Equal(t, "installed", runai["outcome"])
	cp := runai["controlPlane"].(map[string]any)
	assert.Equal(t, true, cp["installed"])
	assert.Equal(t, "https://test.runai.local", cp["url"])
	cl := runai["cluster"].(map[string]any)
	assert.Equal(t, true, cl["installed"])
	assert.Equal(t, true, cl["registered"])
	assert.Equal(t, "test-cluster", cl["name"])

	// The run emits a single terminal run_completed summary.
	runEv := findEventByKind(evs, "run_completed")
	require.NotNil(t, runEv, "a run_completed event must be emitted")
	run := runEv["run"].(map[string]any)
	assert.Equal(t, "succeeded", run["outcome"])
	assert.Equal(t, "install", run["command"])
	assert.Equal(t, false, run["dryRun"])
	pipelines := run["pipelines"].(map[string]any)
	assert.Contains(t, pipelines, "preflight")
	assert.Contains(t, pipelines, "runai")
	assert.Nil(t, run["failure"], "no failure on a successful run")

	// The human titles added for the text UI must NOT leak into the JSON/NDJSON
	// contract: it keeps the raw machine names and carries no *Title keys.
	raw, err := os.ReadFile(eventsPath)
	require.NoError(t, err)
	content := string(raw)
	assert.Contains(t, content, `"step":"preflight-node-cpu"`, "NDJSON keeps the machine step name")
	assert.Contains(t, content, `"pipeline":"preflight"`, "NDJSON keeps the machine pipeline name")
	assert.NotContains(t, content, "pipelineTitle")
	assert.NotContains(t, content, "flowTitle")
	assert.NotContains(t, content, "stepTitle")
}

func TestSpecRun_DryRun_NoSideEffects(t *testing.T) {
	fakeHelm := &harness.FakeHelmInstaller{}
	fakeK8s := &harness.FakeK8sExecutor{}
	term, _, _ := newTestTerminal()

	eventsPath := t.TempDir() + "/events.ndjson"
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	opts.DryRun = true
	s := newInstallRunConfig().ToSpec(opts)

	err := specrun.Execute(context.Background(), s, newTestDepsWith(fakeHelm, fakeK8s), term, execOpts())
	require.NoError(t, err)

	assert.Empty(t, fakeHelm.Calls, "dry-run must not install helm releases")
	assert.Empty(t, fakeK8s.CallsByVerb("apply"), "dry-run must not write to k8s")
	assert.Empty(t, fakeK8s.CallsByVerb("create"), "dry-run must not write to k8s")

	evs := readEvents(t, eventsPath)
	skipped := findEvent(evs, "step_skipped", "install-control-plane-chart")
	require.NotNil(t, skipped)
	assert.Equal(t, "dry_run", skipped["skipReason"])
}

func TestSpecRun_HelmFailure_Aborts(t *testing.T) {
	boom := errors.New("helm boom")
	fakeHelm := &harness.FakeHelmInstaller{InstallErrors: map[string]error{"runai-backend": boom}}
	fakeK8s := &harness.FakeK8sExecutor{}
	term, _, _ := newTestTerminal()

	eventsPath := t.TempDir() + "/events.ndjson"
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	// OnFailure defaults to abort.
	s := newInstallRunConfig().ToSpec(opts)

	err := specrun.Execute(context.Background(), s, newTestDepsWith(fakeHelm, fakeK8s), term, execOpts())
	require.Error(t, err)

	require.Len(t, fakeHelm.CallsByRelease("runai-backend"), 1, "control-plane install must be attempted")
	assert.Empty(t, fakeHelm.CallsByRelease("runai-cluster"), "cluster must not be attempted after abort")
	_, ok := fakeHelm.GetRelease("runai-cluster")
	assert.False(t, ok)

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(evs, "step_failed", "install-control-plane-chart"))

	// A failed install still emits the terminal run_completed, marked failed and
	// locating the failing step.
	runEv := findEventByKind(evs, "run_completed")
	require.NotNil(t, runEv, "run_completed must be emitted even on failure")
	run := runEv["run"].(map[string]any)
	assert.Equal(t, "failed", run["outcome"])
	failure := run["failure"].(map[string]any)
	assert.Equal(t, "runai", failure["pipeline"])
	assert.Equal(t, "install-control-plane-chart", failure["step"])
}

func TestSpecRun_HelmFailure_SkipContinues(t *testing.T) {
	boom := errors.New("helm boom")
	fakeHelm := &harness.FakeHelmInstaller{InstallErrors: map[string]error{"runai-backend": boom}}
	fakeK8s := &harness.FakeK8sExecutor{}
	term, _, _ := newTestTerminal()

	eventsPath := t.TempDir() + "/events.ndjson"
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	opts.OnFailure = spec.FailurePolicySkip
	s := newInstallRunConfig().ToSpec(opts)

	err := specrun.Execute(context.Background(), s, newTestDepsWith(fakeHelm, fakeK8s), term, execOpts())
	require.NoError(t, err)

	require.Len(t, fakeHelm.CallsByRelease("runai-backend"), 1)
	_, ok := fakeHelm.GetRelease("runai-backend")
	assert.False(t, ok, "failed control-plane release must not be recorded")

	require.Len(t, fakeHelm.CallsByRelease("runai-cluster"), 1, "cluster install must still run after skip")
	_, ok = fakeHelm.GetRelease("runai-cluster")
	assert.True(t, ok)

	evs := readEvents(t, eventsPath)
	skipped := findEvent(evs, "step_skipped", "install-control-plane-chart")
	require.NotNil(t, skipped)
	assert.Equal(t, "user_selected", skipped["skipReason"])
}

func TestSpecRun_K8sFailure_Aborts(t *testing.T) {
	boom := errors.New("apply boom")
	fakeHelm := &harness.FakeHelmInstaller{}
	fakeK8s := &harness.FakeK8sExecutor{CreateErrors: map[string]error{"runai-backend": boom}}
	term, _, _ := newTestTerminal()

	eventsPath := t.TempDir() + "/events.ndjson"
	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.EventsFile = eventsPath
	s := newInstallRunConfig().ToSpec(opts)

	err := specrun.Execute(context.Background(), s, newTestDepsWith(fakeHelm, fakeK8s), term, execOpts())
	require.Error(t, err)

	assert.Empty(t, fakeHelm.CallsByRelease("runai-backend"), "helm must not run after the namespace step fails")

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(evs, "step_failed", "create-control-plane-namespace"))

	runEv := findEventByKind(evs, "run_completed")
	require.NotNil(t, runEv, "run_completed must be emitted even on failure")
	run := runEv["run"].(map[string]any)
	assert.Equal(t, "failed", run["outcome"])
	failure := run["failure"].(map[string]any)
	assert.Equal(t, "create-control-plane-namespace", failure["step"])
}

func TestSpecRun_InvalidSpecFailsValidation(t *testing.T) {
	term, _, _ := newTestTerminal()

	opts := baseRunOptions()
	opts.Output = spec.OutputFormatText
	opts.Registry = nil // registry is required for install
	s := newInstallRunConfig().ToSpec(opts)

	err := specrun.Execute(context.Background(), s, newTestDeps(&harness.FakeHelmInstaller{}), term, execOpts())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "registry")
}
