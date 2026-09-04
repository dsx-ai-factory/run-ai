/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package thirdparty

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestFromChart(t *testing.T) {
	chart := t.TempDir()
	values := `
thirdPartyImages:
  git-sync:
    registry: registry.k8s.io
    repository: git-sync/git-sync
    tag: v4.7.0
  keycloak: keycloakx.image
keycloakx:
  image:
    repository: quay.io/keycloak/keycloak
    tag: "26.6"
`
	if err := os.WriteFile(filepath.Join(chart, "values.yaml"), []byte(values), 0o600); err != nil {
		t.Fatal(err)
	}
	actual, err := FromChart(chart)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Image{
		{Name: "git-sync", Repository: "registry.k8s.io/git-sync/git-sync", Tag: "v4.7.0"},
		{Name: "keycloak", Repository: "quay.io/keycloak/keycloak", Tag: "26.6"},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("images = %#v", actual)
	}
}

func TestFromValuesRejectsIncompleteImage(t *testing.T) {
	_, err := fromValues([]byte(`
thirdPartyImages:
  image:
    repository: example/image
`))
	if err == nil || !strings.Contains(err.Error(), "tag is required") {
		t.Fatalf("error = %v", err)
	}
}
