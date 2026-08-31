/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	typeDistributedInferenceRestartPolicy = "distributed-inference-restart-policy"
)

type DistributedInferenceRestartPolicyFlag struct {
	value *openapi.NullableDistributedInferenceRestartPolicy
}

func NewDistributedInferenceRestartPolicyFlag(value *openapi.NullableDistributedInferenceRestartPolicy) *DistributedInferenceRestartPolicyFlag {
	return &DistributedInferenceRestartPolicyFlag{value: value}
}

func (rp *DistributedInferenceRestartPolicyFlag) String() string {
	if !rp.IsSet() {
		return ""
	}
	return string(*rp.value.Get())
}

func (rp *DistributedInferenceRestartPolicyFlag) IsSet() bool {
	return rp.value.IsSet()
}

func (rp *DistributedInferenceRestartPolicyFlag) Set(value string) error {
	restartPolicy, err := openapi.NewDistributedInferenceRestartPolicyFromValue(value)
	if err != nil {
		return err
	}
	rp.value.Set(restartPolicy)
	return nil
}

func (rp *DistributedInferenceRestartPolicyFlag) Value() *openapi.NullableDistributedInferenceRestartPolicy {
	return rp.value
}

func (rp *DistributedInferenceRestartPolicyFlag) Type() string {
	return typeDistributedInferenceRestartPolicy
}
