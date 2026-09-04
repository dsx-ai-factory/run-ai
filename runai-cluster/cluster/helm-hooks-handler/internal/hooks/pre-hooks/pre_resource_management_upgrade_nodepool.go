/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/go-logr/logr"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// nodePoolKind is the OwnerReference Kind on shards owned by a nodepool
	// (unchanged between the old and new CRD).
	nodePoolKind = "NodePool"

	// binpackStrategy is the device-level GPU placement value that toggles the
	// gpupack / gpuspread / gpusharingorder plugins (as the legacy controller does).
	binpackStrategy = "binpack"
)

// migrateNodePools migrates every legacy nodepool (run.ai/v1alpha1) to a kai
// nodepool (kai.resources/v1alpha1), re-parenting the scheduling shards it owns.
// It is idempotent (a nodepool with an existing kai twin is skipped) and each
// nodepool rolls back on failure.
func migrateNodePools(ctx context.Context, k8sClient client.Client) error {
	logger := log.FromContext(ctx)
	created := 0
	skipped := 0

	oldNodePools, err := hooks.ListLegacyNodePools(ctx, k8sClient)
	if err != nil {
		return err
	}

	for i := range oldNodePools {
		old := &oldNodePools[i]

		var existing kaiv1alpha1.NodePool
		err = k8sClient.Get(ctx, types.NamespacedName{Name: old.Name}, &existing)
		if err == nil {
			skipped++
			logger.Info(fmt.Sprintf("kai nodepool %s already exists, skipping", old.Name))
			continue
		}
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get kai nodepool %s: %w", old.Name, err)
		}

		if err = migrateNodePool(ctx, k8sClient, old); err != nil {
			return fmt.Errorf("migrate nodepool %s: %w", old.Name, err)
		}
		created++
		logger.Info(fmt.Sprintf("migrated nodepool %s", old.Name))
	}

	logger.Info(fmt.Sprintf("nodepool migration complete: old=%d created=%d skipped=%d", len(oldNodePools), created, skipped))

	return clearLegacyNodePoolFinalizers(ctx, k8sClient, oldNodePools, logger)
}

// clearLegacyNodePoolFinalizers strips the legacy nodepools' finalizers so the
// post-upgrade cleanup can delete them.
func clearLegacyNodePoolFinalizers(ctx context.Context, k8sClient client.Client,
	oldNodePools []runv1alpha1.NodePool, logger logr.Logger) error {
	for i := range oldNodePools {
		nodePool := &oldNodePools[i]
		if len(nodePool.Finalizers) == 0 {
			continue
		}

		logger.Info(fmt.Sprintf("clearing finalizers on legacy nodepool %s", nodePool.Name), "finalizers", nodePool.Finalizers)
		patch := client.MergeFrom(nodePool.DeepCopy())
		nodePool.Finalizers = nil
		if err := k8sClient.Patch(ctx, nodePool, patch); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("clear finalizers on legacy nodepool %s: %w", nodePool.Name, err)
		}
	}
	return nil
}

// migrateNodePool migrates a single legacy nodepool: it creates the new kai
// nodepool, then re-parents every scheduling shard the old nodepool owns. On any
// failure after the new nodepool is created, it rolls back the whole chain.
func migrateNodePool(ctx context.Context, k8sClient client.Client, old *runv1alpha1.NodePool) error {
	newNodePool := buildNewNodePool(old)

	if err := k8sClient.Create(ctx, newNodePool); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("create kai nodepool: %w", err)
		}
		// Already exists (concurrent migration or retry). Create leaves newNodePool
		// without a UID, so fetch it before reparenting - the shards are stamped with
		// newNodePool.UID below and would otherwise get an empty UID.
		if err = k8sClient.Get(ctx, types.NamespacedName{Name: newNodePool.Name}, newNodePool); err != nil {
			return fmt.Errorf("get existing kai nodepool %s: %w", newNodePool.Name, err)
		}
	}

	var reowned []reownedObject
	if err := reownSchedulingShard(ctx, k8sClient, old, newNodePool, &reowned); err != nil {
		rollbackNodePool(ctx, k8sClient, newNodePool, reowned)
		return err
	}

	return nil
}

// buildNewNodePool maps a legacy nodepool onto a kai nodepool, matching the agent's
// kai nodepool handler: labelKey/labelValue carried over, networkTopologyName ->
// preferredNetworkTopologyName, GPU-network-acceleration fields -> annotations, and
// the scheduling fields folded into schedulingShardConfig (which the controller
// compiles back into the re-parented shard). Labels/annotations are preserved; the
// legacy finalizer is not (the kai nodepool has its own; copying it would wedge
// deletion).
func buildNewNodePool(old *runv1alpha1.NodePool) *kaiv1alpha1.NodePool {
	newNodePool := &kaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:        old.Name,
			Labels:      copyStringMap(old.Labels),
			Annotations: copyStringMap(old.Annotations),
		},
		Spec: kaiv1alpha1.NodePoolSpec{
			LabelKey:                     old.Spec.LabelKey,
			LabelValue:                   old.Spec.LabelValue,
			PreferredNetworkTopologyName: old.Spec.NetworkTopologyName,
			SchedulingShardConfig:        buildNodePoolSchedulingShardConfig(old),
		},
	}

	if newNodePool.Annotations == nil {
		newNodePool.Annotations = map[string]string{}
	}
	for key, value := range nodePoolGPUNetworkAccelerationAnnotations(old) {
		newNodePool.Annotations[key] = value
	}

	return newNodePool
}

