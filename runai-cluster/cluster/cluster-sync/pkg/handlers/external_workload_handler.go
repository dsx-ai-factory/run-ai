/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/OneOfOne/xxhash"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	kubeaischedv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExternalWorkloadHandler struct {
	*WorkloadHandler
}

func (h *ExternalWorkloadHandler) HandleUpdateItem(key client.ObjectKey) error {
	externalWorkload := &v1alpha1.ExternalWorkload{}
	err := h.cachedClient.Get(context.Background(), key, externalWorkload)
	if err != nil {
		return err
	}

	err = h.handleAddUpdate(externalWorkload)
	if err != nil {
		return err
	}

	// The informer drops an object from the cache before calling the delete handler, which doesn't wait for queued
	// updates, so an object that's gone or replaced by now may have had its delete reach the backend before this update.
	current := &v1alpha1.ExternalWorkload{}
	err = h.cachedClient.Get(context.Background(), key, current)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	workloadId := topowner.GetRunaiWorkloadId(externalWorkload.Spec.OriginalWorkloadUID, externalWorkload.Annotations)
	if err == nil && topowner.GetRunaiWorkloadId(current.Spec.OriginalWorkloadUID, current.Annotations) == workloadId {
		return nil
	}

	zap.S().Infof("ExternalWorkload <%s/%s> was deleted while its update was sent - deleting it again",
		externalWorkload.Namespace, externalWorkload.Name)
	return h.HandleDelete(externalWorkload)
}

func NewExternalWorkloadHandler(workloadHandler *WorkloadHandler) *ExternalWorkloadHandler {
	return &ExternalWorkloadHandler{WorkloadHandler: workloadHandler}
}

func (h *WorkloadHandler) HandleAdd(externalWorkload *v1alpha1.ExternalWorkload) error {
	if h.WasSynced(externalWorkload.UID, externalWorkload.Namespace) {
		zap.S().Infof("ExternalWorkload <%s/%s> was already synced - not adding",
			externalWorkload.Namespace, externalWorkload.Name)
		return nil
	}

	return h.handleAddUpdate(externalWorkload)
}

func (h *WorkloadHandler) HandleDelete(externalWorkload *v1alpha1.ExternalWorkload) error {
	workloadId := topowner.GetRunaiWorkloadId(externalWorkload.Spec.OriginalWorkloadUID, externalWorkload.Annotations)
	return h.handleDelete(workloadId, externalWorkload.GetNamespace(),
		externalWorkload.GetName(), constants.ExternalWorkload)
}

func (h *WorkloadHandler) HandleUpdate(externalWorkload *v1alpha1.ExternalWorkload) error {
	return h.handleAddUpdate(externalWorkload)
}

func (h *WorkloadHandler) handleAddUpdate(externalWorkload *v1alpha1.ExternalWorkload) error {
	externalWorkload = externalWorkload.DeepCopy()

	workloadSyncFields, _, podGroupsOfWorkload, err := h.parser.ParseExternalWorkloadToServerScheme(externalWorkload)
	if err != nil {
		return err
	}

	if isExternalWorkloadUpToDateAndSetSyncStatus(&externalWorkload.Status.SyncStatus, externalWorkload.Generation,
		&externalWorkload.Status.ObservedStatus.LastStatusChangeTime, workloadSyncFields) {
		zap.S().Infof("ExternalWorkload <%s/%s> is up to date - not upserting",
			externalWorkload.Namespace, externalWorkload.Name)
		return nil
	}

	var errMsg *string = nil
	err = h.handleUpsert(workloadSyncFields, externalWorkload.Namespace, externalWorkload.Name, constants.ExternalWorkload)
	if err != nil {
		errMsg = ptr.To(err.Error())
	}

	_ = h.updateExternalWorkloadSyncStatus(externalWorkload, errMsg)
	_ = h.addExternalWorkloadNameAnnotationIfNeeded(context.Background(), podGroupsOfWorkload, externalWorkload.Name)
	return err
}

func (h *WorkloadHandler) getParsedExternalWorkloadsInNamespace(namespaceName string,
	syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, error) {

	externalWorkloads := &v1alpha1.ExternalWorkloadList{}
	err := h.cachedClient.List(context.Background(), externalWorkloads, &client.ListOptions{Namespace: namespaceName})
	if err != nil {
		zap.S().Errorf("Failed listing ExternalWorkloads in namespace <%s>, error: %s",
			namespaceName, err.Error())
		return []openapi.WorkloadSyncFields{}, []string{}, []client.Object{}, err
	}

	workloadSyncFields, upToDateIds, updateSyncStatus, _, err := h.parseExternalWorkloads(externalWorkloads.Items, syncedInNamespace)
	return workloadSyncFields, upToDateIds, updateSyncStatus, err
}

