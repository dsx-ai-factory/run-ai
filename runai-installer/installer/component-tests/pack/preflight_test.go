/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pack_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Targets come from the lock, so preflight probes the registries the pack will
// actually pull from, once each.
func TestPack_PreflightProbesEveryRegistryOnce(t *testing.T) {
	res := run(t, packConfig(t), harness.TestLock(), false)
	require.NoError(t, res.err)

	var hosts []string
	for _, call := range res.prober.Calls {
		hosts = append(hosts, call.Host)
	}
	assert.ElementsMatch(t, []string{"nvcr.io", "index.docker.io"}, hosts)

	// The probed scope must be a real repository: a digest is dropped at its
	// "@", not at the colon inside sha256:, or the registry rejects the scope
	// and a valid credential looks invalid.
	var repos []string
	for _, call := range res.prober.Calls {
		repos = append(repos, call.Repo)
	}
	assert.ElementsMatch(t, []string{"nvidia/runai/assets-service", "bitnamilegacy/postgresql"}, repos)

	// The verified sources are named in the output, so the set the bundle pulls
	// from is visible rather than implied.
	assert.Contains(t, res.stdout, "verified registries index.docker.io, nvcr.io; chart repositories helm.ngc.nvidia.com")

	ev := findEvent(res.events, "step_succeeded", "preflight-pack-sources")
	require.NotNil(t, ev)
	data := ev["results"].([]any)[0].(map[string]any)["data"].(map[string]any)
	assert.Equal(t, []any{"index.docker.io", "nvcr.io"}, data["registries"])
	assert.Equal(t, []any{"helm.ngc.nvidia.com"}, data["chartRepositories"])
}

// Charts download after the images, so an unauthorized chart repository must be
// caught before the image set is pulled rather than after.
func TestPack_PreflightFailsOnAnInaccessibleChartRepository(t *testing.T) {
	f := newFakes()
	f.fetcher.ProbeErrors = map[string]error{"control-plane": errors.New("401 Unauthorized")}

	res := runFakes(t, packConfig(t), harness.TestLock(), f)
	require.ErrorIs(t, res.err, specrun.ErrPreflightFailed)

	assert.Equal(t, []string{"control-plane", "runai-cluster"}, res.fetcher.Probes)
	assert.Contains(t, res.stdout, "chart control-plane (helm.ngc.nvidia.com): 401 Unauthorized")
	assert.Empty(t, res.puller.Pulls, "no image is downloaded when a chart repository is unreachable")
	assert.Empty(t, res.fetcher.Fetches)
}

// An OCI chart is served by a registry, so it is probed as one rather than
// reported as a separate chart repository.
func TestPack_PreflightProbesOCIChartsAsRegistries(t *testing.T) {
	lock := harness.TestLock()
	lock.Spec.Charts = lock.Spec.Charts[:1]
	lock.Spec.Charts[0].Source = "oci://ghcr.io/dsx-ai-factory/karta:0.1.0"

	res := run(t, packConfig(t), lock, false)
	require.NoError(t, res.err)

	var hosts []string
	for _, call := range res.prober.Calls {
		hosts = append(hosts, call.Host)
	}
	assert.Contains(t, hosts, "ghcr.io", "the chart's registry is probed")

	ev := findEvent(res.events, "step_succeeded", "preflight-pack-sources")
	require.NotNil(t, ev)
	data := ev["results"].([]any)[0].(map[string]any)["data"].(map[string]any)
	assert.Empty(t, data["chartRepositories"], "an OCI chart is not double-reported")
}

func TestPack_PreflightProbesTaggedOverrides(t *testing.T) {
	cfg := packConfig(t) + "imageOverrides:\n  - name: extra\n    source: harbor.corp.example.com/runai/extra:1.0\n"

	res := run(t, cfg, harness.TestLock(), false)
	require.NoError(t, res.err)

	for _, call := range res.prober.Calls {
		if call.Host == "harbor.corp.example.com" {
			assert.Equal(t, "runai/extra", call.Repo, "a tag is dropped from the probed repository")
			return
		}
	}
	t.Fatal("the override's registry was never probed")
}

