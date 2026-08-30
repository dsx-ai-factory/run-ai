package lock

import (
	"strings"
	"testing"
)

func TestImageLockRoundTrip(t *testing.T) {
	value := ImageLock{
		APIVersion: APIVersion,
		Kind:       ImageKind,
		Metadata:   Metadata{Name: "runai", Version: "3.0.0"},
		Spec: ImageLockSpec{
			Profile:  "standard",
			Platform: Platform{OS: "linux", Architecture: "amd64"},
			Images: []Image{
				{
					Name:        "service",
					Image:       "registry.example.com/service@sha256:" + strings.Repeat("1", 64),
					Source:      "registry.example.com/service:3.0.0",
					IndexDigest: "sha256:" + strings.Repeat("2", 64),
				},
				{
					Name:        "postgres",
					Image:       "docker.io/postgres@sha256:" + strings.Repeat("3", 64),
					Source:      "docker.io/postgres:16-alpine",
					IndexDigest: "sha256:" + strings.Repeat("4", 64),
				},
			},
		},
	}
	contents, err := Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseImageLock(contents)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Spec.Images[0] != value.Spec.Images[0] {
		t.Fatalf("parsed image = %#v, want %#v", parsed.Spec.Images[0], value.Spec.Images[0])
	}
	if parsed.Spec.Images[1] != value.Spec.Images[1] {
		t.Fatalf("parsed third-party image = %#v, want %#v", parsed.Spec.Images[1], value.Spec.Images[1])
	}
}

func TestParseImageLockRejectsUnknownFields(t *testing.T) {
	contents := []byte(`apiVersion: artifacts.run.ai/v1alpha1
kind: ImageLock
metadata:
  name: runai
  version: 3.0.0
unexpected: true
spec:
  profile: standard
  platform:
    os: linux
    architecture: amd64
  images: []
`)
	if _, err := ParseImageLock(contents); err == nil || !strings.Contains(err.Error(), "field unexpected not found") {
		t.Fatalf("error = %v", err)
	}
}

func TestArtifactLockValidatesInstallerSchema(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	value := ArtifactLock{
		APIVersion: APIVersion,
		Kind:       ArtifactKind,
		Metadata:   Metadata{Name: "runai", Version: "3.0.0"},
		Spec: ArtifactLockSpec{
			Profile:  "standard",
			Platform: Platform{OS: "linux", Architecture: "amd64"},
			ImageLocks: []CollectedImageLock{{
				Name:    "runai",
				Version: "3.0.0",
				Digest:  digest,
				Images: []Image{{
					Name:        "service",
					Image:       "registry.example.com/service@" + digest,
					Source:      "registry.example.com/service:3.0.0",
					IndexDigest: digest,
				}},
			}},
			Charts: []Chart{{
				Name:       "control-plane",
				Version:    "3.0.0",
				Source:     "https://charts.example.com/control-plane-3.0.0.tgz",
				BundlePath: "charts/control-plane-3.0.0.tgz",
				Digest:     digest,
				Dependencies: []ChartDependency{{
					Name:         "dependency",
					Version:      "1.0.0",
					Repository:   "https://charts.example.com",
					EmbeddedPath: "charts/dependency-1.0.0.tgz",
					Digest:       digest,
				}},
			}},
			CLIs: []CLI{{
				Name:         "runai-installer",
				Version:      "3.0.0",
				Source:       "https://api.example.com/runai-installer.tgz",
				SourceDigest: digest,
				SourcePath:   "linux/amd64/runai-installer",
				BundlePath:   "cli/runai-installer",
				Digest:       digest,
			}},
		},
	}
	contents, err := Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseArtifactLock(contents); err != nil {
		t.Fatal(err)
	}
}

func TestArtifactLockRejectsLegacyCLIWithoutSourceDigest(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	value := ArtifactLock{
		APIVersion: APIVersion,
		Kind:       ArtifactKind,
		Metadata:   Metadata{Name: "runai", Version: "3.0.0"},
		Spec: ArtifactLockSpec{
			Profile:  "standard",
			Platform: Platform{OS: "linux", Architecture: "amd64"},
			CLIs: []CLI{{
				Name:       "runai-installer",
				Version:    "3.0.0",
				Source:     "https://api.example.com/runai-installer.tgz",
				SourcePath: "linux/amd64/runai-installer",
				BundlePath: "cli/runai-installer",
				Digest:     digest,
			}},
		},
	}
	if err := value.Validate(); err == nil {
		t.Fatal("expected source digest validation error")
	}
}
