/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package event_handlers

import (
	"context"
	"fmt"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/exporters/podgroup_exporter"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/metrics"
)

type MigPodGroupEventHandler struct {
	k8sClient client.Client
}

func NewMigPodGroupEventHandler(k8sClient client.Client) *MigPodGroupEventHandler {
	return &MigPodGroupEventHandler{k8sClient: k8sClient}
}

func (pe *MigPodGroupEventHandler) ResetMetrics(identifier ObjectIdentifier) {
	metrics.ResetMigPodGroupMetrics(identifier.Name)
}

func (pe *MigPodGroupEventHandler) SetMetrics(ctx context.Context, eventObject interface{}) error {
	podGroup, ok := eventObject.(*kaiv2alpha2.PodGroup)
	if !ok {
		return fmt.Errorf("eventObject is not a podgroup")
	}

	exporter := podgroup_exporter.NewPodGroupExporter(ctx, pe.k8sClient, podGroup)
	return metrics.SetMigPodGroupMetrics(exporter)
}
