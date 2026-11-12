/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
)

func (o OrgUnitApiService) SimulateNodePoolFairShare(_ context.Context, _ string, _ api.FairShareSimulationRequest) (api.ImplResponse, error) {
	return api.Response(http.StatusNotImplemented, api.NewError(http.StatusNotImplemented, "fair share simulation is not implemented yet", nil)), nil
}
