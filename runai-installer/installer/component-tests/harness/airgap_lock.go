/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package harness

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
	"gopkg.in/yaml.v3"
)

// TestLock returns a small valid artifact lock for pack tests.
func TestLock() *artifactlock.ArtifactLock {
	return &artifactlock.ArtifactLock{
		APIVersion: artifactlock.APIVersion,
		Kind:       artifactlock.Kind,
		Metadata:   artifactlock.Metadata{Name: "runai", Version: "2.26.13"},
		Spec: artifactlock.Spec{
			Profile:  "standard",
			Platform: artifactlock.Platform{OS: artifactlock.OSLinux, Architecture: artifactlock.ArchAMD64},
			ImageLocks: []artifactlock.ImageLock{{
				Name:    "runai-control-plane",
				Version: "2.26.13",
				Digest:  "sha256:3b6ff656fd44c245b97e2de7b78bfb9264de61a9d34464215da481e4d64c86b3",
				Images: []artifactlock.Image{
					{
						Name:        "assets-service",
						Image:       "nvcr.io/nvidia/runai/assets-service@sha256:9741abc964a10bc41271a1bf2cffd56b4d05664cd78782c71b55469950b572d3",
						Source:      "nvcr.io/nvidia/runai/assets-service:2.26.13",
						IndexDigest: "sha256:1e141841fb6cb989ace44a08710797d59219f79889307a831798707cf58f7387",
					},
					{
						Name:        "postgresql",
						Image:       "docker.io/bitnamilegacy/postgresql@sha256:8f50ea374acc3e8da24b1c2ab0d2340cb98b7045ca9b25eff95d634e4874991d",
						Source:      "docker.io/bitnamilegacy/postgresql:16",
						IndexDigest: "sha256:f17be895df5b436547897b59e8a6d4fbdd7e6375d8a218eabd5c9482fc60f3dd",
					},
				},
			}},
			Charts: []artifactlock.Chart{{
				Name:       "control-plane",
				Version:    "2.26.13",
				Source:     "https://helm.ngc.nvidia.com/nvidia/runai/charts/control-plane-2.26.13.tgz",
				BundlePath: "charts/control-plane-2.26.13.tgz",
				Digest:     ChartDigest("control-plane"),
			}, {
				Name:       "runai-cluster",
				Version:    "2.26.13",
				Source:     "https://helm.ngc.nvidia.com/nvidia/runai/charts/runai-cluster-2.26.13.tgz",
				BundlePath: "charts/runai-cluster-2.26.13.tgz",
				Digest:     ChartDigest("runai-cluster"),
			}},
			CLIs: []artifactlock.CLI{},
		},
	}
}

// WriteLockFile marshals a lock to a temp file, returning the path to pass as
// --airgap-lock-file.
func WriteLockFile(t *testing.T, lock *artifactlock.ArtifactLock) string {
	t.Helper()
	out, err := yaml.Marshal(lock)
	if err != nil {
		t.Fatalf("marshal artifact lock: %v", err)
	}
	path := filepath.Join(t.TempDir(), "artifact-lock.yaml")
	if err := os.WriteFile(path, out, 0o600); err != nil {
		t.Fatalf("write artifact lock: %v", err)
	}
	return path
}
