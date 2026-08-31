/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/run-ai/runai/control-plane/tenants-manager/generated/pkg/api"
	"gorm.io/gorm"
)

type MetricStoreType string

const (
	ExternalMetricStoreType   MetricStoreType = "external"
	GrafanaLabMetricStoreType MetricStoreType = "grafanalab"
)

type MetricStore struct {
	gorm.Model
	TenantId    uint               `gorm:"index;unique"`
	MetricStore MetricStoreDetails `gorm:"serializer:json; type:jsonb;"`
	Info        MetricStoreInfo    `gorm:"serializer:json; type:jsonb;"`
	Active      bool
}

type MetricStoreInfo struct {
	Type                 MetricStoreType `json:"type"` // grafanalab, external
	GrafanaLabIdentifier string          `json:"grafanalab_identifier"`
}

type MetricStoreDetails api.MetricStore

func (metricStore *MetricStore) ToApiModel() api.MetricStore {
	return api.MetricStore(metricStore.MetricStore)
}
