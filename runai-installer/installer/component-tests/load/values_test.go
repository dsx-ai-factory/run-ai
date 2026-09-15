/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package load_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "rewrite the golden values files")

// TestLoad_GeneratedValues loads a bundle holding one image of every kind the
// charts consume differently, and compares both values files to the golden
// ones. The golden files are how a chart-side change to any of those keys is
// caught in review.
func TestLoad_GeneratedValues(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, bindingsLock())

	res := run(t, loadConfig(t, bundle), false)
	require.NoError(t, res.err)
	assert.Contains(t, res.pusher.Pushes, destination+"/karta/operator:0.2.6")
	assert.Contains(t, res.pusher.Pushes, destination+"/kai-scheduler/operator:v0.16.8")

	for _, name := range []string{"control-plane-airgap-values.yaml", "cluster-airgap-values.yaml"} {
		got, err := os.ReadFile(filepath.Join(res.valuesDir, name))
		require.NoError(t, err)

		golden := filepath.Join("testdata", name)
		if *update {
			require.NoError(t, os.MkdirAll("testdata", 0o755))
			require.NoError(t, os.WriteFile(golden, got, 0o644))
		}
		want, err := os.ReadFile(golden)
		require.NoError(t, err)
		assert.Equal(t, string(want), string(got), "%s must match its golden file (re-run with -update)", name)
	}
}

// bindingsLock holds one image of every kind the charts consume differently,
// plus a Run:ai image, which needs no values of its own.
func bindingsLock() *artifactlock.ArtifactLock {
	lock := harness.TestLock()
	lock.Spec.ImageLocks = []artifactlock.ImageLock{
		{
			Name:    "runai-control-plane",
			Version: "2.26.13",
			Images: []artifactlock.Image{
				lockImage("assets-service", "nvcr.io/nvidia/runai/assets-service", "2.26.13"),
				lockImage("postgresql", "docker.io/bitnamilegacy/postgresql", "16"),
				lockImage("thanos", "docker.io/bitnamilegacy/thanos", "0.38.0-debian-12-r3"),
				lockImage("nats", "docker.io/bitnamilegacy/nats", "2.11.8-debian-12-r0"),
				lockImage("keycloak", "quay.io/keycloak/keycloak", "26.6"),
				lockImage("traefik", "docker.io/traefik", "v3.7"),
				lockImage("postgres", "docker.io/postgres", "16-alpine"),
				lockImage("wait-ubi-minimal", "registry.access.redhat.com/ubi9/ubi-minimal", "9.8"),
			},
		},
		{
			Name:    "runai-cluster",
			Version: "2.26.13",
			Images: []artifactlock.Image{
				lockImage("wait-ubi-minimal", "registry.access.redhat.com/ubi9/ubi-minimal", "9.8"),
				lockImage("valkey", "docker.io/valkey/valkey", "9.0.4"),
				lockImage("git-sync", "registry.k8s.io/git-sync/git-sync", "v4.7.0"),
			},
		},
		{
			Name:    "karta",
			Version: "0.2.6",
			Images: []artifactlock.Image{
				lockImage("operator", "ghcr.io/run-ai/karta/karta-operator", "0.2.6"),
			},
		},
		{
			Name:    "kai-scheduler",
			Version: "v0.16.8",
			Images: []artifactlock.Image{
				lockImage("operator", "ghcr.io/kai-scheduler/kai-scheduler/operator", "v0.16.8"),
			},
		},
	}
	return lock
}

func lockImage(name, repository, tag string) artifactlock.Image {
	digest := fmt.Sprintf("sha256:%064x", len(name))
	return artifactlock.Image{
		Name:        name,
		Image:       repository + "@" + digest,
		Source:      repository + ":" + tag,
		IndexDigest: digest,
	}
}
