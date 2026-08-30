package harness

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// PushRegistryMarker is the file the default test setup.sh writes the resolved
// registry coordinates into (relative to the extracted working dir), letting
// tests assert the push step actually ran with the configured mirror. The
// contents are "$REGISTRY_URL $REGISTRY_USERNAME $REGISTRY_PASSWORD".
const PushRegistryMarker = "push-registry.txt"

// defaultSetupScript records the registry URL and credentials it was invoked
// with, so the push step has an observable outcome, then succeeds.
const defaultSetupScript = "#!/bin/sh\necho \"$REGISTRY_URL $REGISTRY_USERNAME $REGISTRY_PASSWORD\" > " + PushRegistryMarker + "\nexit 0\n"

// failingSetupScript extracts fine but exits non-zero, simulating a push that
// fails after a successful extraction.
const failingSetupScript = "#!/bin/sh\necho 'push failed' >&2\nexit 1\n"

// CreateTestTarball builds a minimal airgap package layout (charts/, custom-env.yaml,
// setup.sh) in a t.TempDir() and returns the path to the gzipped tarball.
// Contents are stubs — enough to satisfy the airgap preparation step's
// extraction & directory-shape checks without being valid helm charts.
func CreateTestTarball(t *testing.T) string {
	return createTarball(t, defaultSetupScript)
}

// CreateTarballFailingSetup is like CreateTestTarball but its setup.sh exits
// non-zero, so the push-images step fails after extraction succeeds.
func CreateTarballFailingSetup(t *testing.T) string {
	return createTarball(t, failingSetupScript)
}

func createTarball(t *testing.T, setupScript string) string {
	t.Helper()
	dir := t.TempDir()
	tarballPath := filepath.Join(dir, "runai-airgapped-package-test.tar.gz")

	f, err := os.Create(tarballPath)
	require.NoError(t, err)
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "charts/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}))

	files := map[string]string{
		"charts/control-plane-2.20.0.tgz": "fake-cp-chart",
		"charts/runai-cluster-2.20.0.tgz": "fake-cluster-chart",
		"custom-env.yaml":                 "registry: test\n",
	}
	for name, content := range files {
		require.NoError(t, tw.WriteHeader(&tar.Header{
			Name: name,
			Size: int64(len(content)),
			Mode: 0644,
		}))
		_, err := tw.Write([]byte(content))
		require.NoError(t, err)
	}

	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name: "setup.sh",
		Size: int64(len(setupScript)),
		Mode: 0755,
	}))
	_, err = tw.Write([]byte(setupScript))
	require.NoError(t, err)

	return tarballPath
}
