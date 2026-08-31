/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"

type NodeTelemetriesByType map[openapi.NodeTelemetryType]*openapi.TelemetryResponse
