/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"context"

	"github.com/run-ai/runai/control-plane/tenants-manager/generated/pkg/api"

	"gorm.io/gorm"
)

type TenantStatus string

const (
	TenantCreating TenantStatus = "Creating"
	TenantFailed   TenantStatus = "Failed"
	TenantDeleting TenantStatus = "Deleting"
	TenantReady    TenantStatus = "Ready"
)

const (
	ContractTypeNormal string = "normal"
)

type Tenant struct {
	gorm.Model
	Uuid         string       `gorm:"not null;unique_index"`
	MimirId      string       `gorm:"not null,unique"`
	Name         string       `gorm:"not null;unique"`
	DisplayName  string       `gorm:"not null;unique"`
	Alias        *string      `gorm:"unique"`
	Status       TenantStatus `gorm:"not null"`
	ContractType string       `gorm:"default:'normal'"`
	Smg          bool
}

func (tenant *Tenant) ToApiModel(ctx context.Context) api.Tenant {
	t := api.Tenant{
		Id:           int32(tenant.ID),
		Uuid:         tenant.Uuid,
		MimirId:      tenant.MimirId,
		Name:         tenant.Name,
		DisplayName:  tenant.DisplayName,
		Status:       string(tenant.Status),
		CreatedAt:    tenant.CreatedAt,
		UpdatedAt:    tenant.UpdatedAt,
		ContractType: tenant.ContractType,
		Smg:          tenant.Smg,
	}

	if tenant.Alias != nil {
		t.Alias = *tenant.Alias
	}

	return t
}
