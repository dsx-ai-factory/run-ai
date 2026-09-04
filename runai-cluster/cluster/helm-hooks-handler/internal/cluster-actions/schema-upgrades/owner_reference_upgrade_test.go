/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package schema_upgrades

import (
	"github.com/google/uuid"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestOwnerReferenceNoRefOnRunaiConfig(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	newUUID, _ := uuid.NewUUID()
	typeUId := types.UID(newUUID.String())
	clusterCr := &runaiv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			UID:  typeUId,
			Name: runaiv1.ClusterName,
		},
	}
	ownerRefUpgrader := newOwnerReferenceUpgrader(clusterCr)

	shouldUpgrade := ownerRefUpgrader.shouldPerformUpgrade(runaiConfig)
	assert.Equal(t, shouldUpgrade, true)

	resource, err := ownerRefUpgrader.updateRunaiConfig(runaiConfig)
	assert.Nil(t, err)
	assert.Equal(t, len(resource.ObjectMeta.OwnerReferences), 1)
	assert.Equal(t, resource.ObjectMeta.OwnerReferences[0].UID, typeUId)

	shouldUpgrade = ownerRefUpgrader.shouldPerformUpgrade(resource)
	assert.Equal(t, shouldUpgrade, false)
}

func TestOwnerReferenceRefDefinedOnRunaiConfig(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	newUUID, _ := uuid.NewUUID()
	typeUId := types.UID(newUUID.String())
	clusterCr := &runaiv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			UID:  typeUId,
			Name: runaiv1.ClusterName,
		},
	}
	runaiConfig.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		{
			UID:        typeUId,
			Name:       runaiv1.ClusterName,
			Kind:       runaiv1.ClusterKind,
			APIVersion: runaiv1.GroupVersion.String(),
		},
	}
	ownerRefUpgrader := newOwnerReferenceUpgrader(clusterCr)

	shouldUpgrade := ownerRefUpgrader.shouldPerformUpgrade(runaiConfig)
	assert.Equal(t, shouldUpgrade, false)
}
