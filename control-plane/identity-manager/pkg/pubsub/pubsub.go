/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pubsub

import (
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/services"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

func ErrorHandler(err error) {
	logging.Errorf("Error in pubsub: %v", err)
}

func Register(serviceAccountsService *services.ServiceAccountsApiService) {
	runaipubsub.SubscribeAsync[events.ClusterDelete](serviceAccountsService.ClusterDeleteCallback)
}
