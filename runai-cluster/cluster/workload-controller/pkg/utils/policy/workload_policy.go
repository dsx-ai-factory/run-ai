/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy

import (
	"context"
	"fmt"
	interutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/internal/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api"
	enforcer2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer/factory"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type WorkloadPolicy interface {
	ApplyPolicy(ctx context.Context, workload client.Object) error
}

type WorkloadPolicyApplier struct {
	workloadMergeCreator factory.JobMergeCreator
	subresourcePolicy    enforcer2.SubresourceApplier
}

func NewPolicyApplier(workloadMergeCreator factory.JobMergeCreator, subresourceApplier enforcer2.SubresourceApplier) *WorkloadPolicyApplier {
	return &WorkloadPolicyApplier{
		workloadMergeCreator: workloadMergeCreator,
		subresourcePolicy:    subresourceApplier,
	}
}

func NewSimplePolicyApplier(workloadMergeCreator factory.JobMergeCreator) *WorkloadPolicyApplier {
	return &WorkloadPolicyApplier{
		workloadMergeCreator: workloadMergeCreator,
		subresourcePolicy:    &enforcer2.SubresourceApplierImpl{},
	}
}

func (self *WorkloadPolicyApplier) ApplyPolicy(ctx context.Context, workload client.Object) error {
	log := logf.FromContext(ctx)
	var subResourcePolicyMap = make(map[api.SubresourceComponent][]byte)
	if workload == nil {
		return fmt.Errorf("workload cannot be nil")
	}
	mergeStackMap, err := self.workloadMergeCreator.CreateMergeStack(ctx, workload)
	if err != nil {

		return fmt.Errorf("failed to load workload policy object, error: %v", err)
	}
	for subresource, mergeStack := range mergeStackMap {

		data, mergeStackErr := self.subresourcePolicy.Apply(ctx, mergeStack)
		if mergeStackErr != nil {
			return fmt.Errorf("failed to apply policy to %v, error: %v", subresource, mergeStackErr)
		}
		subResourcePolicyMap[subresource] = data
	}
	if len(subResourcePolicyMap) == 0 {
		log.Info("no policy to apply to workload")
		return nil
	}
	err = interutils.PatchWorkloadPolicy(workload, subResourcePolicyMap)
	if err != nil {
		return fmt.Errorf("failed to patch policy to workload, error: %v", err)
	}

	return nil
}
