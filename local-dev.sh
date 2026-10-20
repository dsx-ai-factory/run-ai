#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# local-dev.sh — Run.ai local development environment
#
# Builds all control-plane and runai-cluster services, spins up a Kind cluster,
# loads images directly via 'kind load docker-image', and deploys Run.ai with a
# self-signed TLS certificate.
#
# Prerequisites (checked at runtime):
#   - Docker Desktop (containerd image store recommended for Mac; required on Linux)
#   - GITHUB_TOKEN env var  — Go private-module access; not needed once repo is public or with --skip-build
#   - kind, kubectl, helm   — installed automatically if missing
#
# Usage:
#   ./local-dev.sh [OPTIONS]
#
# Options:
#   --domain FQDN          Domain to use (default: runai.local)
#   --admin-email EMAIL    Admin account email (default: test@run.ai)
#   --admin-password PASS  Admin password — must meet complexity requirements
#                          (default: auto-generated, shown in the final summary)
#   --parallel N           Parallel build jobs (default: 4)
#   --skip-build           Skip building Docker images
#   --skip-build-cp        Skip building control-plane images only
#   --skip-build-rc        Skip building runai-cluster images only
#   --skip-frontend        Skip building the frontend (JS build is slow)
#   --skip-cluster         Skip Kind cluster creation (reuse existing)
#   --skip-deploy          Skip Helm deployment (build + cluster only)
#   --skip-deploy-cp       Skip deploying the control plane only
#   --skip-deploy-rc       Skip deploying runai-cluster only
#   --skip-prometheus      Skip prometheus-operator install
#   --skip-fake-gpu        Skip fake-gpu-operator install
#   --gpu-operator         Use real NVIDIA GPUs: install NVIDIA GPU Operator instead
#                          of fake-gpu-operator (Linux + Docker + NVIDIA GPUs only)
#   --gpu-driver           Let the GPU Operator install the NVIDIA driver
#                          (driver.enabled=true; default uses the host driver).
#                          Implies --gpu-operator
#   --gpu-operator-version VER
#                          GPU Operator chart version (default: v26.7.1).
#                          Implies --gpu-operator
#   --teardown             Remove the cluster, registry (+ its volume), kind network,
#                          built images, and the /etc/hosts entry, then exit
#   --purge                With --teardown: also clear the build cache, Go/Gradle
#                          caches, and generated .local-dev/ files (implies --teardown)
#   --force-rebuild        Ignore build cache; rebuild and reload all images
#   --help                 Show this help

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ─── Defaults ─────────────────────────────────────────────────────────────────
# Images are tagged with this prefix and loaded directly into Kind via
# 'kind load docker-image'. No registry container is needed.
LOCAL_IMAGE_REPO="runai-local"

# Local registry — runs as a container, connected to the Kind network.
# Nodes pull from it via http://kind-registry:5000; the Mac host pushes to
# localhost:REGISTRY_PORT (forwarded by Podman/Docker port mapping).
REGISTRY_NAME="kind-registry"
REGISTRY_PORT="5001"

# CP chart appVersion is "0.0.0"; images must match.
# RC chart values hardcode "0.0.0-dev" for all explicit image tags.
CP_IMAGE_TAG="0.0.0"
RC_IMAGE_TAG="0.0.0-dev"

KIND_CLUSTER="${KIND_CLUSTER:-runai-local}"
KIND_VERSION="v0.33.0"
# K8s 1.37 image from the kind v0.33.0 release, pinned by digest as kind recommends.
# K8s 1.35+ has no cgroup v1 support: the container runtime host must use cgroup v2.
KIND_NODE_IMAGE="kindest/node:v1.37.0@sha256:a1ed56cfb0e7b93589bdf97c8cd566405a265939e3620fc4f5de89adff580ae5"
GPU_OPERATOR_VERSION="${GPU_OPERATOR_VERSION:-v26.7.1}"

# Linux host floor: 5 Kind nodes + ~50 services + parallel builds share the host
# directly (no VM to size). Disk: a full build leaves ~44 GB of caches (Docker build
# cache ~11 GB, Go module + build caches ~33 GB) plus ~8 GB of images stored twice.
MIN_LINUX_CPUS=4
MIN_LINUX_RAM_GB=16
MIN_LINUX_DISK_GB=100

DOMAIN="${RUNAI_DOMAIN:-runai.local}"
ADMIN_EMAIL="${RUNAI_ADMIN_EMAIL:-test@run.ai}"
# Empty means the chart's secret-generator hook picks a random password; it is read
# back from runai-backend-gen-tenants-manager after deploy.
ADMIN_PASSWORD="${RUNAI_ADMIN_PASSWORD:-}"
CP_NAMESPACE="runai-backend"
RC_NAMESPACE="runai"
CP_TENANT_NAME="runai"
CP_CLUSTER_NAME="${KIND_CLUSTER}"
# Keycloak master admin, set in .local-dev/values-cp.yaml; its password is generated
# into runai-backend-gen-keycloakx. Used to impersonate ADMIN_EMAIL for API tokens.
KC_ADMIN_USER="admin"
PARALLEL_JOBS="${PARALLEL_JOBS:-4}"

# Set by provision_cluster_credentials() after CP is up
CLUSTER_UUID=""
CLUSTER_BOOTSTRAP_TOKEN=""

CERT_DIR="${REPO_ROOT}/.local-dev/certs"
LOG_DIR="${REPO_ROOT}/.local-dev/logs"
HASH_DIR="${REPO_ROOT}/.local-dev/hashes"
CP_REBUILT_FILE="${LOG_DIR}/.cp-rebuilt"
RC_REBUILT_FILE="${LOG_DIR}/.rc-rebuilt"

# ─── CLI flags ─────────────────────────────────────────────────────────────────
SKIP_BUILD=0
SKIP_BUILD_CP=0
SKIP_BUILD_RC=0
SKIP_FRONTEND=0
SKIP_CLUSTER=0
SKIP_DEPLOY=0
SKIP_DEPLOY_CP=0
SKIP_DEPLOY_RC=0
SKIP_PROMETHEUS=0
SKIP_FAKE_GPU=0
GPU_OPERATOR=0
GPU_DRIVER=0
TEARDOWN=0
PURGE=0
FORCE_REBUILD=0

# Set to "podman" by check_prereqs when a Podman-backed socket is detected.
CONTAINER_TOOL="docker"

usage() {
  sed -n '/^# Usage:/,/^[^#]/p' "$0" | grep '^#' | sed 's/^# \?//'
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --domain)          DOMAIN="$2";         shift 2 ;;
    --admin-email)     ADMIN_EMAIL="$2";    shift 2 ;;
    --admin-password)  ADMIN_PASSWORD="$2"; shift 2 ;;
    --parallel)        PARALLEL_JOBS="$2";  shift 2 ;;
    --skip-build)      SKIP_BUILD=1;        shift ;;
    --skip-build-cp)   SKIP_BUILD_CP=1;     shift ;;
    --skip-build-rc)   SKIP_BUILD_RC=1;     shift ;;
    --skip-frontend)   SKIP_FRONTEND=1;     shift ;;
    --skip-cluster)    SKIP_CLUSTER=1;      shift ;;
    --skip-deploy)     SKIP_DEPLOY=1;       shift ;;
    --skip-deploy-cp)  SKIP_DEPLOY_CP=1;    shift ;;
    --skip-deploy-rc)  SKIP_DEPLOY_RC=1;    shift ;;
    --skip-prometheus) SKIP_PROMETHEUS=1;   shift ;;
    --skip-fake-gpu)   SKIP_FAKE_GPU=1;     shift ;;
    --gpu-operator)    GPU_OPERATOR=1;      shift ;;
    --gpu-driver)      GPU_OPERATOR=1; GPU_DRIVER=1; shift ;;
    --gpu-operator-version) GPU_OPERATOR=1; GPU_OPERATOR_VERSION="$2"; shift 2 ;;
    --teardown)        TEARDOWN=1;          shift ;;
    --purge)           TEARDOWN=1; PURGE=1; shift ;;
    --force-rebuild)   FORCE_REBUILD=1;      shift ;;
    --help)            usage; exit 0 ;;
    *) echo "Unknown flag: $1"; usage; exit 1 ;;
  esac
done

# ─── Service lists ─────────────────────────────────────────────────────────────
# Control-plane Go services (build with build-go-amd → docker push)
CP_GO_SERVICES=(
  assets-service
  audit-service
  authorization
  bff-service
  catalog-service
  cli-exposer
  cluster-migrator
  cluster-service
  datavolumes-service
  gateway-hub
  identity-manager
  k8s-objects-tracker
  metric-stores-migrator
  metrics-service
  notifications-proxy
  notifications-service
  org-unit-service
  policy-service
  secret-generator
  temp-backend-settings
  tenants-manager
  workloads
  workloads-manager
)

# runai-cluster Go services
RC_GO_SERVICES=(
  accessrule-controller
  admission-controller
  advice-restart-controller
  agent
  anyworkload-controller
  application-controller
  cluster-api
  cluster-sync
  external-workload-integrator
  go-operator
  helm-hooks-handler
  init-ca
  karta-migrate
  metrics-exporter
  runai-crd-upgrader
  shared-objects-controller
  status-updater
  workload-exporter
  runtime-rules
)

# ─── Helpers ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
log()  { echo -e "${GREEN}[$(date '+%H:%M:%S')]${NC} $*"; }
warn() { echo -e "${YELLOW}[$(date '+%H:%M:%S')] WARN:${NC} $*" >&2; }
die()  { echo -e "${RED}[$(date '+%H:%M:%S')] ERROR:${NC} $*" >&2; exit 1; }

detect_os_arch() {
  case "$(uname -s)" in
    Darwin) OS=darwin ;;
    Linux)  OS=linux  ;;
    *)      die "Unsupported OS: $(uname -s)" ;;
  esac
  case "$(uname -m)" in
    x86_64)          ARCH=amd64 ;;
    aarch64|arm64)   ARCH=arm64 ;;
    *)               die "Unsupported arch: $(uname -m)" ;;
  esac
}

require_cmd() {
  command -v "$1" &>/dev/null || die "'$1' is required but not found. $2"
}

install_kind() {
  log "Installing kind ${KIND_VERSION} (${OS}/${ARCH})..."
  local url="https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-${OS}-${ARCH}"
  local tmp
  tmp="$(mktemp)"
  curl -fsSL "${url}" -o "${tmp}"
  chmod +x "${tmp}"
  if [[ "${OS}" == "darwin" ]]; then
    sudo mv "${tmp}" /usr/local/bin/kind
  else
    sudo mv "${tmp}" /usr/local/bin/kind
  fi
  log "kind installed: $(kind --version)"
}

kind_at_least_pinned() {
  local installed pinned="${KIND_VERSION#v}"
  installed="$(kind --version 2>/dev/null | awk '{print $3}')"
  [[ -n "${installed}" ]] || return 1
  [[ "$(printf '%s\n' "${installed}" "${pinned}" | sort -V | head -1)" == "${pinned}" ]]
}

install_kubectl() {
  log "Installing kubectl (${OS}/${ARCH})..."
  local version
  version="$(curl -fsSL https://dl.k8s.io/release/stable.txt)"
  local url="https://dl.k8s.io/release/${version}/bin/${OS}/${ARCH}/kubectl"
  local tmp
  tmp="$(mktemp)"
  curl -fsSL "${url}" -o "${tmp}"
  chmod +x "${tmp}"
  if [[ "${OS}" == "darwin" ]]; then
    mv "${tmp}" /usr/local/bin/kubectl
  else
    sudo mv "${tmp}" /usr/local/bin/kubectl
  fi
  log "kubectl installed: $(kubectl version --client --short 2>/dev/null || kubectl version --client)"
}

install_helm() {
  log "Installing Helm..."
  curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
  log "Helm installed: $(helm version --short)"
}

# ─── Phase 0: Teardown ────────────────────────────────────────────────────────
do_teardown() {
  if ! command -v docker &>/dev/null && command -v podman &>/dev/null; then
    CONTAINER_TOOL=podman
  fi
  local ct="${CONTAINER_TOOL}"
  log "Tearing down local dev environment..."

  kind delete cluster --name "${KIND_CLUSTER}" 2>/dev/null || true
  # -v: registry:2 keeps every pushed image in an anonymous volume (~1 GB).
  "${ct}" rm -fv "${REGISTRY_NAME}" >/dev/null 2>&1 || true

  local other_clusters
  other_clusters="$(kind get clusters 2>/dev/null || true)"
  local -a images=("registry:2")
  if [[ -z "${other_clusters}" ]]; then
    "${ct}" network rm kind >/dev/null 2>&1 || true
    images+=("${KIND_NODE_IMAGE}")
  else
    log "  Other Kind clusters exist (${other_clusters//$'\n'/, }) — keeping the kind network and node image"
  fi

  local img
  while IFS= read -r img; do images+=("${img}"); done < <(
    "${ct}" images --format '{{.Repository}}:{{.Tag}}' 2>/dev/null \
      | grep -E "(^|/)${LOCAL_IMAGE_REPO}/" || true)
  log "  Removing ${#images[@]} images..."
  for img in "${images[@]}"; do
    "${ct}" rmi -f "${img}" >/dev/null 2>&1 || true
  done

  local hosts_line="127.0.0.1 ${DOMAIN}"
  if grep -qxF "${hosts_line}" /etc/hosts 2>/dev/null; then
    log "  Removing '${hosts_line}' from /etc/hosts (requires sudo)..."
    local hosts_tmp
    hosts_tmp="$(mktemp)"
    grep -vxF "${hosts_line}" /etc/hosts > "${hosts_tmp}" || true
    sudo cp "${hosts_tmp}" /etc/hosts || warn "Could not update /etc/hosts — remove '${hosts_line}' manually"
    rm -f "${hosts_tmp}"
  fi

  rm -rf "${HASH_DIR}" "${CLUSTER_STAMP_FILE}"

  local -a caches=(
    "${HOME}/.cache/go-build-docker-gopath"
    "${HOME}/.cache/go-build-docker-gocache"
    "${REPO_ROOT}/control-plane/keycloak-extensions/.gradle"
    # Not all of .local-dev/: values-cp.yaml and values-rc.yaml there are tracked in git.
    "${CERT_DIR}"
    "${LOG_DIR}"
    "${REPO_ROOT}/.local-dev/bin"
  )

  if [[ $PURGE -eq 1 ]]; then
    log "  Purging build caches..."
    if [[ "${ct}" == "docker" ]]; then
      docker builder prune -af >/dev/null 2>&1 || true
    else
      podman image prune -f >/dev/null 2>&1 || true
    fi
    local dir
    for dir in "${caches[@]}"; do
      [[ -e "${dir}" ]] || continue
      # Go's module cache is read-only, and builds can leave root-owned files behind.
      chmod -R u+w "${dir}" 2>/dev/null || true
      rm -rf "${dir}" 2>/dev/null || true
      if [[ -e "${dir}" ]]; then
        "${ct}" run --rm -v "$(dirname "${dir}"):/p" busybox rm -rf "/p/$(basename "${dir}")" >/dev/null 2>&1 || true
      fi
      [[ -e "${dir}" ]] && warn "  Could not remove ${dir} — try: sudo rm -rf ${dir}"
    done
    log "Done. Everything local-dev.sh created has been removed."
  else
    log "Done. Kept for faster rebuilds (remove with --teardown --purge):"
    if [[ "${ct}" == "docker" ]]; then
      log "  build cache: $(docker system df --format '{{.Type}}\t{{.Size}}' 2>/dev/null | awk -F'\t' '/Build Cache/{print $2}') (Docker's build cache is shared with other projects)"
    fi
    local dir
    for dir in "${caches[@]}"; do
      [[ -e "${dir}" ]] && log "  $(du -sh "${dir}" 2>/dev/null | cut -f1)	${dir}"
    done
  fi
  log "Base images pulled for builds (golang, node, gradle, ...) are not removed."
  exit 0
}

# ─── Phase 1: Prerequisites ───────────────────────────────────────────────────
ensure_gitconfig_file() {
  local f="${HOME}/.gitconfig"
  # Build containers bind-mount ~/.gitconfig. If it does not exist, Docker creates it as
  # a directory (root-owned on Linux), which breaks git for this user in every repo.
  if [[ -d "${f}" && ! -L "${f}" ]]; then
    rmdir "${f}" 2>/dev/null \
      || die "${f} is a directory (left by a Docker bind mount), which breaks git. Fix with: sudo rmdir ${f}"
    warn "Removed ${f}, an empty directory left by an earlier Docker bind mount"
  fi
  [[ -e "${f}" ]] || touch "${f}"
}

