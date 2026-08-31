/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	WorkloadTypeIdLabel = "run.ai/workload-type-id"
	KartaGroupLabel     = "run.ai/karta-group"
	KartaVersionLabel   = "run.ai/karta-version"
	KartaKindLabel      = "run.ai/karta-kind"
	KartaUpdatedAtLabel = "run.ai/karta-updated-at"
)

type KartaConditionType string

const (
	CRDExists      KartaConditionType = "CRDExists"
	RBACReady      KartaConditionType = "RBACReady"
	SyncEnabled    KartaConditionType = "SyncEnabled"
	KartaValidated KartaConditionType = "KartaValidated"
	RunaiReady     KartaConditionType = "RunaiReady"
	Ready          KartaConditionType = "Ready"

	// deletion types
	SyncDisabled        KartaConditionType = "SyncDisabled"
	NoWorkloadsUseKarta KartaConditionType = "NoWorkloadsUseKarta"
	RBACCleanupSuccess  KartaConditionType = "RBACCleanupSuccess"
)

type KartaMatchedPhase string

const (
	KartaMatchedPhaseRunning      KartaMatchedPhase = "Running"
	KartaMatchedPhaseFailed       KartaMatchedPhase = "Failed"
	KartaMatchedPhaseCompleted    KartaMatchedPhase = "Completed"
	KartaMatchedPhaseInitializing KartaMatchedPhase = "Initializing"
	KartaMatchedPhaseUnknown      KartaMatchedPhase = "Unknown"
)

func GetKartaLabels(workloadTypeId string, group, version, kind, updatedAt string) map[string]string {
	return map[string]string{
		WorkloadTypeIdLabel: workloadTypeId,
		KartaGroupLabel:     group,
		KartaVersionLabel:   version,
		KartaKindLabel:      kind,
		KartaUpdatedAtLabel: updatedAt,
	}
}

func GetKartaPerGvkLabelSelectors(gvk schema.GroupVersionKind) *client.ListOptions {
	labelSelector := client.MatchingLabels{
		KartaGroupLabel: gvk.Group,
		KartaKindLabel:  gvk.Kind,
	}
	if gvk.Version != "" {
		labelSelector[KartaVersionLabel] = gvk.Version
	}
	listOpts := &client.ListOptions{}
	labelSelector.ApplyToList(listOpts)
	return listOpts
}
