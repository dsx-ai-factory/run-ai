/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package artifactlock

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/build/airgap/pkg/lock"
)

func TestCollectImageLocksSortsAndHashesInputs(t *testing.T) {
	root := t.TempDir()
	cluster := writeImageLock(t, root, imageLockFixture("runai-cluster", "3.0.0", "cluster", "nvcr.io/nvidia/runai/cluster:3.0.0", "2"))
	controlPlaneLock := imageLockFixture("runai-control-plane", "3.0.0", "control-plane", "nvcr.io/nvidia/runai/control-plane:3.0.0", "1")
	controlPlaneLock.Spec.Images = append(controlPlaneLock.Spec.Images, lock.Image{
		Name:        "postgres",
		Image:       "docker.io/postgres@sha256:" + strings.Repeat("3", 64),
		Source:      "docker.io/postgres:16-alpine",
		IndexDigest: "sha256:" + strings.Repeat("4", 64),
	})
	controlPlane := writeImageLock(t, root, controlPlaneLock)

	collection, err := CollectImageLocks([]string{controlPlane, cluster})
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, imageLock := range collection.Locks {
		names = append(names, imageLock.Name)
	}
	if !reflect.DeepEqual(names, []string{"runai-cluster", "runai-control-plane"}) {
		t.Fatalf("names = %v", names)
	}
	contents, err := os.ReadFile(controlPlane)
	if err != nil {
		t.Fatal(err)
	}
	if collection.Locks[1].Digest != lock.Digest(contents) {
		t.Fatalf("digest = %s", collection.Locks[1].Digest)
	}
	if len(collection.Locks[1].Images) != 2 || collection.Locks[1].Images[1].Name != "postgres" {
		t.Fatalf("images = %#v", collection.Locks[1].Images)
	}
}

func TestCollectImageLocksRejectsPlatformMismatch(t *testing.T) {
	root := t.TempDir()
	first := imageLockFixture("first", "1.0.0", "first", "registry.example.com/first:1.0.0", "1")
	second := imageLockFixture("second", "1.0.0", "second", "registry.example.com/second:1.0.0", "2")
	second.Spec.Platform.Architecture = "arm64"

	_, err := CollectImageLocks([]string{
		writeImageLock(t, root, first),
		writeImageLock(t, root, second),
	})
	if err == nil || !strings.Contains(err.Error(), "same platform") {
		t.Fatalf("error = %v", err)
	}
}

func TestCollectImageLocksRejectsConflictingSources(t *testing.T) {
	root := t.TempDir()
	first := imageLockFixture("first", "1.0.0", "first", "registry.example.com/shared:1.0.0", "1")
	second := imageLockFixture("second", "1.0.0", "second", "registry.example.com/shared:1.0.0", "2")

	_, err := CollectImageLocks([]string{
		writeImageLock(t, root, first),
		writeImageLock(t, root, second),
	})
	if err == nil || !strings.Contains(err.Error(), "conflicting digests") {
		t.Fatalf("error = %v", err)
	}
}

func TestCollectImageLocksDeduplicatesSharedThirdPartyImage(t *testing.T) {
	root := t.TempDir()
	cluster := imageLockFixture("runai-cluster", "3.0.0", "cluster", "registry.example.com/cluster:3.0.0", "1")
	controlPlane := imageLockFixture("runai-control-plane", "3.0.0", "control-plane", "registry.example.com/control-plane:3.0.0", "2")
	shared := lock.Image{
		Name:        "wait-ubi-minimal",
		Image:       "registry.access.redhat.com/ubi9/ubi-minimal@sha256:" + strings.Repeat("3", 64),
		Source:      "registry.access.redhat.com/ubi9/ubi-minimal:9.8",
		IndexDigest: "sha256:" + strings.Repeat("4", 64),
	}
	cluster.Spec.Images = append(cluster.Spec.Images, shared)
	controlPlane.Spec.Images = append(controlPlane.Spec.Images, shared)

	collection, err := CollectImageLocks([]string{
		writeImageLock(t, root, controlPlane),
		writeImageLock(t, root, cluster),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(collection.Locks[0].Images) != 2 {
		t.Fatalf("cluster images = %#v", collection.Locks[0].Images)
	}
	if len(collection.Locks[1].Images) != 1 {
		t.Fatalf("control-plane images = %#v", collection.Locks[1].Images)
	}
}

func imageLockFixture(name, version, imageName, source, digestCharacter string) lock.ImageLock {
	digest := "sha256:" + strings.Repeat(digestCharacter, 64)
	repository := strings.Split(source, ":")[0]
	return lock.ImageLock{
		APIVersion: lock.APIVersion,
		Kind:       lock.ImageKind,
		Metadata:   lock.Metadata{Name: name, Version: version},
		Spec: lock.ImageLockSpec{
			Profile:  "standard",
			Platform: lock.Platform{OS: "linux", Architecture: "amd64"},
			Images: []lock.Image{{
				Name:        imageName,
				Image:       repository + "@" + digest,
				Source:      source,
				IndexDigest: digest,
			}},
		},
	}
}

func writeImageLock(t *testing.T, root string, value lock.ImageLock) string {
	t.Helper()
	contents, err := lock.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, value.Metadata.Name+".yaml")
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
