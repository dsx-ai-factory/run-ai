/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"encoding/json"
	"fmt"

	workloadsManagerApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func convertStorageConfig(apiPvcs []workloadsManagerApi.WorkloadV2UpdateConfigurationPvc, pvcsConfig *workloadsManagerApi.WorkloadV2UpdateConfigurationPvcsConfig) (*dromosv1alpha1.StorageConfig, error) {
	pvcs := make([]corev1.PersistentVolumeClaim, 0, len(apiPvcs))
	for _, apiPvc := range apiPvcs {
		pvc, err := NewPersistentVolumeClaimFromApiPVC(&apiPvc)
		if err != nil {
			return nil, fmt.Errorf("pvc config to PersistentVolumeClaim %q: %w", apiPvc.Metadata.Name, err)
		}
		pvcs = append(pvcs, pvc)
	}

	storageConfig := &dromosv1alpha1.StorageConfig{NewPvcs: pvcs}
	// Leave CleanupOnCompletion nil when unset so the cluster-side default applies.
	if pvcsConfig != nil {
		storageConfig.CleanupOnCompletion = pvcsConfig.CleanupOnCompletion.Get()
	}
	return storageConfig, nil
}

// NewPersistentVolumeClaimFromApiPVC builds a corev1.PersistentVolumeClaim
// The Spec is converted as-is  because the API type is a structural subset of corev1.PersistentVolumeClaimSpec
// Adding a new field to the API auto-propagates as long as it stays a valid subset
// Metadata is mapped explicitly because its surface stays small and may need Run:ai-specific behavior. such as label prefixing or annotation validation.
func NewPersistentVolumeClaimFromApiPVC(apiPvc *workloadsManagerApi.WorkloadV2UpdateConfigurationPvc) (corev1.PersistentVolumeClaim, error) {
	specJSON, err := json.Marshal(&apiPvc.Spec)
	if err != nil {
		return corev1.PersistentVolumeClaim{}, fmt.Errorf("marshal api pvc spec: %w", err)
	}

	var spec corev1.PersistentVolumeClaimSpec
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return corev1.PersistentVolumeClaim{}, fmt.Errorf("unmarshal as PersistentVolumeClaimSpec: %w", err)
	}

	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: apiPvc.Metadata.Name},
		Spec:       spec,
	}

	if apiPvc.Metadata.Labels != nil {
		pvc.Labels = *apiPvc.Metadata.Labels
	}
	if apiPvc.Metadata.Annotations != nil {
		pvc.Annotations = *apiPvc.Metadata.Annotations
	}

	return pvc, nil
}
