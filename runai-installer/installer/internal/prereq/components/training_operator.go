package components

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/ids"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	configinput "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/steps"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

type trainingOperator struct {
	*PreReqImpl
}

func NewTrainingOperator(k8sClient k8s.K8sClients) *trainingOperator {
	return &trainingOperator{
		PreReqImpl: NewPreReqComponentImpl(ids.PrerequisiteIdentifierTrainingOperator, "v1.8.1", false, true, k8sClient),
	}
}

func (t *trainingOperator) GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (t *trainingOperator) GetOptionalInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (t *trainingOperator) GetInstallSteps(k8sClient k8s.K8sClients, _ step.HelmInstaller, _ step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	return []step.InstallStep{
		steps.NewTrainingOperatorInstallStep(k8sClient, t.GetVersion()),
	}
}

func (t *trainingOperator) InstalledOnCluster(runaiData *model.RunAIInstallData) model.InstallPhase {
	return steps.NewTrainingOperatorInstallStep(t.k8sClient, t.GetVersion()).GetStatusSync(runaiData)
}
