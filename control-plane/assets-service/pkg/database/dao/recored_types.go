/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import "github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"

type AssetRecord interface {
	GetApiAssetMeta() api.AssetMeta
	GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes
}