func (h *WorkloadHandler) parseExternalWorkloads(externalWorkloads []v1alpha1.ExternalWorkload, syncedInNamespace map[types.UID]bool) ([]openapi.WorkloadSyncFields, []string, []client.Object, []corev1.Pod, error) {
	workloadsToSync := []openapi.WorkloadSyncFields{}
	upToDateIds := []string{}
	workloadsToUpdateSyncStatus := []client.Object{}
	workloadPods := []corev1.Pod{}

	for i := range externalWorkloads {
		// important: to avoid bug where we insert a pointer of for-looped map value to an array
		externalWorkload := &externalWorkloads[i]
		syncedInNamespace[externalWorkload.UID] = true

		parsed, pods, podGroupsOfWorkload, innerErr := h.parser.ParseExternalWorkloadToServerScheme(externalWorkload)
		if innerErr != nil {
			zap.S().Warnf("Failed parsing ExternalWorkload <%s/%s>, error: %s",
				externalWorkload.Namespace, externalWorkload.Name, innerErr.Error())
			continue
		}

		workloadPods = append(workloadPods, pods...)
		if isExternalWorkloadUpToDateAndSetSyncStatus(&externalWorkload.Status.SyncStatus, externalWorkload.Generation,
			&externalWorkload.Status.ObservedStatus.LastStatusChangeTime, parsed) {
			uid := topowner.GetRunaiWorkloadId(externalWorkload.Spec.OriginalWorkloadUID, externalWorkload.Annotations)
			upToDateIds = append(upToDateIds, uid)
			continue
		}

		workloadsToSync = append(workloadsToSync, *parsed)
		workloadsToUpdateSyncStatus = append(workloadsToUpdateSyncStatus, externalWorkload)

		// nothing to do even if error, also errors are logged inside.
		_ = h.addExternalWorkloadNameAnnotationIfNeeded(context.Background(), podGroupsOfWorkload, externalWorkload.Name)
	}

	return workloadsToSync, upToDateIds, workloadsToUpdateSyncStatus, workloadPods, nil
}

func (h *WorkloadHandler) updateExternalWorkloadSyncStatus(externalWorkload *v1alpha1.ExternalWorkload, errMsg *string) error {
	if errMsg == nil {
		// LastSyncGeneration, LastSyncStatusChangeTime and LastSyncStatusChecksum should already be set
		externalWorkload.Status.SyncStatus.LastSyncTimestamp = metav1.Now()
		externalWorkload.Status.SyncStatus.LastAttemptedSyncGeneration = nil
		externalWorkload.Status.SyncStatus.LastAttemptedSyncStatusChangeTime = nil
		externalWorkload.Status.SyncStatus.SyncErrorMsg = nil
	} else {
		// the Patch operation will ignore LastSyncGeneration, LastSyncStatusChangeTime and LastSyncStatusChecksum fields
		externalWorkload.Status.SyncStatus.LastAttemptedSyncGeneration = &externalWorkload.Generation
		externalWorkload.Status.SyncStatus.LastAttemptedSyncStatusChangeTime = &externalWorkload.Status.ObservedStatus.LastStatusChangeTime
		externalWorkload.Status.SyncStatus.SyncErrorMsg = errMsg
	}

	return h.patchExternalWorkloadSyncStatus(externalWorkload, errMsg == nil)
}

