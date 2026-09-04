/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"fmt"

	swagger "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KartaReadyPhase    = "Ready"
	KartaNotReadyPhase = "NotReady"
	KartaDeletingPhase = "Deleting"
	KartaDeletedPhase  = "Deleted"
)

func ParseBackendKarta(karta *kartav1alpha1.Karta, isDeleted bool) (swagger.WorkloadTypeForSync, error) {
	if karta.Spec.StructureDefinition.RootComponent.Kind == nil {
		return swagger.WorkloadTypeForSync{}, fmt.Errorf("failed to get the version for Karta %s", karta.Name)
	}
	workloadTypeForSyncVersion := &karta.Spec.StructureDefinition.RootComponent.Kind.Version
	workloadTypeForSyncId, ok := karta.Labels[v1alpha1.WorkloadTypeIdLabel]
	if !ok || workloadTypeForSyncId == "" {
		return swagger.WorkloadTypeForSync{}, fmt.Errorf("failed to get backend ID for Karta %s", karta.Name)
	}

	var phase string
	workloadTypeForSyncConditions := []swagger.Condition{}
	if isDeleted {
		phase = KartaDeletedPhase
	} else {
		phase, workloadTypeForSyncConditions = ParseBackendKartaConditions(karta)
	}

	workloadTypeForSync := swagger.WorkloadTypeForSync{
		Phase:      &phase,
		Version:    workloadTypeForSyncVersion,
		Id:         workloadTypeForSyncId,
		Conditions: workloadTypeForSyncConditions,
	}

	return workloadTypeForSync, nil
}

func ParseBackendKartaConditions(karta *kartav1alpha1.Karta) (string, []swagger.Condition) {
	resourceInterfaceConditions := karta.Status.Conditions
	syncEnabled := false
	result := []swagger.Condition{}
	for _, condition := range resourceInterfaceConditions {
		result = append(result, swagger.Condition{
			Type:               condition.Type,
			Status:             string(condition.Status),
			Message:            strPointerIfNotEmpty(condition.Message),
			Reason:             strPointerIfNotEmpty(condition.Reason),
			LastTransitionTime: getApiTime(condition.LastTransitionTime),
		})

		if condition.Type == string(v1alpha1.SyncEnabled) &&
			condition.Status == metav1.ConditionTrue {
			syncEnabled = true
		}
	}

	phase := KartaNotReadyPhase
	if !karta.DeletionTimestamp.IsZero() {
		phase = KartaDeletingPhase
	} else if syncEnabled {
		phase = KartaReadyPhase
	}

	return phase, result
}
