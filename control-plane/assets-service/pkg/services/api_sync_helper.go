/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
)

var (
	departmentScope = api.SCOPE_DEPARTMENT //nolint:unused // shared by service test fixtures
	tenantScope     = api.SCOPE_TENANT     //nolint:unused // shared by service test fixtures
	projectScope    = api.SCOPE_PROJECT
	clusterScope    = api.SCOPE_CLUSTER //nolint:unused // shared by service test fixtures
)