func TestPack_PreflightForwardsCredentials(t *testing.T) {
	cfg := packConfig(t) + "registryAuth:\n  - registry: nvcr.io\n    token: ngc-token\n"

	res := run(t, cfg, harness.TestLock(), false)
	require.NoError(t, res.err)

	for _, call := range res.prober.Calls {
		if call.Host != "nvcr.io" {
			continue
		}
		assert.Equal(t, "$oauthtoken", call.Username, "a token authenticates as $oauthtoken")
		assert.Equal(t, "ngc-token", call.Password)
		return
	}
	t.Fatal("nvcr.io was never probed")
}

func TestPack_JWTWithoutDockerAccessClaim(t *testing.T) {
	res := runPackWithRegistryClaims(t, map[string]any{
		"sub": "ci-user",
		"scp": "member-of-groups:readers",
	})

	require.NoError(t, res.err)
	assert.Len(t, res.puller.Pulls, 1)
	assert.FileExists(t, res.output)
	require.NotNil(t, findEvent(res.events, "step_succeeded", "preflight-pack-sources"))
}

func TestPack_JWTWithInvalidDockerAccessClaimFailsClosed(t *testing.T) {
	for name, access := range map[string]any{
		"null":      nil,
		"malformed": "pull",
	} {
		t.Run(name, func(t *testing.T) {
			res := runPackWithRegistryClaims(t, map[string]any{
				"sub":    "ci-user",
				"access": access,
			})

			require.ErrorIs(t, res.err, specrun.ErrPreflightFailed)
			assert.Empty(t, res.puller.Pulls)
			assert.Contains(t, res.stdout, "no pull access")
		})
	}
}

func runPackWithRegistryClaims(t *testing.T, claims map[string]any) result {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/":
			w.Header().Set("WWW-Authenticate", `Bearer realm="`+server.URL+`/token",service="registry"`)
			w.WriteHeader(http.StatusUnauthorized)
		case "/token":
			username, password, _ := r.BasicAuth()
			if username != "ci-user" || password != "ci-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"token": registryJWT(t, claims)})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	host := strings.TrimPrefix(server.URL, "https://")
	lock := harness.TestLock()
	lock.Spec.ImageLocks[0].Images = []artifactlock.Image{{
		Name:        "app",
		Image:       host + "/runai/app@sha256:8f50ea374acc3e8da24b1c2ab0d2340cb98b7045ca9b25eff95d634e4874991d",
		Source:      host + "/runai/app:1.0",
		IndexDigest: "sha256:f17be895df5b436547897b59e8a6d4fbdd7e6375d8a218eabd5c9482fc60f3dd",
	}}
	config := packConfig(t) + fmt.Sprintf("registryAuth:\n  - registry: %s\n    username: ci-user\n    password: ci-token\n    insecureSkipVerify: true\n", host)
	f := newFakes()
	f.realProber = orchestrator2.NewRegistryProber()

	return runFakes(t, config, lock, f)
}

func registryJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": "RS256"})
	require.NoError(t, err)
	payload, err := json.Marshal(claims)
	require.NoError(t, err)
	encode := base64.RawURLEncoding.EncodeToString
	return encode(header) + "." + encode(payload) + ".signature"
}

func TestPack_PreflightFailsOnAnInaccessibleRegistry(t *testing.T) {
	f := newFakes()
	f.prober.Results = map[string]orchestrator2.RegistryCheckResult{
		"nvcr.io": {Reachable: true, Authenticated: false, Detail: "unauthorized"},
	}

	res := runFakes(t, packConfig(t), harness.TestLock(), f)
	require.ErrorIs(t, res.err, specrun.ErrPreflightFailed)

	assert.Empty(t, res.puller.Pulls, "nothing is downloaded after preflight fails")
	assert.Contains(t, res.stdout, "nvcr.io: credentials rejected")
	assert.Contains(t, res.stdout, "add credentials for private registries under registryAuth")

	_, err := os.Stat(res.output)
	assert.True(t, os.IsNotExist(err))
}

