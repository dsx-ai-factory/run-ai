/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package structs

import (
	openapi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
)

type PolicyWorkloadTypeConfig struct {
	CategoryName   *string
	PriorityName   *string
	Preemptibility *openapi.Preemptibility
}
