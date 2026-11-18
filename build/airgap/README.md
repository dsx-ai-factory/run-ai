# Air-gap lock generators

The air-gap module provides the Run:ai release pipeline with two Go command-line
generators: one resolves published container images into platform-specific
`ImageLock` files, and the other combines image locks, Helm charts, and optional
CLI artifacts into an `ArtifactLock` consumed by `installer airgap pack`. It is
repository build tooling rather than a deployed service; its
[`artifact-metadata.json`](artifact-metadata.json) does not declare an owning
team.

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)

## Overview

The module is responsible for:

- **Container discovery** — walking configured repository roots for
  `artifact-metadata.json` files and selecting enabled `container` artifacts
  that do not set `skip_airgap`.
- **Platform image resolution** — resolving a published multi-architecture
  image tag through the local Docker credential configuration, selecting
  exactly one manifest for `linux/amd64` or `linux/arm64`, and recording both
  the index digest and immutable platform digest.
- **Third-party image inventory** — resolving component-specific public images
  alongside Run:ai-owned entries in each `ImageLock`.
- **Artifact aggregation** — validating and combining Run:ai, KRM, KAI, and
  Karta image locks. KAI Scheduler is packaged as a dependency of KRM but still
  contributes its own lock, since each lock covers one chart's own inventory.
  Every input must use the same profile and platform, lock names must be
  unique, and conflicting digests for the same image source are rejected.
- **Chart locking** — downloading the release's control-plane and cluster Helm
  charts, verifying their versions, and hashing the root archives and all
  packaged or unpacked chart dependencies.
- **Optional CLI locking** — discovering enabled `cli` metadata, downloading
  matching NGC resources, extracting the binary for the lock's platform, and
  recording both archive and binary digests.
- **Lock schemas and validation** — defining strict `artifacts.run.ai/v1alpha1`
  `ImageLock` and `ArtifactLock` YAML schemas in `pkg/lock`, including
  validation for metadata, profiles, platforms, digests, and bundle paths.

Generation is fail-closed: image-lock output is not written when any discovered
image cannot be resolved, and artifact-lock output is not written when an input
lock, chart, dependency, or CLI artifact is inconsistent.

## Repository layout

```text
.
├── cmd/
│   ├── image-lock/          # ImageLock generator entry point
│   └── artifact-lock/       # ArtifactLock generator entry point
├── internal/
│   ├── artifactlock/        # Image-lock collection plus chart and CLI locking
│   ├── imagelock/           # Metadata discovery and registry digest resolution
│   ├── metadata/            # artifact-metadata.json discovery and filtering
│   └── thirdparty/          # Public image extraction from chart values
├── pkg/
│   └── lock/                # Public lock schemas, YAML codec, digesting, validation
├── artifact-metadata.json   # Module artifact identity
├── Makefile                 # Build, generation, test, lint, and validation targets
├── go.mod                   # Go module and dependency versions
└── go.sum
```

There is no `Dockerfile`, OpenAPI specification, integration-test directory, or
generated source tree. Both generators run directly as Go commands.

## Building and running

The module is built with Go 1.26.3 as declared in [`go.mod`](go.mod). Its
[`Makefile`](Makefile) is self-contained and does not include the repository's
shared service makefiles.

| Target | Description |
|---|---|
| `make build` | Compile both command packages with `go build ./cmd/...`. |
| `make image-lock` | Generate control-plane and cluster locks for amd64 and arm64. |
| `make image-lock-control-plane` | Generate control-plane locks for amd64 and arm64. |
| `make image-lock-cluster` | Generate cluster locks for amd64 and arm64. |
| `make image-lock-amd64` | Generate separate control-plane and cluster locks for amd64. |
| `make image-lock-arm64` | Generate separate control-plane and cluster locks for arm64. |
| `make image-lock-control-plane-amd64` | Generate only the control-plane amd64 lock. |
| `make image-lock-control-plane-arm64` | Generate only the control-plane arm64 lock. |
| `make image-lock-cluster-amd64` | Generate only the cluster amd64 lock. |
| `make image-lock-cluster-arm64` | Generate only the cluster arm64 lock. |
| `make artifact-lock` | Validate the supplied image locks, pull and hash release charts, optionally resolve CLI artifacts, and write an `ArtifactLock` under `dist/artifact-locks`. |

