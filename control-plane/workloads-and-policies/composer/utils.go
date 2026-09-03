/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

func IsMainNode(replicaType *workloadsApi.ReplicaType) bool {
	return replicaType == nil || *replicaType == workloadsApi.REPLICATYPE_WORKER
}
