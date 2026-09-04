/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pubsub

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/services"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
)

// Register registers event handlers to the pubsub
func Register() {
	runaipubsub.SubscribeAsync[events.ProjectDeleted](services.HandleProjectDeletedEvent)
	runaipubsub.SubscribeAsync[events.ClusterDelete](services.HandleClusterDeleteEvent)
	runaipubsub.SubscribeAsync[events.TenantDelete](services.HandleTenantDeleteEvent)
}
