/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenants

type TenantsConfiguration struct {
	ServiceUrl string `mapstructure:"service_url" validate:"required,url"`
}

func (tc *TenantsConfiguration) SetDefaults() {
	tc.ServiceUrl = "http://runai-backend-tenants-manager:8080"
}
