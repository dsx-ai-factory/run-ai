# local-dev.sh — Local Development Guide

local-dev.sh spins up a complete, working NVIDIA Run:ai platform on your machine — control plane, cluster services, UI and Command-line interface (CLI). As such, you can develop and test against the real product instead of individual services in isolation. The script builds ~50 Go services from source, creates a self-contained [Kind](https://kind.sigs.k8s.io/) cluster, loads images directly into it via `kind load docker-image`, and deploys the full stack behind a self-signed TLS certificate. Everything runs locally — no external registry, no cloud cluster.

---

## Prerequisites

| Requirement | Notes |
|---|---|
| **OS / Architecture** | `macOS` or `Linux`, `amd64` or `arm64` — the only four supported combinations. Any other platform, such as Windows, stops the script with a clear error before setup begins. |
| **Container Runtime** | Use docker or podman. On Linux, enable the containerd image store. See more info on Podman settings for MacOS below. |
| **Linux host: ≥ 4 CPUs, ≥ 16 GB RAM, ≥ 100 GB free disk** | Linux only. Kind nodes, ~50 services, and parallel builds all share the host directly, with no VM to size. Disk is measured where the container runtime stores data (usually `/var/lib/docker`). One full build leaves about 44 GB of caches: roughly 11 GB of Docker build cache, and 33 GB of Go module and build caches in `~/.cache`. It also needs about 8 GB of images, which are stored both in the registry and in the Kind nodes. The script warns at startup if the host is below any of these; it does not stop. On a tight host, use `--parallel 2` and `--skip-frontend`. |
| **Podman VM ≥ 12 GB RAM** | macOS only. The full stack (NATS×3, PostgreSQL, Keycloak, ~50 services) requires at least 12 GB. With 8 GB the API server starves under QEMU overhead and the cluster never stabilises. Resize with `podman machine stop && podman machine set --memory 12288 && podman machine start`. |
| **`GITHUB_TOKEN`** | Only needed while the repo is private — Go builds pull modules from `github.com/run-ai/*`. Not required with `--skip-build` or once the repo is public. Set it before running: `export GITHUB_TOKEN=ghp_...` |
| **`kind`, `kubectl`, `helm`** | Installed automatically if missing. `kind` is version-pinned (`v0.33.0`, node image `v1.37.0`) and upgraded if an older version is installed; `kubectl` and `helm` always install the latest release. Kubernetes 1.37 needs a cgroup v2 host (current Docker Desktop, Podman machine, and Linux distributions all use it). |
| **`openssl`** | For TLS cert generation. Ships with macOS and most Linux distros; not auto-installed if missing. |
| **`make`, `git`, `curl`, `python3`** | Used directly by the script and the service Makefiles. Not auto-installed; the script stops at startup if any is missing. |

Go, Node, and Java are **not** needed on the host. Every Go, frontend, and Gradle build runs inside a container (`golang`, `node`, `gradle` images).

---

## Quick Start

1. **Run the full setup.** Builds and deploys both the control plane and the cluster services — the complete Run:ai platform, end to end.

   ```bash
   ./local-dev.sh
   ```

2. **The domain is added to `/etc/hosts` automatically.** `runai.local` does not exist on the real internet, so the script resolves it to `127.0.0.1` (itself) for you, prompting for `sudo` if needed. To do it yourself instead:

   ```bash
   sudo sh -c 'echo "127.0.0.1 runai.local" >> /etc/hosts'
   ```

3. **Open `https://runai.local`** and log in with the credentials printed at the end of the run.

First run takes a while — it compiles ~50 Go services inside Docker containers. Subsequent runs with `--skip-build` are much faster.

---

## How It Works

The script runs these phases in order. Each phase is independently skippable.

### Prerequisites Check

- Detects OS/arch (`Darwin`/`Linux` × `amd64`/`arm64`), and stops with a clear error before any setup begins if your platform is not supported — for example, Windows.
- Detects your container runtime — Docker Desktop, Podman, or Rancher Desktop — and auto-creates a `docker` shim pointing at Podman if only `podman` is found.
- Confirms the container daemon is actually running, and on macOS with Podman, auto-switches the VM to rootful mode if needed (Kind requires it).
- Validates `GITHUB_TOKEN` — warns, but does not fail, if it is unset and a build is about to run.
- Auto-installs `kind`, `kubectl`, and `helm` if they are not already on `$PATH`.
- Stops if `make`, `git`, `curl`, `python3`, or `openssl` is missing.
- On Linux, warns if the host has fewer than 4 CPUs, less than 16 GB RAM, or less than 100 GB free for container storage.
- With `--gpu-operator`, checks the NVIDIA host setup, and configures Docker's `nvidia` runtime if needed. See [Real GPUs](#real-gpus---gpu-operator).

### TLS Certificate + `/etc/hosts`

- Generates a local CA (RSA 4096) and a server certificate (RSA 2048) for the configured domain (default `runai.local`) using `openssl`.
- The server cert covers a wildcard SAN (`runai.local`, `*.runai.local`, `127.0.0.1`), not just the bare domain.
- Cert files are written to `.local-dev/certs/` and reused on subsequent runs — the script checks whether the existing cert already covers the current `--domain` value, and regenerates from scratch if you have since switched domains.
- This runs **before** the cluster is created — the CA cert is mounted into the Kind cluster as the API server's OIDC `oidc-ca-file`, which has to exist at cluster-creation time.
- The script also adds the domain to `/etc/hosts` here.

### Kind Cluster

- Creates a 5-node Kind cluster (1 control-plane + 4 workers) named `runai-local`, pinned to Kubernetes `v1.37.0`.
- The control-plane node maps host ports 80/443 for ingress, and is labeled `ingress-ready=true`.
- containerd on each node is pointed at `/etc/containerd/certs.d` for registry settings. In the next phase, each node gets a `hosts.toml` there that sends `localhost:5001` pulls to the local registry container.

### Local Registry

- Starts a `registry:2` container named `kind-registry` on port `5001`, connected to the Kind network so cluster nodes can resolve and pull from it directly (`http://kind-registry:5000` internally) — while your host machine pushes to `localhost:5001` (the forwarded port).
- All built images are pushed here; charts are pointed at `localhost:5001/runai-local`.
- On Podman, the script also registers `localhost:5001` as an insecure (non-TLS) registry inside the Podman VM — Docker Desktop allows insecure localhost registries by default, Podman does not.

### ingress-nginx

- Installs the Kind-specific ingress-nginx manifest (pinned to `controller-v1.12.3`) and waits up to 180s for the controller pod to be ready.
- Immediately patches the controller to cap `worker-processes` at 2 — necessary to avoid a container pids-limit crash under Kind's default cgroup settings. See [nginx Worker Crashes](#nginx-worker-crashes-pthread_create-failed) for the full explanation.

### prometheus-operator
Installs `kube-prometheus-stack` into the `monitoring` namespace — a minimal install: Grafana, Alertmanager, and the node exporter are explicitly disabled; only `kubeStateMetrics` and a bare Prometheus instance are enabled. Skip with `--skip-prometheus`.

### fake-gpu-operator

- Installs `oci://ghcr.io/run-ai/fake-gpu-operator` into `gpu-operator`, configured for 8 × `Tesla-A10` in the default node pool, so GPU workloads can be scheduled without real hardware. Wired to the Prometheus instance installed in the previous phase.
- Labels every worker node with `run.ai/simulated-gpu-node-pool=default` — required for the chart's node-pool topology to recognize which nodes should simulate GPUs.
- Skip with `--skip-fake-gpu`. Replaced by the real NVIDIA GPU Operator when `--gpu-operator` is set — see [Real GPUs](#real-gpus---gpu-operator).

### Build — control-plane (25 Go Services + Frontend)
For each Go service, the script compiles an `amd64` binary inside a golang Docker container (workspace-mounted, uses the Go module cache at `~/.cache/go-build-docker-*`), builds a `linux/amd64` image, and **pushes it to the local registry**. Up to `--parallel N` services build concurrently. When the module cache is empty, the script first pre-warms it with `go mod download` in every Go module, `N` modules at a time, and logs how long that took.

The frontend (React → nginx) runs `npm run build` inside a Node container, then builds and pushes the nginx image.

Build logs land in `.local-dev/logs/<service>.log`.

### Build — runai-cluster (25 Go Services + crd-upgrader)
The same build pattern applies to the cluster-side controllers and operators. `runai-crd-upgrader` is the exception — no Go source, just packages `kubectl` + CRD manifests into a scratch image.

### Deploy — control-plane
- Creates namespace `runai-backend` and loads the TLS cert as a Kubernetes secret.
- Patches the `appVersion` field in `control-plane-chart/Chart.yaml` to match the built image tag.
- Runs `helm dep build` (uses pre-committed subchart tarballs in `charts/`).
- Installs the chart with `global.image.registry=localhost:5001/runai-local` and `imagePullPolicy=IfNotPresent` — the latter keeps Kubernetes from reaching out to an external registry.

### Provision Cluster Credentials

- Gets an API token for the admin user: the Keycloak master admin (password from `runai-backend-gen-keycloakx`) impersonates the user and completes a browser login (authorization code + PKCE) with the `runai-cli` or `runai-admin-ui` client.
- Waits for the Keycloak realm to be ready, then registers the cluster with the control plane via its REST API (reusing the existing registration if one already exists for the cluster name).
- Mints a one-time installer bootstrap token for the cluster from `/api/v1/clusters/{uuid}/cluster-install-info`.
- Feeds both the cluster UID and the bootstrap token into the `runai-cluster` Helm install in the next phase (`controlPlane.bootstrapToken`, `cluster.uid`).

### Deploy — runai-cluster

Installs the cluster chart with `global.image.registry=localhost:5001/runai-local`. `imagePullPolicy=IfNotPresent` keeps Kubernetes from reaching out to an external registry.

The runai-cluster chart configures these scheduling subcharts:

| Subchart | Setting | Notes |
|---|---|---|
| `kai-resource-management` | `enabled=true` | KAI Resource Management — supplies the nodepool-controller, project-controller and pod-group-assigner, and bundles `kai-scheduler` as its own dependency. |
| `karta` | `enabled=true` | CRD-based resource abstraction layer with a standard schema for describing any Kubernetes workload type, so controllers can extract pod specs and status without per-type logic. |

#### KRM and kai-scheduler CRDs

The script applies nothing by hand. Both charts ship their CRDs in their own `crds/` directory, which Helm applies for subcharts, and each carries a crd-upgrader hook for the upgrade path.

#### Scale-Adjust Namespace

Enabling `kai-resource-management` triggers creation of the `runai-scale-adjust` namespace. If a previous failed install left it without Helm ownership labels, the script deletes it first so Helm can recreate it cleanly.

#### Chart Reconciliation

Before installing, the script reconciles `charts/` against `Chart.yaml`, since Helm needs a subchart present for every declared dependency and `charts/*.tgz` is gitignored (a fresh clone has none):

1. **Check** — runs `helm dependency list`. Helm compares `charts/` against `Chart.yaml` itself, so a dependency may use any version constraint Helm supports (`0.2.6`, `~0.18.0`, `^1.2`, `>=1.0 <2.0`). Anything not `ok` (missing, wrong version, too many matches) or a tarball `Chart.yaml` does not declare triggers a fetch. Helm loads *every* tarball in `charts/`, so a leftover from a removed dependency would otherwise render with no condition to gate it.
2. **Fetch** — runs `helm dependency update`, which resolves version ranges, pulls the real subcharts and replaces outdated or undeclared tarballs. Both come from public OCI registries (`ghcr.io/kai-scheduler`, `ghcr.io/dsx-ai-factory/workload-map/karta`) and need no credentials.
3. **Fail** — if a dependency is still not satisfied (offline, registry down, no version matches the constraint), the script stops with the Helm error instead of installing a chart with an empty scheduler.

Empty placeholder tarballs left in `charts/` by older versions of the script are removed first, so the real charts are fetched.

---

## All Options

```
./local-dev.sh [OPTIONS]
```

| Flag | Description |
|---|---|
| `--domain FQDN` | Domain for ingress and TLS (default: `runai.local`) |
| `--admin-email EMAIL` | Initial admin account email (default: `test@run.ai`) |
| `--admin-password PASS` | Initial admin password (default: randomly generated by the chart and printed in the final summary) — min 8 chars, 1 digit, 1 upper, 1 lower, 1 special |
| `--parallel N` | Concurrent build jobs (default: `4`). Set higher on machines with many cores |
| `--skip-build` | Skip all image builds — reuse whatever is already in the Kind cluster |
| `--skip-build-cp` | Skip control-plane builds only |
| `--skip-build-rc` | Skip runai-cluster builds only |
| `--skip-frontend` | Skip the frontend npm build (saves ~5 min). Useful when only working on backend services |
| `--skip-cluster` | Skip Kind cluster creation — reuse an existing `runai-local` cluster |
| `--skip-deploy` | Skip Helm installs entirely (build images and set up cluster only) |
| `--skip-deploy-cp` | Skip deploying the control-plane chart |
| `--skip-deploy-rc` | Skip deploying the runai-cluster chart |
| `--skip-prometheus` | Skip the prometheus-operator install |
| `--skip-fake-gpu` | Skip the fake-gpu-operator install. Without it, GPU workloads cannot be scheduled |
| `--gpu-operator` | Use real NVIDIA GPUs (only NVIDIA is supported): install the NVIDIA GPU Operator instead of fake-gpu-operator. Linux + Docker only; see [Real GPUs](#real-gpus---gpu-operator) |
| `--gpu-driver` | With the GPU Operator, let it install the NVIDIA driver (`driver.enabled=true`) instead of using the host driver. Implies `--gpu-operator` |
| `--gpu-operator-version VER` | GPU Operator chart version (default: `v26.7.1`). Implies `--gpu-operator` |
| `--force-rebuild` | Ignore the build cache — rebuild and re-push every image |
| `--teardown` | Remove the Kind cluster, local registry and its volume, `kind` network, built images, and the `/etc/hosts` entry, then exit. See [Teardown](#teardown) |
| `--purge` | With `--teardown`: also clear the build cache, Go/Gradle caches, and the generated files in `.local-dev/` (certs, logs). Implies `--teardown` |
| `--help` | Print usage |

### Environment Variables

| Variable | Effect |
|---|---|
| `RUNAI_DOMAIN` | Same as `--domain` |
| `RUNAI_ADMIN_EMAIL` | Same as `--admin-email` |
| `RUNAI_ADMIN_PASSWORD` | Same as `--admin-password` |
| `PARALLEL_JOBS` | Same as `--parallel` |
| `KIND_CLUSTER` | Kind cluster name (default: `runai-local`) |
| `GPU_OPERATOR_VERSION` | NVIDIA GPU Operator chart version used with `--gpu-operator` (default: `v26.7.1`) |

---

## Real GPUs (`--gpu-operator`)

By default the cluster uses fake-gpu-operator, so no GPU hardware is needed. On a Linux host with GPUs, `--gpu-operator` installs the real [NVIDIA GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/) instead.

> **Only NVIDIA GPUs are supported.** AMD, Intel, and other accelerators are not. The script checks the host PCI bus for an NVIDIA device (vendor `0x10de`) and stops if it finds none.

| Flag | Effect |
|---|---|
| `--gpu-operator` | Install the NVIDIA GPU Operator using the **host's** NVIDIA driver (`driver.enabled=false`). |
| `--gpu-driver` | Let the GPU Operator install the driver itself (`driver.enabled=true`). Implies `--gpu-operator`. See [Operator-managed driver](#operator-managed-driver---gpu-driver). |
| `--gpu-operator-version VER` | GPU Operator chart version (default: `v26.7.1`; also settable via `GPU_OPERATOR_VERSION`). Implies `--gpu-operator`. |

Kind nodes are containers, so GPUs must first be passed from the host into a node container, and then from the node into pods. That needs a one-time host setup.

### Host driver (default)

**Host requirements:** Linux, Docker (Podman is not supported), an NVIDIA GPU with a working NVIDIA driver (`nvidia-smi` runs), and the [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html). The script checks these before it starts and stops if any is missing.

Docker must also use the `nvidia` runtime by default and accept GPU requests through volume mounts. If it does not, the script configures it for you, prompting for `sudo`, and **restarts Docker**, which stops any running containers. It lists them first. It runs:

```bash
sudo nvidia-ctk runtime configure --runtime=docker --set-as-default
sudo nvidia-ctk config --set accept-nvidia-visible-devices-as-volume-mounts=true --in-place
sudo systemctl restart docker
```

If you prefer, run these yourself before the script. It then skips this step.

**What the script does differently:**

- **Cluster creation:** the first worker (`runai-local-worker`) mounts `/dev/null` at `/var/run/nvidia-container-devices/all`, which makes the nvidia runtime inject all host GPUs into that node. Only one node gets GPUs, because every node would otherwise advertise the same physical devices.
- **Node prep:** inside the GPU node, the script symlinks `/sbin/ldconfig.real` and unmounts `/proc/driver/nvidia`, which the operator's container toolkit needs inside a Kind node.
- **Other nodes:** labeled `nvidia.com/gpu.deploy.operands=false`, because NFD sees the host PCI bus from every node.
- **Install:** installs `nvidia/gpu-operator` with `driver.enabled=false`. The operator's own DCGM exporter `ServiceMonitor` stays off, because the runai-cluster chart creates one; enabling both would scrape GPU metrics twice. The script then waits until the node advertises `nvidia.com/gpu`.
- **Registry fix on the GPU node:** the operator's toolkit adds `/etc/containerd/conf.d/99-nvidia.toml`, which uses containerd config version 3. Once any loaded file is version 3, containerd ignores the node's older-format `io.containerd.grpc.v1.cri` section, including the local-registry setting. Without a fix, the GPU node fails every `localhost:5001` pull with `connection refused`. After the operator is ready, the script adds `zz-runai-registry.toml`, which loads after the nvidia file and restores the registry setting. It then restarts containerd on that node and checks the setting is in effect.

The GPU mount can only be added when the cluster is created. To switch an existing fake-GPU cluster to real GPUs, recreate it:

```bash
./local-dev.sh --teardown
./local-dev.sh --gpu-operator
```

### Operator-managed driver (`--gpu-driver`)

With `--gpu-driver`, the host must have an NVIDIA GPU but **no** NVIDIA driver loaded; the script stops if `/sys/module/nvidia` exists. The NVIDIA Container Toolkit and the Docker runtime setup above are not needed, and the `/dev/null` mount is not added, because there is no host driver for the runtime to inject. Instead, the driver container runs inside the privileged GPU node, builds the kernel module against the host kernel, and creates the `/dev/nvidia*` devices itself. The script waits up to 30 minutes for the GPUs to appear.

> **Experimental on Kind.** The GPU Operator picks the driver image from the node OS, which in Kind is the `kindest/node` image (Debian), not your host distribution. The driver build also needs kernel headers matching the host kernel. If the `nvidia-driver-daemonset` pod fails, check its logs (`kubectl logs -n gpu-operator -l app=nvidia-driver-daemonset`). Use the host-driver mode instead.

---

## Common Workflows

### Rebuild a Single Service

```bash
cd control-plane/assets-service
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY=
docker buildx build --build-arg TARGETARCH=amd64 --platform linux/amd64 \
  --provenance=false --load -t runai-local/assets-service:0.0.0 .
kind load docker-image runai-local/assets-service:0.0.0 --name runai-local

helm upgrade runai-backend ./control-plane/control-plane-chart \
  -n runai-backend --reuse-values
```

### Iterate on the Cluster Chart Only (Images Already Loaded)

```bash
./local-dev.sh --skip-build --skip-cluster --skip-deploy-cp
```

### Rebuild Images Without Touching the Cluster

```bash
./local-dev.sh --skip-cluster --skip-deploy
```

### Rebuild Everything from Scratch

```bash
./local-dev.sh --teardown
./local-dev.sh
```

### Change Domain

```bash
./local-dev.sh --domain mydev.example.com
```

The script adds `mydev.example.com` to `/etc/hosts` automatically. To do it yourself instead:

```bash
sudo sh -c 'echo "127.0.0.1 mydev.example.com" >> /etc/hosts'
```

### Speed Up with More Parallel Jobs

```bash
PARALLEL_JOBS=8 ./local-dev.sh
```


---

## Trusting the Self-Signed Certificate

The CA cert is at `.local-dev/certs/ca.crt`.

### macOS

```bash
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain .local-dev/certs/ca.crt
```

### Linux

```bash
sudo cp .local-dev/certs/ca.crt /usr/local/share/ca-certificates/runai-local.crt
sudo update-ca-certificates
```

### Chrome/Firefox on Linux

Import `.local-dev/certs/ca.crt` via `chrome://settings/certificates` or the Firefox certificate manager.

---

## Podman Disk Management

local-dev.sh uses Podman (not Docker Desktop) on macOS. The VM gets a fixed disk allocated at init time.

### Check VM Disk Usage

```bash
podman machine ssh "df -h /"
```

### Recover from a Full Disk

When the Podman VM disk fills up, builds fail with `no space left on device` and even `podman system prune` cannot run (it needs space to write temp files).

**Fix:**

```bash
# Step 1: free enough space for Podman to operate
podman machine ssh "sudo journalctl --vacuum-size=10M"

# Step 2: prune all unused images, containers, and volumes
podman system prune -af --volumes
```

### Recreate the VM with a Larger Disk

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

`--rootful` is mandatory: kind needs it to create containers with elevated privileges. local-dev.sh will error if the machine is in rootless mode.

---

## nginx Worker Crashes (`pthread_create failed`)

On macOS with Podman + Kind, the nginx ingress controller may log:

```
nginx: worker process exited with code 2 (fatal)
pthread_create failed (11: Resource temporarily unavailable)
```

**Root cause:** Kubernetes sets a `pids.max = 307` cgroup limit on each container. Each nginx worker spawns 32 aio threads (33 pids total). With the default `worker_processes 6`, a config reload runs old and new workers simultaneously: `6 × 33 × 2 = 396 pids > 307`. Workers fail to create their thread pools and crash.

**Fix (already applied by `local-dev.sh`):** `patch_nginx_workers()` caps workers at 2 immediately after installing ingress-nginx. During reload, peak pids = `2 × 33 × 2 = 132 < 307`. The patch is idempotent and survives pod restarts (the ConfigMap persists). If you are debugging a pre-existing cluster without this fix, apply it manually:

```bash
kubectl patch configmap ingress-nginx-controller -n ingress-nginx \
  --type merge \
  -p '{"data":{"worker-processes":"2","max-worker-open-files":"65536","reuse-port":"false"}}'
kubectl rollout restart deployment ingress-nginx-controller -n ingress-nginx
```

---

## Networking Troubleshooting

### SSL Handshake Timeouts / Slow Responses

If `curl https://<domain>/_version` times out at the SSL handshake stage (`time_appconnect: 0`) but TCP connects instantly, the likely cause is **Cisco AnyConnect** (or similar corporate VPN with a socket-level kernel extension). AnyConnect's `acsockext` system extension intercepts all TCP connections, including loopback, and routes them through a low-MTU inspection tunnel, which fragments TLS certificate exchanges.

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

If `docker buildx build` fails with `no such file: bin/service-amd64`, the `build-go-amd` step did not produce a binary. Check the log file for the compile error.

If a service log shows `unknown flag: --platform` followed by the top-level `docker` usage, right after a *successful* `build-go-amd`, the `docker buildx` plugin is missing, so the Docker CLI fell back to the root command. This hits every service, not just the one you happened to read. Docker Desktop bundles buildx; Ubuntu's `docker.io` package does not. Install `docker-buildx-plugin` from Docker's official repo and verify with `docker buildx version`.

If `kind load docker-image` is slow, note that images are loaded into all 5 nodes (1 control-plane + 4 workers). With many services this is expected: it only runs once per build.

---

## Image Tag Reference

| Chart | Tag | Image name |
|---|---|---|
| control-plane | `0.0.0` (matches `Chart.yaml appVersion`) | `runai-local/<service>:0.0.0` |
| runai-cluster operator/hooks | `0.0.0-dev` | `runai-local/<service>:0.0.0-dev` |

Images are pushed to the local `kind-registry` container and referenced as `localhost:5001/runai-local/<service>:<tag>`. Helm deploys with `imagePullPolicy=IfNotPresent`, so Kubernetes pulls from the local registry and never reaches an external one.

---

## Known Issues

### No KRM / kai-scheduler / karta Pods After Deploy

The script stops before installing the chart if it cannot fetch these subcharts:

```
ERROR: runai-cluster chart dependencies are not satisfied (see above and the helm output). ...
```

The Helm error above it says why (no network, registry down, no version matching the constraint in `Chart.yaml`). Check what Helm sees in `charts/`; every dependency should be `ok`:

```bash
helm dependency list runai-cluster/cluster/runai-chart
```

Fix the cause, fetch manually, then redeploy just the chart:

```bash
helm dependency update runai-cluster/cluster/runai-chart
./local-dev.sh --skip-build --skip-cluster
```

With the real charts you should see these pods:

- `kai-operator`
- `kai-admission`
- `kai-binder`
- `kai-podgrouper`
- `kai-queuecontroller`
- `kai-node-scale-adjuster`
- `karta-operator`

```bash
kubectl get pods -n runai | grep -E 'kai|karta'
```

Earlier versions of the script installed an empty placeholder subchart when the fetch failed, so the scheduler silently deployed nothing; they also could not handle a version range in `Chart.yaml` ([#2716](https://github.com/dsx-ai-factory/run-ai/issues/2716)).

### Go Build Failures After an OSS Re-Export

**Applies only to OSS export repos** (those with a `.oss-export-id` file), not to the monorepo itself. Three different-looking errors share one cause:

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

Each export re-points every version tag at a fresh commit. Go assumes `module@version` is immutable and caches it forever, so a machine that built a previous export keeps serving that export's content. Nothing is actually wrong with the repo: the second and third errors are especially misleading, since the symbol and the stdlib package are both perfectly present.

`local-dev.sh` heals this automatically: it compares `.oss-export-id` against a marker in the module cache and evicts the affected caches when they differ. You should not normally have to do anything.

If you hit it anyway (an older checkout, or a cache cleared out from under the marker), purge by hand. **All three locations matter**; clearing only one or two leaves the build broken with no hint as to why.

Run from the repo root: it derives the module namespace itself, so there is nothing to substitute:

```bash
GP=~/.cache/go-build-docker-gopath/pkg/mod
PREFIX=$(awk '/^module /{print $2; exit}' "$(ls runai-common-packages/*/go.mod | head -1)" | cut -d/ -f1-3)
echo "purging $PREFIX"

# 1. git clone cache: Go builds module zips from here, and it holds the OLD
#    tag -> sha map, so it serves stale content even when everything else is clean
for d in "$GP"/cache/vcs/*/; do
  git -C "$d" config --get-regexp '^remote\..*\.url$' 2>/dev/null | grep -qi "${PREFIX##*/}" \
    && { chmod -R u+w "$d" 2>/dev/null; rm -rf "$d"; }
done

# 2. downloaded zips + extracted module trees (mode 0444, hence the chmod)
chmod -R u+w "$GP/$PREFIX" "$GP/cache/download/$PREFIX" 2>/dev/null
rm -rf "$GP/$PREFIX" "$GP/cache/download/$PREFIX"

# 3. build cache: per-module file index, also keyed by module@version
rm -rf ~/.cache/go-build-docker-gocache
```

On `permission denied`, delete as container-root instead; build steps sometimes write as root:

```bash
docker run --rm -v ~/.cache:/c busybox sh -c '
  rm -rf /c/go-build-docker-gocache /c/go-build-docker-gopath/pkg/mod/cache/vcs'
```

Expect the next build to be slow; every cache is cold.

Make sure the checkout is current: a `go.sum` from an older export describes content the tags no longer serve. `local-dev.sh` warns when you are behind, but it will not move your HEAD:

```bash
git pull --ff-only
```

---

## Teardown

```bash
./local-dev.sh --teardown
```

Removes everything the run created that only local-dev.sh uses:

- the Kind cluster and its node containers;
- the `kind-registry` container **and its volume**, which holds every pushed image (~1 GB);
- the `kind` network, and the `kindest/node` image, unless other Kind clusters still exist;
- all `runai-local/*` and `localhost:5001/runai-local/*` images, plus `registry:2`;
- the `127.0.0.1 <domain>` line in `/etc/hosts`, which prompts for `sudo`;
- the build-hash cache in `.local-dev/hashes`.

Caches that make the next run faster are **kept**, and the script prints each one with its size:

- Docker's build cache, which is shared with your other projects;
- `~/.cache/go-build-docker-*`;
- `control-plane/keycloak-extensions/.gradle`;
- `.local-dev/certs/`, `.local-dev/logs/`, and `.local-dev/bin/`.

`.local-dev/values-cp.yaml` and `.local-dev/values-rc.yaml` are tracked in git and are never removed, even with `--purge`.

To remove those too:

```bash
./local-dev.sh --teardown --purge
```

`--purge` runs `docker builder prune -af`, which clears Docker's **entire** build cache, not just this project's. Base images pulled for builds (`golang`, `node`, `gradle`, and so on) are never removed; delete them with `docker rmi` if needed.
