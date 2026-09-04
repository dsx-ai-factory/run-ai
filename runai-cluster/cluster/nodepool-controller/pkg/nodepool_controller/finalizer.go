/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/utils"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
)

func (npc *NodePoolController) addControllerAsFinalizerIfNeeded(ctx context.Context, nodePool *v1alpha1.NodePool) (err error) {
	finalizerName := config.FinalizerName()
	if utils.IsItemInList(nodePool.Finalizers, finalizerName) {
		return nil
	}

	log.Info().Msgf("Adding controller to finalizers list of node pool <%v>", nodePool.Name)
	nodePool.Finalizers = append(nodePool.Finalizers, finalizerName)

	if err = npc.updateNodePoolFinalizers(ctx, nodePool); err != nil {
		log.Error().Msgf("Error adding controller to finalizers list of node pool <%v>, err: %v", nodePool.Name, err.Error())
		return err
	}
	return nil
}

func (npc *NodePoolController) finalize(ctx context.Context, nodePool *v1alpha1.NodePool) (err error) {
	log.Info().Msgf("Finalizing deletion of nodepool <%v>", nodePool.Name)

	if !utils.IsItemInList(nodePool.Finalizers, config.FinalizerName()) {
		log.Info().Msgf("Got a finalization event for NodePool <%v>, but this controller is not in its finalizers list. Skipping finalization.", nodePool.Name)
		return
	}

	isNodePoolDeleted := npc.handleNodePoolDeletionInCaseOfRunaiUninstall(ctx, nodePool)
	if isNodePoolDeleted {
		return
	}

	resultNodes, err := npc.handleNodePoolDeletion(ctx, nodePool)
	if err != nil {
		innerErr := npc.updateNodePoolStatusOnDeletionError(ctx, nodePool, resultNodes)
		err = utils.AppendErrIfNotNil(err, innerErr)
		return fmt.Errorf("errors encountered during deletion of nodepool <%v>, error: %v", nodePool.Name, err)
	}

	phase, err := npc.updateNodePoolStatusOnDeletion(ctx, nodePool, resultNodes)
	if err != nil {
		return fmt.Errorf("errors encountered during deletion of nodepool <%v>, error: %v", nodePool.Name, err.Error())
	}

	nodePoolDeletingMsg := ""
	if phase != v1alpha1.NodePoolDeleting {
		// removing the finalizer from the nodepool object, will cause the resource to be deleted by Kubernetes
		if err = npc.deleteFinalizer(ctx, nodePool); err != nil {
			return
		}
	} else {
		nodePoolDeletingMsg = "; nodepool is in 'Deleting' phase"
		unschedulableErr := errors.New(fmt.Sprintf("NodePool <%v> is in 'Deleting' phase", nodePool.Name))
		err = utils.AppendErrIfNotNil(err, unschedulableErr)
	}

	log.Info().Msgf("Done handling deletion of nodepool <%v>%v", nodePool.Name, nodePoolDeletingMsg)
	return err
}

// handleNodePoolDeletionInCaseOfRunaiUninstall - in case the runai operator is being uninstalled,
// we want to delete the nodepool
// and ignore the pods that are running on the nodes of the nodepool.
// returns true if the nodepool was deleted, false otherwise.
func (npc *NodePoolController) handleNodePoolDeletionInCaseOfRunaiUninstall(ctx context.Context, nodePool *v1alpha1.NodePool) bool {
	clusterResource, err := npc.getClusterResource(ctx)
	if err != nil && !apierrors.IsNotFound(err) {
		log.Error().Msgf("Failed getting cluster resource, err: %s", err.Error())
		return false
	}

	if clusterResource != nil && clusterResource.GetDeletionTimestamp().IsZero() {
		return false
	}

	log.Info().Msgf("Deleting NodePool <%v>; runai cluster is being deleted, ignoring running pods on nodepool's nodes", nodePool.Name)
	_ = npc.deleteFinalizer(ctx, nodePool)
	_ = npc.cleanupDeletedNodePool(ctx, nodePool)
	return true
}

func (npc *NodePoolController) deleteFinalizer(ctx context.Context, nodePool *v1alpha1.NodePool) (err error) {
	finalizerName := config.FinalizerName()
	log.Info().Msgf("Removing finalizer <%v> from nodepool's <%v> finalizers list",
		finalizerName, nodePool.Name)
	nodePool.Finalizers = utils.DeleteFromList(nodePool.Finalizers, finalizerName)
	if err = npc.updateNodePoolFinalizers(ctx, nodePool); err != nil && !apierrors.IsNotFound(err) {
		log.Error().Msgf("Error removing controller from finalizer list for nodepool <%v>; err: %v",
			nodePool.Name, err.Error())
		return err
	}

	return nil
}

func (npc *NodePoolController) updateNodePoolFinalizers(ctx context.Context, nodePool *v1alpha1.NodePool) (err error) {
	patchBytes, err := updateNodePoolFinalizersPatchBytes(nodePool.Finalizers)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - update finalizers of nodepool <%v>, err: %v", nodePool.Name, err)
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = npc.Client.Patch(ctx, nodePool, patch)
	if err != nil {
		return err
	}
	return nil
}

// The run.ai Cluster CR belongs to the proprietary packaging, so it is read by
// GVK rather than by Go type to keep this module free of run.ai API imports.
var (
	clusterGVK  = schema.GroupVersionKind{Group: "run.ai", Version: "v1", Kind: "Cluster"}
	clusterName = "cluster"
)

func (npc *NodePoolController) getClusterResource(ctx context.Context) (cluster *unstructured.Unstructured, err error) {
	cluster = &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(clusterGVK)
	objectKey := types.NamespacedName{Name: clusterName}
	err = npc.Client.Get(ctx, objectKey, cluster)
	if err != nil {
		log.Error().Msgf("Failed getting cluster resource <%v>, error: %v", clusterName, err.Error())
		return nil, err
	}
	return cluster, nil
}

func updateNodePoolFinalizersPatchBytes(finalizers []string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{
		"finalizers": finalizers}})
}
