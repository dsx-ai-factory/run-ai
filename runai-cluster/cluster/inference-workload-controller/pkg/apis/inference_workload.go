/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apis

import (
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetNewInferenceWorkloadFn() client.Object {
	return &v2alpha1.InferenceWorkload{}
}
