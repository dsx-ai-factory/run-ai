package components

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/install/steps"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

type airgapPreparation struct{}

func NewAirgapPreparation() step.InstallationComponent {
	return &airgapPreparation{}
}

func (a *airgapPreparation) GetName() string {
	return "Airgap Preparation"
}

func (a *airgapPreparation) ComponentType() step.ComponentType {
	return step.ComponentTypeAirgapPreparation
}

func (a *airgapPreparation) GetMandatoryInputs(_ *model.RunAIInstallData) []configinputs.InstallInputType {
	return []configinputs.InstallInputType{
		configinputs.InstallInputAirgapPackagePath,
		configinputs.InstallInputRegistryURL,
	}
}

func (a *airgapPreparation) GetOptionalInputs(_ *model.RunAIInstallData) []configinputs.InstallInputType {
	return []configinputs.InstallInputType{
		configinputs.InstallInputRegistryUsername,
		configinputs.InstallInputRegistryPassword,
	}
}

func (a *airgapPreparation) GetInstallSteps(_ k8s.K8sClients, _ step.HelmInstaller, _ step.RunAIAPIClientProvider, _ *model.RunAIInstallData) []step.InstallStep {
	return []step.InstallStep{
		steps.NewExtractAirgapPackage(),
		steps.NewPushAirgapImages(),
	}
}