Prerequisites are Go, Helm for artifact-lock generation, registry credentials
available through the Docker credential configuration for image resolution, and
an NGC API key for published charts and optional CLI resources.

### Local build (local-dev.sh)

[`local-dev.sh`](../../local-dev.sh) only builds deployable control-plane and
runai-cluster images. It does not list or invoke `build/airgap`, so there is no
local-dev build command, container image, or fixed Helm image tag for this
module. Build it directly from the repository root:

```bash
make -C build/airgap build
```

This compiles for the current host; the `linux/amd64` and `linux/arm64` variants
below describe the platform recorded in generated locks, not the architecture
of a container built from this module.

### Generate Run:ai image locks

```bash
make -C build/airgap image-lock \
  VERSION=3.0.0 \
  REPOSITORY=nvcr.io/nvidia/runai
```

The combined target writes:

```text
dist/image-locks/runai-control-plane-3.0.0-linux-amd64.yaml
dist/image-locks/runai-control-plane-3.0.0-linux-arm64.yaml
dist/image-locks/runai-cluster-3.0.0-linux-amd64.yaml
dist/image-locks/runai-cluster-3.0.0-linux-arm64.yaml
```

Use `image-lock-control-plane` or `image-lock-cluster` to generate one component
for both architectures. Add the architecture suffix to generate one component
and one architecture. `image-lock-amd64` and `image-lock-arm64` generate both
components for one architecture.

Control-plane and cluster accept independent versions and registries:

```bash
make -C build/airgap image-lock-amd64 \
  CONTROL_PLANE_VERSION=3.0.0 \
  CONTROL_PLANE_REPOSITORY=nvcr.io/nvidia/runai \
  CLUSTER_VERSION=3.0.0 \
  CLUSTER_REPOSITORY=nvcr.io/nvidia/runai
```

`VERSION` and `REPOSITORY` are shared defaults. The generator does not discover
versions or repositories; callers supply concrete values globally or per
component. Resolution attempts every discovered image and emits no lock if any
image fails.

### Third-party images

Generated Run:ai image locks contain Run:ai-owned and public dependencies in
the existing `images` list without changing the installer-compatible lock
schema. Each chart declares its inventory under `thirdPartyImages`, and the
generator reads the chart supplied with `--chart`.

Entries can define an image inline or point to the values path used by a
subchart. Every image must define a repository and tag.

### Generate an artifact lock

```bash
export NGC_API_KEY="<api-key>"

make -C build/airgap artifact-lock \
  VERSION=3.0.0 \
  ARTIFACT_LOCK_IMAGE_LOCKS="\
dist/image-locks/runai-control-plane-3.0.0-linux-amd64.yaml \
dist/image-locks/runai-cluster-3.0.0-linux-amd64.yaml \
dist/image-locks/kai-resource-management-v0.0.3-linux-amd64.yaml \
dist/image-locks/kai-scheduler-v0.16.9-linux-amd64.yaml \
dist/image-locks/karta-0.2.6-linux-amd64.yaml"
```

This example writes:

```text
dist/artifact-locks/runai-3.0.0-standard-linux-amd64.yaml
```

The artifact lock version is independent from its chart versions. For branch
bundles or components released at different versions, supply chart sources and
versions in matching order:

```bash
make -C build/airgap artifact-lock \
  VERSION=master \
  ARTIFACT_LOCK_IMAGE_LOCKS="/absolute/path/cp.yaml /absolute/path/cluster.yaml" \
  ARTIFACT_LOCK_CHART_URLS="$CONTROL_PLANE_CHART_URL $CLUSTER_CHART_URL" \
  ARTIFACT_LOCK_CHART_VERSIONS="$CONTROL_PLANE_VERSION $CLUSTER_VERSION"
```

