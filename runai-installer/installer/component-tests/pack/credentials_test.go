package pack_test

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/artifacts"
	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPack_CredentialIsNotSentToAnotherPortOnTheSameHost(t *testing.T) {
	lock := harness.TestLock()
	lock.Spec.ImageLocks[0].Images = []artifactlock.Image{{
		Name:        "app",
		Image:       "registry.example.com:5000/team/app@sha256:8f50ea374acc3e8da24b1c2ab0d2340cb98b7045ca9b25eff95d634e4874991d",
		Source:      "registry.example.com:5000/team/app:1.0",
		IndexDigest: "sha256:f17be895df5b436547897b59e8a6d4fbdd7e6375d8a218eabd5c9482fc60f3dd",
	}}
	cfg := packConfig(t) + "registryAuth:\n  - registry: registry.example.com\n    username: u\n    password: p\n"

	res := run(t, cfg, lock, false)
	require.NoError(t, res.err)

	for _, call := range res.prober.Calls {
		if call.Host == "registry.example.com:5000" {
			assert.Empty(t, call.Username, "a credential for the bare host must not reach a different port")
			assert.Empty(t, call.Password)
			return
		}
	}
	t.Fatal("the registry was never probed")
}

// The chart download runs for real against a local server, so what reaches the
// wire is what is asserted.
func chartServer(t *testing.T, body []byte, seen *[]string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*seen = append(*seen, r.Header.Get("Authorization"))
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)
	return server
}

func plaintextChartLock(t *testing.T, url string, body []byte) *artifactlock.ArtifactLock {
	t.Helper()
	lock := harness.TestLock()
	lock.Spec.Charts = []artifactlock.Chart{{
		Name:       "control-plane",
		Version:    "2.26.13",
		Source:     url,
		BundlePath: "charts/control-plane-2.26.13.tgz",
		Digest:     fmt.Sprintf("sha256:%x", sha256.Sum256(body)),
	}}
	return lock
}

func TestPack_CredentialIsWithheldFromAPlaintextChartURL(t *testing.T) {
	body := []byte("chart-archive")
	var seen []string
	server := chartServer(t, body, &seen)

	host := strings.TrimPrefix(server.URL, "http://")
	cfg := packConfig(t) + "registryAuth:\n  - registry: " + host + "\n    username: u\n    password: p\n"

	f := newFakes()
	f.realFetcher = artifacts.NewChartFetcher(artifacts.NewKeychain(artifacts.Credentials{{
		Prefix: host, Username: "u", Password: "p",
	}}))

	res := runFakes(t, cfg, plaintextChartLock(t, server.URL+"/control-plane.tgz", body), f)
	require.NoError(t, res.err)

	require.NotEmpty(t, seen, "the chart server was never contacted")
	for _, auth := range seen {
		assert.Empty(t, auth, "credentials must not travel over a plaintext URL")
	}
}

func TestPack_PlaintextChartCredentialRequiresInsecure(t *testing.T) {
	body := []byte("chart-archive")
	var seen []string
	server := chartServer(t, body, &seen)

	host := strings.TrimPrefix(server.URL, "http://")
	cfg := packConfig(t) + "registryAuth:\n  - registry: " + host + "\n    username: u\n    password: p\n    insecure: true\n"

	f := newFakes()
	f.realFetcher = artifacts.NewChartFetcher(artifacts.NewKeychain(artifacts.Credentials{{
		Prefix: host, Username: "u", Password: "p", Insecure: true,
	}}))

	res := runFakes(t, cfg, plaintextChartLock(t, server.URL+"/control-plane.tgz", body), f)
	require.NoError(t, res.err)

	require.NotEmpty(t, seen)
	for _, auth := range seen {
		assert.True(t, strings.HasPrefix(auth, "Basic "), "an explicit insecure opt-in still sends credentials")
	}
}
