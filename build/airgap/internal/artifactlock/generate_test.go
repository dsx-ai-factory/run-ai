/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package artifactlock

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/build/airgap/pkg/lock"
)

func TestBuildCreatesValidArtifactLock(t *testing.T) {
	root := t.TempDir()
	imagePath := writeImageLock(t, root, imageLockFixture(
		"runai",
		"3.0.0",
		"service",
		"nvcr.io/nvidia/runai/service:3.0.0",
		"1",
	))
	runner := &fakeHelmRunner{
		archive:  chartArchive(t, "control-plane", "3.0.0", nil, false),
		filename: "control-plane-3.0.0.tgz",
	}

	result, err := Build(context.Background(), Config{
		Name:          "runai",
		Version:       "master",
		ImageLocks:    []string{imagePath},
		ChartSources:  []string{"https://charts.example.com/control-plane-3.0.0.tgz"},
		ChartVersions: []string{"3.0.0"},
	}, Dependencies{
		HelmRunner: runner,
	})
	if err != nil {
		t.Fatal(err)
	}
	contents, err := lock.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := lock.ParseArtifactLock(contents)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Metadata.Version != "master" || len(parsed.Spec.ImageLocks) != 1 || len(parsed.Spec.Charts) != 1 {
		t.Fatalf("artifact lock = %#v", parsed)
	}
	if parsed.Spec.CLIs == nil {
		t.Fatal("CLIs must serialize as an empty list")
	}
}
