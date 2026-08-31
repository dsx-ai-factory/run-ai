/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package karta

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Karta GVK-routing labels. These are applied by upstream producers onto Karta
// CRs so we can look up the right Karta for a given workload GVK.
const (
	KartaGroupLabel   = "run.ai/karta-group"
	KartaVersionLabel = "run.ai/karta-version"
	KartaKindLabel    = "run.ai/karta-kind"
)

// KartaConditionType enumerates the condition types we read off Karta CRs.
type KartaConditionType string

const (
	Ready KartaConditionType = "Ready"
)

// GetKartaPerGvkLabelSelectors returns a ListOptions that filters Karta CRs by
// matching their GVK labels against the supplied GVK.
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
