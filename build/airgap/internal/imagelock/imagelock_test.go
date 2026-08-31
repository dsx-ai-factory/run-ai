/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package imagelock

import (
	"context"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/run-ai/runai/build/airgap/internal/thirdparty"
	"github.com/run-ai/runai/build/airgap/pkg/lock"
)

type fakeResolver struct {
	mutex    sync.Mutex
	calls    []string
	failures map[string]error
}

func (resolver *fakeResolver) Resolve(_ context.Context, source string, _ lock.Platform) (string, string, error) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.calls = append(resolver.calls, source)
	if err := resolver.failures[source]; err != nil {
		return "", "", err
	}
	return "sha256:" + strings.Repeat("1", 64), "sha256:" + strings.Repeat("2", 64), nil
}

func TestBuildCreatesDeterministicPlatformLock(t *testing.T) {
	root := t.TempDir()
	writeContainerMetadata(t, root, "z-service")
	writeContainerMetadata(t, root, "a-service")
	resolver := &fakeResolver{}

	result, err := Build(context.Background(), Config{
		Name:       "runai-control-plane",
		Version:    "3.0.0",
		Registry:   "nvcr.io/nvidia/runai/",
		Platform:   lock.Platform{OS: "linux", Architecture: "arm64"},
		Roots:      []string{root},
		MaxWorkers: 2,
		ThirdParty: []thirdparty.Image{{
			Name:       "postgres",
			Repository: "docker.io/postgres",
			Tag:        "16-alpine",
		}},
	}, resolver)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, image := range result.Spec.Images {
		names = append(names, image.Name)
	}
	if !reflect.DeepEqual(names, []string{"a-service", "postgres", "z-service"}) {
		t.Fatalf("names = %v", names)
	}
	if result.Metadata.Name != "runai-control-plane" {
		t.Fatalf("name = %s", result.Metadata.Name)
	}
	if result.Spec.Platform.Architecture != "arm64" {
		t.Fatalf("platform = %#v", result.Spec.Platform)
	}
	if result.Spec.Images[0].Source != "nvcr.io/nvidia/runai/a-service:3.0.0" {
		t.Fatalf("source = %s", result.Spec.Images[0].Source)
	}
	if result.Spec.Images[0].Image != "nvcr.io/nvidia/runai/a-service@sha256:"+strings.Repeat("2", 64) {
		t.Fatalf("image = %s", result.Spec.Images[0].Image)
	}
	thirdPartyImage := result.Spec.Images[1]
	if thirdPartyImage.Name != "postgres" || thirdPartyImage.Source != "docker.io/postgres:16-alpine" {
		t.Fatalf("third-party image = %#v", thirdPartyImage)
	}
}

func TestBuildAttemptsEveryImageAndFailsClosed(t *testing.T) {
	root := t.TempDir()
	writeContainerMetadata(t, root, "first")
	writeContainerMetadata(t, root, "second")
	resolver := &fakeResolver{failures: map[string]error{
		"registry.example.com/first:3.0.0":  errors.New("first failed"),
		"registry.example.com/second:3.0.0": errors.New("second failed"),
	}}

	_, err := Build(context.Background(), Config{
		Name:       "runai-cluster",
		Version:    "3.0.0",
		Registry:   "registry.example.com",
		Platform:   lock.Platform{OS: "linux", Architecture: "amd64"},
		Roots:      []string{root},
		MaxWorkers: 2,
	}, resolver)
	if err == nil || !strings.Contains(err.Error(), "first failed") || !strings.Contains(err.Error(), "second failed") {
		t.Fatalf("error = %v", err)
	}
	if len(resolver.calls) != 2 {
		t.Fatalf("calls = %v", resolver.calls)
	}
}

func TestBuildRejectsLatestVersion(t *testing.T) {
	_, err := Build(context.Background(), Config{Name: "runai", Version: "latest"}, &fakeResolver{})
	if err == nil || !strings.Contains(err.Error(), "latest") {
		t.Fatalf("error = %v", err)
	}
}

func TestBuildRequiresName(t *testing.T) {
	if _, err := Build(context.Background(), Config{}, &fakeResolver{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoteResolverSelectsPlatformDigest(t *testing.T) {
	server := httptest.NewTLSServer(registry.New())
	defer server.Close()
	source := strings.TrimPrefix(server.URL, "https://") + "/runai/service:3.0.0"
	reference, err := name.ParseReference(source)
	if err != nil {
		t.Fatal(err)
	}
	amd64, err := random.Image(512, 1)
	if err != nil {
		t.Fatal(err)
	}
	arm64, err := random.Image(512, 1)
	if err != nil {
		t.Fatal(err)
	}
	index := mutate.AppendManifests(
		empty.Index,
		mutate.IndexAddendum{
			Add: amd64,
			Descriptor: v1.Descriptor{
				Platform: &v1.Platform{OS: "linux", Architecture: "amd64"},
			},
		},
		mutate.IndexAddendum{
			Add: arm64,
			Descriptor: v1.Descriptor{
				Platform: &v1.Platform{OS: "linux", Architecture: "arm64"},
			},
		},
	)
	transport := server.Client().Transport
	if err := remote.WriteIndex(reference, index, remote.WithTransport(transport)); err != nil {
		t.Fatal(err)
	}

	indexDigest, platformDigest, err := (RemoteResolver{
		Options: []remote.Option{remote.WithTransport(transport)},
	}).Resolve(context.Background(), source, lock.Platform{OS: "linux", Architecture: "arm64"})
	if err != nil {
		t.Fatal(err)
	}
	expectedIndexDigest, err := index.Digest()
	if err != nil {
		t.Fatal(err)
	}
	expectedPlatformDigest, err := arm64.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if indexDigest != expectedIndexDigest.String() || platformDigest != expectedPlatformDigest.String() {
		t.Fatalf("digests = %s, %s", indexDigest, platformDigest)
	}
}

func TestRemoteResolverAcceptsSingleArchitectureImage(t *testing.T) {
	server := httptest.NewTLSServer(registry.New())
	defer server.Close()
	source := strings.TrimPrefix(server.URL, "https://") + "/third-party/postgres:16"
	reference, err := name.ParseReference(source)
	if err != nil {
		t.Fatal(err)
	}
	image, err := random.Image(512, 1)
	if err != nil {
		t.Fatal(err)
	}
	config, err := image.ConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	config.OS = "linux"
	config.Architecture = "amd64"
	image, err = mutate.ConfigFile(image, config)
	if err != nil {
		t.Fatal(err)
	}
	transport := server.Client().Transport
	if err := remote.Write(reference, image, remote.WithTransport(transport)); err != nil {
		t.Fatal(err)
	}

	indexDigest, platformDigest, err := (RemoteResolver{
		Options: []remote.Option{remote.WithTransport(transport)},
	}).Resolve(context.Background(), source, lock.Platform{OS: "linux", Architecture: "amd64"})
	if err != nil {
		t.Fatal(err)
	}
	expected, err := image.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if indexDigest != expected.String() || platformDigest != expected.String() {
		t.Fatalf("digests = %s, %s", indexDigest, platformDigest)
	}
}

func writeContainerMetadata(t *testing.T, root, name string) {
	t.Helper()
	directory := filepath.Join(root, name)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	contents := `{"service_name":"` + name + `","artifact_type":"container"}`
	if err := os.WriteFile(filepath.Join(directory, "artifact-metadata.json"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
