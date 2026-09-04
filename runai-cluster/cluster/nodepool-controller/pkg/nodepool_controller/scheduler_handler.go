/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/operands"
	runai_scheduler "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/operands/runai-scheduler"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/utils"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	ctrl_log "sigs.k8s.io/controller-runtime/pkg/log"
)

func (npc *NodePoolController) reconcileScheduler(ctx context.Context, nodePool *v1alpha1.NodePool) (err error) {
	log.Info().Msgf("Reconciling Scheduler for Node Pool <%v>", nodePool.Name)

	createOrUpdateResources, deleteResources, err := npc.getSchedulerResources(ctx, nodePool, npc.params)
	if err != nil {
		return
	}

	logger := ctrl_log.FromContext(ctx)

	err = utils.CreateOrUpdateResources(npc.Client, npc.Scheme, ctx,
		logger, nodePool, createOrUpdateResources)
	if err != nil {
		err = fmt.Errorf("enabled scheduler for nodepool <%v> handling error: %v", nodePool.Name, err)
		return
	}

	// Delete resources of disabled operands
	err = utils.DeleteResources(npc.Client, ctx, deleteResources)
	if err != nil {
		err = fmt.Errorf("disabled scheduler for nodepool <%v> cleanup error: %v", nodePool.Name, err)
		return
	}

	return
}

func (npc *NodePoolController) getSchedulerResources(ctx context.Context,
	nodePool *v1alpha1.NodePool, params *common.NodePoolControllerParams) (
	createOrUpdateResources []operands.ResourceOld, deleteResources []operands.Resource, err error) {
	schedulerOperand := npc.getSchedulerOperand(nodePool)

	schedulerResources, err := schedulerOperand.ResourcesForNodePool(ctx,
		npc.Client, nodePool.DeepCopy(), params.DeepCopy())
	if err != nil {
		log.Error().Msgf("Failed getting scheduler resources of nodepool <%v>, err: %v", nodePool.Name, err.Error())
		return
	}

	for _, resource := range schedulerResources {
		if resource.Object == nil {
			continue
		}
		createOrUpdateResources = append(createOrUpdateResources, resource)
	}

	return
}

func (npc *NodePoolController) getNodePoolStatusByScheduler(ctx context.Context, nodePool *v1alpha1.NodePool) (
	schedulerReady bool, err error) {
	scheduler := npc.getSchedulerOperand(nodePool)

	status, err := scheduler.Status(ctx, npc.Client, nodePool, npc.params)
	if err != nil {
		return schedulerReady, fmt.Errorf("get status of scheduler for nodepool <%v> error: %v", nodePool.Name, err)
	}

	if !status.Ready {
		nodePool.Status.Phase = v1alpha1.NodePoolUnschedulable
		nodePool.Status.Message = fmt.Sprintf("%v, reason: %v", common.SchedulerNotReadyMessage, strings.Join(status.Reasons, ", "))
	}
	return status.Ready, nil
}

func (npc *NodePoolController) getSchedulerOperand(nodePool *v1alpha1.NodePool) operands.NodePoolOperand {
	return runai_scheduler.Operand(nodePool.Name, npc.serviceMonitorEnabled)
}
