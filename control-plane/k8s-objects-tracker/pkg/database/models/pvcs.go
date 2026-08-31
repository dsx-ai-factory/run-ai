/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"

	"github.com/google/uuid"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm"
	"k8s.io/utils/ptr"
)

const (
	PvcsTable           = "pvcs"
	PvcsNameColumn      = "pvc_name"
	PvcOriginColumn     = "origin"
	pvcsTenantIdColumn  = "tenant_id"
	pvcsClusterIdColumn = "cluster_id"
	pvcsScopeTypeColumn = "scope_type"
	pvcsScopeIdColumn   = "scope_id"
)

type PvcFilters struct {
	TenantId   int32
	ClusterId  *string
	PvcName    *string
	Origin     *string
	ScopeQuery *ScopeQueryParams `gorm:"embedded"`
}

func (p *PvcFilters) Clone() *PvcFilters {
	return &PvcFilters{
		TenantId:   p.TenantId,
		ClusterId:  p.ClusterId,
		PvcName:    p.PvcName,
		Origin:     p.Origin,
		ScopeQuery: p.ScopeQuery,
	}
}

func (p *PvcFilters) Apply(ctx context.Context, db orm.DB, includeAncestors bool, includeDescendants bool) (orm.DB, *api.Error) {
	// we need all the filters except the scope query that we handle separately
	clonedFilters := p.Clone()
	clonedFilters.ScopeQuery = nil

	db = db.Where(clonedFilters)

	if p.ScopeQuery != nil {
		scopeQuery, args, err := p.ScopeQuery.BuildScopeQuery(ctx, includeAncestors, includeDescendants)
		if err != nil {
			return nil, err
		}
		db = db.Where(scopeQuery, args...)
	}

	return db, nil
}

func (p *PvcFilters) TableName() string {
	return PvcsTable
}

type Pvc struct {
	ID uuid.UUID `gorm:"primaryKey"`

	PvcFilters `gorm:"embedded"`

	Phase *string

	StorageClassName *string

	RequestedCapacity *string
	ActualCapacity    *string
	DataSharing       *bool
}

func (p *Pvc) ScopeParams() *Pvc {
	return &Pvc{
		PvcFilters: PvcFilters{
			TenantId:  p.TenantId,
			ClusterId: p.ClusterId,
			ScopeQuery: &ScopeQueryParams{
				ScopeType: p.ScopeQuery.ScopeType,
				ScopeId:   p.ScopeQuery.ScopeId,
			},
			Origin: p.Origin,
		},
	}
}

func NewPvc(filters *PvcFilters) *Pvc {
	var result Pvc
	if filters != nil {
		result.PvcFilters = *filters
	}

	return &result
}

func (p *Pvc) BeforeCreate(_ *gorm.DB) error {
	p.ID = uuid.New()
	return nil
}

func (p *Pvc) TableName() string {
	return PvcsTable
}

func (p *Pvc) TenantIdColumn() string {
	return pvcsTenantIdColumn
}

func (p *Pvc) ClusterIdColumn() string {
	return pvcsClusterIdColumn
}

func (p *Pvc) ScopeTypeColumn() string {
	return pvcsScopeTypeColumn
}

func (p *Pvc) ScopeIdColumn() string {
	return pvcsScopeIdColumn
}

func (p *Pvc) ErrorDetails() string {
	var scopeType string
	var scopeId string

	if p.ScopeQuery != nil {
		scopeType = p.ScopeQuery.ScopeType
		scopeId = p.ScopeQuery.ScopeId
	}

	return fmt.Sprintf("tenant[%d] cluster[%s] %s[%s] %s pvc[%s]",
		p.TenantId,
		ptr.Deref(p.ClusterId, ""),
		scopeType,
		scopeId,
		ptr.Deref(p.Origin, ""),
		ptr.Deref(p.PvcName, ""))
}

func (p *Pvc) ErrorMessage(title string) string {
	return fmt.Sprintf("failed to %s pvc %s", title, p.ErrorDetails())
}
