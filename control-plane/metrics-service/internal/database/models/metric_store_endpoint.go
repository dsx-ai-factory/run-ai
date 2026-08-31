/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type MetricStoreEndpoint struct {
	Url     string                   `gorm:"not null;"`
	Auth    *MetricStoreEndpointAuth `gorm:"type:jsonb;not null;"`
	Headers map[string]string
}
