package workflow

import (
	"context"
	"fmt"
	"time"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-installer/installer/pkg/api"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/config"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/state"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/types"
)

const (
	InstallationTimeout = 10 * time.Minute
)

var installerAPI InstallerAPIInterface = &InstallerAPI{}

func SetInstallerAPI(api InstallerAPIInterface) {
	installerAPI = api
}

func InstallRunAI(ctx context.Context, envID types.EnvironmentID, mode types.ProvisioningMode, domain string, wildcardDomain string, kubernetesDistribution *string, email *string, projectName *string) error {
	state.UpdateMessage(envID, "Initializing Run:AI installation")
	logging.Infof("Initializing Run:AI installation for env %s", envID)

	config, err := setupRunAIInstallConfig(domain, wildcardDomain, mode, kubernetesDistribution, email, projectName)
	if err != nil {
		return fmt.Errorf("failed to setup runai data: %w", err)
	}

	eventsChan, err := startInstallation(config)
	if err != nil {
		return fmt.Errorf("failed to start installation: %w", err)
	}

	return handleInstallationEvents(ctx, envID, eventsChan)
}

func setupRunAIInstallConfig(domain string, wildcardDomain string, mode types.ProvisioningMode, kubernetesDistribution *string, email *string, projectName *string) (*api.Config, error) {
	tlsConfig := runai_config.Get[*config.TLS]()
	jfrogConfig := runai_config.Get[*config.JFrog]()
	installConfig := runai_config.Get[*config.Installation]()

	configValues := map[string]string{
		api.FlagJfrogCreds:          jfrogConfig.Token,
		api.FlagClusterName:         installConfig.ClusterName,
		api.FlagFQDN:                domain,
		api.FlagWildcardFQDN:        wildcardDomain,
		api.FlagSSLCertPath:         tlsConfig.CertPath,
		api.FlagSSLKeyPath:          tlsConfig.KeyPath,
		api.FlagWildcardCertPath:    tlsConfig.WildcardCertPath,
		api.FlagWildcardCertKeyPath: tlsConfig.WildcardKeyPath,
	}

	if tlsConfig.CAChainPath != "" {
		configValues[api.FlagCAChainPath] = tlsConfig.CAChainPath
		logging.Infof("CA chain certificate configured at: %s", tlsConfig.CAChainPath)
	}

	if kubernetesDistribution != nil && *kubernetesDistribution != "" {
		configValues[api.FlagKubernetesDistribution] = *kubernetesDistribution
	}

	if email != nil && *email != "" {
		configValues[api.FlagResearcherUserEmail] = *email
	}

	if projectName != nil && *projectName != "" {
		configValues[api.FlagResearcherProjectName] = *projectName
	}
	runaiInstallFormat := api.SHWithResearcherSetup
	if mode == types.ProvisioningModeControlPlaneOnly {
		runaiInstallFormat = api.CPOnly
	}

	cfg := &api.Config{
		StepTimeout:        time.Duration(installConfig.StepTimeoutMinutes) * time.Minute,
		ExecutionTimeout:   time.Duration(installConfig.ExecutionTimeoutMinutes) * time.Minute,
		GPUType:            api.GPUTypeReal,
		CpVersion:          installConfig.CpVersion,
		ClusterVersion:     installConfig.ClusterVersion,
		RunaiInstallFormat: runaiInstallFormat,
		ConfigValues:       configValues,
	}

	return cfg, nil
}

func startInstallation(config *api.Config) (<-chan *api.FlowStatusReportingMsg, error) {
	err := installerAPI.Init(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize installer: %w", err)
	}

	installer := installerAPI.Get()
	eventsChan, err := installer.RunInstallationWithEventListner()
	if err != nil {
		return nil, fmt.Errorf("failed to start installation: %w", err)
	}

	return eventsChan, nil
}

func handleInstallationEvents(ctx context.Context, envID types.EnvironmentID, eventsChan <-chan *api.FlowStatusReportingMsg) error {
	hasFailure := false
	for {
		select {
		case <-ctx.Done():
			state.SetFailed(envID, "Run:AI installation canceled")
			return ctx.Err()
		case event, ok := <-eventsChan:
			if !ok {
				if hasFailure {
					state.SetFailed(envID, "Run:AI installation failed")
					return fmt.Errorf("Run:AI installation failed")
				}
				state.UpdateMessage(envID, "Run:AI installation completed successfully")
				state.SetSucceeded(envID, "Run:AI installation completed successfully")
				return nil
			}

			logging.Infof("Installation event: %+v", event)
			if processFlowStatus(envID, event) {
				hasFailure = true
			}
		}
	}
}

func processFlowStatus(envID types.EnvironmentID, event *api.FlowStatusReportingMsg) bool {
	switch event.FlowStatus {
	case api.Installing:
		state.UpdateMessage(envID, fmt.Sprintf("Installing %s", event.FlowName))
	case api.Installed:
		state.UpdateMessage(envID, fmt.Sprintf("Installed %s", event.FlowName))
	case api.Failed, api.Timeout:
		state.UpdateMessage(envID, fmt.Sprintf("Failed to install %s", event.FlowName))
		return true
	}
	return false
}
