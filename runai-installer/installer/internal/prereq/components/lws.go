package components

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/ids"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	configinput "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/steps"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

type lws struct {
	*PreReqImpl
}

func NewLWS(k8sClient k8s.K8sClients) *lws {
	return &lws{
		PreReqImpl: NewPreReqComponentImpl(ids.PrerequisiteIdentifierLWS, "0.7.0", false, false, k8sClient),
	}
}

func (l *lws) GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (l *lws) GetOptionalInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (l *lws) GetInstallSteps(k8sClient k8s.K8sClients, _ step.HelmInstaller, _ step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	return []step.InstallStep{
		steps.NewLWSInstallStep(k8sClient, l.GetVersion()),
	}
}

func (l *lws) InstalledOnCluster(runaiData *model.RunAIInstallData) model.InstallPhase {
	return steps.NewLWSInstallStep(l.k8sClient, l.GetVersion()).GetStatusSync(runaiData)
}
