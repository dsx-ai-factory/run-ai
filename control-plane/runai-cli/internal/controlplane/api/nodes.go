/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"

type NodeTelemetriesByType map[openapi.NodeTelemetryType]*openapi.TelemetryResponse
