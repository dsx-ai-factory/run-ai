// Package pack_test drives the airgap pack command end to end
// (packconfig.Load -> artifactlock.ParseFile -> ToSpec -> specrun.Execute). No
// test reaches into the internals; assertions are made against the run's error,
// the terminal output, the NDJSON events file, and the files on disk.
package pack_test

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/pterm/pterm"
	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/artifacts"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/plan"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/packconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/run-ai/runai/runai-installer/installer/internal/terminal"
	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
	"github.com/stretchr/testify/require"
)

func init() {
	pterm.DisableColor()
	pterm.DisableStyling()
}

// result is everything one pack run produced.
type result struct {
	err     error
	stdout  string
	events  []map[string]any
	output  string
	workDir string
	puller  *harness.FakeImagePuller
	fetcher *harness.FakeChartFetcher
	cli     *harness.FakeCLIFetcher
	prober  *harness.FakeRegistryProber
}

// fakes are the engine seams one run executes against, for a test to break in
// one specific way.
type fakes struct {
	puller  *harness.FakeImagePuller
	fetcher *harness.FakeChartFetcher
	cli     *harness.FakeCLIFetcher
	prober  *harness.FakeRegistryProber
	// realFetcher, when set, replaces the fake so a test can drive the actual
	// download path against a local server.
	realFetcher artifacts.ChartFetcher
	realCLI     artifacts.CLIFetcher
	realProber  orchestrator2.RegistryProber
}

func newFakes() *fakes {
	return &fakes{
		puller:  &harness.FakeImagePuller{},
		fetcher: &harness.FakeChartFetcher{},
		cli:     &harness.FakeCLIFetcher{},
		prober: &harness.FakeRegistryProber{
			Default: orchestrator2.RegistryCheckResult{Reachable: true, Authenticated: true, CanPull: true},
		},
	}
}

// run executes the pack command with the given config YAML against a lock,
// returning the run's error and everything it wrote.
func run(t *testing.T, configYAML string, lock *artifactlock.ArtifactLock, dryRun bool) result {
	t.Helper()
	return runWith(t, configYAML, harness.WriteLockFile(t, lock), dryRun, newFakes())
}

// runFakes is run against fakes a test has configured to fail.
func runFakes(t *testing.T, configYAML string, lock *artifactlock.ArtifactLock, f *fakes) result {
	t.Helper()
	return runWith(t, configYAML, harness.WriteLockFile(t, lock), false, f)
}

// runWithLockFile is run against a lock already on disk, so a real generated
// lock can be packed as-is.
func runWithLockFile(t *testing.T, configYAML, lockPath string, dryRun bool) result {
	t.Helper()
	return runWith(t, configYAML, lockPath, dryRun, newFakes())
}

// runStaged is run with the preflight staging a test wants to exercise.
func runStaged(t *testing.T, configYAML string, lock *artifactlock.ArtifactLock, f *fakes, adjust func(*specrun.ExecOptions)) result {
	t.Helper()
	return runWith(t, configYAML, harness.WriteLockFile(t, lock), false, f, adjust)
}

func runWith(t *testing.T, configYAML, lockPath string, dryRun bool, f *fakes, adjust ...func(*specrun.ExecOptions)) result {
	t.Helper()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "airgap-pack.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configYAML), 0o600))
	eventsPath := filepath.Join(dir, "events.ndjson")

	cfg, err := packconfig.Load(configPath)
	if err != nil {
		return result{err: err}
	}

	parsed, err := artifactlock.ParseFile(lockPath)
	if err != nil {
		return result{err: err}
	}

	s, err := cfg.ToSpec(parsed, packconfig.RunOptions{
		DryRun:     dryRun,
		Yes:        true,
		Output:     spec.OutputFormatText,
		EventsFile: eventsPath,
	})
	if err != nil {
		return result{err: err}
	}

	var out bytes.Buffer
	term := terminal.New(
		terminal.WithWriter(&out),
		terminal.WithErrWriter(&out),
		terminal.WithYes(true),
	)

	deps := orchestrator2.Deps{
		ImagePuller:    f.puller,
		ChartFetcher:   f.fetcher,
		CLIFetcher:     f.cli,
		RegistryProber: f.prober,
	}
	if f.realFetcher != nil {
		deps.ChartFetcher = f.realFetcher
	}
	if f.realCLI != nil {
		deps.CLIFetcher = f.realCLI
	}
	if f.realProber != nil {
		deps.RegistryProber = f.realProber
	}
	execOpts := specrun.ExecOptions{PlanBuilder: plan.Resolve}
	for _, a := range adjust {
		a(&execOpts)
	}
	runErr := specrun.Execute(context.Background(), s, deps, term, execOpts)

	return result{
		err:     runErr,
		stdout:  out.String(),
		events:  readEvents(t, eventsPath),
		output:  s.AirgapPack.OutputPath,
		workDir: s.AirgapPack.WorkDir,
		puller:  f.puller,
		fetcher: f.fetcher,
		cli:     f.cli,
		prober:  f.prober,
	}
}

// packConfig is a valid config for the harness lock, for tests to break in one
// specific way.
func packConfig(t *testing.T) string {
	t.Helper()
	return "settings:\n  runaiVersion: 2.26.13\n  architecture: amd64\n  output: " + filepath.Join(t.TempDir(), "runai-airgap.tar") + "\n"
}

// tarEntries lists the names of every file in an archive.
func tarEntries(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	var names []string
	tr := tar.NewReader(f)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if header.Typeflag != tar.TypeDir {
			names = append(names, header.Name)
		}
	}
	return names
}

// tarFile returns the contents of one file inside an archive.
func tarFile(t *testing.T, path, name string) []byte {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	tr := tar.NewReader(f)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			t.Fatalf("%s not found in %s", name, path)
		}
		require.NoError(t, err)
		if header.Name == name {
			content, err := io.ReadAll(tr)
			require.NoError(t, err)
			return content
		}
	}
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o600)
}

func filepathBase(path string) string { return filepath.Base(path) }

func sha256Hex(content []byte) string {
	return hex.EncodeToString(sha256Sum(content))
}

func sha256Sum(content []byte) []byte {
	sum := sha256.Sum256(content)
	return sum[:]
}

func readEvents(t *testing.T, path string) []map[string]any {
	t.Helper()
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	var out []map[string]any
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var got map[string]any
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &got), "each line must be valid JSON")
		out = append(out, got)
	}
	require.NoError(t, scanner.Err())
	return out
}

func findEvent(events []map[string]any, kind, step string) map[string]any {
	for _, e := range events {
		if e["kind"] == kind && e["step"] == step {
			return e
		}
	}
	return nil
}

func findEventByKind(events []map[string]any, kind string) map[string]any {
	for _, e := range events {
		if e["kind"] == kind {
			return e
		}
	}
	return nil
}

// findings returns the messages of the input-validation event, keyed by field.
func findings(t *testing.T, events []map[string]any) map[string]string {
	t.Helper()
	ev := findEventByKind(events, "input_validation_failed")
	require.NotNil(t, ev, "an input_validation_failed event must be emitted")

	out := map[string]string{}
	for _, raw := range ev["findings"].([]any) {
		f := raw.(map[string]any)
		out[f["field"].(string)], _ = f["message"].(string)
	}
	return out
}
