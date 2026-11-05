#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

log() { echo "[cloud-agent-start] $*"; }

start_dockerd() {
  if sudo docker info >/dev/null 2>&1; then
    return 0
  fi
  log "Starting dockerd"
  if ! pgrep -x dockerd >/dev/null 2>&1; then
    sudo nohup dockerd >>/tmp/dockerd.log 2>&1 &
  fi
  for _ in $(seq 1 60); do
    if sudo docker info >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  log "dockerd did not become ready; see /tmp/dockerd.log"
  return 1
}

start_dockerd

if ! grep -q '[[:space:]]runai\.local' /etc/hosts 2>/dev/null; then
  echo "127.0.0.1 runai.local" | sudo tee -a /etc/hosts >/dev/null
fi

# Resume Kind cluster containers if a prior local-dev run created them.
if command -v docker >/dev/null 2>&1; then
  mapfile -t kind_containers < <(sudo docker ps -aq --filter "label=io.x-k8s.kind.cluster=${KIND_CLUSTER:-runai-local}" 2>/dev/null || true)
  if ((${#kind_containers[@]} > 0)); then
    log "Starting ${#kind_containers[@]} Kind node containers for cluster ${KIND_CLUSTER:-runai-local}"
    sudo docker start "${kind_containers[@]}" >/dev/null 2>&1 || true
  fi
fi

# Keep registry alive when present (local-dev.sh).
if sudo docker inspect kind-registry >/dev/null 2>&1; then
  state=$(sudo docker inspect -f '{{.State.Running}}' kind-registry 2>/dev/null || echo false)
  if [[ "$state" != "true" ]]; then
    log "Starting kind-registry"
    sudo docker start kind-registry >/dev/null 2>&1 || true
  fi
fi

log "Docker ready"
