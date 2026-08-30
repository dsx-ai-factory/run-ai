package harness

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/artifacts"
	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
)

// FakeImagePusher records pushes and reports the digest each layout holds, as
// a real push does.
type FakeImagePusher struct {
	// PushErrors fails one push, keyed by destination reference.
	PushErrors map[string]error
	CheckError error
	// Digests overrides the digest a push reports, keyed by destination
	// reference, so a mismatch can be staged.
	Digests map[string]string

	mu     sync.Mutex
	Pushes []string
	Checks []string
	// Layouts records the directory each push read, keyed by destination
	// reference.
	Layouts map[string]string
}

func (f *FakeImagePusher) CheckPushPermission(_ context.Context, ref string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Checks = append(f.Checks, ref)
	return f.CheckError
}

func (f *FakeImagePusher) Push(_ context.Context, dir, ref, expectedDigest string) error {
	if expectedDigest == "" {
		return fmt.Errorf("image %s has no expected digest", ref)
	}
	digest, ok := f.Digests[ref]
	if !ok {
		var err error
		digest, err = layoutDigest(dir)
		if err != nil {
			return err
		}
	}
	if digest != expectedDigest {
		return fmt.Errorf("image %s has digest %s but the lock pins %s", ref, digest, expectedDigest)
	}

	f.mu.Lock()
	f.Pushes = append(f.Pushes, ref)
	if f.Layouts == nil {
		f.Layouts = map[string]string{}
	}
	f.Layouts[dir] = ref
	f.mu.Unlock()

	if err := f.PushErrors[ref]; err != nil {
		return err
	}
	return nil
}

// layoutDigest reads the digest an OCI layout holds.
func layoutDigest(dir string) (string, error) {
	content, err := os.ReadFile(filepath.Join(dir, "index.json"))
	if err != nil {
		return "", err
	}
	var index struct {
		Manifests []struct {
			Digest string `json:"digest"`
		} `json:"manifests"`
	}
	if err := json.Unmarshal(content, &index); err != nil {
		return "", err
	}
	if len(index.Manifests) != 1 {
		return "", fmt.Errorf("layout %s holds %d image(s), expected 1", dir, len(index.Manifests))
	}
	return index.Manifests[0].Digest, nil
}

// CreateLoadBundle writes the bundle a lock describes, as the pack does, and
// returns the path to the archive.
func CreateLoadBundle(t *testing.T, lock *artifactlock.ArtifactLock) string {
	t.Helper()

	workDir := t.TempDir()
	if err := artifacts.InitBundleDir(workDir); err != nil {
		t.Fatalf("init bundle: %v", err)
	}
	if err := artifacts.WriteLock(workDir, lock); err != nil {
		t.Fatalf("write lock: %v", err)
	}
	for _, entry := range lock.ImageEntries() {
		writeLayout(t, artifacts.ImageDir(workDir, entry.Group, entry.Image.Name), imageDigest(entry.Image))
	}
	for _, chart := range lock.Spec.Charts {
		if err := os.WriteFile(artifacts.ChartPath(workDir, chart.BundlePath), ChartContent(chart.Name), 0o644); err != nil {
			t.Fatalf("write chart: %v", err)
		}
	}
	for _, cli := range lock.Spec.CLIs {
		files := cli.Files
		if len(files) == 0 && cli.SourcePath != "" {
			files = []artifactlock.CLIFile{{SourcePath: cli.SourcePath, BundlePath: cli.BundlePath, Digest: cli.Digest, Mode: 0o755}}
		}
		for _, file := range files {
			path, err := artifacts.CLIPath(workDir, file.BundlePath)
			if err != nil {
				t.Fatalf("resolve CLI path: %v", err)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatalf("create CLI directory: %v", err)
			}
			if err := os.WriteFile(path, CLIContent(cli.Name, file.SourcePath), os.FileMode(file.Mode)); err != nil {
				t.Fatalf("write CLI: %v", err)
			}
		}
	}

	bundle := filepath.Join(t.TempDir(), "runai-airgap.tar")
	if _, err := artifacts.TarDir(context.Background(), artifacts.BundleDir(workDir), bundle); err != nil {
		t.Fatalf("write bundle: %v", err)
	}
	return bundle
}

func CLIContent(name, sourcePath string) []byte {
	return []byte("cli:" + name + ":" + sourcePath)
}

func CLIDigest(name, sourcePath string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(CLIContent(name, sourcePath)))
}

// GzipFile compresses a file, for asserting that a compressed bundle loads the
// same as a plain one.
func GzipFile(t *testing.T, path string) string {
	t.Helper()

	in, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer in.Close()

	out, err := os.Create(path + ".gz")
	if err != nil {
		t.Fatalf("create %s.gz: %v", path, err)
	}
	defer out.Close()

	gz := gzip.NewWriter(out)
	if _, err := io.Copy(gz, in); err != nil {
		t.Fatalf("compress %s: %v", path, err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("compress %s: %v", path, err)
	}
	return path + ".gz"
}

func writeLayout(t *testing.T, dir, digest string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(dir, "blobs", "sha256"), 0o755); err != nil {
		t.Fatalf("create layout: %v", err)
	}
	index := fmt.Sprintf(`{"schemaVersion":2,"manifests":[{"digest":%q}]}`, digest)
	if err := os.WriteFile(filepath.Join(dir, "index.json"), []byte(index), 0o644); err != nil {
		t.Fatalf("write layout: %v", err)
	}
}

// imageDigest is the digest a locked reference pins, which the bundled layout
// holds.
func imageDigest(image artifactlock.Image) string {
	for i := 0; i < len(image.Image); i++ {
		if image.Image[i] == '@' {
			return image.Image[i+1:]
		}
	}
	return ""
}
