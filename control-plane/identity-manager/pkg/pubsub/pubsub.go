/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pubsub

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/services"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
)

func ErrorHandler(err error) {
	logging.Errorf("Error in pubsub: %v", err)
}

func Register(serviceAccountsService *services.ServiceAccountsApiService) {
	runaipubsub.SubscribeAsync[events.ClusterDelete](serviceAccountsService.ClusterDeleteCallback)
}