// nodePoolGPUNetworkAccelerationAnnotations moves the legacy GPU-network-acceleration
// spec fields onto the kai annotations, copying each through when set. The legacy
// CRD defaults detection to Auto so it is always populated in practice; the label
// key is optional.
func nodePoolGPUNetworkAccelerationAnnotations(old *runv1alpha1.NodePool) map[string]string {
	annotations := map[string]string{}
	if old.Spec.GPUNetworkAccelerationDetection != nil {
		annotations[kaiv1alpha1.AnnotationGPUNetworkAccelerationDetection] = string(*old.Spec.GPUNetworkAccelerationDetection)
	}
	if old.Spec.GPUNetworkAccelerationLabelKey != "" {
		annotations[kaiv1alpha1.AnnotationGPUNetworkAccelerationLabelKey] = old.Spec.GPUNetworkAccelerationLabelKey
	}
	return annotations
}

// buildNodePoolSchedulingShardConfig folds the legacy scheduling fields into a kai
// schedulingShardConfig, mirroring the legacy controller's shard builder (placement,
// min runtime, device/numa plugins) and the agent's time-based-fair-share mapping.
func buildNodePoolSchedulingShardConfig(old *runv1alpha1.NodePool) *kaiv1alpha1.SchedulingShardConfig {
	return &kaiv1alpha1.SchedulingShardConfig{
		PlacementStrategy:  convertNodePoolPlacementStrategy(old),
		MinRuntime:         convertNodePoolMinRuntime(old.Spec.MinGuaranteedRuntime),
		Plugins:            convertNodePoolPlugins(old),
		TimeBasedFairShare: convertNodePoolTimeBasedFairShare(old.Spec.TimeBasedFairShare),
	}
}

// convertNodePoolPlacementStrategy maps the node-level GPU/CPU placement strategy;
// the device-level strategy is expressed via plugins (see convertNodePoolPlugins).
func convertNodePoolPlacementStrategy(old *runv1alpha1.NodePool) *kaiv1.PlacementStrategy {
	return &kaiv1.PlacementStrategy{
		GPU: ptr.To(old.Spec.PlacementStrategy.Gpu),
		CPU: ptr.To(old.Spec.PlacementStrategy.Cpu),
	}
}

func convertNodePoolMinRuntime(minGuaranteedRuntime string) *kaiv1.MinRuntime {
	if minGuaranteedRuntime == "" {
		return nil
	}
	return &kaiv1.MinRuntime{
		PreemptMinRuntime: ptr.To(minGuaranteedRuntime),
		ReclaimMinRuntime: ptr.To(minGuaranteedRuntime),
	}
}

// convertNodePoolPlugins reproduces the legacy controller's plugin overrides: the
// device-level GPU placement plugins (when the device strategy differs from the node
// strategy) and the numa plugin (when numa-aware scheduling is enabled).
func convertNodePoolPlugins(old *runv1alpha1.NodePool) map[string]kaiv1.PluginConfig {
	plugins := convertNodePoolDevicePlacementPlugins(old)
	if old.Spec.NumaAwareScheduling {
		if plugins == nil {
			plugins = map[string]kaiv1.PluginConfig{}
		}
		plugins["numa"] = kaiv1.PluginConfig{Enabled: ptr.To(true)}
	}
	return plugins
}

func convertNodePoolDevicePlacementPlugins(old *runv1alpha1.NodePool) map[string]kaiv1.PluginConfig {
	gpuDeviceStrategy := old.Spec.PlacementStrategy.GpuDevice
	if gpuDeviceStrategy == "" {
		return nil
	}
	isDeviceBinpack := gpuDeviceStrategy == binpackStrategy
	return map[string]kaiv1.PluginConfig{
		"gpupack":         {Enabled: ptr.To(isDeviceBinpack)},
		"gpuspread":       {Enabled: ptr.To(!isDeviceBinpack)},
		"gpusharingorder": {Enabled: ptr.To(isDeviceBinpack)},
	}
}