func (h *WorkloadHandler) patchExternalWorkloadSyncStatus(externalWorkload *v1alpha1.ExternalWorkload, successful bool) error {
	var patchBytes []byte
	var err error = nil

	if successful {
		patchBytes, err = updateSuccessfulSyncStatusPatchBytes(&externalWorkload.Status.SyncStatus)
	} else {
		patchBytes, err = updateFailedSyncStatusPatchBytes(&externalWorkload.Status.SyncStatus)
	}
	if err != nil {
		zap.S().Errorf("Failed to json.Marshal patch - "+
			"update sync status of externalworkload <%s/%s>, error: %s",
			externalWorkload.Namespace, externalWorkload.Name, err.Error())
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = h.cachedClient.Status().Patch(context.Background(), externalWorkload, patch)
	if err != nil {
		zap.S().Errorf("Failed to patch external workload sync status <%s/%s>, error: %s",
			externalWorkload.GetNamespace(), externalWorkload.GetName(), err.Error())
		return err
	}
	zap.S().Debugf("Patched externalworkload <%s/%s> sync status",
		externalWorkload.Namespace, externalWorkload.Name)
	return nil
}

func isExternalWorkloadUpToDateAndSetSyncStatus(syncStatus *v1alpha1.WorkloadSyncStatus, generation int64,
	lastStatusChangeTime *metav1.Time, parsedWorkload *openapi.WorkloadSyncFields) bool {
	syncStatusCopy := syncStatus.DeepCopy()
	err := setExternalWorkloadSuccessfulSyncStatus(syncStatus, generation, lastStatusChangeTime, parsedWorkload)
	if err != nil {
		return false
	}

	return syncStatusCopy.LastSyncGeneration == syncStatus.LastSyncGeneration &&
		(&syncStatusCopy.LastSyncStatusChangeTime).Equal(&syncStatus.LastSyncStatusChangeTime) &&
		syncStatusCopy.LastSyncStatusChecksum == syncStatus.LastSyncStatusChecksum
}

func setExternalWorkloadSuccessfulSyncStatus(syncStatus *v1alpha1.WorkloadSyncStatus, generation int64,
	lastStatusChangeTime *metav1.Time, parsedWorkload *openapi.WorkloadSyncFields) error {
	parsedWorkloadBytes, err := json.Marshal(parsedWorkload)
	if err != nil {
		zap.S().Errorf("Failed to json.Marshal parsed %s <%s/%s>, error: %s",
			parsedWorkload.Type, parsedWorkload.Namespace, parsedWorkload.Name, err.Error())
		return err
	}

	checksum := xxhash.Checksum32(parsedWorkloadBytes)

	syncStatus.LastSyncGeneration = generation
	syncStatus.LastSyncStatusChangeTime = timeDeref(lastStatusChangeTime)
	syncStatus.LastSyncStatusChecksum = checksum
	return nil
}

func updateSuccessfulSyncStatusPatchBytes(syncStatus *v1alpha1.WorkloadSyncStatus) ([]byte, error) {
	return json.Marshal(map[string]any{"status": map[string]any{
		"syncStatus": map[string]any{
			"lastSyncTimestamp":                 syncStatus.LastSyncTimestamp,
			"lastSyncGeneration":                syncStatus.LastSyncGeneration,
			"lastSyncStatusChangeTime":          syncStatus.LastSyncStatusChangeTime,
			"lastSyncStatusChecksum":            syncStatus.LastSyncStatusChecksum,
			"lastAttemptedSyncGeneration":       nil,
			"lastAttemptedSyncStatusChangeTime": nil,
			"syncErrorMsg":                      nil}}})
}

func updateFailedSyncStatusPatchBytes(syncStatus *v1alpha1.WorkloadSyncStatus) ([]byte, error) {
	return json.Marshal(map[string]any{"status": map[string]any{
		"syncStatus": map[string]any{
			"lastAttemptedSyncGeneration":       syncStatus.LastAttemptedSyncGeneration,
			"lastAttemptedSyncStatusChangeTime": syncStatus.LastAttemptedSyncStatusChangeTime,
			"syncErrorMsg":                      syncStatus.SyncErrorMsg}}})
}

func (h *WorkloadHandler) addExternalWorkloadNameAnnotationIfNeeded(
	ctx context.Context,
	podGroups []*kubeaischedv2alpha2.PodGroup,
	externalWorkloadName string) error {
	for _, podGroup := range podGroups {
		err := h.addEWNameAnnotationIfNeeded(ctx, podGroup, externalWorkloadName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *WorkloadHandler) addEWNameAnnotationIfNeeded(
	ctx context.Context,
	podGroup *kubeaischedv2alpha2.PodGroup,
	externalWorkloadName string) error {
	// If the object already has the external workload name annotation, no need to patch
	if podGroup.GetAnnotations()[commonconstants.ExternalWorkloadName] == externalWorkloadName {
		return nil
	}

	patchBytes, err := addAnnotationPatchBytes(commonconstants.ExternalWorkloadName, externalWorkloadName)
	if err != nil {
		zap.S().Errorf("Failed to json.Marshal patch - add external workload name annotation <%s> on podgroup <%s/%s>, error: %s",
			externalWorkloadName, podGroup.Namespace, podGroup.Name, err.Error())
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = h.cachedClient.Patch(ctx, podGroup, patch)
	if err != nil {
		zap.S().Errorf("Failed to patch podgroup <%s/%s> with external workload name annotation <%s>, error: %s",
			podGroup.Namespace, podGroup.Name, externalWorkloadName, err.Error())
		return err
	}
	return nil
}

func addAnnotationPatchBytes(labelKey, labelVal string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{
		"annotations": map[string]string{
			labelKey: labelVal}}})
}