Every `ARTIFACT_LOCK_CHART_URLS` entry must have a corresponding
`ARTIFACT_LOCK_CHART_VERSIONS` entry. The generator verifies each downloaded
chart against its supplied version.

Set `ARTIFACT_LOCK_CLI_ROOTS=../../runai-installer` when the matching installer
resource has been published to NGC. Root chart digests cover the downloaded
archives; dependency digests cover either the packaged archive or the
deterministic contents of an unpacked dependency directory.

## Testing

Tests are local Go unit tests and use fakes for remote registry, Helm, and HTTP
operations; no docker-compose backing services are present.

| Target | Description |
|---|---|
| `make test` | Run all module tests with `go test ./...`. |
| `make lint` | Run `go vet ./...`, then fail if `gofmt -l .` reports any file. |
| `make validate` | Verify that `go.mod` and `go.sum` are already tidy with `go mod tidy -diff`. |

Run the complete local verification set from the repository root:

```bash
make -C build/airgap test
make -C build/airgap lint
make -C build/airgap validate
```

## Configuration

The generators have no YAML configuration file. The Makefile maps the following
variables to command flags and output paths:

| Variable | Description | Default |
|---|---|---|
| `GO` | Go executable used for build, run, test, vet, formatting, and module validation. | `go` |
| `VERSION` | Shared default for control-plane and cluster versions. | `0.0.0` |
| `REPOSITORY` | Shared default for control-plane and cluster image repositories. | `nvcr.io/nvidia/runai-oss` |
| `CONTROL_PLANE_VERSION` | Version used for control-plane images. | `VERSION` |
| `CLUSTER_VERSION` | Version used for cluster images. | `VERSION` |
| `CONTROL_PLANE_REPOSITORY` | Repository prefix for control-plane images. | `REPOSITORY` |
| `CLUSTER_REPOSITORY` | Repository prefix for cluster images. | `REPOSITORY` |
| `MAX_WORKERS` | Maximum concurrent image resolutions. | `10` |
| `REPOSITORY_ROOT` | Repository root searched for artifact metadata. | `../..` |
| `IMAGE_LOCK_OUTPUT_DIR` | Destination for generated `ImageLock` files. | `../../dist/image-locks` |
| `ARTIFACT_LOCK_OUTPUT_DIR` | Destination for generated `ArtifactLock` files. | `../../dist/artifact-locks` |
| `ARTIFACT_LOCK_IMAGE_LOCKS` | Space-separated input lock paths; required by `make artifact-lock`. | Empty |
| `NGC_HELM_BASE_URL` | Base URL for published Run:ai charts. | `https://helm.ngc.nvidia.com/nvidia/runai-oss/charts` |
| `ARTIFACT_LOCK_CHART_URLS` | Space-separated root chart URLs. | Control-plane and runai-cluster charts for `VERSION` |
| `ARTIFACT_LOCK_CHART_VERSIONS` | Space-separated chart versions corresponding by position to `ARTIFACT_LOCK_CHART_URLS`. | `VERSION VERSION` |
| `ARTIFACT_LOCK_CLI_ROOTS` | Space-separated roots searched for optional `cli` metadata. | Empty |
| `ARTIFACT_LOCK_HELM_USERNAME` | Username passed to `helm pull`. | `$oauthtoken` |
| `ARTIFACT_LOCK_HELM_PASSWORD_ENV` | Environment variable containing the Helm repository password. | `NGC_API_KEY` |
| `NGC_RESOURCE_BASE_URL` | Base URL used to download optional NGC CLI resources. | `https://api.ngc.nvidia.com/v2/org/nvidia/team/runai/resources` |

The Makefile requires the environment variable selected by
`ARTIFACT_LOCK_HELM_PASSWORD_ENV`. With defaults:

| Variable | Description | Default |
|---|---|---|
| `NGC_API_KEY` | Password for `helm pull` and bearer token for optional NGC CLI downloads; required by `make artifact-lock`. | — |

The selected password is also used as the bearer token for optional CLI
downloads.
