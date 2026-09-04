/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package page

type IdsMapper struct {
	maxItems     int
	duplicateMap map[string]bool
}

func NewIdsMapper(maxItems int) *IdsMapper {
	return &IdsMapper{maxItems: maxItems, duplicateMap: make(map[string]bool)}
}

func (im *IdsMapper) ShouldAdd(id string) bool {
	if _, found := im.duplicateMap[id]; !found {
		im.duplicateMap[id] = true
		return true
	}
	return false
}

func (im *IdsMapper) ReachedLimit() bool {
	if im.maxItems < 1 {
		return false
	}
	return im.GetLen() >= im.maxItems
}

func (im *IdsMapper) GetLen() int {
	return len(im.duplicateMap)
}
