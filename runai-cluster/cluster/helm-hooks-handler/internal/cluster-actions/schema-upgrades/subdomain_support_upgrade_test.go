/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package schema_upgrades

import (
	"testing"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestSubdomainSupportNotSet_Pre225(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.Spec.Global = &global.Global{}
	upgrader := newSubdomainSupportUpgrader("2.24.0")

	assert.True(t, upgrader.shouldPerformUpgrade(runaiConfig))

	resource, err := upgrader.updateRunaiConfig(runaiConfig)
	assert.Nil(t, err)
	assert.Equal(t, ptr.To(false), resource.Spec.Global.SubdomainSupport)

	assert.False(t, upgrader.shouldPerformUpgrade(resource))
}

func TestSubdomainSupportNilGlobal_Pre225(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	upgrader := newSubdomainSupportUpgrader("2.24.0")

	assert.True(t, upgrader.shouldPerformUpgrade(runaiConfig))

	resource, err := upgrader.updateRunaiConfig(runaiConfig)
	assert.Nil(t, err)
	assert.NotNil(t, resource.Spec.Global)
	assert.Equal(t, ptr.To(false), resource.Spec.Global.SubdomainSupport)
}

func TestSubdomainSupportAlreadyFalse(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.Spec.Global = &global.Global{
		SubdomainSupport: ptr.To(false),
	}
	upgrader := newSubdomainSupportUpgrader("2.24.0")

	assert.False(t, upgrader.shouldPerformUpgrade(runaiConfig))
}

func TestSubdomainSupportAlreadyTrue(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.Spec.Global = &global.Global{
		SubdomainSupport: ptr.To(true),
	}
	upgrader := newSubdomainSupportUpgrader("2.24.0")

	assert.False(t, upgrader.shouldPerformUpgrade(runaiConfig))
}

func TestSubdomainSupportNotSet_Post225(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.Spec.Global = &global.Global{}
	upgrader := newSubdomainSupportUpgrader("2.25.0")

	assert.False(t, upgrader.shouldPerformUpgrade(runaiConfig))
}

func TestSubdomainSupportNotSet_Post225Patch(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.Spec.Global = &global.Global{}
	upgrader := newSubdomainSupportUpgrader("2.25.1")

	assert.False(t, upgrader.shouldPerformUpgrade(runaiConfig))
}

func TestSubdomainSupportNotSet_EmptyVersion(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.Spec.Global = &global.Global{}
	upgrader := newSubdomainSupportUpgrader("")

	assert.True(t, upgrader.shouldPerformUpgrade(runaiConfig))
}

func TestSubdomainSupportDoubleUpgrade(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.Spec.Global = &global.Global{}

	firstUpgrader := newSubdomainSupportUpgrader("2.24.0")
	assert.True(t, firstUpgrader.shouldPerformUpgrade(runaiConfig))
	resource, err := firstUpgrader.updateRunaiConfig(runaiConfig)
	assert.Nil(t, err)
	assert.Equal(t, ptr.To(false), resource.Spec.Global.SubdomainSupport)

	secondUpgrader := newSubdomainSupportUpgrader("2.25.0")
	assert.False(t, secondUpgrader.shouldPerformUpgrade(resource))
}

func TestSubdomainSupportNotSet_DevVersion(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.Spec.Global = &global.Global{}
	upgrader := newSubdomainSupportUpgrader("0.0.0-20012642.jonathanrun-37063-ena-c7ac382")

	assert.False(t, upgrader.shouldPerformUpgrade(runaiConfig))
}
