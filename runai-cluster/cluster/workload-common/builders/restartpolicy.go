/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"fmt"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	v1 "k8s.io/api/core/v1"
)

func GetRestartPolicy(value *flds.StringField) (*v1.RestartPolicy, error) {
	if value == nil || value.Value == nil {
		return nil, nil
	}

	switch resType := v1.RestartPolicy(value.Deref()); resType {
	case v1.RestartPolicyOnFailure, v1.RestartPolicyAlways, v1.RestartPolicyNever:
		return &resType, nil
	default:
		return nil, fmt.Errorf("invalid restartPolicy value %s", value.Deref())
	}
}
