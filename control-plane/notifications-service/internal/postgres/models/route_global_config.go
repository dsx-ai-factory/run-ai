/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	RouteGlobalConfigTable = "route_global_config"
	RouteUserConfigTable   = "route_user_config"
)

type RouteGlobalConfig struct {
	gorm.Model
	TenantId    string         `gorm:"not null;primaryKey" json:"tenant_id"`
	RouteId     string         `gorm:"not null;primaryKey;unique" json:"route_id"`
	ListenerId  string         `gorm:"not null;primaryKey" json:"listener_id"`
	RouteConfig datatypes.JSON `gorm:"type:jsonb,not null,serializer:json" json:"route_config"`
}

func NewEmptyRouteGlobalConfig() *RouteGlobalConfig {
	return &RouteGlobalConfig{RouteConfig: datatypes.JSON("{}")}
}

func (r *RouteGlobalConfig) TableName() string {
	return RouteGlobalConfigTable
}

func (r *RouteGlobalConfig) GetConfig() datatypes.JSON {
	return r.RouteConfig
}

func (r *RouteGlobalConfig) SetConfig(config datatypes.JSON) {
	r.RouteConfig = config
}

type RouteUserConfig struct {
	gorm.Model
	TenantId    string         `gorm:"not null;primaryKey" json:"tenant_id"`
	UserId      string         `gorm:"not null;primaryKey" json:"user_id"`
	RouteId     string         `gorm:"not null;primaryKey;unique" json:"route_id"`
	ListenerId  string         `gorm:"not null;primaryKey" json:"listener_id"`
	RouteConfig datatypes.JSON `gorm:"type:jsonb,not null,serializer:json" json:"route_config"`
}

func (r *RouteUserConfig) TableName() string {
	return RouteUserConfigTable
}

func (r *RouteUserConfig) GetConfig() datatypes.JSON {
	return r.RouteConfig
}

func (r *RouteUserConfig) SetConfig(config datatypes.JSON) {
	r.RouteConfig = config
}
