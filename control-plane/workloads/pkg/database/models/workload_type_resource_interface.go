/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
	"time"

	"github.com/google/uuid"
	karta "github.com/run-ai/karta/pkg/api/runai/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type WorkloadTypeKarta struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantId       *string   `json:"tenant_id"`
	ScopeType      string    `json:"scope_type"`
	WorkloadTypeId uuid.UUID `json:"workload_type_id"`
	Version        string    `json:"version"`
	Karta          Karta     `gorm:"type:jsonb" json:"karta"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      *string   `json:"created_by"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (WorkloadTypeKarta) TableName() string {
	return "workload_type_kartas"
}

func (wtk *WorkloadTypeKarta) BeforeCreate(tx *gorm.DB) (err error) {
	wtk.ID = uuid.New()
	return
}

type Karta karta.Karta

func (r Karta) GroupVersionKind() *schema.GroupVersionKind {
	gvk := r.Spec.StructureDefinition.RootComponent.Kind
	if gvk == nil {
		return nil
	}
	return &schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}
}

func (r Karta) IsEmpty() bool {
	return r.Spec.StructureDefinition.RootComponent.Kind == nil || r.Spec.StructureDefinition.RootComponent.Kind.Kind == "" ||
		r.Spec.StructureDefinition.RootComponent.Kind.Version == ""
}

func (r Karta) Value() (driver.Value, error) {
	valueString, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return string(valueString), err
}

func (r *Karta) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &r); err != nil {
		return err
	}
	return nil
}

// WorkloadTypeKartaArray is a JSON array wrapper for scanning into []WorkloadTypeKarta
type WorkloadTypeKartaArray []WorkloadTypeKarta

func (a WorkloadTypeKartaArray) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (a *WorkloadTypeKartaArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	if err := json.Unmarshal(value.([]byte), &a); err != nil {
		return err
	}
	return nil
}
