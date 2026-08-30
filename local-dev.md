# local-dev.sh — Local Development Guide

Builds all Run.ai services from source, creates a [Kind](https://kind.sigs.k8s.io/) cluster, loads images directly into the cluster via `kind load docker-image`, and deploys the full stack with a self-signed TLS certificate. Everything runs locally — no external registry, no cloud cluster.

---

## Prerequisites

| Requirement | Notes |
|---|---|
| **Docker Desktop** | Must be running. On Linux, enable the containerd image store. |
| **Podman VM ≥ 12 GB RAM** | macOS only. The full stack (NATS×3, PostgreSQL, Keycloak, ~50 services) requires at least 12 GB. With 8 GB the API server starves under QEMU overhead and the cluster never stabilises. Resize with `podman machine stop && podman machine set --memory 12288 && podman machine start`. |
| **`GITHUB_TOKEN`** | Only needed while the repo is private — Go builds pull modules from `github.com/run-ai/*`. Not required with `--skip-build` or once the repo is public. |
| `kind`, `kubectl`, `helm` | Installed automatically if missing. |
| `openssl` | For TLS cert generation. Ships with macOS and most Linux distros. |

```bash
export GITHUB_TOKEN=ghp_...
./local-dev.sh
```

---

## Quick Start

```bash
# Full setup: build everything, create cluster, deploy
./local-dev.sh

# Add the domain to /etc/hosts
sudo sh -c 'echo "127.0.0.1 runai.local" >> /etc/hosts'

# Open https://runai.local  (admin@run.ai / Admin@1234!)
```

First run takes a while — it compiles ~50 Go services inside Docker containers. Subsequent runs with `--skip-build` are much faster.

---

## What It Does

The script runs these phases in order. Each phase is independently skippable.

### 1. Prerequisites check
Detects OS/arch (Linux/Darwin × amd64/arm64), validates `GITHUB_TOKEN`, and auto-installs `kind`/`kubectl`/`helm` if they're not on `$PATH`.

### 2. TLS certificate + `/etc/hosts`
Generates a local CA and a server cert for the configured domain (default `runai.local`) using openssl. Cert files are written to `.local-dev/certs/` and reused on subsequent runs.

This runs **before** the cluster is created — the CA cert is mounted into the Kind cluster as the API server's OIDC `oidc-ca-file`, which has to exist at cluster-creation time. The script also adds the domain to `/etc/hosts` here.

### 3. Kind cluster
Creates a 5-node Kind cluster (1 control-plane + 4 workers) named `runai-local` with host ports 80/443 mapped for ingress.

### 4. Local registry
Starts a `registry:2` container named `kind-registry` on port `5001` and connects it to the Kind network, so cluster nodes can resolve it by name. All built images are pushed here; charts are pointed at `localhost:5001/runai-local`.

### 5. ingress-nginx
Installs the Kind-specific ingress-nginx manifest and waits for the controller pod to be ready.

### 6. prometheus-operator
Installs `kube-prometheus-stack` into the `monitoring` namespace. Skip with `--skip-prometheus`.

### 7. fake-gpu-operator
Installs `oci://ghcr.io/run-ai/fake-gpu-operator` into `gpu-operator`, configured for 8 × `Tesla-A10` in the default node pool, so GPU workloads can be scheduled without real hardware. Skip with `--skip-fake-gpu`.

### 8. Build — control-plane (25 Go services + frontend)
For each Go service: compiles an `amd64` binary inside a golang Docker container (workspace-mounted, uses the Go module cache at `~/.cache/go-build-docker-*`), builds a `linux/amd64` image, and **pushes it to the local registry**. Up to `--parallel N` services build concurrently.

The frontend (React → nginx) runs `npm run build` inside a Node container, then builds and pushes the nginx image.

Build logs land in `.local-dev/logs/<service>.log`.

### 9. Build — runai-cluster (25 Go services + crd-upgrader)
Same pattern for the cluster-side controllers and operators. `runai-crd-upgrader` is special — no Go source, just packages `kubectl` + CRD manifests into a scratch image.

### 10. Deploy — control-plane
- Creates namespace `runai` and loads the TLS cert as a Kubernetes secret
- Patches `control-plane-chart/Chart.yaml` `appVersion` to match the built image tag
- Runs `helm dep build` (uses pre-committed subchart tarballs in `charts/`)
- Installs the chart with `global.image.registry=localhost:5001/runai-local` and `imagePullPolicy=IfNotPresent`

### 11. Deploy — runai-cluster
Installs the cluster chart with `global.image.registry=localhost:5001/runai-local`. `imagePullPolicy=IfNotPresent` keeps Kubernetes from reaching out to an external registry.

Scheduling subcharts are set as follows:

| Subchart | Setting | Notes |
|---|---|---|
| `runai-engine` | `enabled=false` | Legacy engine — superseded by KAI Scheduler and no longer a chart dependency |
| `kai-scheduler` | `enabled=true` | KAI Scheduler |
| `karta` | `enabled=true` | |

The script also applies the kai-scheduler CRDs (`podgroups`, `queues`, `schedulingshards`, `configs`, `topologies`) directly from the Go module cache, because the `crd-upgrader` job does not manage them. The module version is pinned in `runai-cluster/cluster/go-operator/go.mod`. If that module isn't in the cache you'll see `kai-scheduler Go module not found; CRDs may be missing` — run `go mod download` in `go-operator` and re-deploy.

Enabling `kai-scheduler` is also what causes the `runai-scale-adjust` namespace to be created; the script deletes it first if a previous failed install left it without Helm ownership labels.

Before installing, the script reconciles `charts/` against `Chart.yaml` — Helm needs a subchart present for every declared dependency, and `charts/*.tgz` is gitignored, so a fresh clone has none:

1. **Prune** — deletes any tarball that no longer matches a declared dependency. Helm loads *every* tarball in `charts/`, so a leftover from a previous version bump would make the resolved subchart ambiguous, and a tarball for a removed dependency would render with no condition to gate it.
2. **Fetch** — runs `helm dependency build` to pull the real subcharts. Both come from public OCI registries (`ghcr.io/kai-scheduler`, `ghcr.io/run-ai/karta`) and need no credentials.
3. **Fall back** — if the fetch fails (offline, registry down), generates a minimal placeholder so the install can still proceed, and warns that the subchart will be empty.

Names and versions are read from `Chart.yaml`, so bumping a dependency can't leave the script stale.

> **If you see the placeholder warning,** `kai-scheduler` and `karta` will be enabled but deploy nothing — Helm matches dependencies by name only, so a placeholder satisfies the check and the subchart installs empty. Re-run once you have network access, or fetch manually with `helm dependency build runai-cluster/cluster/runai-chart`. See [Known Issues](#known-issues).

---

## All Options

```
./local-dev.sh [OPTIONS]
```

| Flag | Description |
|---|---|
| `--domain FQDN` | Domain for ingress and TLS (default: `runai.local`) |
| `--admin-email EMAIL` | Initial admin account email (default: `test@run.ai`) |
| `--admin-password PASS` | Initial admin password — min 8 chars, 1 digit, 1 upper, 1 lower, 1 special |
| `--parallel N` | Concurrent build jobs (default: `4`). Set higher on machines with many cores. |
| `--skip-build` | Skip all image builds — reuse whatever is already in the Kind cluster. |
| `--skip-build-cp` | Skip control-plane builds only. |
| `--skip-build-rc` | Skip runai-cluster builds only. |
| `--skip-frontend` | Skip the frontend npm build (saves ~5 min). Useful when only working on backend services. |
| `--skip-cluster` | Skip Kind cluster creation — reuse an existing `runai-local` cluster. |
| `--skip-deploy` | Skip Helm installs entirely (build images and set up cluster only). |
| `--skip-deploy-cp` | Skip deploying the control-plane chart. |
| `--skip-deploy-rc` | Skip deploying the runai-cluster chart. |
| `--skip-prometheus` | Skip the prometheus-operator install. |
| `--skip-fake-gpu` | Skip the fake-gpu-operator install. Without it, GPU workloads cannot be scheduled. |
| `--force-rebuild` | Ignore the build cache — rebuild and re-push every image. |
| `--teardown` | Delete the Kind cluster and local registry, then exit. |
| `--help` | Print usage. |

### Environment variables

| Variable | Effect |
|---|---|
| `RUNAI_DOMAIN` | Same as `--domain` |
| `RUNAI_ADMIN_EMAIL` | Same as `--admin-email` |
| `RUNAI_ADMIN_PASSWORD` | Same as `--admin-password` |
| `PARALLEL_JOBS` | Same as `--parallel` |
| `KIND_CLUSTER` | Kind cluster name (default: `runai-local`) |
| `RUNAI_CLUSTER_UID` | Fixed cluster UUID for the runai-cluster chart (auto-generated if unset) |

---

## Common Workflows

### Rebuild a single service

```bash
cd control-plane/assets-service
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY=
docker buildx build --build-arg TARGETARCH=amd64 --platform linux/amd64 \
  --provenance=false --load -t runai-local/assets-service:0.0.0 .
kind load docker-image runai-local/assets-service:0.0.0 --name runai-local

helm upgrade runai-cp ./control-plane/control-plane-chart \
  -n runai --reuse-values
```

### Iterate on the cluster chart only (images already loaded)

```bash
./local-dev.sh --skip-build --skip-cluster --skip-deploy-cp
```

### Rebuild images without touching the cluster

```bash
./local-dev.sh --skip-cluster --skip-deploy
```

### Rebuild everything from scratch

```bash
./local-dev.sh --teardown
./local-dev.sh
```

### Change domain

```bash
./local-dev.sh --domain mydev.example.com
sudo sh -c 'echo "127.0.0.1 mydev.example.com" >> /etc/hosts'
```

### Speed up with more parallel jobs

```bash
PARALLEL_JOBS=8 ./local-dev.sh
```

---

## Connecting the Cluster

After the control plane is up, connect the cluster through the CP UI (`https://runai.local`) to get a `clientSecret` and `clusterUID`. Then upgrade the cluster chart:

```bash
helm upgrade runai-cluster ./runai-cluster/cluster/runai-chart \
  -n runai \
  --set controlPlane.clientSecret=<SECRET> \
  --set cluster.uid=<UID> \
  --reuse-values
```

---

## Trusting the Self-Signed Certificate

The CA cert is at `.local-dev/certs/ca.crt`.

**macOS:**
```bash
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain .local-dev/certs/ca.crt
```

**Linux:**
```bash
sudo cp .local-dev/certs/ca.crt /usr/local/share/ca-certificates/runai-local.crt
sudo update-ca-certificates
```

**Chrome/Firefox on Linux:** Import `.local-dev/certs/ca.crt` via `chrome://settings/certificates` or the Firefox certificate manager.

---

## Podman Disk Management

local-dev.sh uses Podman (not Docker Desktop) on macOS. The VM gets a fixed disk allocated at init time.

### Check VM disk usage

```bash
podman machine ssh "df -h /"
```

### Recover from a full disk

When the Podman VM disk fills up, builds fail with `no space left on device` and even `podman system prune` can't run (it needs space to write temp files). Fix:

```bash
# Step 1: free enough space for Podman to operate
podman machine ssh "sudo journalctl --vacuum-size=10M"

# Step 2: prune all unused images, containers, and volumes
podman system prune -af --volumes
```

### Recreate the VM with a larger disk

If disk usage after a full prune is already above ~80%, recreate the VM at a larger size. The default is 200 GB; 300 GB is a reasonable minimum for active development.

```bash
podman machine stop
podman machine rm
podman machine init --disk-size 300 --cpus 6 --memory 12288
podman machine start
podman machine stop
podman machine set --rootful   # required by kind on macOS
podman machine start
```

`--rootful` is mandatory — kind needs it to create containers with elevated privileges. local-dev.sh will error if the machine is in rootless mode.

---

## nginx Worker Crashes (`pthread_create failed`)

On macOS with Podman + Kind, the nginx ingress controller may log:

```
nginx: worker process exited with code 2 (fatal)
pthread_create failed (11: Resource temporarily unavailable)
```

**Root cause:** Kubernetes sets a `pids.max = 307` cgroup limit on each container. Each nginx worker spawns 32 aio threads (33 pids total). With the default `worker_processes 6`, a config reload runs old and new workers simultaneously: `6 × 33 × 2 = 396 pids > 307`. Workers fail to create their thread pools and crash.

**Fix (already applied by `local-dev.sh`):** `patch_nginx_workers()` caps workers at 2 immediately after installing ingress-nginx. During reload, peak pids = `2 × 33 × 2 = 132 < 307`. The patch is idempotent and survives pod restarts (the ConfigMap persists). If you're debugging a pre-existing cluster without this fix, apply it manually:

```bash
kubectl patch configmap ingress-nginx-controller -n ingress-nginx \
  --type merge \
  -p '{"data":{"worker-processes":"2","max-worker-open-files":"65536","reuse-port":"false"}}'
kubectl rollout restart deployment ingress-nginx-controller -n ingress-nginx
```

---

## Networking Troubleshooting

### SSL handshake timeouts / slow responses

If `curl https://<domain>/_version` times out at the SSL handshake stage (`time_appconnect: 0`) but TCP connects instantly, the likely cause is **Cisco AnyConnect** (or similar corporate VPN with a socket-level kernel extension). AnyConnect's `acsockext` system extension intercepts all TCP connections — including loopback — and routes them through a low-MTU inspection tunnel, which fragments TLS certificate exchanges.

To confirm: disconnect from AnyConnect and retry. If it responds instantly, AnyConnect is the cause.

**Fix:** Ask IT to add split-tunnel exclusions for `127.0.0.1/8` and `192.168.127.0/24` (the Podman VM network) in the AnyConnect profile. In the meantime, disconnect from VPN during local dev sessions.

---

## Build Logs and Debugging

All per-service build logs are in `.local-dev/logs/`:
```bash
# Watch a specific service build
tail -f .local-dev/logs/assets-service.log

# Find what failed
ls .local-dev/logs/*.log | xargs grep -l "Error\|FAIL"
```

If a Go build fails with a module download error and the repo is still private, check that `GITHUB_TOKEN` is set and that `git config --list` shows the `github.com` URL substitution.

If `docker buildx build` fails with `no such file: bin/service-amd64`, the `build-go-amd` step didn't produce a binary — check the log file for the compile error.

If a service log shows `unknown flag: --platform` followed by the top-level `docker` usage — right after a *successful* `build-go-amd` — the `docker buildx` plugin is missing, so the Docker CLI fell back to the root command. This hits every service, not just the one you happened to read. Docker Desktop bundles buildx; Ubuntu's `docker.io` package does not. Install `docker-buildx-plugin` from Docker's official repo and verify with `docker buildx version`.

If `kind load docker-image` is slow, note that images are loaded into all 5 nodes (1 control-plane + 4 workers). With many services this is expected — it only runs once per build.

---

## Image Tag Reference

| Chart | Tag | Image name |
|---|---|---|
| control-plane | `0.0.0` (matches `Chart.yaml appVersion`) | `runai-local/<service>:0.0.0` |
| runai-cluster operator/hooks | `0.0.0-dev` | `runai-local/<service>:0.0.0-dev` |

Images are pushed to the local `kind-registry` container and referenced as `localhost:5001/runai-local/<service>:<tag>`. Helm deploys with `imagePullPolicy=IfNotPresent`, so Kubernetes pulls from the local registry and never reaches an external one.

---

## Known Issues

### No kai-scheduler / karta pods after deploy

The script fetches these subcharts automatically, but if the fetch failed it falls back to an empty placeholder and the scheduler silently deploys nothing. Look for this in the output:

```
[warn]   Could not fetch kai-scheduler@v0.16.8 — installing an EMPTY placeholder.
```

Confirm by size — a real chart is tens of KB, a placeholder a few hundred bytes:

```bash
ls -l runai-cluster/cluster/runai-chart/charts/*.tgz
# kai-scheduler-v0.16.8.tgz   ~84 KB   real
# karta-0.2.1.tgz             ~13 KB   real
```

Fix by fetching manually, then redeploying just the chart:

```bash
helm dependency build runai-cluster/cluster/runai-chart
./local-dev.sh --skip-build --skip-cluster
```

With the real charts you should see `kai-operator`, `kai-admission`, `kai-binder`, `kai-podgrouper`, `kai-queuecontroller`, `kai-node-scale-adjuster` and `karta-operator`:

```bash
kubectl get pods -n runai | grep -E 'kai|karta'
```

Historically this was a hard failure rather than a fallback: the placeholder versions were hardcoded in `local-dev.sh` and drifted from `Chart.yaml` (which moved to kai-scheduler `v0.16.8` and karta `0.2.1`, and dropped `runai-engine` entirely), so every fresh clone got empty subcharts. The script now derives names and versions from `Chart.yaml`.

### Go build failures after an OSS re-export

**Applies only to OSS export repos** (those with a `.oss-export-id` file), not to
the monorepo itself. Three different-looking errors share one cause:

```
verifying <module>@v1.0.0: checksum mismatch
        downloaded: h1:z5IDz77...
        go.sum:     h1:htuQpzz...
SECURITY ERROR
```
```
generated/api_service_accounts.go:132:80: undefined: InstallerBootstrapResponse
```
```
gojsonq/runai.go:5:2: could not import math (open : no such file or directory)
```

Each export re-points every version tag at a fresh commit. Go assumes
`module@version` is immutable and caches it forever, so a machine that built a
previous export keeps serving that export's content. Nothing is actually wrong
with the repo — the second and third errors are especially misleading, since the
symbol and the stdlib package are both perfectly present.

`local-dev.sh` heals this automatically: it compares `.oss-export-id` against a
marker in the module cache and evicts the affected caches when they differ. You
should not normally have to do anything.

If you hit it anyway — an older checkout, or a cache cleared out from under the
marker — purge by hand. **All three locations matter**; clearing only one or two
leaves the build broken with no hint as to why:

Run from the repo root — it derives the module namespace itself, so there is
nothing to substitute:

```bash
GP=~/.cache/go-build-docker-gopath/pkg/mod
PREFIX=$(awk '/^module /{print $2; exit}' "$(ls runai-common-packages/*/go.mod | head -1)" | cut -d/ -f1-3)
echo "purging $PREFIX"

# 1. git clone cache — Go builds module zips from here, and it holds the OLD
#    tag -> sha map, so it serves stale content even when everything else is clean
for d in "$GP"/cache/vcs/*/; do
  git -C "$d" config --get-regexp '^remote\..*\.url$' 2>/dev/null | grep -qi "${PREFIX##*/}" \
    && { chmod -R u+w "$d" 2>/dev/null; rm -rf "$d"; }
done

# 2. downloaded zips + extracted module trees (mode 0444, hence the chmod)
chmod -R u+w "$GP/$PREFIX" "$GP/cache/download/$PREFIX" 2>/dev/null
rm -rf "$GP/$PREFIX" "$GP/cache/download/$PREFIX"

# 3. build cache — per-module file index, also keyed by module@version
rm -rf ~/.cache/go-build-docker-gocache
```

On `permission denied`, delete as container-root instead — build steps sometimes
write as root:

```bash
docker run --rm -v ~/.cache:/c busybox sh -c '
  rm -rf /c/go-build-docker-gocache /c/go-build-docker-gopath/pkg/mod/cache/vcs'
```

Expect the next build to be slow; every cache is cold.

Also make sure the checkout is current — a `go.sum` from an older export describes
content the tags no longer serve. `local-dev.sh` warns when you are behind, but it
will not move your HEAD:

```bash
git pull --ff-only
```

---

## Teardown

```bash
./local-dev.sh --teardown
```

Deletes the Kind cluster and the `kind-registry` container. Files in `.local-dev/` are left in place — remove manually if needed.
