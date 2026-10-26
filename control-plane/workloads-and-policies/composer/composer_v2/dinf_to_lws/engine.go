/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dinf_to_lws

import (
	"strings"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

const legacyNIMWorkloadLabel = "runai.nim/workload"

// isNIMLeader follows the same label precedence and image detection rules as
// regular inference: an explicit label wins, including false suppressing image detection.
func isNIMLeader(leader *workloadsApi.DistributedInferenceLeaderWorkerSpec) bool {
	if leader == nil {
		return false
	}

	nimValue := false
	nimSpecified := false
	for _, label := range leader.Labels {
		if label.GetExclude() {
			continue
		}
		name := label.Name.Get()
		if name == nil || (*name != constants.LabelNIMWorkload && *name != legacyNIMWorkloadLabel) {
			continue
		}
		nimSpecified = true
		value := label.Value.Get()
		nimValue = nimValue || (value != nil && *value == constants.BoolTrueAsString)
	}
	if nimValue {
		return true
	}
	if nimSpecified {
		return false
	}

	image := strings.ToLower(strings.TrimSpace(leader.GetImage()))
	return strings.HasPrefix(image, "nvcr.io/nim/") ||
		strings.HasPrefix(image, "nvcr.io/nvidia/nim:") ||
		strings.HasPrefix(image, "nvcr.io/nvidia/nim@")
}