check_prereqs() {
  log "Checking prerequisites..."

  detect_os_arch
  log "Platform: ${OS}/${ARCH}"

  ensure_gitconfig_file

  # Podman without a real `docker` command (common on macOS, where `docker` is
  # often just a shell alias that scripts can't see). Create a shim so the prereq
  # check, podman detection (`docker info`), and the Makefile's `docker run`
  # Go-build steps all resolve to podman — no `brew install podman-docker` needed.
  # The shim also drops buildx-only flags podman rejects (--provenance/--load).
  if ! command -v docker &>/dev/null && command -v podman &>/dev/null; then
    mkdir -p "${REPO_ROOT}/.local-dev/bin"
    cat > "${REPO_ROOT}/.local-dev/bin/docker" <<'SHIM'
#!/usr/bin/env bash
# docker -> podman shim (auto-created by local-dev.sh)
_build() {
  local a=()
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --provenance=*|--sbom=*|--attest=*) shift ;;
      --provenance|--sbom|--attest) shift 2 ;;
      --load) shift ;;
      *) a+=("$1"); shift ;;
    esac
  done
  exec podman build "${a[@]}"
}
if [[ "${1:-}" == "buildx" && "${2:-}" == "build" ]]; then shift 2; _build "$@"
elif [[ "${1:-}" == "build" ]]; then shift; _build "$@"
else exec podman "$@"; fi
SHIM
    chmod +x "${REPO_ROOT}/.local-dev/bin/docker"
    export PATH="${REPO_ROOT}/.local-dev/bin:${PATH}"
    log "docker not found — created a podman shim at .local-dev/bin/docker"
  fi

  require_cmd docker "Install Docker Desktop from https://docs.docker.com/get-docker/, or install Podman"
  local _tool
  for _tool in make git curl python3; do
    require_cmd "${_tool}" "Install ${_tool} via brew (macOS) or your package manager"
  done

  if ! docker info &>/dev/null; then
    die "Docker daemon is not running. Start Docker Desktop and retry."
  fi

  # On macOS: if podman CLI is present and the machine is rootless, auto-switch
  # to rootful before anything else — kind cannot create clusters in rootless mode.
  if [[ "${OS}" == "darwin" ]] && command -v podman &>/dev/null; then
    local is_rootful
    is_rootful="$(podman machine inspect --format '{{.Rootful}}' 2>/dev/null || echo "true")"
    if [[ "${is_rootful}" == "false" ]]; then
      log "Podman is in rootless mode — switching to rootful (required by kind on macOS)..."
      podman machine stop \
        && podman machine set --rootful --disk-size 200 \
        && podman machine start \
        || die "Failed to switch Podman to rootful mode. Run manually:
    podman machine stop && podman machine set --rootful --disk-size 200 && podman machine start"
      log "Podman machine restarted in rootful mode"
    fi
  fi

  # Detect Podman-backed Docker socket and configure kind accordingly.
  if docker info 2>/dev/null | grep -qiE "rootless|containers/storage" || \
     echo "${DOCKER_HOST:-}" | grep -qi "podman"; then
    log "Detected Podman provider — setting KIND_EXPERIMENTAL_PROVIDER=podman"
    export KIND_EXPERIMENTAL_PROVIDER=podman
    CONTAINER_TOOL=podman
    # Rootless Podman blocks ports <1024; lower the threshold inside the VM so
    # kind can bind ports 80/443 on the host.
    local current_port_start
    current_port_start=$(podman machine ssh "sysctl -n net.ipv4.ip_unprivileged_port_start" 2>/dev/null | tr -d '[:space:]')
    if [[ "${current_port_start}" != "80" ]]; then
      log "Configuring Podman VM to allow binding ports 80/443 (current: ${current_port_start:-unknown})..."
      podman machine ssh "sudo sysctl -w net.ipv4.ip_unprivileged_port_start=80" || \
        warn "Could not lower unprivileged_port_start — ports 80/443 may fail"
    fi
    # Raise the Linux VM kernel keyring limit so Kind pods don't stall when
    # 100+ containers compete for keyring slots (default maxkeys is 200).
    local vm_maxkeys
    vm_maxkeys=$(podman machine ssh "cat /proc/sys/kernel/keys/maxkeys" 2>/dev/null | tr -d '[:space:]')
    if [[ -n "${vm_maxkeys}" && ${vm_maxkeys} -lt 800 ]]; then
      log "Raising Podman VM kernel keyring limit (current: ${vm_maxkeys})..."
      podman machine ssh "sudo sysctl -w kernel.keys.maxkeys=800 kernel.keys.maxbytes=200000" 2>/dev/null || \
        warn "Could not raise Podman VM kernel keyring limit — pods may stall with 100+ containers"
    fi
  elif [[ "${OS}" == "darwin" ]] && command -v rdctl &>/dev/null && \
       { docker context show 2>/dev/null | grep -qi "rancher-desktop" || \
         echo "${DOCKER_HOST:-}" | grep -qi "\.rd/docker\.sock"; }; then
    log "Detected Rancher Desktop provider"
    # Rancher Desktop's dockerd runs inside its own Linux VM (moby/QEMU), same as
    # Podman above — the macOS host has no meaningful /proc to read, so (like
    # Podman) the kernel limits below have to be read and raised inside the VM,
    # reached via `rdctl shell` instead of `podman machine ssh`. Not persisted
    # across VM restarts (same as the Podman branch): cheap to re-check and
    # re-apply on every run instead.
    #
    # `rdctl shell` can fail (VM not running, rdctl hiccup) and print an error
    # string to stdout instead of a number. Under `set -u`, comparing that
    # directly (`$val -lt N`) crashes the whole script with "unbound variable" —
    # bash's arithmetic context treats a non-numeric word as a variable
    # reference. Route every read through this guard so a bad read just skips
    # that one check instead of aborting check_prereqs.
    _rd_is_numeric() { [[ "$1" =~ ^[0-9]+$ ]]; }

    # kernel.keys.maxkeys AND kernel.keys.maxbytes both gate the same failure
    # (pods stalling once 100+ containers compete for keyring slots) — raise if
    # EITHER is below target, not just maxkeys, or a low maxbytes with an
    # already-sufficient maxkeys silently never gets fixed.
    local rd_maxkeys rd_maxbytes
    rd_maxkeys=$(rdctl shell cat /proc/sys/kernel/keys/maxkeys 2>/dev/null | tr -d '[:space:]')
    rd_maxbytes=$(rdctl shell cat /proc/sys/kernel/keys/maxbytes 2>/dev/null | tr -d '[:space:]')
    if { _rd_is_numeric "${rd_maxkeys}" && [[ ${rd_maxkeys} -lt 800 ]]; } || \
       { _rd_is_numeric "${rd_maxbytes}" && [[ ${rd_maxbytes} -lt 200000 ]]; }; then
      log "Raising Rancher Desktop VM kernel keyring limit (current: maxkeys=${rd_maxkeys:-?}, maxbytes=${rd_maxbytes:-?})..."
      if ! rdctl shell sudo sysctl -w kernel.keys.maxkeys=800 kernel.keys.maxbytes=200000 2>/dev/null; then
        warn "Could not raise Rancher Desktop VM kernel keyring limit — pods may stall with 100+ containers."
        warn "  Run once, then re-create the cluster:"
        warn "    rdctl shell sudo sysctl -w kernel.keys.maxkeys=800 kernel.keys.maxbytes=200000"
      fi
    fi

    # inotify limits are per-VM and shared by every Kind node — see the detailed
    # failure-mode comment on the native-Linux branch below; same root cause,
    # just hit through Rancher Desktop's VM instead of the host kernel directly.
    local rd_inotify_need_instances=8192 rd_inotify_need_watches=524288
    local rd_cur_inst rd_cur_watch rd_inotify_args=""
    rd_cur_inst=$(rdctl shell cat /proc/sys/fs/inotify/max_user_instances 2>/dev/null | tr -d '[:space:]')
    rd_cur_watch=$(rdctl shell cat /proc/sys/fs/inotify/max_user_watches 2>/dev/null | tr -d '[:space:]')
    _rd_is_numeric "${rd_cur_inst}" && [[ ${rd_cur_inst} -lt ${rd_inotify_need_instances} ]] && \
      rd_inotify_args+=" fs.inotify.max_user_instances=${rd_inotify_need_instances}"
    _rd_is_numeric "${rd_cur_watch}" && [[ ${rd_cur_watch} -lt ${rd_inotify_need_watches} ]] && \
      rd_inotify_args+=" fs.inotify.max_user_watches=${rd_inotify_need_watches}"
    if [[ -n "${rd_inotify_args}" ]]; then
      log "Raising Rancher Desktop VM inotify limits (instances: ${rd_cur_inst:-?}, watches: ${rd_cur_watch:-?})..."
      if ! rdctl shell sudo sysctl -w ${rd_inotify_args} >/dev/null 2>&1; then
        warn "Could not raise Rancher Desktop VM inotify limits."
        warn "  Without this, Kind nodes can go NotReady with 'container runtime is down',"
        warn "  and PVCs stay Pending because the provisioner cannot run on them."
        warn "  Run once, then re-create the cluster:"
        warn "    rdctl shell sudo sysctl -w${rd_inotify_args}"
      fi
    fi
    unset -f _rd_is_numeric
  else
    # On Linux with Docker, the host kernel keyring limit applies directly.
    # The default (200 keys) is exhausted when Kind runs 100+ containers.
    if [[ -f /proc/sys/kernel/keys/maxkeys ]]; then
      local cur_maxkeys
      cur_maxkeys=$(cat /proc/sys/kernel/keys/maxkeys)
      if [[ ${cur_maxkeys} -lt 800 ]]; then
        log "Raising kernel keyring limit (current: ${cur_maxkeys}) for Kind cluster..."
        # -n so a missing sudo credential fails immediately instead of blocking the run
        # on a password prompt. On Ubuntu this surfaces via polkit as an opaque
        # "Authentication is required to start transient unit run-p...scope", which gives
        # no hint that it is only a keyring tweak and is safe to decline.
        if ! sudo -n sysctl -w kernel.keys.maxkeys=800 kernel.keys.maxbytes=200000 >/dev/null 2>&1; then
          warn "Could not raise kernel keyring limit (needs sudo) — pods may stall with 100+ containers."
          warn "  Optional; run once if you hit stalled pods:"
          warn "    sudo sysctl -w kernel.keys.maxkeys=800 kernel.keys.maxbytes=200000"
        fi
      fi
    fi

    # inotify limits are per-host and shared by every Kind node. Ubuntu defaults to 128
    # instances, which 5 nodes running ~50 services exhaust. When it runs out,
    # containerd starts but its CRI plugin fails to load:
    #   failed to create CRI service: failed to create cni conf monitor for default:
    #   failed to create fsnotify watcher: too many open files
    # The node then reports "container runtime is down, PLEG is not healthy" and goes
    # NotReady with CONTAINER-RUNTIME containerd://Unknown. Downstream that looks like
    # anything but inotify: pods stuck ContainerCreating, FailedCreatePodSandBox with
    # "error reading from server: EOF", and local-path PVCs stuck Pending because the
    # provisioner's helper pod cannot run on the dead node ("create process timeout").
    # Values are the ones kind documents.
    local _inotify_need_instances=8192 _inotify_need_watches=524288
    local _cur_inst _cur_watch _inotify_args=""
    _cur_inst=$(cat /proc/sys/fs/inotify/max_user_instances 2>/dev/null || echo "")
    _cur_watch=$(cat /proc/sys/fs/inotify/max_user_watches 2>/dev/null || echo "")
    [[ -n "${_cur_inst}"  && ${_cur_inst}  -lt ${_inotify_need_instances} ]] && \
      _inotify_args+=" fs.inotify.max_user_instances=${_inotify_need_instances}"
    [[ -n "${_cur_watch}" && ${_cur_watch} -lt ${_inotify_need_watches} ]] && \
      _inotify_args+=" fs.inotify.max_user_watches=${_inotify_need_watches}"
    if [[ -n "${_inotify_args}" ]]; then
      log "Raising inotify limits (instances: ${_cur_inst}, watches: ${_cur_watch}) for Kind nodes..."
      if ! sudo -n sysctl -w ${_inotify_args} >/dev/null 2>&1; then
        warn "Could not raise inotify limits (needs sudo)."
        warn "  Without this, Kind nodes can go NotReady with 'container runtime is down',"
        warn "  and PVCs stay Pending because the provisioner cannot run on them."
        warn "  Run once, then re-create the cluster:"
        warn "    sudo sysctl -w${_inotify_args}"
        warn "  To persist across reboots:"
        warn "    echo 'fs.inotify.max_user_instances=${_inotify_need_instances}' | sudo tee /etc/sysctl.d/99-kind.conf"
        warn "    echo 'fs.inotify.max_user_watches=${_inotify_need_watches}'  | sudo tee -a /etc/sysctl.d/99-kind.conf"
      fi
    fi
  fi

  if ! command -v kind &>/dev/null; then
    install_kind
  elif ! kind_at_least_pinned; then
    # Node images are built for a kind release; older kind binaries may not run them.
    warn "kind $(kind --version | awk '{print $3}') is older than ${KIND_VERSION} (needed for ${KIND_NODE_IMAGE%@*}) — upgrading"
    install_kind
    kind_at_least_pinned || die "An older kind still comes first in PATH at $(command -v kind). Upgrade or remove it, then re-run."
  else
    log "kind: $(kind --version)"
  fi

  if ! command -v kubectl &>/dev/null; then
    install_kubectl
  else
    log "kubectl: $(kubectl version --client --short 2>/dev/null | head -1 || echo 'found')"
  fi

  if ! command -v helm &>/dev/null; then
    install_helm
  else
    log "helm: $(helm version --short)"
  fi

  require_cmd openssl "Install openssl via brew (macOS) or your package manager"

  [[ $GPU_OPERATOR -eq 1 ]] && check_gpu_host_prereqs

  # GITHUB_TOKEN is only needed when building Go services (private modules from github.com/run-ai/*).
  # Once the repo is public this check becomes unnecessary.
  if [[ $SKIP_BUILD -eq 0 && ( $SKIP_BUILD_CP -eq 0 || $SKIP_BUILD_RC -eq 0 ) && -z "${GITHUB_TOKEN:-}" ]]; then
    warn "GITHUB_TOKEN is not set. Go builds may fail if github.com/run-ai/* modules are private."
    warn "Set GITHUB_TOKEN or use --skip-build to skip compilation."
  fi

  mkdir -p "${CERT_DIR}" "${LOG_DIR}"

  # Warn early if the container VM disk is nearly full — parallel builds each
  # write several GB of BuildKit/overlay snapshots and will fail mid-build.
  local vm_avail_gb=999
  if [[ "${CONTAINER_TOOL}" == "podman" ]]; then
    vm_avail_gb=$(podman machine ssh \
      "df / | awk 'NR==2{printf \"%d\", \$4/1024/1024}'" 2>/dev/null || echo 999)
  else
    vm_avail_gb=$(docker run --rm busybox \
      df / 2>/dev/null | awk 'NR==2{printf "%d", $4/1024/1024}' || echo 999)
  fi
  if (( vm_avail_gb < 20 && vm_avail_gb != 999 )); then
    warn "Container VM has only ${vm_avail_gb} GB free (need ≥20 GB for parallel builds)."
    warn "Free space with: docker system prune -f && docker volume prune -f"
    if [[ "${CONTAINER_TOOL}" == "podman" ]]; then
      warn "Or increase disk: podman machine stop && podman machine set --disk-size 200 && podman machine start"
    else
      warn "Or increase Docker Desktop → Settings → Resources → Virtual Disk Limit"
    fi
    if (( vm_avail_gb < 5 )); then
      die "Less than 5 GB free in container VM — builds will fail. Free space first."
    fi
  fi

  if [[ "${OS}" == "linux" ]]; then
    local host_cpus host_ram_gb
    local -a below_floor=()
    host_cpus=$(nproc 2>/dev/null || echo 0)
    host_ram_gb=$(awk '/^MemTotal/{printf "%d", $2/1024/1024}' /proc/meminfo 2>/dev/null || echo 0)
    if (( host_cpus > 0 && host_cpus < MIN_LINUX_CPUS )); then
      below_floor+=("CPU:  ${host_cpus} cores (minimum ${MIN_LINUX_CPUS})")
    fi
    # MemTotal excludes kernel-reserved memory, so a 16 GB host reports ~15 GB.
    if (( host_ram_gb > 0 && host_ram_gb < MIN_LINUX_RAM_GB - 1 )); then
      below_floor+=("RAM:  ${host_ram_gb} GB (minimum ${MIN_LINUX_RAM_GB} GB)")
    fi
    if (( vm_avail_gb != 999 && vm_avail_gb < MIN_LINUX_DISK_GB )); then
      below_floor+=("Disk: ${vm_avail_gb} GB free for container storage (minimum ${MIN_LINUX_DISK_GB} GB)")
    fi
    if (( ${#below_floor[@]} > 0 )); then
      warn "Host is below the documented minimum for the full stack:"
      local _item
      for _item in "${below_floor[@]}"; do warn "  ${_item}"; done
      warn "  Expect OOM-killed builds or NotReady nodes. Continuing anyway;"
      warn "  --parallel 2 and --skip-frontend reduce peak CPU/RAM during builds."
    fi
  fi

  # The full stack (NATS×3 + PostgreSQL + Keycloak + ~50 services under QEMU on
  # ARM64) needs at least 12 GB of VM RAM. With 8 GB the API server gets starved,
  # kube-controller-manager loses leader election repeatedly, and the cluster
  # never stabilises.
  if [[ "${CONTAINER_TOOL}" == "podman" ]]; then
    local vm_total_gb
    vm_total_gb=$(podman machine ssh \
      "awk '/^MemTotal/{printf \"%d\", \$2/1024/1024}' /proc/meminfo" 2>/dev/null)
    if [[ -n "${vm_total_gb}" ]] && (( vm_total_gb < 11 )); then
      die "Podman VM has only ${vm_total_gb} GB RAM (minimum 12 GB required).
Resize with:
  podman machine stop
  podman machine set --memory 12288
  podman machine start"
    fi
  fi

  ensure_insecure_registry
  ensure_amd64_emulation

  log "Prerequisites OK"
}

# Every service image is built for linux/amd64 (see container_build), so on a non-amd64
# host the amd64 RUN steps need binfmt/qemu emulation. Without it the Go build succeeds
# — cross-compiling needs no emulation — and the image build then dies on the first RUN:
#
#   #3 [base 2/2] RUN mkdir -p /temp/tmp && ...
#   #3 0.215 exec /bin/sh: exec format error
#
# Docker Desktop and the podman machine ship this emulation, which is why an arm64 Mac
# builds amd64 images happily; a bare arm64 Linux host has nothing registered in
# binfmt_misc and fails on every service.
ensure_amd64_emulation() {
  [[ $SKIP_BUILD -eq 0 ]] || return 0
  [[ "${ARCH}" == "amd64" ]] && return 0

  # Cheapest possible probe: run an amd64 image and see whether the kernel can exec it.
  if "${CONTAINER_TOOL}" run --rm --platform linux/amd64 busybox true >/dev/null 2>&1; then
    return 0
  fi

  log "Host is ${OS}/${ARCH} but images are built for linux/amd64 — installing binfmt emulation..."
  "${CONTAINER_TOOL}" run --privileged --rm tonistiigi/binfmt --install amd64 >/dev/null 2>&1 || true

  if "${CONTAINER_TOOL}" run --rm --platform linux/amd64 busybox true >/dev/null 2>&1; then
    log "  amd64 emulation ready (builds will be slower than native)"
    return 0
  fi

  die "Cannot execute linux/amd64 images on this ${ARCH} host — every image build would
fail with 'exec /bin/sh: exec format error'.

Install emulation manually, then re-run:
  ${CONTAINER_TOOL} run --privileged --rm tonistiigi/binfmt --install amd64

On Debian/Ubuntu the package works too:
  sudo apt-get install -y qemu-user-static binfmt-support

Verify with:
  ${CONTAINER_TOOL} run --rm --platform linux/amd64 busybox true"
}

# The local registry serves plain HTTP on localhost:${REGISTRY_PORT}. Podman rejects
# a non-TLS registry unless it is declared insecure, and its Docker-compat API
# reports that as a bare `500 (Internal Server Error) but provided no error-message`
# on push — no mention of TLS, the registry, or the image. Every service then fails
# at "kind load", which sends you looking at kind, disk or the images instead.
#
# This config lives in the Podman VM, not the repo, so it does not survive
# `podman machine rm` — hence checking on every run rather than documenting it.
ensure_insecure_registry() {
  [[ "${CONTAINER_TOOL}" == "podman" ]] || return 0     # Docker Desktop allows localhost by default
  local conf="/etc/containers/registries.conf.d/100-kind-registry.conf"
  local target="localhost:${REGISTRY_PORT}"

  if podman machine ssh "grep -rqsF '${target}' /etc/containers/registries.conf.d/ /etc/containers/registries.conf" 2>/dev/null; then
    return 0
  fi

  log "Registering ${target} as an insecure registry in the Podman VM..."
  if ! podman machine ssh "printf '[[registry]]\nlocation = \"${target}\"\ninsecure = true\n' | sudo tee ${conf} >/dev/null" 2>/dev/null; then
    warn "Could not write ${conf} in the Podman VM."
    warn "  Pushes to ${target} will fail with an unexplained HTTP 500. Add it manually:"
    warn "    podman machine ssh \"printf '[[registry]]\\nlocation = \\\"${target}\\\"\\ninsecure = true\\n' | sudo tee ${conf}\""
    warn "    podman machine ssh 'sudo systemctl restart podman.socket podman.service'"
    return 0
  fi
  # Restarting drops in-flight connections, so this must happen here in the
  # prerequisite phase — never once builds are running.
  podman machine ssh "sudo systemctl restart podman.socket podman.service" 2>/dev/null || true
  sleep 2
  log "  Insecure registry configured"
}

# Ensure all Kind node containers are running.
# When using Podman, restarting the Podman machine (e.g. after a disk resize)
# stops all containers.  This function detects and restarts exited Kind nodes.
ensure_kind_nodes_running() {
  kind get clusters 2>/dev/null | grep -q "^${KIND_CLUSTER}$" || return 0

  local -a stopped=()
  while IFS= read -r name; do
    stopped+=("${name}")
  done < <("${CONTAINER_TOOL}" ps -a \
    --filter "name=${KIND_CLUSTER}" \
    --filter "status=exited" \
    --format "{{.Names}}" 2>/dev/null)

  if (( ${#stopped[@]} > 0 )); then
    log "Restarting ${#stopped[@]} stopped Kind node(s): ${stopped[*]}"
    "${CONTAINER_TOOL}" start "${stopped[@]}" >/dev/null
    sleep 2
    kind export kubeconfig --name "${KIND_CLUSTER}" 2>/dev/null || true
    log "Kind nodes running"
  fi
}

# ─── Phase 3: Kind Cluster ────────────────────────────────────────────────────
setup_local_registry() {
  # Check if already running.
  if "${CONTAINER_TOOL}" ps --format '{{.Names}}' 2>/dev/null | grep -q "^${REGISTRY_NAME}$"; then
    log "Local registry '${REGISTRY_NAME}' already running"
  else
    # Check if a stopped container with our name exists.
    local existing_id
    existing_id="$("${CONTAINER_TOOL}" ps -a --filter "name=^${REGISTRY_NAME}$" --format '{{.ID}}' 2>/dev/null | head -1)"
    if [[ -n "${existing_id}" ]]; then
      log "Restarting stopped registry container '${REGISTRY_NAME}'..."
      "${CONTAINER_TOOL}" start "${REGISTRY_NAME}" >/dev/null
    else
      log "Starting local image registry on port ${REGISTRY_PORT}..."
      "${CONTAINER_TOOL}" run -d --restart=always \
        -p "${REGISTRY_PORT}:5000" \
        --name "${REGISTRY_NAME}" \
        registry:2 >/dev/null
    fi
  fi
  # Connect registry into the Kind network so nodes resolve it by name.
  "${CONTAINER_TOOL}" network connect kind "${REGISTRY_NAME}" 2>/dev/null || true
  configure_node_registry_hosts
}

# Route localhost:REGISTRY_PORT pulls to the registry container via certs.d/hosts.toml
# (Kind's documented local-registry setup). Unlike registry.mirrors, the certs.d path can
# be re-declared in a v3 drop-in, which GPU nodes need — see restore_gpu_node_registry.
configure_node_registry_hosts() {
  local node
  for node in $(kind get nodes --name "${KIND_CLUSTER}" 2>/dev/null); do
    "${CONTAINER_TOOL}" exec "${node}" sh -c "
      mkdir -p '/etc/containerd/certs.d/localhost:${REGISTRY_PORT}' &&
      printf '[host.\"http://${REGISTRY_NAME}:5000\"]\n' > '/etc/containerd/certs.d/localhost:${REGISTRY_PORT}/hosts.toml'
    " || warn "Could not configure registry mirror on node ${node}"
  done
}

create_cluster() {
  if kind get clusters 2>/dev/null | grep -q "^${KIND_CLUSTER}$"; then
    log "Kind cluster '${KIND_CLUSTER}' already exists — reusing"
    # The cluster can exist while its node containers are stopped — restarting the
    # Podman/Docker VM leaves them exited. `kind export kubeconfig` shells into the
    # control-plane node, so it fails ("can only create exec sessions on running
    # containers") unless the nodes are started first.
    ensure_kind_nodes_running
    kind export kubeconfig --name "${KIND_CLUSTER}"
    return
  fi
  write_cluster_stamp

  log "Creating Kind cluster '${KIND_CLUSTER}'..."

  # On Linux, rootless runtimes need a systemd scope with Delegate=yes so
  # Kind nodes can manage their own cgroups.
  #
  # A SYSTEM scope requires the polkit privilege org.freedesktop.systemd1.manage-units,
  # and the probe itself pops an interactive prompt:
  #   ==== AUTHENTICATING FOR org.freedesktop.systemd1.manage-units ====
  #   Authentication is required to manage system services or other units.
  # Redirecting output does not hide it — polkit's agent talks to the TTY directly — so
  # merely testing for the capability blocked the run with a prompt that never explains
  # which step wants root, or that declining is harmless.
  #
  # A user scope needs no authorisation and is the right placement for a rootless
  # runtime, so probe that unless we are already root. If neither works we run kind
  # unwrapped, exactly as before when the probe failed.
  local kind_run=()
  if command -v systemd-run &>/dev/null; then
    if [[ "$(id -u)" -eq 0 ]]; then
      if systemd-run --scope -p Delegate=yes true &>/dev/null; then
        kind_run=(systemd-run --scope -p Delegate=yes --)
        log "  Using systemd-run --scope -p Delegate=yes for rootless runtime"
      fi
    elif systemd-run --user --scope -p Delegate=yes true &>/dev/null; then
      kind_run=(systemd-run --user --scope -p Delegate=yes --)
      log "  Using systemd-run --user --scope -p Delegate=yes for rootless runtime"
    fi
  fi

  # With accept-nvidia-visible-devices-as-volume-mounts=true on the host, this mount
  # makes the nvidia runtime inject every host GPU into this one node container.
  # Only one worker gets it: each node would otherwise advertise the same GPUs.
  # Not used with --gpu-driver: there is no host driver for the runtime to inject.
  local gpu_worker_mounts=""
  if [[ $GPU_OPERATOR -eq 1 && $GPU_DRIVER -eq 0 ]]; then
    gpu_worker_mounts=$'\n  extraMounts:\n  - hostPath: /dev/null\n    containerPath: /var/run/nvidia-container-devices/all'
  fi

  local kind_config
  kind_config=$(cat <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry]
    config_path = "/etc/containerd/certs.d"
nodes:
- role: control-plane
  image: ${KIND_NODE_IMAGE}
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
- role: worker
  image: ${KIND_NODE_IMAGE}${gpu_worker_mounts}
- role: worker
  image: ${KIND_NODE_IMAGE}
- role: worker
  image: ${KIND_NODE_IMAGE}
- role: worker
  image: ${KIND_NODE_IMAGE}
EOF
)

  if ! echo "${kind_config}" | \
       "${kind_run[@]+"${kind_run[@]}"}" kind create cluster --name "${KIND_CLUSTER}" --config=- 2>&1; then
    die "Kind cluster creation failed (see output above)."
  fi

  log "Kind cluster ready"
}

add_host_entry() {
  local ip="127.0.0.1"
  if grep -qF "${DOMAIN}" /etc/hosts; then
    log "/etc/hosts already has entry for ${DOMAIN} — skipping"
  else
    log "Adding '${ip} ${DOMAIN}' to /etc/hosts (requires sudo)..."
    sudo -- sh -c "echo '${ip} ${DOMAIN}' >> /etc/hosts"
    log "/etc/hosts updated"
  fi

}

install_ingress_nginx() {
  log "Installing ingress-nginx for Kind..."
  kubectl apply -f \
    https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.3/deploy/static/provider/kind/deploy.yaml

  log "Waiting for ingress-nginx controller..."
  kubectl wait --namespace ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=180s

  patch_nginx_workers
  log "ingress-nginx ready"
}

patch_nginx_workers() {
  # The nginx container inside Kind has pids.max=307 (set by Kubernetes cgroup accounting).
  # Each nginx worker spawns 32 aio threads, so each worker occupies 33 pids.
  # During a config reload, old and new workers coexist briefly:
  #   6 workers × 33 pids × 2 generations = 396 pids > 307 → workers crash with
  #   "pthread_create failed (11: Resource temporarily unavailable)"
  # Limiting to 2 workers keeps the reload peak at 2×33×2=132 pids, safely under 307.
  log "Patching ingress-nginx: capping worker-processes=2 to stay within cgroup pids limit..."
  kubectl patch configmap ingress-nginx-controller -n ingress-nginx \
    --type merge \
    -p '{"data":{"worker-processes":"2","max-worker-open-files":"65536","reuse-port":"false"}}'

  # Restart so the new config is picked up from a clean state (no lingering crashed workers).
  # Use rollout status instead of kubectl wait: after a restart, kubectl wait --selector
  # tracks the old pod by name and emits "Error from server (NotFound)" when it is deleted,
  # causing the script to fail under set -e even though the new pod is healthy.
  kubectl rollout restart deployment ingress-nginx-controller -n ingress-nginx
  kubectl rollout status deployment ingress-nginx-controller -n ingress-nginx --timeout=120s
  log "ingress-nginx worker config applied (2 workers, pids-safe)"
}

install_prometheus_operator() {
  log "Installing prometheus-operator..."
  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 2>/dev/null || true
  helm repo update prometheus-community

  # Clear any stuck pending-install/pending-upgrade state from a previously interrupted run
  local status
  status=$(helm status prometheus-operator -n monitoring -o json 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin)['info']['status'])" 2>/dev/null || true)
  if [[ "${status}" == "pending-install" || "${status}" == "pending-upgrade" || "${status}" == "pending-rollback" ]]; then
    warn "Found stuck helm release (${status}), removing before retry..."
    helm delete prometheus-operator -n monitoring 2>/dev/null || true
  fi

  # Install operator + CRDs + a minimal Prometheus instance; no Grafana/Alertmanager.
  helm upgrade --install prometheus-operator prometheus-community/kube-prometheus-stack \
    --namespace monitoring \
    --create-namespace \
    --set grafana.enabled=false \
    --set alertmanager.enabled=false \
    --set nodeExporter.enabled=false \
    --set kubeStateMetrics.enabled=true \
    --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
    --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false \
    --wait \
    --timeout 5m
  log "prometheus-operator ready"
}

install_fake_gpu_operator() {
  log "Installing fake-gpu-operator..."
  helm upgrade --install fake-gpu-operator \
    oci://ghcr.io/run-ai/fake-gpu-operator/fake-gpu-operator \
    --namespace gpu-operator \
    --create-namespace \
    --set gpuOperator.enabled=false \
    --set nvidiaDraDriver.enabled=false \
    --set prometheus.url="http://prometheus-operated.monitoring:9090" \
    --set topology.nodePools.default.gpuCount=8 \
    --set topology.nodePools.default.gpuType=Tesla-A10 \
    --wait \
    --timeout 5m

  # Label worker nodes so the fake-gpu-operator simulates GPUs on them.
  # The default node pool (Tesla-K80, 2 GPUs/node) is defined in the chart's
  # topology.nodePools.default — worker nodes need this label to be recognised.
  log "  Labeling worker nodes for fake GPU simulation..."
  kubectl get nodes --no-headers -o custom-columns="NAME:.metadata.name" 2>/dev/null \
    | grep worker \
    | xargs -I{} kubectl label node {} run.ai/simulated-gpu-node-pool=default --overwrite 2>/dev/null || true

  log "fake-gpu-operator ready"
}

check_gpu_host_prereqs() {
  log "Checking host prerequisites for --gpu-operator..."
  [[ "${OS}" == "linux" ]] || die "--gpu-operator requires a Linux host with NVIDIA GPUs (got ${OS})."
  [[ "${CONTAINER_TOOL}" == "docker" ]] || die "--gpu-operator requires Docker; Podman is not supported for GPU passthrough into Kind."
  grep -qx 0x10de /sys/bus/pci/devices/*/vendor 2>/dev/null \
    || die "No NVIDIA GPU found on this host. Only NVIDIA GPUs are supported; use the default fake GPUs instead."

  if [[ $GPU_DRIVER -eq 1 ]]; then
    [[ ! -d /sys/module/nvidia ]] \
      || die "--gpu-driver: the NVIDIA driver is already loaded on the host. Drop --gpu-driver to use it, or unload it first."
    log "  NVIDIA GPU present, no host driver loaded — the GPU Operator will install the driver"
    return 0
  fi

  nvidia-smi -L &>/dev/null || die "nvidia-smi failed — install the NVIDIA driver on the host, or use --gpu-driver to let the GPU Operator install it."
  command -v nvidia-ctk &>/dev/null || die "nvidia-ctk not found — install the NVIDIA Container Toolkit: https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html"

  # Kind node containers receive GPUs only when Docker's default runtime is nvidia and
  # the runtime honours the /var/run/nvidia-container-devices volume-mount request.
  local default_runtime
  default_runtime="$(docker info --format '{{.DefaultRuntime}}' 2>/dev/null || true)"
  if [[ "${default_runtime}" != "nvidia" ]] || \
     ! grep -qE '^\s*accept-nvidia-visible-devices-as-volume-mounts\s*=\s*true' /etc/nvidia-container-runtime/config.toml 2>/dev/null; then
    log "  Configuring Docker for GPUs in Kind (requires sudo; restarts Docker)..."
    local running
    running="$(docker ps --format '{{.Names}}' 2>/dev/null | paste -sd, - || true)"
    [[ -n "${running}" ]] && warn "  Restarting Docker stops running containers: ${running}"
    if ! { sudo nvidia-ctk runtime configure --runtime=docker --set-as-default >/dev/null && \
           sudo nvidia-ctk config --set accept-nvidia-visible-devices-as-volume-mounts=true --in-place && \
           sudo systemctl restart docker; }; then
      warn "Automatic configuration failed. Run once, then re-run this script:"
      warn "  sudo nvidia-ctk runtime configure --runtime=docker --set-as-default"
      warn "  sudo nvidia-ctk config --set accept-nvidia-visible-devices-as-volume-mounts=true --in-place"
      warn "  sudo systemctl restart docker"
      die "Host is not configured for GPUs in Kind (default runtime: '${default_runtime:-unknown}')."
    fi
    local _i
    for _i in $(seq 1 30); do docker info &>/dev/null && break; sleep 1; done
    default_runtime="$(docker info --format '{{.DefaultRuntime}}' 2>/dev/null || true)"
    [[ "${default_runtime}" == "nvidia" ]] \
      || die "Docker restarted but its default runtime is '${default_runtime:-unknown}', not nvidia. Check /etc/docker/daemon.json."
  fi
  log "  GPUs: $(nvidia-smi -L | wc -l | tr -d ' ') detected, nvidia runtime is Docker default"
}

install_gpu_operator() {
  local gpu_node="${KIND_CLUSTER}-worker"
  local driver_enabled=false wait_minutes=10
  if [[ $GPU_DRIVER -eq 1 ]]; then
    driver_enabled=true
    wait_minutes=30   # driver container compiles kernel modules before anything else starts
  fi
  log "Installing NVIDIA GPU Operator ${GPU_OPERATOR_VERSION} (GPUs on ${gpu_node}, driver.enabled=${driver_enabled})..."

  if [[ $GPU_DRIVER -eq 0 ]] && ! docker exec "${gpu_node}" nvidia-smi -L &>/dev/null; then
    die "No GPUs visible inside ${gpu_node}. The Kind cluster was probably created without --gpu-operator; recreate it: ./local-dev.sh --teardown && ./local-dev.sh --gpu-operator"
  fi

  # The toolkit the operator installs inside the node expects Ubuntu's ldconfig.real, and
  # needs to mount /proc/driver/nvidia itself rather than inherit the outer runtime's mount.
  docker exec "${gpu_node}" sh -c 'ln -sf /sbin/ldconfig /sbin/ldconfig.real; umount -R /proc/driver/nvidia 2>/dev/null || true'

  # NFD sees the host PCI bus from every node, so it would mark all of them as GPU nodes.
  kubectl get nodes --no-headers -o custom-columns=NAME:.metadata.name \
    | grep -vx "${gpu_node}" \
    | xargs -I{} kubectl label node {} nvidia.com/gpu.deploy.operands=false --overwrite >/dev/null

  helm repo add nvidia https://helm.ngc.nvidia.com/nvidia 2>/dev/null || true
  helm repo update nvidia
  helm upgrade --install gpu-operator nvidia/gpu-operator \
    --version "${GPU_OPERATOR_VERSION}" \
    --namespace gpu-operator \
    --create-namespace \
    --set driver.enabled="${driver_enabled}" \
    --wait \
    --timeout 10m

  log "  Waiting for ${gpu_node} to advertise nvidia.com/gpu..."
  local i gpus=""
  for i in $(seq 1 $(( wait_minutes * 6 ))); do
    gpus="$(kubectl get node "${gpu_node}" -o jsonpath='{.status.allocatable.nvidia\.com/gpu}' 2>/dev/null || true)"
    [[ -n "${gpus}" && "${gpus}" != "0" ]] && break
    sleep 10
  done
  [[ -n "${gpus}" && "${gpus}" != "0" ]] || die "GPU Operator installed but ${gpu_node} advertises no GPUs. Check: kubectl get pods -n gpu-operator"
  log "GPU Operator ready — ${gpu_node} advertises ${gpus} GPU(s)"

  restore_gpu_node_registry "${gpu_node}"
}

# The toolkit adds /etc/containerd/conf.d/99-nvidia.toml in config version 3. Once any
# imported file is v3, containerd ignores the node's v2 "io.containerd.grpc.v1.cri"
# section, including registry config_path, so the GPU node can no longer pull
# localhost:5001 images. A v3 drop-in sorting after 99-nvidia.toml restores it. It must
# be written after the toolkit, or it would disable the v2 section before nvidia's exists.
restore_gpu_node_registry() {
  local node="$1"
  log "  Restoring local-registry access on ${node} (toolkit containerd drop-in)..."
  docker exec -i "${node}" sh -s <<'EOF' || die "Could not update containerd config on ${node}"
cat > /etc/containerd/conf.d/zz-runai-registry.toml <<'TOML'
version = 3
[plugins.'io.containerd.cri.v1.images'.registry]
  config_path = '/etc/containerd/certs.d'
TOML
systemctl restart containerd
for i in $(seq 1 30); do crictl info >/dev/null 2>&1 && exit 0; sleep 1; done
exit 1
EOF

  docker exec "${node}" containerd config dump 2>/dev/null \
    | grep -q "config_path = '/etc/containerd/certs.d'" \
    || die "${node}: registry config_path not in effect. Check: docker exec ${node} containerd config dump | grep config_path"
  log "  ${node} pulls from localhost:${REGISTRY_PORT} again"
}

# ─── Build cache ──────────────────────────────────────────────────────────────
# Each built service writes a hash file: "<svc>:<cluster-stamp>:<src-hash>".
# On re-runs we skip build+load when both the source and the target cluster
# are unchanged.  --force-rebuild bypasses the check entirely.
#
# The cluster stamp is written when a NEW kind cluster is created; reusing an
# existing cluster leaves it intact so hashes stay valid across script re-runs.

CLUSTER_STAMP_FILE="${REPO_ROOT}/.local-dev/cluster-stamp"

write_cluster_stamp() {
  mkdir -p "$(dirname "${CLUSTER_STAMP_FILE}")"
  # Use a random token so each new cluster gets a unique stamp.
  printf '%s' "$(uuidgen 2>/dev/null || date +%s%N)" > "${CLUSTER_STAMP_FILE}"
}

get_cluster_stamp() {
  [[ -f "${CLUSTER_STAMP_FILE}" ]] && cat "${CLUSTER_STAMP_FILE}" || echo "none"
}

# Hash the git-tracked content of one or more directories, including any
# uncommitted changes.  openssl is already a required dependency.
compute_dir_hash() {
  local dir
  # git's stderr is left visible and its failure is fatal: this runs before any
  # per-service log exists, so a swallowed error (e.g. a broken ~/.gitconfig) used to
  # surface only as "check the logs" pointing at an empty directory.
  for dir in "$@"; do
    git -C "${REPO_ROOT}" ls-files -s -- "${dir}" || exit
    git -C "${REPO_ROOT}" diff HEAD -- "${dir}" || exit
  done | openssl dgst -sha256 | awk '{print $NF}' \
    || die "git failed while hashing ${*} — see the git error above"
}

should_skip_build() {
  local svc_name="$1"
  local src_hash="$2"
  [[ $FORCE_REBUILD -eq 1 ]] && return 1
  local hash_file="${HASH_DIR}/${svc_name}.hash"
  [[ ! -f "${hash_file}" ]] && return 1
  [[ "$(cat "${hash_file}")" == "${svc_name}:$(get_cluster_stamp):${src_hash}" ]]
}

save_build_hash() {
  local svc_name="$1"
  local src_hash="$2"
  mkdir -p "${HASH_DIR}"
  printf '%s' "${svc_name}:$(get_cluster_stamp):${src_hash}" > "${HASH_DIR}/${svc_name}.hash"
}

# ─── Phase 4: Build ───────────────────────────────────────────────────────────

# Derive SERVICE_NAME from the service directory's Makefile
get_service_name() {
  local svc_dir="$1"
  grep -m1 '^SERVICE_NAME' "${svc_dir}/Makefile" 2>/dev/null \
    | awk -F'=' '{gsub(/[[:space:]]/, "", $2); print $2}' \
    || basename "${svc_dir}"
}

# Run a build step with up to 3 attempts.  Transient Podman socket EOF errors
# (exit 125) fail the docker run/wait call without corrupting any state, so
# retrying is always safe.
retry_build() {
  local attempt
  for attempt in 1 2 3; do
    if "$@"; then
      return 0
    fi
    (( attempt < 3 )) && warn "Build attempt ${attempt} failed — retrying in 10s..." && sleep 10
  done
  return 1
}

build_image_only_service() {
  local svc_dir="$1"
  local image_tag="$2"
  local rebuilt_file="${3:-}"
  local svc_name
  svc_name="$(get_service_name "${svc_dir}")"

  local src_hash
  src_hash="$(compute_dir_hash "${svc_dir}")"
  if should_skip_build "${svc_name}" "${src_hash}"; then
    log "  ${svc_name} — unchanged, skipping"
    return 0
  fi

  local log_file="${LOG_DIR}/${svc_name}.log"
  log "  Building ${svc_name}..."

  local image_ref="${LOCAL_IMAGE_REPO}/${svc_name}:${image_tag}"
  if ! retry_build container_build "${image_ref}" --build-arg TARGETARCH=amd64 "${svc_dir}" \
        >"${log_file}" 2>&1; then
    die "docker build failed for ${svc_name}. See ${log_file}"
  fi
  if ! container_load "${image_ref}" "${log_file}"; then
    die "kind load failed for ${svc_name}. See ${log_file}"
  fi

  [[ -n "${rebuilt_file}" ]] && touch "${rebuilt_file}"
  save_build_hash "${svc_name}" "${src_hash}"
  log "  ${svc_name} → ${image_ref} (pushed to registry)"
}

build_go_service() {
  local svc_dir="$1"
  local image_tag="$2"
  local rebuilt_file="${3:-}"
  local svc_name
  svc_name="$(get_service_name "${svc_dir}")"

  local src_hash
  src_hash="$(compute_dir_hash "${svc_dir}")"
  if should_skip_build "${svc_name}" "${src_hash}"; then
    log "  ${svc_name} — unchanged, skipping"
    return 0
  fi

  local log_file="${LOG_DIR}/${svc_name}.log"
  log "  Building ${svc_name}..."

  # Step 1: Compile amd64 Go binary inside Docker (workspace mounted via Makefile).
  # Use (cd && make) instead of make -C so that $PWD reflects the service dir —
  # base.mk computes WORKSPACE_HOST_PATH from ${PWD}/../.. and make -C doesn't
  # update the shell's $PWD, causing the volume mount to resolve incorrectly.
  # DOCKERHUB_PROXY= pulls the golang builder image directly from Docker Hub.
  local attempt
  for attempt in 1 2 3; do
    if (cd "${svc_dir}" && make build-go-amd \
          OS=linux ARCH=amd64 \
          DOCKERHUB_PROXY= \
          GITHUB_TOKEN= \
          >"${log_file}" 2>&1); then
      break
    fi
    if (( attempt == 3 )); then
      die "build-go-amd failed for ${svc_name}. See ${log_file}"
    fi
    warn "  ${svc_name} build-go-amd attempt ${attempt} failed (transient?) — retrying in 10s..."
    sleep 10
  done

  # Step 2: Build image into local container store, then load it into every Kind node.
  local image_ref="${LOCAL_IMAGE_REPO}/${svc_name}:${image_tag}"
  if ! retry_build container_build "${image_ref}" --build-arg TARGETARCH=amd64 "${svc_dir}" \
        >>"${log_file}" 2>&1; then
    die "docker build failed for ${svc_name}. See ${log_file}"
  fi
  if ! container_load "${image_ref}" "${log_file}"; then
    die "kind load failed for ${svc_name}. See ${log_file}"
  fi

  [[ -n "${rebuilt_file}" ]] && touch "${rebuilt_file}"
  save_build_hash "${svc_name}" "${src_hash}"
  log "  ${svc_name} → ${image_ref} (pushed to registry)"
}

# cli-exposer embeds the runai-cli binary (linux/amd64) and its docs/version
# artifacts in its Docker image. The normal build-go-amd step only compiles
# the cli-exposer server; this function adds the runai-cli compilation step so
# external developers don't need Go installed on the host.
build_cli_exposer() {
  local svc_dir="${REPO_ROOT}/control-plane/cli-exposer"
  local cli_dir="${REPO_ROOT}/control-plane/runai-cli"
  local image_tag="$1"
  local rebuilt_file="${2:-}"

  local src_hash
  src_hash="$(compute_dir_hash "${svc_dir}" "${cli_dir}")"
  if should_skip_build "cli-exposer" "${src_hash}"; then
    log "  cli-exposer — unchanged, skipping"
    return 0
  fi

  local log_file="${LOG_DIR}/cli-exposer.log"
  log "  Building cli-exposer..."

  # Step 1: compile cli-exposer server binary inside Docker (same as build_go_service).
  local attempt
  for attempt in 1 2 3; do
    if (cd "${svc_dir}" && make build-go-amd \
          OS=linux ARCH=amd64 \
          DOCKERHUB_PROXY= \
          GITHUB_TOKEN= \
          >"${log_file}" 2>&1); then
      break
    fi
    if (( attempt == 3 )); then
      die "build-go-amd failed for cli-exposer. See ${log_file}"
    fi
    warn "  cli-exposer build-go-amd attempt ${attempt} failed (transient?) — retrying in 10s..."
    sleep 10
  done

  # Step 2: compile runai-cli for linux/amd64 inside Docker. Only the linux/amd64
  # binary is needed for the Kind cluster; building all platforms (make dist) is
  # reserved for release pipelines.
  local cli_out="${cli_dir}/bin/dist/linux/amd64"
  mkdir -p "${cli_out}"

  # Read CLI version from the version file (RUNAI_CLI_MAJOR/MINOR/PATCH).
  local cli_version="0.0.0-local"
  if [[ -f "${cli_dir}/version" ]]; then
    local major minor patch
    major=$(grep RUNAI_CLI_MAJOR "${cli_dir}/version" | cut -d= -f2)
    minor=$(grep RUNAI_CLI_MINOR "${cli_dir}/version" | cut -d= -f2)
    patch=$(grep RUNAI_CLI_PATCH "${cli_dir}/version" | cut -d= -f2)
    cli_version="${major}.${minor}.${patch}-local"
  fi

  local gopath_vol="${HOME}/.cache/go-build-docker-gopath"
  local gocache_vol="${HOME}/.cache/go-build-docker-gocache"
  mkdir -p "${gopath_vol}" "${gocache_vol}"

  if ! docker run --rm \
        -w /local/control-plane/runai-cli \
        -v "${REPO_ROOT}:/local" \
        -u "$(id -u):$(id -g)" \
        -v "${HOME}/.gitconfig"://.gitconfig \
        -e HOME=/go \
        -e CGO_ENABLED=0 \
        -e GO111MODULE=on \
        -e GOFIPS140=v1.0.0 \
        -e GOOS=linux \
        -e GOARCH=amd64 \
        -e GOPRIVATE="github.com/dsx-ai-factory/run-ai/*" \
        `# Unlike the per-service builds, this one compiles a PACKAGE ("." below)` \
        `# rather than a file list, so Go runs buildvcs and shells out to git.` \
        `# Podman presents the bind-mounted repo as root-owned while the build runs` \
        `# as $(id -u), so git aborts with "detected dubious ownership" and the` \
        `# build dies on "error obtaining VCS status: exit status 128". Declaring` \
        `# the path safe via GIT_CONFIG_* works regardless of HOME (this container` \
        `# sets HOME=/go, so a mounted ~/.gitconfig would not be read anyway).` \
        -e GIT_CONFIG_COUNT=1 \
        -e GIT_CONFIG_KEY_0=safe.directory \
        -e GIT_CONFIG_VALUE_0=/local \
        -v "${gopath_vol}:/go:z" \
        -v "${gocache_vol}:/tmp/.cache:z" \
        -e GOPATH=/go \
        -e GOCACHE=/tmp/.cache \
        "library/golang:$(awk '$1=="go"{print $2;exit}' "${cli_dir}/go.mod")" \
        go build \
          -ldflags "-X github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/version.version=${cli_version} -extldflags -static -s -w" \
          -o bin/dist/linux/amd64/runai \
          . \
        >>"${log_file}" 2>&1; then
    die "runai-cli linux/amd64 build failed for cli-exposer. See ${log_file}"
  fi

  # Step 3: populate cli-exposer/bin/ with CLI artifacts expected by the Dockerfile.
  cp -r "${cli_dir}/bin/dist" "${svc_dir}/bin/dist"
  cp -r "${cli_dir}/docs"     "${svc_dir}/bin/docs"
  cp    "${cli_dir}/version"  "${svc_dir}/bin/version"

  # Step 4: build image and load into Kind.
  local image_ref="${LOCAL_IMAGE_REPO}/cli-exposer:${image_tag}"
  if ! container_build "${image_ref}" --build-arg TARGETARCH=amd64 "${svc_dir}" \
        >>"${log_file}" 2>&1; then
    die "docker build failed for cli-exposer. See ${log_file}"
  fi
  if ! container_load "${image_ref}" "${log_file}"; then
    die "kind load failed for cli-exposer. See ${log_file}"
  fi

  [[ -n "${rebuilt_file}" ]] && touch "${rebuilt_file}"
  save_build_hash "cli-exposer" "${src_hash}"
  log "  cli-exposer → ${image_ref} (pushed to registry)"
}

build_keycloak_extensions() {
  local svc_dir="${REPO_ROOT}/control-plane/keycloak-extensions"
  local rebuilt_file="${1:-}"
  local svc_name="keycloak-extensions"

  local src_hash
  src_hash="$(compute_dir_hash "${svc_dir}")"
  if should_skip_build "${svc_name}" "${src_hash}"; then
    log "  ${svc_name} — unchanged, skipping"
    return 0
  fi

  local log_file="${LOG_DIR}/${svc_name}.log"
  log "  Building ${svc_name} (Gradle + Go + Docker)..."

  # Retried with a longer backoff than retry_build's flat 10s. Maven Central
  # rate-limits per source IP, so on a shared/NAT egress downloads can fail with
  # "Received status code 429 from server". Dependencies in build.gradle are pinned
  # (no dynamic ranges) to avoid the per-build maven-metadata.xml lookups that
  # used to trigger it; a cold cache can still hit it.
  local _kc_attempt _kc_delay=20
  for _kc_attempt in 1 2 3; do
    if (cd "${svc_dir}" && make build-gradle DOCKERHUB_PROXY= >"${log_file}" 2>&1); then
      break
    fi
    if (( _kc_attempt == 3 )); then
      if grep -q "429\|Too Many Requests" "${log_file}" 2>/dev/null; then
        warn "keycloak-extensions: Maven Central returned 429 (rate limited) on every attempt."
        warn "  This is your egress IP being throttled, not a build error. Options:"
        warn "    - wait a few minutes and re-run"
        warn "    - point Gradle at an internal Maven mirror"
      fi
      die "keycloak-extensions Gradle build failed. See ${log_file}"
    fi
    warn "  ${svc_name} build attempt ${_kc_attempt} failed — retrying in ${_kc_delay}s..."
    sleep "${_kc_delay}"
    _kc_delay=$(( _kc_delay * 2 ))
  done

  # Step 2: compile the Go init binary (amd64 only — this is local dev, not a
  # release build) that replaces the old copy-extensions.sh/write-keycloak-
  # conf.sh scripts inside the scratch runtime image. Same pattern as
  # build_go_service's step 1.
  local _kc_go_attempt
  for _kc_go_attempt in 1 2 3; do
    if (cd "${svc_dir}" && make build-go-amd \
          OS=linux ARCH=amd64 \
          DOCKERHUB_PROXY= \
          GITHUB_TOKEN= \
          >>"${log_file}" 2>&1); then
      break
    fi
    if (( _kc_go_attempt == 3 )); then
      die "build-go-amd failed for ${svc_name}. See ${log_file}"
    fi
    warn "  ${svc_name} build-go-amd attempt ${_kc_go_attempt} failed (transient?) — retrying in 10s..."
    sleep 10
  done

  local image_ref="${LOCAL_IMAGE_REPO}/${svc_name}:${CP_IMAGE_TAG}"
  if ! container_build "${image_ref}" --build-arg TARGETARCH=amd64 "${svc_dir}" >>"${log_file}" 2>&1; then
    die "docker build failed for ${svc_name}. See ${log_file}"
  fi
  if ! container_load "${image_ref}" "${log_file}"; then
    die "kind load failed for ${svc_name}. See ${log_file}"
  fi

  [[ -n "${rebuilt_file}" ]] && touch "${rebuilt_file}"
  save_build_hash "${svc_name}" "${src_hash}"
  log "  ${svc_name} → ${image_ref} (pushed to registry)"
}

build_frontend() {
  local svc_dir="${REPO_ROOT}/control-plane/frontend"

  local src_hash
  src_hash="$(compute_dir_hash "${svc_dir}")"
  if should_skip_build "frontend" "${src_hash}"; then
    log "  frontend — unchanged, skipping"
    return 0
  fi

  local log_file="${LOG_DIR}/frontend.log"
  log "  Building frontend (npm + nginx)..."

  # Step 1: Build React app inside Node Docker container (outputs to dist/)
  if ! (cd "${svc_dir}" && make build-js \
        DOCKERHUB_PROXY= \
        >"${log_file}" 2>&1); then
    die "build-js failed for frontend. See ${log_file}"
  fi

  # Step 2: Build nginx image and load into Kind
  local image_ref="${LOCAL_IMAGE_REPO}/frontend:${CP_IMAGE_TAG}"
  if ! container_build "${image_ref}" "${svc_dir}" >>"${log_file}" 2>&1; then
    die "docker build failed for frontend. See ${log_file}"
  fi
  if ! container_load "${image_ref}" "${log_file}"; then
    die "kind load failed for frontend. See ${log_file}"
  fi

  save_build_hash "frontend" "${src_hash}"
  log "  frontend → ${image_ref} (pushed to registry)"
}

# Build an image and (for Docker) load it into the local daemon store.
# Usage: container_build IMAGE_REF [extra docker build flags...] CONTEXT_DIR
#
# Docker:  docker buildx build --platform linux/amd64 --provenance=false --load
# Podman:  podman build --platform linux/amd64 (images land in local store automatically)
#          + re-tags with docker.io/ prefix so containerd inside Kind resolves the
#            short image name correctly (pods reference "runai-local/svc:tag" which
#            containerd normalises to "docker.io/runai-local/svc:tag").
container_build() {
  local image_ref="$1"; shift
  if [[ "${CONTAINER_TOOL}" == "podman" ]]; then
    # Reuse Docker's credential store for registry auth so the user doesn't
    # need to run `podman login` separately for every private registry.
    local authfile_args=()
    if [[ -f "${HOME}/.docker/config.json" ]]; then
      authfile_args=(--authfile "${HOME}/.docker/config.json")
    fi
    podman build --platform linux/amd64 --retry 3 --retry-delay 5s "${authfile_args[@]}" -t "${image_ref}" "$@"
    podman tag "localhost/${image_ref}" "docker.io/${image_ref}" 2>/dev/null || true
  else
    docker buildx build --platform linux/amd64 --provenance=false --load -t "${image_ref}" "$@"
  fi
}

# Load an image into every Kind node.
# When using Podman, use the docker.io/-prefixed tag so containerd imports the
# image under the name that pods reference.
container_load() {
  local image_ref="$1"
  local log_file="$2"
  local registry_ref="localhost:${REGISTRY_PORT}/${image_ref}"

  local attempt
  for attempt in 1 2 3; do
    if [[ "${CONTAINER_TOOL}" == "podman" ]]; then
      podman tag "localhost/${image_ref}" "${registry_ref}" >>"${log_file}" 2>&1 || \
        podman tag "${image_ref}" "${registry_ref}" >>"${log_file}" 2>&1
      if podman push --tls-verify=false "${registry_ref}" >>"${log_file}" 2>&1; then
        return 0
      fi
    else
      docker tag "${image_ref}" "${registry_ref}" >>"${log_file}" 2>&1
      if docker push "${registry_ref}" >>"${log_file}" 2>&1; then
        return 0
      fi
    fi
    (( attempt < 3 )) && sleep $(( attempt * 5 ))
  done
  return 1
}

# Boot the buildx builder before parallel builds so workers don't race to
# create the builder container simultaneously (causes "name already in use").
# No-op when using Podman (no buildx builder needed).
ensure_buildx() {
  [[ "${CONTAINER_TOOL}" == "podman" ]] && return

  # Distinguish "buildx plugin absent" from "buildx present but bootstrap failed".
  # Without the plugin, `docker buildx build --platform ...` (see container_build)
  # falls back to the root docker command, which reports the misleading
  # "unknown flag: --platform" once per service instead of naming the real cause.
  # Docker Desktop bundles buildx; distro packages such as Ubuntu's docker.io do not.
  if ! docker buildx version &>/dev/null; then
    die "docker buildx is not installed, but every image build needs it.

Install the plugin, then re-run:
  Ubuntu/Debian (Docker's official repo — docker.io from the distro has no buildx):
    sudo apt-get install docker-buildx-plugin
  RHEL/Fedora:
    sudo dnf install docker-buildx-plugin
  Manual:
    https://github.com/docker/buildx#manual-download  (install to ~/.docker/cli-plugins/docker-buildx)

Verify with: docker buildx version"
  fi

  if ! docker buildx inspect --bootstrap &>/dev/null; then
    warn "docker buildx bootstrap failed — parallel builds may encounter container name conflicts"
  fi
}

# Run up to PARALLEL_JOBS builds concurrently, collecting errors
run_parallel_builds() {
  local -a build_fns=("$@")
  # FIFO sliding window: as soon as the oldest job finishes, start the next one.
  # The previous batch-barrier approach waited for ALL N jobs before starting the
  # next N, leaving slots idle whenever one job was slower than the rest.
  local -a queue=()
  local any_failed=0

  for fn_call in "${build_fns[@]}"; do
    if (( ${#queue[@]} >= PARALLEL_JOBS )); then
      wait "${queue[0]}" || any_failed=1
      queue=("${queue[@]:1}")
    fi
    eval "${fn_call}" &
    queue+=($!)
  done

  for pid in "${queue[@]+"${queue[@]}"}"; do
    wait "${pid}" || any_failed=1
  done

  if (( any_failed )); then
    die "One or more builds failed. Check logs in ${LOG_DIR}/"
  fi
}

# Pull all unique base images from the given list of Dockerfile dirs in parallel
# so that parallel builds don't race to pull the same image from the registry.
prefetch_base_images() {
  local -a dirs=("$@")
  local -a images=()

  for dir in "${dirs[@]}"; do
    [[ -f "${dir}/Dockerfile" ]] || continue
    local img
    img=$(awk '/^FROM /{print $2; exit}' "${dir}/Dockerfile")
    # Skip scratch and unexpanded shell variables
    [[ -z "${img}" || "${img}" == "scratch" || "${img}" == *'${'* ]] && continue
    images+=("${img}")
  done

  local -a unique=()
  while IFS= read -r img; do unique+=("${img}"); done \
    < <(printf '%s\n' "${images[@]+"${images[@]}"}" | sort -u)

  [[ ${#unique[@]} -eq 0 ]] && return 0
  log "  Pre-pulling ${#unique[@]} base images in parallel..."

  local -a pull_pids=()
  for img in "${unique[@]}"; do
    if [[ "${CONTAINER_TOOL}" == "podman" ]]; then
      local af_args=()
      [[ -f "${HOME}/.docker/config.json" ]] && af_args=(--authfile "${HOME}/.docker/config.json")
      ( podman pull "${af_args[@]+"${af_args[@]}"}" "${img}" >/dev/null 2>&1 || true ) &
    else
      ( docker pull "${img}" >/dev/null 2>&1 || true ) &
    fi
    pull_pids+=($!)
  done
  for pid in "${pull_pids[@]+"${pull_pids[@]}"}"; do wait "${pid}" || true; done
  log "  Base image prefetch done"
}

# Pre-populate the shared Go module cache before parallel builds begin.
# On the first run the cache is empty; 12+ builds racing to download the same
# modules creates network contention and redundant work.  Running go mod download
# once, serially, in a single container fills the cache for everyone.
# Repair a Go cache left root-owned by an earlier run (see the ownership note in
# prefetch_go_modules).  Build containers run as $(id -u), so any root-owned dir
# in the cache makes every subsequent build fail on the first write it needs —
# typically `mkdir /go/pkg/mod/cache/download/golang.org/fips140: permission
# denied`, since GOFIPS140 is only fetched at build time, never by the pre-warm.
# Only Linux is affected; macOS bind mounts squash ownership.
repair_go_cache_ownership() {
  [[ "$(uname -s)" != "Linux" ]] && return 0

  local uid gid stray=""
  uid="$(id -u)"; gid="$(id -g)"

  local dir
  for dir in "$@"; do
    [[ -d "${dir}" ]] || continue
    if [[ -n "$(find "${dir}" ! -user "${uid}" -print -quit 2>/dev/null)" ]]; then
      stray+=" ${dir}"
    fi
  done
  [[ -z "${stray}" ]] && return 0

  warn "Go cache contains files not owned by you — repairing (builds run as ${uid}:${gid})"
  local d
  for d in ${stray}; do
    "${CONTAINER_TOOL}" run --rm -v "${d}:/target:z" \
      busybox \
      chown -R "${uid}:${gid}" /target >/dev/null 2>&1 || \
      warn "  Could not chown ${d} — try: sudo chown -R ${uid}:${gid} ${d}"
  done
}

# Derive this repo's Go module namespace (github.com/<owner>/<repo>) from an in-tree
# go.mod. Never hardcode it — the OSS export target is a parameter of the migration
# script, and this file is copied verbatim into whatever repo it produces.
oss_module_prefix() {
  local gm mod
  for gm in "${REPO_ROOT}"/runai-common-packages/*/go.mod \
            "${REPO_ROOT}"/runai-application-packages/*/go.mod \
            "${REPO_ROOT}"/control-plane/*/go.mod; do
    [[ -f "${gm}" ]] || continue
    mod="$(awk '/^module /{print $2; exit}' "${gm}" 2>/dev/null)"
    [[ -n "${mod}" && "${mod}" != *"{{"* ]] || continue   # skip cookiecutter templates
    printf '%s\n' "${mod}" | cut -d/ -f1-3
    return 0
  done
  return 1
}

# Warn when the checkout predates the current export.
#
# Every export re-points all version tags at fresh commits, so a go.sum from an older
# export describes content the tags no longer serve — builds then die with
# "SECURITY ERROR ... does NOT match an earlier download recorded in go.sum".
# Deliberately advisory: silently moving HEAD out from under someone mid-branch is a
# far worse surprise than a warning.
warn_if_stale_checkout() {
  [[ -f "${REPO_ROOT}/.oss-export-id" ]] || return 0      # not an OSS export
  git -C "${REPO_ROOT}" rev-parse --git-dir >/dev/null 2>&1 || return 0
  local local_head upstream
  local_head="$(git -C "${REPO_ROOT}" rev-parse HEAD 2>/dev/null)" || return 0
  upstream="$(git -C "${REPO_ROOT}" rev-parse '@{upstream}' 2>/dev/null)" || return 0
  [[ "${local_head}" == "${upstream}" ]] && return 0
  # Only complain when we are strictly behind; ahead/diverged is the user's own work.
  if git -C "${REPO_ROOT}" merge-base --is-ancestor "${local_head}" "${upstream}" 2>/dev/null; then
    warn "This checkout is behind its upstream — go.sum here may predate the current tags."
    warn "  If builds fail with 'checksum mismatch', run: git pull --ff-only"
  fi
}

# Self-heal the Go caches after an OSS re-export.
#
# The export re-uses version tags but re-points them at new content, which Go cannot
# detect: module@version is assumed immutable and cached forever. THREE caches must be
# cleared, and missing any one of them still breaks the build:
#   1. cache/vcs/<hash>  - the git clone Go builds module zips from. Holds the OLD
#                          tag -> sha map, so it zips the previous export's content.
#   2. cache/download/   - downloaded zips + their hashes.
#   3. GOCACHE           - per-module file INDEX keyed by module@version; a stale one
#                          yields bogus "undefined: <Symbol>" / "could not import math".
#
# No-op in the source monorepo: there is no .oss-export-id there, and its tags are
# genuinely immutable.
purge_stale_oss_caches() {
  local gopath_vol="$1" gocache_vol="$2"
  local stamp="${REPO_ROOT}/.oss-export-id"
  [[ -f "${stamp}" ]] || return 0

  local prefix
  prefix="$(oss_module_prefix)" || return 0
  # Guard: a malformed prefix must never reach `rm -rf`. Require host/owner/repo.
  [[ "${prefix}" == */*/* && "${prefix}" != */*/*/* ]] || return 0

  local moddir="${gopath_vol}/pkg/mod"
  local marker="${moddir}/.oss-export-id"
  local want have
  want="$(cat "${stamp}" 2>/dev/null)"
  have="$(cat "${marker}" 2>/dev/null || true)"
  [[ -n "${want}" ]] || return 0
  [[ "${want}" == "${have}" ]] && return 0

  # Build containers can leave root-owned entries in these volumes. prefetch_go_modules
  # repairs that, but it runs later and only on the control-plane path, so repair here
  # too — otherwise every removal below fails on a machine where a container ran as root.
  repair_go_cache_ownership "${gopath_vol}" "${gocache_vol}" 2>/dev/null || true

  # Nothing below is fatal. A cache we cannot clear is a slow build or a confusing
  # failure later, but aborting the whole run here is strictly worse — and under
  # `set -e` an unguarded rm/redirect does exactly that.
  local _failed=0
  if [[ -d "${moddir}" ]]; then
    warn "OSS export changed (${have:-none} -> ${want})"
    warn "  Purging cached ${prefix} modules — tags now point at different content."
    local d
    # 1. git clone cache — identified by remote URL, not by path guessing.
    for d in "${moddir}"/cache/vcs/*/; do
      [[ -d "${d}" ]] || continue
      if git -C "${d}" config --get-regexp '^remote\..*\.url$' 2>/dev/null \
           | grep -qiF "${prefix#*/}"; then
        chmod -R u+w "${d}" 2>/dev/null || true
        rm -rf "${d}" 2>/dev/null || _failed=1
      fi
    done
    # 2. downloaded zips + extracted module trees (mode 0444, hence the chmod).
    for d in "${moddir}/${prefix}" "${moddir}/cache/download/${prefix}"; do
      [[ -e "${d}" ]] || continue
      chmod -R u+w "${d}" 2>/dev/null || true
      rm -rf "${d}" 2>/dev/null || _failed=1
    done
  fi
  # 3. build cache — holds the stale module file index.
  if [[ -n "${gocache_vol}" && -d "${gocache_vol}" ]]; then
    chmod -R u+w "${gocache_vol}" 2>/dev/null || true
    rm -rf "${gocache_vol}" 2>/dev/null || _failed=1
  fi

  if [[ "${_failed}" -eq 1 ]]; then
    warn "  Some cached entries could not be removed (permission denied)."
    warn "  Clear them as root, then re-run:"
    warn "    ${CONTAINER_TOOL:-docker} run --rm -v ${gopath_vol%/pkg/mod}:/gp -v ${gocache_vol}:/gc busybox \\"
    warn "      sh -c 'rm -rf /gc/* /gp/pkg/mod/cache/vcs /gp/pkg/mod/${prefix} /gp/pkg/mod/cache/download/${prefix}'"
    return 0                                  # marker not written: retry next run
  fi

  # Record which export these caches now correspond to. Best-effort: Go marks module
  # cache directories read-only, and the volume may be root-owned, so a failure here
  # must not abort the run — it only means the purge repeats on the next run.
  # The subshell matters: a failed redirection is reported by the SHELL before the
  # command runs, so `printf ... 2>/dev/null` does not suppress it — only redirecting
  # the whole compound does.
  mkdir -p "${moddir}" 2>/dev/null || true
  if ! ( printf '%s\n' "${want}" > "${marker}" ) 2>/dev/null; then
    chmod u+w "${moddir}" 2>/dev/null || true
    ( printf '%s\n' "${want}" > "${marker}" ) 2>/dev/null \
      || warn "  Could not record the export stamp — caches will be re-purged next run."
  fi
  return 0
}

prefetch_go_modules() {
  local gopath_vol="${HOME}/.cache/go-build-docker-gopath"
  local gocache_vol="${HOME}/.cache/go-build-docker-gocache"

  repair_go_cache_ownership "${gopath_vol}" "${gocache_vol}"

  # Must run BEFORE the "cache already populated" early-return below — a populated
  # cache is exactly the case that needs eviction after a re-export.
  purge_stale_oss_caches "${gopath_vol}" "${gocache_vol}"

  # Skip if the module cache already has content.
  if [[ -d "${gopath_vol}/pkg/mod/cache/download" && \
        -n "$(ls -A "${gopath_vol}/pkg/mod/cache/download" 2>/dev/null)" ]]; then
    return 0
  fi

  log "First run — pre-warming Go module cache, ${PARALLEL_JOBS} modules at a time (saves redundant downloads across parallel builds)..."
  mkdir -p "${gopath_vol}" "${gocache_vol}"

  # Build a script that runs go mod download in each service directory.
  local -a mod_dirs=()
  for base_dir in "${REPO_ROOT}/control-plane" "${REPO_ROOT}/runai-cluster/cluster"; do
    while IFS= read -r -d '' go_mod; do
      local rel
      rel="$(dirname "${go_mod}")"
      mod_dirs+=("${rel#${REPO_ROOT}/}")
    done < <(find "${base_dir}" -maxdepth 2 -name "go.mod" -print0 2>/dev/null)
  done

  [[ ${#mod_dirs[@]} -eq 0 ]] && return 0

  # The newest Go any of them asks for; the image's Go will not fetch a newer one.
  local go_version
  go_version=$(for rel in "${mod_dirs[@]}"; do awk '$1=="go"{print $2;exit}' "${REPO_ROOT}/${rel}/go.mod"; done | sort -V | tail -1)

  # A cold cache is several GB, so download PARALLEL_JOBS modules at a time (Go locks
  # the shared module cache, so concurrent downloads are safe). Each module prints its
  # path when done, numbered below in completion order, so the step does not look hung.
  # No `all`: that fetches every module version in the dependency graph, ~2-3x more
  # than the modules go.mod requires, which is everything the builds need.
  local cmds i=0 prewarm_start=${SECONDS}
  cmds="printf '%s\n' ${mod_dirs[*]} | xargs -n 1 -P ${PARALLEL_JOBS} sh -c 'cd /repo/\$1 && go mod download 2>/dev/null; echo \$1' _"

  local af_args=()
  [[ "${CONTAINER_TOOL}" == "podman" && -f "${HOME}/.docker/config.json" ]] && \
    af_args=(--authfile "${HOME}/.docker/config.json")

  # Run as the invoking user, exactly like the per-service build containers
  # (build/makefile/base.mk uses `docker run -u $(id -u):$(id -g)`).  On Linux
  # bind mounts preserve real uids: if this container ran as root it would fill
  # ~/.cache/go-build-docker-gopath with root-owned dirs, and every later build
  # would fail with "mkdir /go/pkg/mod/...: permission denied".  macOS hides the
  # bug because Docker Desktop's VM squashes bind-mount ownership.
  # Rootless podman already maps container-root to the host user, so only
  # docker needs the explicit -u.  HOME=/go keeps Go's config writable.
  local user_args=()
  [[ "${CONTAINER_TOOL}" == "docker" ]] && user_args=(-u "$(id -u):$(id -g)" -e HOME=/go)

  "${CONTAINER_TOOL}" run --rm \
    "${af_args[@]+"${af_args[@]}"}" \
    "${user_args[@]+"${user_args[@]}"}" \
    -v "${REPO_ROOT}:/repo:z" \
    -v "${HOME}/.gitconfig":/.gitconfig \
    -v "${gopath_vol}:/go:z" \
    -v "${gocache_vol}:/tmp/.cache:z" \
    -e GIT_CONFIG_GLOBAL=/.gitconfig \
    -e GOPATH=/go \
    -e GOCACHE=/tmp/.cache \
    -e GOPRIVATE="github.com/run-ai/*" \
    -e CGO_ENABLED=0 \
    -e GO111MODULE=on \
    "library/golang:${go_version}" \
    bash -c "${cmds}" 2>/dev/null \
    | while IFS= read -r line; do i=$((i + 1)); log "  [${i}/${#mod_dirs[@]}] ${line}"; done || \
    warn "Module pre-warm incomplete — remaining modules will download during build"

  local prewarm_elapsed=$(( SECONDS - prewarm_start ))
  log "Go module cache warm ($(( prewarm_elapsed / 60 ))m $(( prewarm_elapsed % 60 ))s)"
}

setup_git_token() {
  # Build containers read ~/.gitconfig via GIT_CONFIG_GLOBAL=/.gitconfig (set by the
  # Makefile).  Any auth we configure must land in ~/.gitconfig on the HOST so the
  # container sees it.  We support two strategies:
  #
  #   1. GITHUB_TOKEN set  → HTTPS + token URL rewrite (recommended)
  #   2. No token          → forward the host SSH private key via the GOPATH volume
  #                          (already mounted at /go inside every build container)
  #
  local gopath_dir="${HOME}/.cache/go-build-docker-gopath"
  mkdir -p "${gopath_dir}"

  # Remove url.*github.com* rewrites written by earlier runs before writing new ones.
  # Each token produces its own section keyed by the token, so rotating one leaves
  # several sections claiming the same insteadOf; git then reports "has multiple
  # values" and picks one arbitrarily — possibly a revoked token.
  # `git config --get-regexp` exits 1 when nothing matches, and under `set -o pipefail`
  # that becomes the pipeline's status — so with `set -e` "no rewrites to clean up",
  # the normal case, aborted the whole run silently right after the cluster came up.
  # Hence the explicit `|| true` and `return 0`.
  _purge_github_rewrites() {                 # "$@" = --global | --file <path>
    git config "$@" --name-only --get-regexp '^url\..*github\.com.*\.insteadof$' 2>/dev/null \
      | sed 's/\.insteadof$//' | sort -u \
      | while read -r _s; do
          [[ -n "${_s}" ]] && git config "$@" --remove-section "${_s}" 2>/dev/null || true
        done || true
    return 0
  }
  # Two insteadOf values on one section need --replace-all then --add. Two plain
  # `git config` calls do NOT work: the second REPLACES the first, which silently
  # dropped the https:// rewrite and left only git@ — see the credential-helper note
  # below for why that was so hard to spot.
  _write_rewrite() {                         # $1 = url prefix; rest = --global | --file <path>
    local _pfx="$1"; shift
    git config "$@" --replace-all url."${_pfx}".insteadOf "https://github.com/"
    git config "$@" --add         url."${_pfx}".insteadOf "git@github.com:"
  }

  # Once this repo is public no credentials are needed at all, so probe anonymously
  # first and skip auth entirely.
  #
  # The probe must be genuinely credential-free or it is worthless: with the caller's
  # config it would follow an existing url.insteadOf rewrite and report a private repo
  # as public, after which we would strip the very credentials the build needs.
  # Global and system config are disabled AND credential.helper is cleared via
  # GIT_CONFIG_* (highest precedence) — on macOS, git also reads Xcode's own gitconfig,
  # which enables osxkeychain and answers from the keychain regardless of the other two.
  local _slug _anon
  # `|| true`: oss_module_prefix exits 1 when it finds no in-tree go.mod, which pipefail
  # would turn into an abort here rather than the intended "skip the probe".
  _slug="$(oss_module_prefix 2>/dev/null | sed 's#^github\.com/##' || true)"
  if [[ -n "${_slug}" ]]; then
    if GIT_TERMINAL_PROMPT=0 GIT_ASKPASS=/usr/bin/false \
       GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_SYSTEM=/dev/null \
       GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=credential.helper GIT_CONFIG_VALUE_0= \
       git ls-remote "https://github.com/${_slug}" HEAD >/dev/null 2>&1; then
      _anon=1
    fi
  fi

  if [[ "${_anon:-0}" == "1" ]]; then
    # Public repo: drop stale token rewrites so a revoked token cannot break fetches.
    _purge_github_rewrites --global
    _purge_github_rewrites --file "${gopath_dir}/.gitconfig"
    log "github.com/${_slug} is readable anonymously — no Git credentials needed"
  elif [[ -n "${GITHUB_TOKEN:-}" ]]; then
    # user:password form, NOT https://<token>@github.com/. With only a username git
    # still has no password, so it consults credential.helper — and on a host set up
    # via `gh auth login` that is `!/usr/bin/gh auth git-credential`, which does not
    # exist inside the build image (the container has its own filesystem; only the
    # gitconfig is mounted, never /usr/bin). Every Go build then died with:
    #   /usr/bin/gh: not found
    #   fatal: could not read Username for 'https://github.com'
    # Supplying both fields means no helper is ever consulted. A machine with no
    # helper configured works either way, which is why this hit Linux and not macOS.
    local pfx="https://x-access-token:${GITHUB_TOKEN}@github.com/"
    _purge_github_rewrites --global
    _purge_github_rewrites --file "${gopath_dir}/.gitconfig"
    # ~/.gitconfig — read by Makefile containers via GIT_CONFIG_GLOBAL=/.gitconfig
    _write_rewrite "${pfx}" --global
    # ${gopath_dir}/.gitconfig — read by the runai-cli direct docker run (HOME=/go)
    _write_rewrite "${pfx}" --file "${gopath_dir}/.gitconfig"
    log "Git HTTPS credential written to ~/.gitconfig and GOPATH/.gitconfig"
    local _probe="https://github.com/"
    if [[ "$(git ls-remote --get-url "${_probe}" 2>/dev/null)" == "${_probe}" ]]; then
      warn "HTTPS rewrite for github.com is not in effect — private module fetches will fail."
      warn "  Inspect with: git config --global --get-regexp '^url\\..*insteadof'"
    fi
  else
    # No GITHUB_TOKEN: inject the host SSH key into the containers via the GOPATH
    # volume mount.  The GOPATH dir is mounted as /go inside every container, so a
    # key placed at ${gopath_dir}/.ssh_id is visible at /go/.ssh_id in-container.
    # We configure core.sshCommand in ~/.gitconfig (which the container reads via
    # GIT_CONFIG_GLOBAL) to point git at that key and silence host-key prompts.
    local ssh_key=""
    for candidate in "${HOME}/.ssh/id_ed25519" "${HOME}/.ssh/id_rsa" "${HOME}/.ssh/id_ecdsa"; do
      [[ -f "${candidate}" ]] && ssh_key="${candidate}" && break
    done

    if [[ -n "${ssh_key}" ]]; then
      cp "${ssh_key}" "${gopath_dir}/.ssh_id"
      chmod 600 "${gopath_dir}/.ssh_id"
      local _sshcmd="ssh -i /go/.ssh_id -o StrictHostKeyChecking=accept-new -o UserKnownHostsFile=/dev/null"
      git config --global core.sshCommand "${_sshcmd}"
      git config --file "${gopath_dir}/.gitconfig" core.sshCommand "${_sshcmd}"
      # Go requests github.com/run-ai/* over HTTPS, so without this rewrite the key is
      # never consulted and git falls through to the host credential helper instead —
      # the SSH fallback silently did nothing for module downloads.
      _purge_github_rewrites --global
      _purge_github_rewrites --file "${gopath_dir}/.gitconfig"
      git config --global --replace-all url."git@github.com:".insteadOf "https://github.com/"
      git config --file "${gopath_dir}/.gitconfig" --replace-all url."git@github.com:".insteadOf "https://github.com/"
      log "SSH key ($(basename "${ssh_key}")) forwarded into containers via GOPATH volume"
    else
      warn "GITHUB_TOKEN not set and no SSH key found at ~/.ssh/{id_ed25519,id_rsa,id_ecdsa}"
      warn "Private module downloads (github.com/run-ai/*) will fail inside containers"
    fi
  fi
}

build_all_cp() {
  log "Building control-plane services (${#CP_GO_SERVICES[@]} Go + frontend)..."
  log "Parallel jobs: ${PARALLEL_JOBS} | Logs: ${LOG_DIR}/"
  ensure_kind_nodes_running
  ensure_buildx
  rm -f "${CP_REBUILT_FILE}"

  local -a cp_dirs=()
  for svc in "${CP_GO_SERVICES[@]}"; do
    cp_dirs+=("${REPO_ROOT}/control-plane/${svc}")
  done
  # cp_dirs+=("${REPO_ROOT}/control-plane/keycloak")
  prefetch_base_images "${cp_dirs[@]}"
  prefetch_go_modules

  local -a jobs=()

  for svc in "${CP_GO_SERVICES[@]}"; do
    local svc_dir="${REPO_ROOT}/control-plane/${svc}"
    if [[ ! -d "${svc_dir}" ]]; then
      warn "Skipping ${svc}: directory not found"
      continue
    fi
    if [[ "${svc}" == "cli-exposer" ]]; then
      jobs+=("build_cli_exposer '${CP_IMAGE_TAG}' '${CP_REBUILT_FILE}'")
    else
      jobs+=("build_go_service '${svc_dir}' '${CP_IMAGE_TAG}' '${CP_REBUILT_FILE}'")
    fi
  done

  jobs+=("build_keycloak_extensions '${CP_REBUILT_FILE}'")

  run_parallel_builds "${jobs[@]}"

  if [[ $SKIP_FRONTEND -eq 0 ]]; then
    build_frontend
  else
    warn "Skipping frontend build (--skip-frontend)"
  fi

  log "Control-plane images pushed to registry"
}

build_all_rc() {
  log "Building runai-cluster services (${#RC_GO_SERVICES[@]} Go)..."
  log "Parallel jobs: ${PARALLEL_JOBS} | Logs: ${LOG_DIR}/"
  ensure_kind_nodes_running
  ensure_buildx
  rm -f "${RC_REBUILT_FILE}"

  local -a rc_dirs=()
  for svc in "${RC_GO_SERVICES[@]}"; do
    rc_dirs+=("${REPO_ROOT}/runai-cluster/cluster/${svc}")
  done
  prefetch_base_images "${rc_dirs[@]}"

  local -a jobs=()
  for svc in "${RC_GO_SERVICES[@]}"; do
    local svc_dir="${REPO_ROOT}/runai-cluster/cluster/${svc}"
    if [[ ! -d "${svc_dir}" ]]; then
      warn "Skipping ${svc}: directory not found"
      continue
    fi
    # Library-only modules (no Makefile) are compiled by their consumers.
    if [[ ! -f "${svc_dir}/Makefile" ]]; then
      warn "Skipping ${svc}: no Makefile (library module, not a service)"
      continue
    fi
    jobs+=("build_go_service '${svc_dir}' '${RC_IMAGE_TAG}' '${RC_REBUILT_FILE}'")
  done

  run_parallel_builds "${jobs[@]}"

  # init-ca Makefile uses SERVICE_NAME=init-ca but the operator's SDK hardcodes
  # "init-ca-base" as the image name. Push both names to the local registry so
  # the cluster chart (which pulls from localhost:REGISTRY_PORT/runai-local/...) finds it.
  local init_ca_ref="${LOCAL_IMAGE_REPO}/init-ca:${RC_IMAGE_TAG}"
  local init_ca_base_ref="${LOCAL_IMAGE_REPO}/init-ca-base:${RC_IMAGE_TAG}"
  if "${CONTAINER_TOOL}" image inspect "${init_ca_ref}" &>/dev/null || \
     "${CONTAINER_TOOL}" image inspect "localhost/${init_ca_ref}" &>/dev/null; then
    if [[ "${CONTAINER_TOOL}" == "podman" ]]; then
      podman tag "localhost/${init_ca_ref}" "${init_ca_base_ref}" 2>/dev/null || \
        podman tag "${init_ca_ref}" "${init_ca_base_ref}" 2>/dev/null || true
    else
      "${CONTAINER_TOOL}" tag "${init_ca_ref}" "${init_ca_base_ref}"
    fi
    container_load "${init_ca_base_ref}" "/dev/null"
    log "  init-ca-base alias pushed to registry"
  fi

  log "runai-cluster images pushed to registry"
}

# ─── Phase 5: TLS Certificate ─────────────────────────────────────────────────
generate_tls() {
  if [[ -f "${CERT_DIR}/tls.crt" && -f "${CERT_DIR}/tls.key" ]]; then
    # Reuse only if the cert actually covers the current domain.
    local san
    san=$(openssl x509 -noout -text -in "${CERT_DIR}/tls.crt" 2>/dev/null \
          | grep -o "DNS:[^,]*" | grep -o "[^:]*$" | tr '\n' ' ')
    if echo "${san}" | grep -qF "${DOMAIN}"; then
      log "TLS certs already exist in ${CERT_DIR} — reusing"
      return
    fi
    log "Existing TLS cert does not cover '${DOMAIN}' (covers: ${san}) — regenerating..."
    rm -f "${CERT_DIR}"/tls.* "${CERT_DIR}"/ca.*
  fi

  log "Generating self-signed TLS certificate for ${DOMAIN}..."
  mkdir -p "${CERT_DIR}"

  # CA
  openssl req -x509 -nodes -days 3650 -newkey rsa:4096 \
    -keyout "${CERT_DIR}/ca.key" \
    -out    "${CERT_DIR}/ca.crt" \
    -subj   "/CN=RunAI Local CA/O=RunAI Dev" \
    2>/dev/null

  # Server key + CSR
  openssl req -nodes -newkey rsa:2048 \
    -keyout "${CERT_DIR}/tls.key" \
    -out    "${CERT_DIR}/tls.csr" \
    -subj   "/CN=${DOMAIN}" \
    2>/dev/null

  # Sign with CA (include wildcard SAN for subdomains)
  openssl x509 -req -days 3650 \
    -in     "${CERT_DIR}/tls.csr" \
    -CA     "${CERT_DIR}/ca.crt" \
    -CAkey  "${CERT_DIR}/ca.key" \
    -CAcreateserial \
    -out    "${CERT_DIR}/tls.crt" \
    -extfile <(printf "subjectAltName=DNS:%s,DNS:*.%s,IP:127.0.0.1" \
                      "${DOMAIN}" "${DOMAIN}") \
    2>/dev/null

  log "TLS certificate generated in ${CERT_DIR}/"
}

# ─── Phase 6: Deploy ──────────────────────────────────────────────────────────
create_namespace_and_secrets() {
  # Control-plane namespace: TLS secret for ingress + CA secret
  kubectl create namespace "${CP_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
  kubectl create secret tls runai-backend-tls \
    --cert="${CERT_DIR}/tls.crt" \
    --key="${CERT_DIR}/tls.key" \
    --namespace="${CP_NAMESPACE}" \
    --dry-run=client -o yaml | kubectl apply -f -
  kubectl create secret generic runai-ca-cert \
    --from-file=runai-ca.pem="${CERT_DIR}/ca.crt" \
    --namespace="${CP_NAMESPACE}" \
    --dry-run=client -o yaml | kubectl apply -f -

  # runai-cluster namespace: CA secret so the cluster can verify CP's TLS
  kubectl create namespace "${RC_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
  kubectl create secret generic runai-ca-cert \
    --from-file=runai-ca.pem="${CERT_DIR}/ca.crt" \
    --namespace="${RC_NAMESPACE}" \
    --dry-run=client -o yaml | kubectl apply -f -

  # The operator hardcodes runai-reg-creds as an imagePullSecret on every pod.
  # Create a placeholder so kubelet stops warning about a missing secret.
  kubectl create secret docker-registry runai-reg-creds \
    --docker-server=runai-local \
    --docker-username=local \
    --docker-password=local \
    --namespace="${RC_NAMESPACE}" \
    --dry-run=client -o yaml | kubectl apply -f -

  log "Namespaces and TLS secrets ready (${CP_NAMESPACE}, ${RC_NAMESPACE})"
}

deploy_control_plane() {
  log "Deploying control-plane chart..."

  local cp_chart="${REPO_ROOT}/control-plane/control-plane-chart"
  local chart_yaml="${cp_chart}/Chart.yaml"

  # Patch Chart.yaml appVersion to match the image tag we built.
  local orig_app_version
  orig_app_version="$(grep '^appVersion:' "${chart_yaml}" | awk '{print $2}')"
  if [[ "${orig_app_version}" != "${CP_IMAGE_TAG}" ]]; then
    if [[ "${OS}" == "darwin" ]]; then
      sed -i '' "s/^appVersion:.*/appVersion: ${CP_IMAGE_TAG}/" "${chart_yaml}"
    else
      sed -i "s/^appVersion:.*/appVersion: ${CP_IMAGE_TAG}/" "${chart_yaml}"
    fi
    log "Patched Chart.yaml appVersion: ${orig_app_version} → ${CP_IMAGE_TAG}"
    trap "cd '${cp_chart}' && \
      if [[ '${OS}' == 'darwin' ]]; then \
        sed -i '' 's/^appVersion:.*/appVersion: ${orig_app_version}/' Chart.yaml; \
      else \
        sed -i 's/^appVersion:.*/appVersion: ${orig_app_version}/' Chart.yaml; \
      fi && cd '${REPO_ROOT}'" EXIT
  fi

  # Clear any stuck pending-install/upgrade state from a previously interrupted run.
  local cp_status
  cp_status=$(helm status runai-backend -n "${CP_NAMESPACE}" -o json 2>/dev/null \
    | python3 -c "import sys,json; print(json.load(sys.stdin)['info']['status'])" 2>/dev/null || true)
  if [[ "${cp_status}" == "pending-install" || "${cp_status}" == "pending-upgrade" || "${cp_status}" == "pending-rollback" ]]; then
    warn "Found stuck helm release (${cp_status}), removing before retry..."
    helm delete runai-backend -n "${CP_NAMESPACE}" 2>/dev/null || true
    cp_status=""
  fi

  # On a fresh install (no prior release), stale PVCs from an old run would cause
  # bitnami postgresql to fail — its stored password won't match the newly generated
  # secret.  Delete any leftover PVCs so postgres initialises clean.
  if [[ -z "${cp_status}" ]]; then
    local stale_pvcs
    stale_pvcs=$(kubectl get pvc -n "${CP_NAMESPACE}" --no-headers -o name 2>/dev/null || true)
    if [[ -n "${stale_pvcs}" ]]; then
      warn "Deleting stale PVCs in ${CP_NAMESPACE} before fresh install..."
      kubectl delete pvc -n "${CP_NAMESPACE}" --all 2>/dev/null || true
    fi
  fi

  # On upgrades (secrets already exist), skip helm hooks — migrator jobs use
  # pullPolicy:Always which breaks in kind, and password-generator hooks would
  # overwrite existing secrets making Keycloak credentials inconsistent.
  # On a fresh install (no secrets yet), let hooks run so secrets get created.
  local cp_no_hooks=""
  if kubectl get secret runai-backend-gen-keycloakx -n "${CP_NAMESPACE}" &>/dev/null; then
    cp_no_hooks="--no-hooks"
    log "Existing secrets detected — skipping helm hooks (passwords preserved)"
    # Patch the tenants-manager secret so ADMIN_PASSWORD matches an explicitly requested
    # value. When hooks are skipped the secret keeps its old value; tenants-manager
    # creates the admin user with whatever is in this secret, so we must keep them in sync.
    if [[ -n "${ADMIN_PASSWORD}" ]]; then
      local encoded_pw
      encoded_pw="$(printf '%s' "${ADMIN_PASSWORD}" | base64)"
      kubectl patch secret runai-backend-gen-tenants-manager \
        -n "${CP_NAMESPACE}" \
        --type=merge \
        -p "{\"data\":{\"ADMIN_PASSWORD\":\"${encoded_pw}\"}}" \
        2>/dev/null || true
    fi
  fi

  local password_args=()
  if [[ -n "${ADMIN_PASSWORD}" ]]; then
    password_args=(
      --set-string "global.management.password=${ADMIN_PASSWORD}"
      --set-string "tenantsManager.config.adminPassword=${ADMIN_PASSWORD}"
    )
  fi

  # Skip helm dep build when all tarballs are already present locally.
  # helm upgrade --install reads charts/ tarballs directly — no dep build needed.
  # Also checks that Chart.lock is not stale relative to Chart.yaml: if Chart.yaml
  # lists a dependency absent from Chart.lock, the lock needs to be regenerated.
  if ! python3 - "${cp_chart}/Chart.yaml" "${cp_chart}/Chart.lock" "${cp_chart}/charts" <<'PYEOF' 2>/dev/null
import sys, re, os
chart_yaml_file, lock_file, charts_dir = sys.argv[1], sys.argv[2], sys.argv[3]
if not os.path.exists(lock_file):
    sys.exit(1)
# Extract dependency names from Chart.yaml to detect stale Chart.lock.
with open(chart_yaml_file) as f:
    yaml_content = f.read()
yaml_dep_names = set(re.findall(r'^\s*-\s*name:\s*(\S+)', yaml_content, re.MULTILINE))
with open(lock_file) as f:
    lock_content = f.read()
lock_dep_names = set(re.findall(r'name:\s*(\S+)', lock_content))
stale = yaml_dep_names - lock_dep_names
if stale:
    print(f"Chart.lock is stale — missing deps: {stale}", file=sys.stderr)
    sys.exit(1)
blocks = re.split(r'\n- ', lock_content)
for block in blocks[1:]:
    name = re.search(r'name:\s*(\S+)', block)
    version = re.search(r'version:\s*(\S+)', block)
    if not (name and version):
        continue
    tarball = f"{name.group(1)}-{version.group(1)}.tgz"
    if not os.path.exists(os.path.join(charts_dir, tarball)):
        print(f"Missing: {tarball}", file=sys.stderr)
        sys.exit(1)
PYEOF
  then
    log "Chart tarballs missing — fetching dependencies..."
    helm repo add bitnami       https://charts.bitnami.com/bitnami             2>/dev/null || true
    helm repo add codecentric   https://codecentric.github.io/helm-charts       2>/dev/null || true
    helm repo add unleash       https://docs.getunleash.io/helm-charts           2>/dev/null || true
    helm dependency build "${cp_chart}" 2>&1 \
      | grep -v '^Saving\|^Deleting\|^$' || true
  else
    log "All chart tarballs present locally — skipping helm dep build"
  fi

  local is_fresh_install=0
  [[ -z "${cp_status}" ]] && is_fresh_install=1

  helm upgrade --install runai-backend "${cp_chart}" \
    --namespace "${CP_NAMESPACE}" \
    --create-namespace \
    -f "${REPO_ROOT}/.local-dev/values-cp.yaml" \
    --set global.domain="${DOMAIN}" \
    --set global.image.registry="localhost:${REGISTRY_PORT}/${LOCAL_IMAGE_REPO}" \
    --set global.image.pullPolicy=IfNotPresent \
    --set global.ingress.ingressClass=nginx \
    --set global.ingress.tlsSecretName=runai-backend-tls \
    --set global.customCA.enabled=true \
    --set "global.management.user=${ADMIN_EMAIL}" \
    --set "tenantsManager.config.adminUsername=${ADMIN_EMAIL}" \
    ${password_args[@]+"${password_args[@]}"} \
    --set global.config.kubernetesDistribution=kubernetes \
    --set nats.jetstream.maxMemory=4G \
    --set nats.resources.requests.cpu=250m \
    --set nats.resources.requests.memory=2Gi \
    --set nats.resources.limits.cpu=500m \
    --set nats.resources.limits.memory=4Gi \
    --set unleashEdge.enabled=false \
    --set global.image.tag="${CP_IMAGE_TAG}" \
    ${cp_no_hooks} \
    --timeout 6m

  # On upgrades (REVISION > 1) all local images share the same tag (0.0.0) so
  # helm has no spec change to trigger a pod restart — do it explicitly.
  # On a fresh install (REVISION 1) pods are created new by helm and start on
  # their own; a rollout restart here would just churn them unnecessarily.
  if [[ "${is_fresh_install}" -eq 0 ]]; then
    log "Restarting CP pods (upgrade — image tag unchanged, forcing rollout)..."
    kubectl rollout restart deployment -n "${CP_NAMESPACE}" 2>/dev/null || true
    kubectl rollout restart statefulset -n "${CP_NAMESPACE}" 2>/dev/null || true
  fi

  log "Control-plane deployed successfully"
  resolve_admin_password || die "Could not read ADMIN_PASSWORD from secret runai-backend-gen-tenants-manager in ${CP_NAMESPACE}"
  wait_for_cp_health
}

# Fills ADMIN_PASSWORD from the tenants-manager secret when it wasn't passed explicitly.
resolve_admin_password() {
  [[ -n "${ADMIN_PASSWORD}" ]] && return 0
  ADMIN_PASSWORD="$(kubectl get secret runai-backend-gen-tenants-manager \
    -n "${CP_NAMESPACE}" -o jsonpath='{.data.ADMIN_PASSWORD}' 2>/dev/null \
    | base64 --decode 2>/dev/null || true)"
  [[ -n "${ADMIN_PASSWORD}" ]]
}

# Fetches an access token for ADMIN_EMAIL: the Keycloak master admin impersonates
# the user, then runs the authorization code + PKCE flow of a browser login client
# of the tenant realm (the realm has no password-grant client). Prints the token on
# stdout and returns non-zero if Keycloak or the realm is not (yet) available.
get_cp_token() {
  local kc="https://${DOMAIN}/auth"
  local kc_admin_password admin_token user_id client_info client_id client_uuid redirect_uri
  local jar verifier challenge location code token
  local secret_args=()

  kc_admin_password="$(kubectl get secret runai-backend-gen-keycloakx -n "${CP_NAMESPACE}" \
    -o jsonpath='{.data.KEYCLOAK_ADMIN_PASSWORD}' 2>/dev/null | base64 --decode 2>/dev/null || true)"
  [[ -n "${kc_admin_password}" ]] || return 1

  admin_token="$(curl --insecure --silent --fail --max-time 60 \
    -X POST "${kc}/realms/master/protocol/openid-connect/token" \
    --data-urlencode 'grant_type=password' \
    --data-urlencode 'client_id=admin-cli' \
    --data-urlencode "username=${KC_ADMIN_USER}" \
    --data-urlencode "password=${kc_admin_password}" \
    | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])" 2>/dev/null)" || return 1
  [[ -n "${admin_token}" ]] || return 1

  # The user exists only once tenants-manager has provisioned the realm.
  user_id="$(curl --insecure --silent --fail --max-time 60 \
    -G "${kc}/admin/realms/${CP_TENANT_NAME}/users" \
    --data-urlencode 'exact=true' \
    --data-urlencode "username=${ADMIN_EMAIL}" \
    -H "Authorization: Bearer ${admin_token}" \
    | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['id'])" 2>/dev/null)" || return 1

  # Browser login client that redirects back to the domain: prefer runai-cli, then
  # runai-admin-ui. Prints "clientId|uuid (confidential clients only)|redirect_uri".
  client_info="$(curl --insecure --silent --fail --max-time 60 \
    "${kc}/admin/realms/${CP_TENANT_NAME}/clients?max=500" \
    -H "Authorization: Bearer ${admin_token}" \
    | python3 -c "
import sys, json
base = 'https://${DOMAIN}'
def redirect(c):
    for u in c.get('redirectUris') or []:
        if u == base or u.startswith(base + '/'):
            return u.rstrip('*')
cs = [c for c in json.load(sys.stdin)
      if c.get('enabled') and c.get('standardFlowEnabled')
      and not c['clientId'].startswith('runai-internal.') and redirect(c)]
cs.sort(key=lambda c: {'runai-cli': 0, 'runai-admin-ui': 1}.get(c['clientId'], 2))
c = cs[0]
print(c['clientId'], '' if c.get('publicClient') else c['id'], redirect(c), sep='|')
" 2>/dev/null)" || return 1
  IFS='|' read -r client_id client_uuid redirect_uri <<< "${client_info}"

  if [[ -n "${client_uuid}" ]]; then
    secret_args=(--data-urlencode "client_secret=$(curl --insecure --silent --fail --max-time 60 \
      "${kc}/admin/realms/${CP_TENANT_NAME}/clients/${client_uuid}/client-secret" \
      -H "Authorization: Bearer ${admin_token}" \
      | python3 -c "import sys,json; print(json.load(sys.stdin)['value'])" 2>/dev/null)")
  fi

  read -r verifier challenge < <(python3 -c "
import base64, hashlib, secrets
v = secrets.token_urlsafe(48)
print(v, base64.urlsafe_b64encode(hashlib.sha256(v.encode()).digest()).rstrip(b'=').decode())
")

  # Impersonation sets the user's SSO session cookies, so the authorization request
  # redirects straight back with a code instead of showing the login page.
  jar="$(mktemp)"
  curl --insecure --silent --max-time 60 -o /dev/null -c "${jar}" -b "${jar}" \
    -X POST "${kc}/admin/realms/${CP_TENANT_NAME}/users/${user_id}/impersonation" \
    -H "Authorization: Bearer ${admin_token}" || true
  location="$(curl --insecure --silent --max-time 60 -o /dev/null -c "${jar}" -b "${jar}" \
    -w '%{redirect_url}' \
    -G "${kc}/realms/${CP_TENANT_NAME}/protocol/openid-connect/auth" \
    --data-urlencode "client_id=${client_id}" \
    --data-urlencode "redirect_uri=${redirect_uri}" \
    --data-urlencode 'response_type=code' \
    --data-urlencode 'scope=openid' \
    --data-urlencode 'state=local-dev' \
    --data-urlencode "code_challenge=${challenge}" \
    --data-urlencode 'code_challenge_method=S256')" || location=""
  rm -f "${jar}"
  code="$(printf '%s' "${location}" | sed -n 's/.*[?&#]code=\([^&]*\).*/\1/p')"
  [[ -n "${code}" ]] || return 1

  token="$(curl --insecure --silent --fail --max-time 60 \
    -X POST "${kc}/realms/${CP_TENANT_NAME}/protocol/openid-connect/token" \
    --data-urlencode 'grant_type=authorization_code' \
    --data-urlencode "code=${code}" \
    --data-urlencode "redirect_uri=${redirect_uri}" \
    --data-urlencode "client_id=${client_id}" \
    ${secret_args[@]+"${secret_args[@]}"} \
    --data-urlencode "code_verifier=${verifier}" \
    | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])" 2>/dev/null)" || return 1
  [[ -n "${token}" ]] || return 1
  printf '%s' "${token}"
}

# /_version is an authenticated API — probing it requires a bearer token, so this
# also implicitly waits for the Keycloak realm to be provisioned.
cp_version_ready() {
  local token
  token="$(get_cp_token)" || return 1
  curl --insecure --silent --fail --max-time 60 \
    -H "Authorization: Bearer ${token}" \
    "https://${DOMAIN}/_version" >/dev/null 2>&1
}

wait_for_cp_health() {
  log "Waiting for control plane at https://${DOMAIN}/_version..."
  local attempt=0
  until cp_version_ready; do
    attempt=$(( attempt + 1 ))
    if (( attempt >= 60 )); then
      die "Control plane did not become healthy after ~10 min. Check pods in ${CP_NAMESPACE}."
    fi
    log "  /_version not ready yet (attempt ${attempt}/60), retrying in 10s..."
    sleep 10
  done
  log "Control plane is healthy"
}

# Registers the cluster in the CP via the REST API and mints the one-time
# installer bootstrap token.  Sets CLUSTER_UUID and CLUSTER_BOOTSTRAP_TOKEN.
provision_cluster_credentials() {
  log "Provisioning cluster credentials via CP API..."

  # Keycloak realm provisioning is async — wait until the token endpoint succeeds.
  local token
  local attempt=0
  log "Waiting for Keycloak realm '${CP_TENANT_NAME}' to be ready..."
  until token="$(get_cp_token)"; do
    attempt=$(( attempt + 1 ))
    if (( attempt >= 60 )); then
      die "Failed to obtain Keycloak token after ~10 min — is the control plane healthy?"
    fi
    log "  Keycloak realm not ready yet (attempt ${attempt}/60), retrying in 10s..."
    sleep 10
  done
  log "Keycloak realm ready"

  # Create the cluster (idempotent: reuse if already registered by name).
  local existing_uuid
  existing_uuid="$(curl --insecure --silent --fail --max-time 60 \
    "https://${DOMAIN}/v1/k8s/clusters" \
    -H 'accept: application/json' \
    -H "Authorization: Bearer ${token}" \
    | python3 -c "
import sys, json
cs = [c for c in json.load(sys.stdin) if c.get('name') == '${CP_CLUSTER_NAME}']
print(cs[0]['uuid'] if cs else '')
" 2>/dev/null)" || existing_uuid=""

  if [[ -n "${existing_uuid}" ]]; then
    CLUSTER_UUID="${existing_uuid}"
    log "  Cluster '${CP_CLUSTER_NAME}' already registered (uuid: ${CLUSTER_UUID})"
  else
    # Retry: backend services may still be warming up even after /_version returns.
    # Refresh the token every 10 attempts (~50s) to stay ahead of the default
    # Keycloak 5-minute token TTL — without this, a slow warm-up causes silent 401s.
    local reg_attempt=0
    until CLUSTER_UUID="$(curl --insecure --silent --fail --max-time 60 \
        -X POST "https://${DOMAIN}/v1/k8s/clusters" \
        -H 'accept: application/json' \
        -H "Authorization: Bearer ${token}" \
        -H 'Content-Type: application/json' \
        -d "{\"name\": \"${CP_CLUSTER_NAME}\"}" \
        | python3 -c "import sys,json; print(json.load(sys.stdin)['uuid'])" 2>/dev/null)" \
        && [[ -n "${CLUSTER_UUID}" ]]; do
      reg_attempt=$(( reg_attempt + 1 ))
      if (( reg_attempt >= 36 )); then
        die "Failed to register cluster '${CP_CLUSTER_NAME}' in control plane after 3 min"
      fi
      if (( reg_attempt % 10 == 0 )); then
        local refreshed
        refreshed="$(get_cp_token)" || refreshed=""
        [[ -n "${refreshed}" ]] && token="${refreshed}"
      fi
      sleep 5
    done
    log "  Cluster '${CP_CLUSTER_NAME}' registered (uuid: ${CLUSTER_UUID})"
  fi

  # Mint the installer bootstrap token for this cluster (a new one on every call).
  local attempt=0
  while true; do
    CLUSTER_BOOTSTRAP_TOKEN="$(curl --insecure --silent --fail --max-time 60 \
      -G "https://${DOMAIN}/api/v1/clusters/${CLUSTER_UUID}/cluster-install-info" \
      --data-urlencode "version=${RC_IMAGE_TAG}" \
      -H "Authorization: Bearer ${token}" \
      | python3 -c "import sys,json; print(json.load(sys.stdin).get('bootstrapToken',''))" 2>/dev/null)" \
      || CLUSTER_BOOTSTRAP_TOKEN=""
    [[ -n "${CLUSTER_BOOTSTRAP_TOKEN}" ]] && break
    attempt=$(( attempt + 1 ))
    (( attempt >= 30 )) && die "Timed out waiting for installer bootstrap token from control plane"
    if (( attempt % 10 == 0 )); then
      local refreshed
      refreshed="$(get_cp_token)" || refreshed=""
      [[ -n "${refreshed}" ]] && token="${refreshed}"
    fi
    log "  Waiting for installer bootstrap token (${attempt}/30)..."
    sleep 5
  done

  log "Cluster credentials provisioned"
}

deploy_runai_cluster() {
  log "Deploying runai-cluster chart..."

  local rc_chart="${REPO_ROOT}/runai-cluster/cluster/runai-chart"
  local chart_yaml="${rc_chart}/Chart.yaml"
  if [[ "${OS}" == "darwin" ]]; then
    sed -i '' "s/^appVersion:.*/appVersion: ${RC_IMAGE_TAG}/" "${chart_yaml}"
  else
    sed -i "s/^appVersion:.*/appVersion: ${RC_IMAGE_TAG}/" "${chart_yaml}"
  fi

  # Patch CoreDNS so pods inside the Kind cluster can resolve runai.local.
  # Use the ingress-nginx ClusterIP so that kube-proxy on each node handles DNAT+SNAT
  # properly — pods receive responses from the same IP they connected to.
  # (Node IP via hostPort breaks because nginx responds directly through the pod overlay,
  # bypassing the conntrack entry that would un-DNAT the reply.)
  local ingress_cluster_ip
  ingress_cluster_ip="$(kubectl get svc ingress-nginx-controller -n ingress-nginx \
    -o jsonpath='{.spec.clusterIP}' 2>/dev/null || true)"
  if [[ -n "${ingress_cluster_ip}" ]]; then
    local current_corefile new_corefile
    current_corefile="$(kubectl get configmap coredns -n kube-system \
      -o jsonpath='{.data.Corefile}' 2>/dev/null || true)"
    if [[ "${current_corefile}" != *"${DOMAIN}:53"* ]]; then
      new_corefile="${current_corefile}
${DOMAIN}:53 {
    hosts {
        ${ingress_cluster_ip} ${DOMAIN}
        fallthrough
    }
    cache 30
}
"
      kubectl patch configmap coredns -n kube-system --type merge \
        -p "{\"data\":{\"Corefile\":$(printf '%s' "${new_corefile}" | python3 -c 'import sys,json; print(json.dumps(sys.stdin.read()))')}}" \
        2>/dev/null
      log "  CoreDNS patched: ${DOMAIN} → ${ingress_cluster_ip} (ingress-nginx ClusterIP)"
      # Restart CoreDNS so it picks up the new Corefile immediately — without a
      # restart the reload plugin polls every ~30s, which causes DNS timeouts
      # during the runai-cluster helm install that follows.
      kubectl rollout restart deployment/coredns -n kube-system 2>/dev/null || true
      kubectl rollout status deployment/coredns -n kube-system --timeout=30s 2>/dev/null || true
    else
      # Upgrade path: config already present — restart to force immediate reload.
      kubectl rollout restart deployment/coredns -n kube-system 2>/dev/null || true
      kubectl rollout status deployment/coredns -n kube-system --timeout=30s 2>/dev/null || true
      log "  CoreDNS already configured for ${DOMAIN}, restarted to reload"
    fi
  fi

  # runai-scale-adjust is created by the chart (kai-resource-management.enabled=true). If it exists
  # without helm ownership labels (leftover from a failed install), helm refuses to adopt it.
  # Delete it so helm can recreate it cleanly.
  if kubectl get namespace runai-scale-adjust &>/dev/null; then
    if [[ -z "$(kubectl get namespace runai-scale-adjust -o jsonpath='{.metadata.labels.app\.kubernetes\.io/managed-by}' 2>/dev/null)" ]]; then
      kubectl delete namespace runai-scale-adjust 2>/dev/null || true
    fi
  fi
  # runai-reservation has no chart template but the post-install RBAC hook creates objects inside it.
  kubectl create namespace runai-reservation --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null || true

  local rc_status
  rc_status=$(helm status runai-cluster -n "${RC_NAMESPACE}" -o json 2>/dev/null \
    | python3 -c "import sys,json; print(json.load(sys.stdin)['info']['status'])" 2>/dev/null || true)
  if [[ "${rc_status}" == "pending-install" || "${rc_status}" == "pending-upgrade" || \
        "${rc_status}" == "pending-rollback" || "${rc_status}" == "failed" || \
        "${rc_status}" == "uninstalling" ]]; then
    warn "Found stuck helm release (${rc_status}), removing before retry..."
    helm delete runai-cluster -n "${RC_NAMESPACE}" 2>/dev/null || true
    local wait_i=0
    until ! helm status runai-cluster -n "${RC_NAMESPACE}" &>/dev/null; do
      sleep 2; (( wait_i++ > 30 )) && break
    done
  fi

  # Helm requires a subchart in charts/ for every dependency declared in Chart.yaml,
  # regardless of its condition — conditions are only evaluated later, at render time.
  # charts/*.tgz is gitignored, so a fresh clone has none.
  #
  # Helm itself checks charts/ against Chart.yaml, so a dependency may use any version
  # constraint Helm supports (exact, "~0.18.0", "^1.2", ">=1.0 <2.0", ...).
  local charts_dir="${rc_chart}/charts"
  mkdir -p "${charts_dir}"

  # Older versions of this script installed an empty placeholder when a fetch failed.
  # Helm reports a placeholder as "ok", so drop it, or the real chart is never fetched.
  local chart_tgz
  for chart_tgz in "${charts_dir}"/*.tgz; do
    [[ -e "${chart_tgz}" ]] || continue
    if helm show chart "${chart_tgz}" 2>/dev/null \
        | grep -q 'Placeholder — real chart could not be fetched'; then
      log "  Removing placeholder subchart tarball: $(basename "${chart_tgz}")"
      rm -f "${chart_tgz}"
    fi
  done

  # Prints every dependency Helm does not report as ok/unpacked (missing, wrong version,
  # too many matches, ...), and every tarball in charts/ that Chart.yaml does not declare.
  # Helm loads every tarball in charts/, so an undeclared one (e.g. runai-engine) has no
  # condition to gate it and renders unconditionally.
  _rc_dep_problems() {
    local out
    if ! out="$(helm dependency list "${rc_chart}" 2>&1)"; then
      echo "${out}"
      return 0
    fi
    awk '/^WARNING: .* is not in Chart\.yaml/ { print; next }
         NR > 1 && NF >= 4 && $NF != "ok" && $NF != "unpacked" { print }' <<<"${out}"
  }

  local dep_problems
  dep_problems="$(_rc_dep_problems)"
  if [[ -n "${dep_problems}" ]]; then
    log "  Fetching runai-cluster chart dependencies..."
    echo "${dep_problems}"
    # Pulls the subcharts from their repositories, resolving version ranges, and replaces
    # anything outdated or undeclared in charts/.
    # `update`, not `build`: Chart.lock is gitignored, so a local copy left over from an
    # older Chart.yaml makes `build` refuse to run ("lock file is out of sync").
    helm dependency update "${rc_chart}" 2>&1 \
      | grep -v '^Saving\|^Deleting\|^$' || true
    dep_problems="$(_rc_dep_problems)"
    if [[ -n "${dep_problems}" ]]; then
      echo "${dep_problems}" >&2
      die "runai-cluster chart dependencies are not satisfied (see above and the helm output). Fix the cause, then retry: helm dependency update ${rc_chart}"
    fi
  fi

  local scheduler_args=(
    --set runai-engine.enabled=false
    --set kai-resource-management.enabled=true
    --set karta.enabled=true
  )

  # KRM and the KAI scheduler it bundles both ship their CRDs in their charts'
  # crds/ directories, which Helm applies for subcharts, and each carries its own
  # crd-upgrader hook for the upgrade path. Nothing to apply by hand here.

  # helm post-install/upgrade hooks run queue-version validation which can take >6m to
  # converge. Allow the hook timeout, then fix the cluster in post_deploy_fixes below.
  helm upgrade --install runai-cluster "${rc_chart}" \
    --namespace "${RC_NAMESPACE}" \
    --create-namespace \
    --set "controlPlane.url=https://${DOMAIN}" \
    --set "controlPlane.clientID=runai-internal.installer" \
    --set "controlPlane.bootstrapToken=${CLUSTER_BOOTSTRAP_TOKEN}" \
    --set "cluster.uid=${CLUSTER_UUID}" \
    --set "cluster.url=https://${DOMAIN}" \
    --set "global.customCA.enabled=true" \
    --set "global.image.registry=localhost:${REGISTRY_PORT}/${LOCAL_IMAGE_REPO}" \
    --set "runai-operator.image.registry=localhost:${REGISTRY_PORT}/${LOCAL_IMAGE_REPO}" \
    --set "runai-operator.image.tag=${RC_IMAGE_TAG}" \
    --set "crdUpgrader.image.registry=localhost:${REGISTRY_PORT}/${LOCAL_IMAGE_REPO}" \
    --set "crdUpgrader.image.tag=${RC_IMAGE_TAG}" \
    --set "helmHooksHandler.image.registry=localhost:${REGISTRY_PORT}/${LOCAL_IMAGE_REPO}" \
    --set "helmHooksHandler.image.tag=${RC_IMAGE_TAG}" \
    --set "kartaMigrate.image.registry=localhost:${REGISTRY_PORT}/${LOCAL_IMAGE_REPO}" \
    --set "kartaMigrate.image.tag=${RC_IMAGE_TAG}" \
    --values "${REPO_ROOT}/.local-dev/values-rc.yaml" \
    --set "development.includeCRDsInReleaseManifest=true" \
    "${scheduler_args[@]}" \
    --timeout 10m \
    --wait \
    --wait-for-jobs || warn "helm install completed with warnings (hook timeout is normal)"

  if [[ -f "${RC_REBUILT_FILE}" ]]; then
    log "Restarting RC pods (images updated for some services)..."
    kubectl rollout restart deployment -n "${RC_NAMESPACE}" 2>/dev/null || true
    kubectl rollout restart statefulset -n "${RC_NAMESPACE}" 2>/dev/null || true
  fi

  log "runai-cluster deployed successfully"
}

# ─── Phase 7: /etc/hosts hint ────────────────────────────────────────────────
print_access_info() {
  echo ""
  echo "═══════════════════════════════════════════════════════════════"
  echo "  Run.ai local environment is ready"
  echo "═══════════════════════════════════════════════════════════════"
  echo ""
  echo "  URL:      https://${DOMAIN}"
  echo "  Admin:    ${ADMIN_EMAIL}"
  echo "  Password: ${ADMIN_PASSWORD:-<unavailable — control plane not deployed>}"
  echo "  Trust the self-signed CA (macOS):"
  echo "    sudo security add-trusted-cert -d -r trustRoot \\"
  echo "      -k /Library/Keychains/System.keychain ${CERT_DIR}/ca.crt"
  echo ""
  echo "  Trust the self-signed CA (Linux):"
  echo "    sudo cp ${CERT_DIR}/ca.crt /usr/local/share/ca-certificates/runai-local.crt"
  echo "    sudo update-ca-certificates"
  echo ""
  local cli_bin="${REPO_ROOT}/control-plane/runai-cli/bin/dist/${OS}/${ARCH}/runai"
  echo "  runai CLI (the UI download is linux/amd64 only, so build one for this machine):"
  echo "    make -C ${REPO_ROOT}/control-plane/runai-cli cli-${OS}-${ARCH}"
  echo "    ${cli_bin} config set --cp-url https://${DOMAIN}"
  echo "    ${cli_bin} login user -u ${ADMIN_EMAIL}"
  echo "    Requires the trusted CA above (the CLI has no skip-TLS option)."
  echo "    Usage: ${REPO_ROOT}/control-plane/runai-cli/docs/runai.md"
  echo ""
  echo "  Teardown:  ./local-dev.sh --teardown"
  echo "  Logs:      ${LOG_DIR}/"
  echo "═══════════════════════════════════════════════════════════════"
  echo ""
  echo "  Cluster uuid:         ${CLUSTER_UUID:-<see runai-cluster-details secret>}"
  echo "  CP namespace:         ${CP_NAMESPACE}"
  echo "  Cluster namespace:    ${RC_NAMESPACE}"
  echo ""
  echo "  Scheduler:  kai-resource-management (bundles kai-scheduler) + karta enabled"
  echo "  Re-deploy cluster with updated credentials:"
  echo "    helm upgrade runai-cluster ${REPO_ROOT}/runai-cluster/cluster/runai-chart \\"
  echo "      -n ${RC_NAMESPACE} \\"
  echo "      --set controlPlane.bootstrapToken=<TOKEN> \\"
  echo "      --set cluster.uid=<UID> \\"
  echo "      --reuse-values"
}

# ─── Main ─────────────────────────────────────────────────────────────────────
main() {
  local _start_ts
  _start_ts=$(date +%s)

  if [[ $TEARDOWN -eq 1 ]]; then
    do_teardown
  fi

  echo "═══════════════════════════════════════════════════════════════"
  echo "  Run.ai Local Development Setup"
  echo "  Domain:   ${DOMAIN}"
  echo "  Images:   localhost:${REGISTRY_PORT}/${LOCAL_IMAGE_REPO}/<name>:<tag> (local registry)"
  echo "  Cluster:  ${KIND_CLUSTER}"
  echo "  Parallel: ${PARALLEL_JOBS} build jobs"
  echo "═══════════════════════════════════════════════════════════════"
  echo ""

  check_prereqs

  # Both run unconditionally: prefetch_go_modules only covers the control-plane build
  # path, so --skip-build-cp / RC-only runs would otherwise keep a poisoned cache.
  warn_if_stale_checkout
  purge_stale_oss_caches "${HOME}/.cache/go-build-docker-gopath" \
                         "${HOME}/.cache/go-build-docker-gocache"

  # TLS is generated early so the CA cert is available for the Kind cluster's
  # OIDC oidc-ca-file mount (required at cluster creation time).
  generate_tls
  add_host_entry

  if [[ $SKIP_CLUSTER -eq 0 ]]; then
    create_cluster
    setup_local_registry
    install_ingress_nginx
    if [[ $SKIP_PROMETHEUS -eq 0 ]]; then
      install_prometheus_operator
    fi
    if [[ $GPU_OPERATOR -eq 1 ]]; then
      install_gpu_operator
    elif [[ $SKIP_FAKE_GPU -eq 0 ]]; then
      install_fake_gpu_operator
    fi
  fi

  if [[ $SKIP_CLUSTER -eq 1 ]]; then
    setup_local_registry
  fi

  if [[ $SKIP_BUILD -eq 0 ]]; then
    setup_git_token
    if [[ $SKIP_BUILD_CP -eq 0 ]]; then
      build_all_cp
    fi
    if [[ $SKIP_BUILD_RC -eq 0 ]]; then
      build_all_rc
    fi
  fi

  if [[ $SKIP_DEPLOY -eq 0 ]]; then
    create_namespace_and_secrets

    if [[ $SKIP_DEPLOY_CP -eq 0 ]]; then
      deploy_control_plane
    fi

    if [[ $SKIP_DEPLOY_RC -eq 0 ]]; then
      # CP must be healthy before we can register the cluster and get credentials.
      if [[ $SKIP_DEPLOY_CP -eq 1 ]]; then
        resolve_admin_password || die "Could not read ADMIN_PASSWORD from secret runai-backend-gen-tenants-manager in ${CP_NAMESPACE} — is the control plane deployed?"
        wait_for_cp_health
      fi
      provision_cluster_credentials
      deploy_runai_cluster
    fi
  fi

  resolve_admin_password || true
  print_access_info

  local _elapsed=$(( $(date +%s) - _start_ts ))
  printf "\n⏱  Total time: %dm %02ds\n" $(( _elapsed / 60 )) $(( _elapsed % 60 ))
}

main "$@"
