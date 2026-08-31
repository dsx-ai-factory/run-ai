/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package components

import (
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestCombinePhases_NoArgs_ReturnsNotInstalled(t *testing.T) {
	phase := CombinePhases()
	assert.Equal(t, model.NotInstalled, phase)
}

func TestCombinePhases_AllNotInstalled(t *testing.T) {
	phase := CombinePhases(model.NotInstalled, model.NotInstalled)
	assert.Equal(t, model.NotInstalled, phase)
}

func TestCombinePhases_InstallingWinsOverInstalled(t *testing.T) {
	phase := CombinePhases(model.Installed, model.Installing)
	assert.Equal(t, model.Installing, phase)
}

func TestCombinePhases_FailedWins(t *testing.T) {
	phase := CombinePhases(model.Installed, model.Failed, model.Installing)
	assert.Equal(t, model.Failed, phase)
}

func TestCombinePhases_Installed(t *testing.T) {
	phase := CombinePhases(model.Installed, model.Installed)
	assert.Equal(t, model.Installed, phase)
}

func TestCombinePhases_MixedWithNotInstalled(t *testing.T) {
	phase := CombinePhases(model.NotInstalled, model.Installed)
	assert.Equal(t, model.Installed, phase)
}

func TestCombinePhases_SingleFailed(t *testing.T) {
	phase := CombinePhases(model.Failed)
	assert.Equal(t, model.Failed, phase)
}

func TestCombinePhases_SingleInstalling(t *testing.T) {
	phase := CombinePhases(model.Installing)
	assert.Equal(t, model.Installing, phase)
}

func TestCombinePhases_SingleInstalled(t *testing.T) {
	phase := CombinePhases(model.Installed)
	assert.Equal(t, model.Installed, phase)
}

func TestCombinePhases_ThreePhases_FailedHasPriority(t *testing.T) {
	phase := CombinePhases(model.NotInstalled, model.Installing, model.Failed)
	assert.Equal(t, model.Failed, phase)
}

func TestCombinePhases_ThreePhases_InstallingSecondPriority(t *testing.T) {
	phase := CombinePhases(model.NotInstalled, model.Installing, model.Installed)
	assert.Equal(t, model.Installing, phase)
}

func TestCombinePhases_InstallingAndNotInstalled(t *testing.T) {
	phase := CombinePhases(model.Installing, model.NotInstalled)
	assert.Equal(t, model.Installing, phase)
}

func TestCombinePhases_FailedAndNotInstalled(t *testing.T) {
	phase := CombinePhases(model.Failed, model.NotInstalled)
	assert.Equal(t, model.Failed, phase)
}
