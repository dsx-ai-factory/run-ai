package components

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/ids"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	configinput "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/steps"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

type prometheus struct {
	*PreReqImpl
}

func NewPrometheus(k8sClient k8s.K8sClients) *prometheus {
	return &prometheus{
		PreReqImpl: NewPreReqComponentImpl(ids.PrerequisiteIdentifierPrometheus, "76.2.0", true, true, k8sClient),
	}
}

func (p *prometheus) GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (p *prometheus) GetOptionalInputs(runaiData *model.RunAIInstallData) []configinput.InstallInputType {
	return []configinput.InstallInputType{}
}

func (p *prometheus) GetInstallSteps(k8sClient k8s.K8sClients, _ step.HelmInstaller, _ step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	return []step.InstallStep{
		steps.NewPrometheusInstallStep(k8sClient, p.GetVersion()),
	}
}

func (p *prometheus) InstalledOnCluster(runaiData *model.RunAIInstallData) model.InstallPhase {
	return steps.NewPrometheusInstallStep(p.k8sClient, p.GetVersion()).GetStatusSync(runaiData)
}
