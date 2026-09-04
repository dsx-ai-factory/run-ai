/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metadata

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDiscoverFiltersAndSortsArtifacts(t *testing.T) {
	root := t.TempDir()
	writeMetadata(t, root, "z", `{"service_name":"z","artifact_type":"container"}`)
	writeMetadata(t, root, "a", `{"service_name":"a","artifact_type":"container","enabled":true}`)
	writeMetadata(t, root, "disabled", `{"service_name":"disabled","artifact_type":"container","enabled":false}`)
	writeMetadata(t, root, "skipped", `{"service_name":"skipped","artifact_type":"container","skip_airgap":true}`)
	writeMetadata(t, root, "library", `{"service_name":"library","artifact_type":"go"}`)

	artifacts, err := Discover([]string{root}, "container")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, artifact := range artifacts {
		names = append(names, artifact.ServiceName)
	}
	if !reflect.DeepEqual(names, []string{"a", "z"}) {
		t.Fatalf("names = %v", names)
	}
}

func TestDiscoverRejectsDuplicateNames(t *testing.T) {
	root := t.TempDir()
	value := `{"service_name":"duplicate","artifact_type":"container"}`
	writeMetadata(t, root, "first", value)
	writeMetadata(t, root, "second", value)

	_, err := Discover([]string{root}, "container")
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("error = %v", err)
	}
}

func TestDiscoverRepositoryArtifacts(t *testing.T) {
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	containers, err := Discover([]string{
		filepath.Join(repository, "control-plane"),
		filepath.Join(repository, "runai-cluster"),
	}, "container")
	if err != nil {
		t.Fatal(err)
	}
	if len(containers) == 0 {
		t.Fatal("no repository container artifacts discovered")
	}
	excluded := map[string]struct{}{
		"new-service-template": {},
	}
	required := map[string]bool{
		"gateway-hub":         false,
		"keycloak-extensions": false,
		"runtime-rules":       false,
	}
	for _, container := range containers {
		if _, exists := excluded[container.ServiceName]; exists {
			t.Errorf("air-gap excluded artifact discovered: %s", container.ServiceName)
		}
		if _, exists := required[container.ServiceName]; exists {
			required[container.ServiceName] = true
		}
	}
	for name, discovered := range required {
		if !discovered {
			t.Errorf("required air-gap artifact not discovered: %s", name)
		}
	}
	clis, err := Discover([]string{filepath.Join(repository, "runai-installer")}, "cli")
	if err != nil {
		t.Fatal(err)
	}
	if len(clis) != 1 || clis[0].ServiceName != "runai-installer" {
		t.Fatalf("CLIs = %#v", clis)
	}
}

func writeMetadata(t *testing.T, root, directory, contents string) {
	t.Helper()
	path := filepath.Join(root, directory)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "artifact-metadata.json"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