func TestPack_PreflightFailsOnAnUnresolvableOverride(t *testing.T) {
	f := newFakes()
	f.puller.ResolveErrors = map[string]error{
		"harbor.corp.example.com/runai/missing:1.0": errors.New("manifest unknown"),
	}
	cfg := packConfig(t) + "imageOverrides:\n  - name: missing\n    source: harbor.corp.example.com/runai/missing:1.0\n"

	res := runFakes(t, cfg, harness.TestLock(), f)
	require.ErrorIs(t, res.err, specrun.ErrPreflightFailed)

	assert.Contains(t, res.stdout, `image override "missing": manifest unknown`)
	assert.Empty(t, res.puller.Pulls)
}

func TestPack_SkipPreflightPacksWithoutProbing(t *testing.T) {
	f := newFakes()
	f.prober.Results = map[string]orchestrator2.RegistryCheckResult{
		"nvcr.io": {Reachable: false, Detail: "unauthorized"},
	}

	res := runStaged(t, packConfig(t), harness.TestLock(), f, func(o *specrun.ExecOptions) {
		o.SkipPreflight = true
	})
	require.NoError(t, res.err)

	assert.Empty(t, res.prober.Calls, "no registry is probed")
	assert.Len(t, res.puller.Pulls, 2)
}

func TestPack_PreflightOnlyStopsBeforeDownloading(t *testing.T) {
	res := runStaged(t, packConfig(t), harness.TestLock(), newFakes(), func(o *specrun.ExecOptions) {
		o.PreflightOnly = true
	})
	require.NoError(t, res.err)

	assert.NotEmpty(t, res.prober.Calls, "the checks still run")
	assert.Empty(t, res.puller.Pulls, "nothing is downloaded")

	_, err := os.Stat(res.output)
	assert.True(t, os.IsNotExist(err))
}

func TestPack_PreflightAppliesDockerHubCredentials(t *testing.T) {
	cfg := packConfig(t) + "registryAuth:\n  - registry: docker.io\n    username: hub-user\n    password: hub-pass\n"

	res := run(t, cfg, harness.TestLock(), false)
	require.NoError(t, res.err)

	for _, call := range res.prober.Calls {
		if call.Host != "index.docker.io" {
			continue
		}
		assert.Equal(t, "hub-user", call.Username, "a docker.io entry must apply to index.docker.io")
		assert.Equal(t, "hub-pass", call.Password)
		return
	}
	t.Fatal("Docker Hub was never probed")
}

func TestPack_PreflightSkipsOverriddenRegistries(t *testing.T) {
	f := newFakes()
	f.prober.Results = map[string]orchestrator2.RegistryCheckResult{
		"index.docker.io": {Reachable: false, Detail: "blocked by the corporate proxy"},
	}
	cfg := packConfig(t) + "imageOverrides:\n  - name: postgresql\n    source: harbor.corp.example.com/runai/postgresql:16\n"

	res := runFakes(t, cfg, harness.TestLock(), f)
	require.NoError(t, res.err, "the blocked registry is replaced by the override, so it must not be probed")

	var hosts []string
	for _, call := range res.prober.Calls {
		hosts = append(hosts, call.Host)
	}
	assert.NotContains(t, hosts, "index.docker.io")
	assert.Contains(t, hosts, "harbor.corp.example.com")
}

func TestPack_PreflightProbesRegistrylessOverrides(t *testing.T) {
	cfg := packConfig(t) + "imageOverrides:\n  - name: postgresql\n    source: postgres:16-alpine\n"

	res := run(t, cfg, harness.TestLock(), false)
	require.NoError(t, res.err)

	for _, call := range res.prober.Calls {
		if call.Host == "index.docker.io" {
			assert.Equal(t, "library/postgres", call.Repo, "an official image resolves to the library namespace")
			return
		}
	}
	t.Fatal("Docker Hub was never probed")
}
