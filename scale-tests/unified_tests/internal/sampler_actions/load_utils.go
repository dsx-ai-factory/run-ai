/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sampler_actions

import (
	"github.com/mroth/weightedrand"
)

type IndexMapper interface {
	GetIndex(currentIndex int) int
}

type UniformIndexMapper struct {
	targetMappingLength int
}

func NewUniformIndexMapper(targetMappingLength int) *UniformIndexMapper {
	return &UniformIndexMapper{targetMappingLength: targetMappingLength}
}

func (u *UniformIndexMapper) GetIndex(currentIndex int) int {
	return currentIndex % u.targetMappingLength
}

type HeavyIndexMapper struct {
	targetMappingLength int
	numberOfIndexToLoad int
	chooser             *weightedrand.Chooser
}

func NewHeavyIndexMapper(targetMappingLength int, numberOfIndexToLoad int) *HeavyIndexMapper {
	choices := []weightedrand.Choice{}
	for i := 0; i < targetMappingLength; i++ {
		if i < numberOfIndexToLoad {
			choices = append(choices, weightedrand.Choice{Item: i, Weight: 5})
		} else {
			choices = append(choices, weightedrand.Choice{Item: i, Weight: 1})
		}
	}
	chooser, _ := weightedrand.NewChooser(choices...)
	return &HeavyIndexMapper{targetMappingLength: targetMappingLength, numberOfIndexToLoad: numberOfIndexToLoad, chooser: chooser}
}

func (h *HeavyIndexMapper) GetIndex(currentIndex int) int {
	return h.chooser.Pick().(int)
}
