/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package load_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/artifacts"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoad_PushesToARealRegistry runs the command against a real registry with
// the real engine, so the parts a fake cannot stand in for — reading an OCI
// layout, the upload itself, and the insecure-registry handling — are covered.
func TestLoad_PushesToARealRegistry(t *testing.T) {
	server := httptest.NewServer(registry.New(registry.Logger(log.New(io.Discard, "", 0))))
	t.Cleanup(server.Close)
	host := strings.TrimPrefix(server.URL, "http://")
	mirror := host + "/runai"

	bundle, digests := realBundle(t)

	config := fmt.Sprintf(`settings:
  bundlePath: %s
  workingDir: %s
registry:
  url: %s
  insecure: true
`, bundle, filepath.Join(t.TempDir(), "airgap"), mirror)

	pusher, err := artifacts.NewImagePusher(artifacts.Credential{Prefix: mirror, Insecure: true})
	require.NoError(t, err)
	res := runWithDeps(t, config, false, orchestrator2.Deps{ImagePusher: pusher})
	require.NoError(t, res.err)

	for imageName, want := range digests {
		ref, err := name.ParseReference(mirror+"/runai-control-plane/"+imageName+":2.26.13", name.Insecure)
		require.NoError(t, err)

		descriptor, err := remote.Get(ref, remote.WithContext(context.Background()))
		require.NoError(t, err, "%s must be in the registry", imageName)
		assert.Equal(t, want, descriptor.Digest.String(), "%s must be stored byte for byte", imageName)
	}
}

func TestLoad_AuthenticatesPreflightAndPush(t *testing.T) {
	const username = "admin"
	const password = "secret"
	registryHandler := registry.New(registry.Logger(log.New(io.Discard, "", 0)))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUsername, gotPassword, ok := r.BasicAuth()
		if !ok || gotUsername != username || gotPassword != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="registry"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		registryHandler.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	host := strings.TrimPrefix(server.URL, "http://")
	mirror := host + "/runai"
	bundle, digests := realBundle(t)
	config := fmt.Sprintf(`settings:
  bundlePath: %s
  workingDir: %s
registry:
  url: %s
  username: %s
  password: %s
  insecure: true
`, bundle, filepath.Join(t.TempDir(), "airgap"), mirror, username, password)

	pusher, err := artifacts.NewImagePusher(artifacts.Credential{
		Prefix: mirror, Username: username, Password: password, Insecure: true,
	})
	require.NoError(t, err)
	res := runWithDeps(t, config, false, orchestrator2.Deps{ImagePusher: pusher})
	require.NoError(t, res.err)

	for imageName, want := range digests {
		ref, err := name.ParseReference(mirror+"/runai-control-plane/"+imageName+":2.26.13", name.Insecure)
		require.NoError(t, err)
		descriptor, err := remote.Get(ref, remote.WithAuth(&authn.Basic{Username: username, Password: password}))
		require.NoError(t, err)
		assert.Equal(t, want, descriptor.Digest.String())
	}

	rejectedPusher, err := artifacts.NewImagePusher(artifacts.Credential{
		Prefix: mirror, Username: username, Password: "wrong", Insecure: true,
	})
	require.NoError(t, err)
	rejected := runWithDeps(t, config, false, orchestrator2.Deps{ImagePusher: rejectedPusher})
	require.ErrorIs(t, rejected.err, specrun.ErrPreflightFailed)
}

func TestLoad_DigestMismatchDoesNotReachRegistry(t *testing.T) {
	server := httptest.NewServer(registry.New(registry.Logger(log.New(io.Discard, "", 0))))
	t.Cleanup(server.Close)
	host := strings.TrimPrefix(server.URL, "http://")
	mirror := host + "/runai"

	bundle, _ := realBundleWithLock(t, func(lock *artifactlock.ArtifactLock) {
		lock.Spec.ImageLocks[0].Images[0].Image = "nvcr.io/nvidia/runai/assets-service@sha256:0000000000000000000000000000000000000000000000000000000000000000"
	})

	config := fmt.Sprintf(`settings:
  bundlePath: %s
  workingDir: %s
registry:
  url: %s
  insecure: true
`, bundle, filepath.Join(t.TempDir(), "airgap"), mirror)

	pusher, err := artifacts.NewImagePusher(artifacts.Credential{Prefix: mirror, Insecure: true})
	require.NoError(t, err)
	res := runWithDeps(t, config, false, orchestrator2.Deps{ImagePusher: pusher})
	require.Error(t, res.err)

	ref, err := name.ParseReference(mirror+"/runai-control-plane/assets-service:2.26.13", name.Insecure)
	require.NoError(t, err)
	_, err = remote.Get(ref, remote.WithContext(context.Background()))
	require.Error(t, err)
}

// realBundle writes a bundle holding real images, returning its path and the
// digest of each image by name.
func realBundle(t *testing.T) (string, map[string]string) {
	return realBundleWithLock(t, nil)
}

func realBundleWithLock(t *testing.T, mutate func(*artifactlock.ArtifactLock)) (string, map[string]string) {
	t.Helper()

	workDir := t.TempDir()
	require.NoError(t, artifacts.InitBundleDir(workDir))

	digests := map[string]string{}
	var images []artifactlock.Image
	for _, imageName := range []string{"assets-service", "frontend"} {
		image, err := random.Image(256, 2)
		require.NoError(t, err)
		digest, err := image.Digest()
		require.NoError(t, err)

		path, err := layout.Write(artifacts.ImageDir(workDir, "runai-control-plane", imageName), empty.Index)
		require.NoError(t, err)
		require.NoError(t, path.AppendImage(image))

		digests[imageName] = digest.String()
		images = append(images, artifactlock.Image{
			Name:        imageName,
			Image:       "nvcr.io/nvidia/runai/" + imageName + "@" + digest.String(),
			Source:      "nvcr.io/nvidia/runai/" + imageName + ":2.26.13",
			IndexDigest: digest.String(),
		})
	}

	lock := harness.TestLock()
	lock.Spec.ImageLocks[0].Images = images
	for _, chart := range lock.Spec.Charts {
		require.NoError(t, os.WriteFile(artifacts.ChartPath(workDir, chart.BundlePath), harness.ChartContent(chart.Name), 0o644))
	}
	if mutate != nil {
		mutate(lock)
	}
	require.NoError(t, artifacts.WriteLock(workDir, lock))

	bundle := filepath.Join(t.TempDir(), "runai-airgap.tar")
	_, err := artifacts.TarDir(context.Background(), artifacts.BundleDir(workDir), bundle)
	require.NoError(t, err)
	return bundle, digests
}
