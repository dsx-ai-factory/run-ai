/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy

import (
	"encoding/json"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	distributedpolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type DistributedAdapter struct {
	policy *distributedpolicy.DistributedPolicy
}

func NewDistributedAdapter(policy *distributedpolicy.DistributedPolicy) *DistributedAdapter {
	return &DistributedAdapter{
		policy,
	}
}

func InitDistributedPolicyFromReq(req admission.Request) (Adapter, error) {
	policy := distributedpolicy.NewEmptyDistributedPolicy()
	err := json.Unmarshal(req.Object.Raw, &policy)
	if err != nil {
		return nil, err
	}
	return NewDistributedAdapter(policy), nil
}

func (m *DistributedAdapter) GetMeta() metav1.ObjectMeta {
	return m.policy.ObjectMeta
}

func (m *DistributedAdapter) NewPolicyList() client.ObjectList {
	return &distributedpolicy.DistributedPolicyList{}
}

func (m *DistributedAdapter) NewEmptyPolicy() interface{} {
	return distributedpolicy.NewEmptyDistributedPolicy()
}

func (m *DistributedAdapter) GetSpec() interface{} {
	return m.policy.Spec
}

func (m *DistributedAdapter) GetContainerFields() *field_groups.ContainerFields {
	return &m.policy.Spec.ContainerFields
}

func (m *DistributedAdapter) GetAccessControlFields() *field_groups.AccessControlFields {
	return &m.policy.Spec.AccessControlFields
}

func (m *DistributedAdapter) GetRunaiSchedulerFields() *field_groups.RunaiSchedulerFields {
	return &m.policy.Spec.RunaiSchedulerFields
}

func (m *DistributedAdapter) Kind() client.Object {
	return &distributedpolicy.DistributedPolicy{}
}
