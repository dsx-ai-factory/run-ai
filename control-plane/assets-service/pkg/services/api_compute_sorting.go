/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"sort"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/docker/go-units"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
)

func SortComputeAssets(assets []api.ComputeAsset) {
	sort.Sort(&ComputeAssetSorter{assets: assets})
}

// GpuRequestType holds the type of GPU request and determines the primary sorting order:
type GpuRequestType int

const (
	NoGpuRequest         GpuRequestType = 0 // Assets with no GPU spec will appear first
	GpuRequestByMemory   GpuRequestType = 1 // Assets specifying GPU by memory size should appear second
	GpuRequestByQuantity GpuRequestType = 2 // Assets specifying GPU by quantity should appear third
)

func GpuRequestTypeOf(asset *api.ComputeAsset) GpuRequestType {
	gpuRequestType := ""
	if asset.Spec.GpuRequestType != nil {
		gpuRequestType = string(*asset.Spec.GpuRequestType)
	}
	switch gpuRequestType {
	case GpuRequestTypeMemory:
		return GpuRequestByMemory
	case GpuRequestTypePortion:
		return GpuRequestByQuantity
	default:
		return NoGpuRequest
	}
}

// ComputeAssetSorter type is intended for sorting compute assets
type ComputeAssetSorter struct {
	sort.Interface // the sorting interface that dictates the functions to be implemented
	assets         []api.ComputeAsset
}

func (s *ComputeAssetSorter) Len() int {
	return len(s.assets)
}

func (s *ComputeAssetSorter) Swap(i, j int) {
	s.assets[i], s.assets[j] = s.assets[j], s.assets[i]
}

func (s *ComputeAssetSorter) Less(i, j int) bool {
	ci := s.assets[i]
	cj := s.assets[j]

	//
	//   if two GPU requests are not of the same type, the request type determines the order
	//   between them (see GpuRequestType type)
	//
	ciGpuReqType := GpuRequestTypeOf(&ci)
	cjGpuReqType := GpuRequestTypeOf(&cj)

	if ciGpuReqType != cjGpuReqType {
		return ciGpuReqType < cjGpuReqType
	}

	//
	//   for requests of the same type, the quantity determines the internal order between them
	//
	switch ciGpuReqType {
	case GpuRequestByQuantity:
		return convutils.Deref(ci.Spec.GpuPortionRequest, 0) < convutils.Deref(cj.Spec.GpuPortionRequest, 0)

	case GpuRequestByMemory:
		ciMem, ciMemErr := units.FromHumanSize(convutils.Deref(ci.Spec.GpuMemoryRequest, "0"))
		cjMem, cjMemErr := units.FromHumanSize(convutils.Deref(cj.Spec.GpuMemoryRequest, "0"))
		if ciMemErr == nil && cjMemErr == nil {
			return ciMem < cjMem
		}
	}

	//
	//   in any case of fallback, sort by name
	//
	return ci.Meta.Name < cj.Meta.Name
}

func GpuRequestTypePtr(originalValue string) *api.GpuRequestType {
	result := api.GpuRequestType(originalValue)
	return &result
}
