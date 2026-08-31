#!/bin/sh

# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <project_name> <asset_name>"
    exit 1
fi

cat > /tmp/pvc.$$ << AaA
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: $2
  namespace: runai-$1
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1M
  storageClassName: standard
  volumeMode: Filesystem
AaA

kubectl apply -f /tmp/pvc.$$
rm /tmp/pvc.$$