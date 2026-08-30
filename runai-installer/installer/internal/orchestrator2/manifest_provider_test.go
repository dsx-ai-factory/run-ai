package orchestrator2

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tarGz builds a gzipped tar archive from the given path->content map, for
// serving as a fake GitHub source tarball.
func tarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		require.NoError(t, tw.WriteHeader(&tar.Header{
			Name:     name,
			Mode:     0600,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}))
		_, err := tw.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

// TestManifestProvider_RenderKustomize renders an overlay from a tarball that
// nests everything under a single top-level directory (as GitHub archives do),
// verifying that directory is stripped and the kustomize build resolves the
// overlay at the requested subPath.
func TestManifestProvider_RenderKustomize(t *testing.T) {
	tarball := tarGz(t, map[string]string{
		"training-operator-1.8.1/manifests/overlays/standalone/kustomization.yaml": "resources:\n- namespace.yaml\n",
		"training-operator-1.8.1/manifests/overlays/standalone/namespace.yaml":     "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: kubeflow\n",
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(tarball)
	}))
	defer srv.Close()

	p := &realManifestProvider{client: srv.Client(), attempts: 1, backoff: time.Millisecond}
	out, err := p.RenderKustomize(context.Background(), srv.URL, "manifests/overlays/standalone")
	require.NoError(t, err)
	assert.Contains(t, string(out), "kind: Namespace")
	assert.Contains(t, string(out), "name: kubeflow")
}

// TestManifestProvider_RenderKustomizeBadSubPath verifies a missing overlay
// surfaces a kustomize build error rather than empty output.
func TestManifestProvider_RenderKustomizeBadSubPath(t *testing.T) {
	tarball := tarGz(t, map[string]string{
		"repo-1.0/config/kustomization.yaml": "resources: []\n",
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(tarball)
	}))
	defer srv.Close()

	p := &realManifestProvider{client: srv.Client(), attempts: 1, backoff: time.Millisecond}
	_, err := p.RenderKustomize(context.Background(), srv.URL, "does/not/exist")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kustomize build")
}

// TestManifestProvider_RenderKustomizeTarballNotFound verifies a 4xx on the
// tarball download surfaces as an error (and is not retried away).
func TestManifestProvider_RenderKustomizeTarballNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	p := &realManifestProvider{client: srv.Client(), attempts: 3, backoff: time.Millisecond}
	_, err := p.RenderKustomize(context.Background(), srv.URL, "any")
	require.Error(t, err)
}

// TestManifestProvider_FetchRetriesTransient verifies a transient 5xx is retried
// until it succeeds.
func TestManifestProvider_FetchRetriesTransient(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok-manifest"))
	}))
	defer srv.Close()

	p := &realManifestProvider{client: srv.Client(), attempts: 3, backoff: time.Millisecond}
	body, err := p.FetchManifest(context.Background(), srv.URL)
	require.NoError(t, err)
	assert.Equal(t, "ok-manifest", string(body))
	assert.EqualValues(t, 3, atomic.LoadInt32(&calls), "fetch retried until success")
}

// TestManifestProvider_FetchExhaustsRetries verifies persistent 5xx fails after
// the configured attempts.
func TestManifestProvider_FetchExhaustsRetries(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	p := &realManifestProvider{client: srv.Client(), attempts: 3, backoff: time.Millisecond}
	_, err := p.FetchManifest(context.Background(), srv.URL)
	require.Error(t, err)
	assert.EqualValues(t, 3, atomic.LoadInt32(&calls), "all attempts used")
}

// TestManifestProvider_FetchNoRetryOn4xx verifies a 4xx is not retried (it will
// not recover).
func TestManifestProvider_FetchNoRetryOn4xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	p := &realManifestProvider{client: srv.Client(), attempts: 3, backoff: time.Millisecond}
	_, err := p.FetchManifest(context.Background(), srv.URL)
	require.Error(t, err)
	assert.EqualValues(t, 1, atomic.LoadInt32(&calls), "4xx must not be retried")
}
