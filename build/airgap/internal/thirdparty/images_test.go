/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package thirdparty

import (
	"reflect"
	"testing"
)

func TestControlPlaneImages(t *testing.T) {
	expected := []Image{
		{Name: "keycloak", Repository: "quay.io/keycloak/keycloak", Tag: "26.6"},
		{Name: "nats", Repository: "docker.io/bitnamilegacy/nats", Tag: "2.11.8-debian-12-r0"},
		{Name: "postgres", Repository: "docker.io/postgres", Tag: "16-alpine"},
		{Name: "postgresql", Repository: "docker.io/bitnamilegacy/postgresql", Tag: "16"},
		{Name: "thanos", Repository: "docker.io/bitnamilegacy/thanos", Tag: "0.38.0-debian-12-r3"},
		{Name: "traefik", Repository: "docker.io/traefik", Tag: "v3.7"},
		{Name: "wait-ubi-minimal", Repository: "registry.access.redhat.com/ubi9/ubi-minimal", Tag: "9.8"},
	}
	if actual := ForLock("runai-control-plane"); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("images = %#v", actual)
	}
}

func TestClusterImages(t *testing.T) {
	expected := []Image{
		{Name: "git-sync", Repository: "registry.k8s.io/git-sync/git-sync", Tag: "v4.7.0"},
		{Name: "valkey", Repository: "docker.io/valkey/valkey", Tag: "9.0.4"},
		{Name: "wait-ubi-minimal", Repository: "registry.access.redhat.com/ubi9/ubi-minimal", Tag: "9.8"},
	}
	if actual := ForLock("runai-cluster"); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("images = %#v", actual)
	}
}

func TestUnknownLockHasNoThirdPartyImages(t *testing.T) {
	if images := ForLock("external"); len(images) != 0 {
		t.Fatalf("images = %#v", images)
	}
}
