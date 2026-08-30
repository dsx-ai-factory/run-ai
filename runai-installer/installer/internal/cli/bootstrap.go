package cli

import (
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/session"
	installerk8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

func logBanner(phase string) {
	logging.Infof("================================================================================")
	logging.Infof("=              NVIDIA Run:ai Installer - %-20s              =", phase)
	logging.Infof("================================================================================")
}

func logConfig(cfg *config.InstallConfig) {
	logging.Infof("Command line options: %+v", cfg)
	if cfg.DevMode {
		logging.Infof("Running in DEVELOPMENT MODE")
	}
}

func initK8sClient(kubeconfigPath string) (installerk8s.K8sClients, error) {
	k8sClients, err := installerk8s.GetK8sClientsWithKubeconfig(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}
	if err := installerk8s.ValidateK8sConnection(k8sClients); err != nil {
		return nil, fmt.Errorf("cannot connect to Kubernetes cluster '%s': %w", k8sClients.GetContextName(), err)
	}
	logging.Infof("Connected to Kubernetes cluster: %s", k8sClients.GetContextName())
	return k8sClients, nil
}

func initRunaiData(cfg *config.InstallConfig, k8sClients installerk8s.K8sClients) (*model.RunAIInstallData, error) {
	runaiData, err := session.SetupSessionLoggingIfFresh(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup session logging: %w", err)
	}
	AlignInstallData(runaiData, k8sClients)
	return runaiData, nil
}
