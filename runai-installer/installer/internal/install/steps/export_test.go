package steps

import (
	"github.com/run-ai/runai/runai-installer/clients/runai/cluster"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
)

func NewGetClusterInstallationDetailsPreset(clusterUuid, clientSecret string) *GetClusterInstallationDetails {
	createCluster := &CreateCluster{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Create cluster (preset)"),
		clusterUuid:          clusterUuid,
	}
	createCluster.SetStatus(model.Installed, nil)

	details := &GetClusterInstallationDetails{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Get cluster installation details (preset)"),
		createClusterStep:    createCluster,
		installInfo:          cluster.NewClusterInstallationInfoResponse("", "", "", clientSecret),
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
