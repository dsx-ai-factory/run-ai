/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package iw_tw_workloads

import (
	"maps"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	batchv1 "k8s.io/api/batch/v1"
)

func jobPodLabelsToMap(src map[string]*string, target any, _ any) *workloadsApi.Error {
	// Kubernetes generates these labels and rejects Jobs with conflicting pod template values.
	labels := maps.Clone(src)
	delete(labels, batchv1.ControllerUidLabel)
	delete(labels, batchv1.JobNameLabel)
	delete(labels, "controller-uid")
	delete(labels, "job-name")
	return composer.MapToMapString(labels, target, nil)
}
