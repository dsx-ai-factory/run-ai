# airgap — Build Context

**Type:** go
**Team:** Not declared in `artifact-metadata.json`
**Module:** `github.com/run-ai/runai/build/airgap`

## Make targets

The module `Makefile` is self-contained; it does not inherit targets from
`build/makefile/*.mk`.

| Target | What it does |
|---|---|
| `build` | Runs `go build ./cmd/...` to compile the `image-lock` and `artifact-lock` command packages. |
| `image-lock` | Generates separate control-plane and cluster locks for amd64 and arm64. |
| `image-lock-control-plane` | Generates control-plane locks for amd64 and arm64. |
| `image-lock-cluster` | Generates cluster locks for amd64 and arm64. |
| `image-lock-amd64` | Generates separate control-plane and cluster locks for amd64. |
| `image-lock-arm64` | Generates separate control-plane and cluster locks for arm64. |
| `image-lock-control-plane-amd64` | Generates only the control-plane amd64 lock. |
| `image-lock-control-plane-arm64` | Generates only the control-plane arm64 lock. |
| `image-lock-cluster-amd64` | Generates only the cluster amd64 lock. |
| `image-lock-cluster-arm64` | Generates only the cluster arm64 lock. |
| `artifact-lock` | Combines supplied image locks with downloaded Helm charts and optional CLI resources, then writes a platform/profile-specific lock under `dist/artifact-locks`. Requires `ARTIFACT_LOCK_IMAGE_LOCKS` and `NGC_API_KEY`. |
| `test` | Runs all Go tests with `go test ./...`. |
| `lint` | Runs `go vet ./...` and checks that `gofmt -l .` returns no files. |
| `validate` | Runs `go mod tidy -diff` to detect untidy module files without rewriting them. |

## Key build notes

- Go version: `go.mod` requires Go 1.26.3.
- Generated code: none. The generated products are YAML lock files under
  `dist/image-locks` and `dist/artifact-locks`, outside this module.
- Container build: none. There is no `Dockerfile`, runtime base image, Docker
  build argument, local image reference, or Helm-coupled image tag.
- Build command: `make -C build/airgap build` compiles both commands directly
  with the host Go toolchain.
- Platform handling: control-plane and cluster are always emitted as separate
  image locks. Component and architecture-specific targets can run alone or
  through aggregate targets. These targets do not cross-compile this module.
- Version handling: the generator uses caller-supplied concrete control-plane
  and cluster versions without remote version discovery.
- Third-party images: component-specific inventories live in
  `internal/thirdparty`. The generator resolves them separately and merges them
  into the existing installer-compatible `spec.images` list.
- Image credentials: registry resolution uses the Docker credential
  configuration available to the process. Repository prefixes are supplied
  globally or independently for control-plane and cluster.
- Artifact credentials: `make artifact-lock` reads the environment variable
  selected by `ARTIFACT_LOCK_HELM_PASSWORD_ENV` and passes
  `ARTIFACT_LOCK_HELM_USERNAME` to Helm. Defaults remain `$oauthtoken` and
  `NGC_API_KEY`.
- Chart versions: artifact-lock metadata version, chart sources, and chart
  versions are independent caller inputs. Chart source and version lists are
  positional and must have equal lengths.
- Release invariants: input image locks must have one shared profile and
  platform; duplicate names, conflicting image digests, chart-version
  mismatches, and missing or unexpected embedded dependencies fail generation.
- Tests: unit tests use fakes and local fixtures; there is no integration-test
  or docker-compose suite.

## local-dev.sh iteration

`local-dev.sh` builds deployable control-plane and runai-cluster container
images only. `build/airgap` is not in either service list and is never invoked
or tagged by that script. Iterate directly:

```bash
make -C build/airgap build
make -C build/airgap test
make -C build/airgap lint
make -C build/airgap validate
```
