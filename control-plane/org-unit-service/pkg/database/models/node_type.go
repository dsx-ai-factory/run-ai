/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NodeType struct {
	IntBaseModel
	Name      string
	ClusterId uuid.UUID
	TenantId  string
}

func (NodeType) TableName() string {
	return NodeTypeTable
}

func (t NodeType) Clone() NodeType {
	return NodeType{
		IntBaseModel: IntBaseModel{
			ID:        t.ID,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
			DeletedAt: gorm.DeletedAt{
				Time:  t.DeletedAt.Time,
				Valid: t.DeletedAt.Valid,
			},
			CreatedBy: t.CreatedBy,
			UpdatedBy: t.UpdatedBy,
			DeletedBy: t.DeletedBy,
		},
		Name:      t.Name,
		ClusterId: t.ClusterId,
		TenantId:  t.TenantId,
	}
}

func NodeTypesArrayEqual(nt1, nt2 *[]NodeType) bool {
	if nt1 == nil {
		return nt2 == nil
	}
	if nt2 == nil {
		return false
	}
	if len(*nt1) != len(*nt2) {
		return false
	}

	nt1Map := map[int64]NodeType{}

	for _, nodeType := range *nt1 {
		nt1Map[nodeType.ID] = nodeType
	}

	for _, nt2NodeType := range *nt2 {
		nt1NodeType, found := nt1Map[nt2NodeType.ID]
		if !found {
			return false
		}

		if !NodeTypesEqual(nt1NodeType, nt2NodeType) {
			return false
		}
	}

	return true
}

func NodeTypesEqual(nt1, nt2 NodeType) bool {
	if nt1.Name != nt2.Name ||
		nt1.ClusterId != nt2.ClusterId ||
		nt1.TenantId != nt2.TenantId {
		return false
	}

	return true
}

func convertRawToNodeType(raw map[string]interface{}) NodeType {
	nodeType := NodeType{}
	nodeType.IntBaseModel = convertRawToIntBaseModel(raw)

	if raw["name"] != nil {
		nodeType.Name = raw["name"].(string)
	}
	if raw["cluster_id"] != nil {
		nodeType.ClusterId = uuid.MustParse(raw["cluster_id"].(string))
	}
	if raw["tenant_id"] != nil {
		nodeType.TenantId = raw["tenant_id"].(string)
	}

	return nodeType
}
