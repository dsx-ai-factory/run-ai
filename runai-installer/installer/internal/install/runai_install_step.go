package install

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/install/steps"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

func GetCpInstallSteps(k8sClient k8s.K8sClients, helmInstaller step.HelmInstaller, runaiClientProvider step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	installSteps := []step.InstallStep{}

	createNameSpaceStep := step.NewCreateNamespace(step.CPNamespace, k8sClient)
	installSteps = append(installSteps, createNameSpaceStep)

	installRepoAccessSecretStep := steps.NewInstallRepoAccessSecret(k8sClient)

	installSteps = append(installSteps, installRepoAccessSecretStep)

	installTlsSecretStep := step.NewInstallTlsSecret(steps.RunaiBackendTlsSecretName,
		step.CPNamespace,
		configinputs.InstallInputSSLPath,
		configinputs.InstallInputSSLKeyPath,
		configinputs.InstallInputFQDN,
		k8sClient)
	installSteps = append(installSteps, installTlsSecretStep)

	if runaiData.GetInputValue(configinputs.InstallInputCAChainPath) != "" {
		installSteps = append(installSteps, steps.NewInstallCAChainSecret(steps.RunaiBackendCAChainSecret, step.CPNamespace, step.CPReleaseName, nil, k8sClient))
	}

	installCPChartStep := steps.NewInstallCPChart(helmInstaller, runaiClientProvider)
	installSteps = append(installSteps, installCPChartStep)

	return installSteps
}

func GetClusterInstallSteps(k8sClient k8s.K8sClients, helmInstaller step.HelmInstaller, runaiClientProvider step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	installSteps := []step.InstallStep{}

	createNameSpaceStep := step.NewCreateNamespace(steps.RunaiClusterNamespace, k8sClient)
	installSteps = append(installSteps, createNameSpaceStep)

	installTlsSecretStep := step.NewInstallTlsSecret(steps.RunaiClusterTlsSecretName,
		steps.RunaiClusterNamespace,
		configinputs.InstallInputSSLPath,
		configinputs.InstallInputSSLKeyPath,
		configinputs.InstallInputFQDN,
		k8sClient)
	installSteps = append(installSteps, installTlsSecretStep)

	if runaiData.GetInputValue(configinputs.InstallInputCAChainPath) != "" {
		installSteps = append(installSteps, steps.NewInstallCAChainSecret(steps.RunaiClusterCAChainSecret, steps.RunaiClusterNamespace, step.ClusterReleaseName, steps.GetImmutableRunaiClusterCAChainSecretLablesMap(), k8sClient))
	}

	createClusterStep := steps.NewCreateCluster(runaiClientProvider)
	installSteps = append(installSteps, createClusterStep)

	getClusterInstallationDetailsStep := steps.NewGetClusterInstallationDetails(createClusterStep, runaiClientProvider)
	installSteps = append(installSteps, getClusterInstallationDetailsStep)

	installClusterChartStep := steps.NewInstallClusterChart(getClusterInstallationDetailsStep, helmInstaller, runaiClientProvider)
	installSteps = append(installSteps, installClusterChartStep)

	return installSteps
}
