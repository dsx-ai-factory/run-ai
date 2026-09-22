/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	cluster "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
)

func NewGetClusterInstallationDetailsPreset(clusterUuid, clientSecret string) *GetClusterInstallationDetails {
	createCluster := &CreateCluster{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Create cluster (preset)"),
		clusterUuid:          clusterUuid,
	}
	createCluster.SetStatus(model.Installed, nil)

	installInfo := cluster.NewClusterInstallationInfoResponse("", "", "", "")
	installInfo.SetClientSecret(clientSecret)
	details := &GetClusterInstallationDetails{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Get cluster installation details (preset)"),
		createClusterStep:    createCluster,
		installInfo:          installInfo,
	}
	details.SetStatus(model.Installed, nil)
	return details
}

func NewCreateClusterPreset(clusterUuid string) *CreateCluster {
	c := &CreateCluster{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Create cluster (preset)"),
		clusterUuid:          clusterUuid,
	}
	c.SetStatus(model.Installed, nil)
	return c
}
