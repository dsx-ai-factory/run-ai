/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pubsubpackage

import (
	"github.com/run-ai/runai/control-plane/workloads/pkg/common"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_crud"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

// Register registers event handlers to the pubsub
func Register() {
	runaipubsub.SubscribeAsync[events.AssetModified](services_crud.AssetModified)
	runaipubsub.SubscribeAsync[events.AssetDeleted](services_crud.AssetDeleted)
	runaipubsub.SubscribeAsync[events.TenantDelete](common.HandleTenantDeleteRequest)
	runaipubsub.SubscribeAsync[events.ClusterDelete](common.HandleClusterDeleteRequest)
	runaipubsub.SubscribeAsync[events.ProjectDeleteRequest](common.HandleProjectDeleteRequest)
}
