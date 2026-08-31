/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy

import (
	"encoding/json"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	inferencepolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	workloadspolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads/policy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type InferenceAdapter struct {
	policy *inferencepolicy.InferencePolicy
}

func NewInferenceAdapter(policy *inferencepolicy.InferencePolicy) *InferenceAdapter {
	return &InferenceAdapter{
		policy,
	}

}

func InitInferencePolicyFromReq(req admission.Request) (workloadspolicy.Adapter, error) {
	policy := inferencepolicy.NewEmptyInferencePolicy()
	err := json.Unmarshal(req.Object.Raw, &policy)
	if err != nil {
		return nil, err
	}
	return NewInferenceAdapter(policy), nil
}

func (m *InferenceAdapter) NewPolicyList() client.ObjectList {
	return &inferencepolicy.InferencePolicyList{}
}

func (m *InferenceAdapter) NewEmptyPolicy() interface{} {
	return inferencepolicy.NewEmptyInferencePolicy()
}

func (m *InferenceAdapter) GetSpec() interface{} {
	return m.policy.Spec
}

func (m *InferenceAdapter) GetContainerFields() *field_groups.ContainerFields {
	return &m.policy.Spec.ContainerFields
}

func (m *InferenceAdapter) GetAccessControlFields() *field_groups.AccessControlFields {
	return &field_groups.AccessControlFields{}
}

func (m *InferenceAdapter) GetRunaiSchedulerFields() *field_groups.RunaiSchedulerFields {
	return &m.policy.Spec.RunaiSchedulerFields
}

func (m *InferenceAdapter) GetMeta() metav1.ObjectMeta {
	return m.policy.ObjectMeta
}