// convertNodePoolTimeBasedFairShare maps the legacy fair-share spec onto the kai
// shape, copying values through as-is: disabled becomes nil, and unset fields stay
// unset (the nodepool-controller applies its own defaults - no need to repeat
// them here).
func convertNodePoolTimeBasedFairShare(old runv1alpha1.TimeBasedFairShare) *kaiv1alpha1.TimeBasedFairShare {
	if !old.Enabled {
		return nil
	}
	tbfs := &kaiv1alpha1.TimeBasedFairShare{
		Enabled:               ptr.To(true),
		HistoricalUsageWeight: old.HistoricalUsageWeight,
		HalfLifePeriod:        old.DecayHalfLife,
	}
	if old.WindowType != nil || old.WindowDuration != nil || old.TumblingWindowStartTime != nil {
		tbfs.Window = &kaiv1alpha1.FairShareWindow{
			Size:              old.WindowDuration,
			TumblingStartTime: old.TumblingWindowStartTime,
		}
		if old.WindowType != nil {
			tbfs.Window.Type = ptr.To(string(*old.WindowType))
		}
	}
	return tbfs
}

// reownSchedulingShard re-points the scheduling shard owned by the old nodepool at
// the new kai nodepool, recording the change in reowned so migrateNodePool can roll
// back on failure. The controller creates one shard per nodepool named after it
// (shard.Name == nodePool.Name), so it is fetched by name, verifies owner reference
// before repointing.
//
// It is a no-op when the kai SchedulingShard CRD is not served yet: on the
// engine->kai cutover the kai-scheduler subchart applies that CRD in a pre-upgrade
// hook that runs after this one, so the Get fails on the RESTMapper rather than
// returning NotFound. There is nothing to re-parent in that window - the engine
// shards were already deleted by pre-crd-upgrade, and after the upgrade the kai
// nodepool-controller creates the shard owned by the kai nodepool directly.
func reownSchedulingShard(ctx context.Context, k8sClient client.Client, old *runv1alpha1.NodePool,
	newNodePool *kaiv1alpha1.NodePool, reowned *[]reownedObject) error {
	var shard kaiv1.SchedulingShard
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: old.Name}, &shard); err != nil {
		if apierrors.IsNotFound(err) {
			return nil // no shard for this nodepool, nothing to reparent
		}
		if meta.IsNoMatchError(err) {
			log.FromContext(ctx).Info("kai SchedulingShard CRD not served yet, skipping shard reparent",
				"nodepool", old.Name)
			return nil
		}
		return fmt.Errorf("get scheduling shard for nodepool %s: %w", old.Name, err)
	}

	oldOwners := shard.GetOwnerReferences()
	updated, replaced := replaceNodePoolOwner(oldOwners, old, newNodePool)
	if !replaced {
		return nil
	}

	patch := client.MergeFrom(shard.DeepCopy())
	shard.SetOwnerReferences(updated)
	if err := k8sClient.Patch(ctx, &shard, patch); err != nil {
		return fmt.Errorf("patch owner of scheduling shard %s: %w", shard.Name, err)
	}
	*reowned = append(*reowned, reownedObject{
		obj:       shard.DeepCopy(),
		oldOwners: oldOwners,
	})
	return nil
}

// replaceNodePoolOwner replaces any OwnerReference to the old nodepool with one to
// the new kai nodepool (keeping the Controller / BlockOwnerDeletion flags), and
// reports whether one was found.
func replaceNodePoolOwner(owners []metav1.OwnerReference, old *runv1alpha1.NodePool,
	newNodePool *kaiv1alpha1.NodePool) ([]metav1.OwnerReference, bool) {
	result := make([]metav1.OwnerReference, 0, len(owners))
	replaced := false
	for _, owner := range owners {
		if owner.Kind == nodePoolKind && owner.UID == old.UID {
			newOwner := owner
			newOwner.APIVersion = kaiv1alpha1.GroupVersion.String()
			newOwner.Name = newNodePool.Name
			newOwner.UID = newNodePool.UID
			result = append(result, newOwner)
			replaced = true
			continue
		}
		result = append(result, owner)
	}
	return result, replaced
}

// rollbackNodePool undoes a partially-completed nodepool migration: it restores the
// owner references and labels of every re-parented shard, then deletes the newly
// created kai nodepool (stripping any finalizer first, since its controller is not
// running during the migration window).
func rollbackNodePool(ctx context.Context, k8sClient client.Client, newNodePool *kaiv1alpha1.NodePool, reowned []reownedObject) {
	logger := log.FromContext(ctx)
	for i := range reowned {
		obj := reowned[i].obj
		patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
		obj.SetOwnerReferences(reowned[i].oldOwners)
		if err := k8sClient.Patch(ctx, obj, patch); err != nil {
			logger.Error(err, fmt.Sprintf("rollback: failed to restore %T %s", obj, obj.GetName()))
		}
	}

	if len(newNodePool.Finalizers) > 0 {
		newNodePool.Finalizers = nil
		if err := k8sClient.Update(ctx, newNodePool); err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, fmt.Sprintf("rollback: failed to strip new nodepool %s finalizers", newNodePool.Name))
		}
	}

	if err := k8sClient.Delete(ctx, newNodePool); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, fmt.Sprintf("rollback: failed to delete new nodepool %s", newNodePool.Name))
	}
}
