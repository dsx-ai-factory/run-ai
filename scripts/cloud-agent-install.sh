#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

# Idempotent bootstrap for Cloud Agents: container runtime and local-dev prerequisites.
# Full platform build/deploy is done explicitly via ./local-dev.sh (see AGENTS.md).

export DEBIAN_FRONTEND=noninteractive

need_pkgs=(docker.io fuse-overlayfs iptables openssl make git curl python3 ca-certificates)
missing=()
for pkg in "${need_pkgs[@]}"; do
  dpkg -s "$pkg" >/dev/null 2>&1 || missing+=("$pkg")
done
if ((${#missing[@]} > 0)); then
  sudo apt-get update -qq
  sudo apt-get install -y -qq -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" "${missing[@]}"
fi

if ! docker buildx version >/dev/null 2>&1 && ! sudo docker buildx version >/dev/null 2>&1; then
  buildx_ver=v0.23.0
  tmp_buildx="$(mktemp)"
  curl -fsSL -o "${tmp_buildx}" \
    "https://github.com/docker/buildx/releases/download/${buildx_ver}/buildx-${buildx_ver}.linux-amd64"
  chmod +x "${tmp_buildx}"
  sudo mkdir -p /usr/libexec/docker/cli-plugins
  sudo mv "${tmp_buildx}" /usr/libexec/docker/cli-plugins/docker-buildx
  mkdir -p "${HOME}/.docker/cli-plugins"
  cp /usr/libexec/docker/cli-plugins/docker-buildx "${HOME}/.docker/cli-plugins/docker-buildx"
fi

sudo mkdir -p /etc/docker
if ! grep -q fuse-overlayfs /etc/docker/daemon.json 2>/dev/null; then
  printf '%s\n' '{"storage-driver":"fuse-overlayfs"}' | sudo tee /etc/docker/daemon.json >/dev/null
fi

if command -v update-alternatives >/dev/null 2>&1; then
  sudo update-alternatives --set iptables /usr/sbin/iptables-legacy 2>/dev/null || true
fi

if ! getent group docker >/dev/null; then
  sudo groupadd docker
fi
if ! id -nG "${USER:-ubuntu}" | grep -qw docker; then
  sudo usermod -aG docker "${USER:-ubuntu}"
fi

# Kubernetes tooling used by local-dev.sh (install from GitHub/release endpoints, not kind.sigs.k8s.io).
if ! command -v kind >/dev/null 2>&1; then
  curl -fsSL -o /tmp/kind "https://github.com/kubernetes-sigs/kind/releases/download/v0.33.0/kind-linux-amd64"
  chmod +x /tmp/kind
  sudo mv /tmp/kind /usr/local/bin/kind
fi
if ! command -v kubectl >/dev/null 2>&1; then
  kver="$(curl -fsSL https://dl.k8s.io/release/stable.txt)"
  curl -fsSL -o /tmp/kubectl "https://dl.k8s.io/release/${kver}/bin/linux/amd64/kubectl"
  chmod +x /tmp/kubectl
  sudo mv /tmp/kubectl /usr/local/bin/kubectl
fi
if ! command -v helm >/dev/null 2>&1; then
  curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
fi

# Docker Hub is often rate-limited in Cloud Agent VMs; mirror.gcr.io is on the default allowlist.
if [[ -z "${DOCKERHUB_PROXY:-}" ]]; then
  export DOCKERHUB_PROXY=mirror.gcr.io/
fi
grep -q DOCKERHUB_PROXY /etc/profile.d/runai-cloud-agent.sh 2>/dev/null || \
  echo 'export DOCKERHUB_PROXY=mirror.gcr.io/' | sudo tee /etc/profile.d/runai-cloud-agent.sh >/dev/null

# Pre-pull heavy images via mirror.gcr.io (Docker Hub often fails in Cloud Agent VMs).
if command -v docker >/dev/null 2>&1 && sudo docker info >/dev/null 2>&1; then
  for img in \
    "mirror.gcr.io/library/golang:1.27.3" \
    "mirror.gcr.io/kindest/node:v1.37.0" \
    "mirror.gcr.io/library/registry:2"; do
    sudo docker pull "$img" >/dev/null 2>&1 || true
  done
  sudo docker tag mirror.gcr.io/library/golang:1.27.3 library/golang:1.27.3 2>/dev/null || true
  sudo docker tag mirror.gcr.io/kindest/node:v1.37.0 kindest/node:v1.37.0 2>/dev/null || true
  sudo docker tag mirror.gcr.io/library/registry:2 registry:2 2>/dev/null || true
  if curl -fsSI --max-time 10 https://registry.access.redhat.com/v2/ >/dev/null 2>&1; then
    sudo docker pull registry.access.redhat.com/ubi9/ubi-minimal:latest >/dev/null 2>&1 || true
  fi
fi

# Warm a minimal Go module cache for fast per-service make test (optional, best-effort).
if command -v go >/dev/null 2>&1 && [[ -d /workspace/runai-application-packages/validation ]]; then
  (cd /workspace/runai-application-packages/validation && go mod download) || true
fi
