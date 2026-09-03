# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Creating an import file got code-generator

cat <<EOF > generate-dep.go
package main

import (
	_ "k8s.io/code-generator"
)
EOF

go mod tidy
go mod vendor
SCRIPT_DIR="$(cd "$(dirname "$(readlink "$0" || echo "$0")")"; pwd)"
chmod +x ${SCRIPT_DIR}/../vendor/k8s.io/code-generator/generate-groups.sh
${SCRIPT_DIR}/../vendor/k8s.io/code-generator/generate-groups.sh client,lister,informer \
  github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/client \
  github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis \
 "scheduling:v1,v2alpha1,v2alpha2,v2 run:v1,v2alpha1,v2alpha2" \
 --go-header-file ${SCRIPT_DIR}/boilerplate.go.txt

rm -f generate-dep.go && go mod tidy && go mod vendor
