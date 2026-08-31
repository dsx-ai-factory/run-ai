/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dynamic_schema_upgrades

import (
	ra "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions"
	testu "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/test-util"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestCurrentRunaiConfig(t *testing.T) {
	dynamicClient := testu.NewTestDynamicClient(false, false)
	runaiConf := &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      runaiv1.DefaultName,
		},
	}
	runaiConf.SetDefaultsWhereNeeded()
	obj := testu.AddToDynamicClient(ra.RunaiConfigResource, runaiConf, dynamicClient)
	mpsUpgrader := newMpsServerUpgrader()
	shouldUpgrade := mpsUpgrader.shouldPerformUpgrade(obj)
	assert.Equal(t, shouldUpgrade, false)
}

func TestRecoverRunaiConfigWithBackup1(t *testing.T) {
	dynamicClient := testu.NewTestDynamicClient(false, false)
	obj := testu.AddYamlRunaiConf(dynamicClient, "../../test-util/test_resources/runaiconfig2_16.yaml")
	mpsUpgrader := newMpsServerUpgrader()
	shouldUpgrade := mpsUpgrader.shouldPerformUpgrade(obj)
	assert.Equal(t, shouldUpgrade, true)
	imageStr := mpsUpgrader.extractImageStringAfterValidation(obj)
	assert.Equal(t, imageStr, "registry.access.redhat.com/ubi8/ubi-minimal:8.9")
}
