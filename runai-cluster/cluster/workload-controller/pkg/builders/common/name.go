/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

type WorkloadInitials string

const (
	KindInteractiveWorkloadInitial = "iw"
	KindInferenceWorkloadInitial   = "infw"
	KindTrainingWorkloadInitial    = "tw"
	KindDistributedWorkloadInitial = "dw"
)

func (wi WorkloadInitials) String() string {
	return string(wi)
}

func GetWorkloadInitials(kind string) (WorkloadInitials, error) {
	switch kind {
	case constants.KindInteractiveWorkload:
		return KindInteractiveWorkloadInitial, nil
	case constants.KindInferenceWorkload:
		return KindInferenceWorkloadInitial, nil
	case constants.KindTrainingWorkload:
		return KindTrainingWorkloadInitial, nil
	case constants.KindDistributedWorkload:
		return KindDistributedWorkloadInitial, nil
	default:
		return "", fmt.Errorf("unknown workload kind: %s", kind)
	}
}
