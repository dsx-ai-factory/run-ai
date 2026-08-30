package harness

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
)

// FakeImagePuller records pulls and digest resolutions, writing a marker file
// into each layout so bundle contents can be asserted on.
type FakeImagePuller struct {
	// PullErrors and ResolveErrors fail one reference, keyed by image reference.
	PullErrors    map[string]error
	ResolveErrors map[string]error
	// Digests overrides the digest a reference resolves to, keyed by reference.
	Digests map[string]string

	mu       sync.Mutex
	Pulls    []string
	Resolves []string
}

func (f *FakeImagePuller) Pull(_ context.Context, ref, dir string, _ artifactlock.Platform) error {
	f.mu.Lock()
	f.Pulls = append(f.Pulls, ref)
	f.mu.Unlock()

	if err := f.PullErrors[ref]; err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "blobs", "sha256"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "index.json"), []byte(`{"schemaVersion":2,"manifests":[]}`), 0o644)
}

func (f *FakeImagePuller) ResolveDigests(_ context.Context, ref string, _ artifactlock.Platform) (string, string, error) {
	f.mu.Lock()
	f.Resolves = append(f.Resolves, ref)
	f.mu.Unlock()

	if err := f.ResolveErrors[ref]; err != nil {
		return "", "", err
	}
	if digest, ok := f.Digests[ref]; ok {
		return digest, digest, nil
	}
	digest := fmt.Sprintf("sha256:%064x", len(ref))
	return digest, digest, nil
}

// FakeChartFetcher writes canned archive bytes, so a digest mismatch can be
// staged without a registry.
type FakeChartFetcher struct {
	// Content is the archive bytes served for a chart, keyed by chart name.
	// A chart with no entry gets bytes matching its locked digest.
	Content map[string][]byte
	// FetchErrors fails one chart, keyed by chart name.
	FetchErrors map[string]error
	// ProbeErrors fails one chart's preflight probe, keyed by chart name.
	ProbeErrors map[string]error

	mu      sync.Mutex
	Fetches []string
	Probes  []string
}

type FakeCLIFetcher struct {
	FetchErrors map[string]error
	ProbeErrors map[string]error
	Fetches     []string
	Probes      []string
}

func (f *FakeCLIFetcher) Probe(_ context.Context, cli artifactlock.CLI) error {
	f.Probes = append(f.Probes, cli.Name)
	return f.ProbeErrors[cli.Name]
}

func (f *FakeCLIFetcher) Fetch(_ context.Context, cli artifactlock.CLI, workDir string) error {
	f.Fetches = append(f.Fetches, cli.Name)
	if err := f.FetchErrors[cli.Name]; err != nil {
		return err
	}
	files := cli.Files
	if len(files) == 0 && cli.SourcePath != "" {
		files = []artifactlock.CLIFile{{SourcePath: cli.SourcePath, BundlePath: cli.BundlePath, Digest: cli.Digest, Mode: 0o755}}
	}
	for _, file := range files {
		path := filepath.Join(workDir, "bundle", filepath.FromSlash(file.BundlePath))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, CLIContent(cli.Name, file.SourcePath), os.FileMode(file.Mode)); err != nil {
			return err
		}
	}
	return nil
}

func (f *FakeChartFetcher) Probe(_ context.Context, chart artifactlock.Chart) error {
	f.mu.Lock()
	f.Probes = append(f.Probes, chart.Name)
	f.mu.Unlock()

	return f.ProbeErrors[chart.Name]
}

func (f *FakeChartFetcher) Fetch(_ context.Context, chart artifactlock.Chart, dest string) error {
	f.mu.Lock()
	f.Fetches = append(f.Fetches, chart.Name)
	f.mu.Unlock()

	if err := f.FetchErrors[chart.Name]; err != nil {
		return err
	}
	content, ok := f.Content[chart.Name]
	if !ok {
		content = ChartContent(chart.Name)
	}
	return os.WriteFile(dest, content, 0o644)
}

// ChartContent is the archive a lock's chart digest is computed over, so a
// fetched chart verifies against the lock.
func ChartContent(name string) []byte {
	return []byte("chart-archive:" + name)
}

// ChartDigest is the digest of a chart's canned content, for building a lock
// whose charts verify.
func ChartDigest(name string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(ChartContent(name)))
}
