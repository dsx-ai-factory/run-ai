package cli

import (
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator"
	"github.com/run-ai/runai/runai-installer/installer/internal/upgrade"
	"github.com/spf13/cobra"
)

func redactUpgradeConfig(cfg *config.UpgradeConfig) *config.UpgradeConfig {
	safe := *cfg
	if safe.RegistryUsername != "" {
		safe.RegistryUsername = "<redacted>"
	}
	if safe.RegistryPassword != "" {
		safe.RegistryPassword = "<redacted>"
	}
	return &safe
}

func newUpgradeCmd() *cobra.Command {
	var (
		executionTimeout    time.Duration
		env                 config.RunaiEnvType
		clusterEnv          config.RunaiEnvType
		controlPlaneVersion string
		clusterVersion      string
		upgradeControlPlane bool
		upgradeCluster      bool
		fqdn                string
		clusterUUID         string
		caCertPath          string
		airgapPackagePath   string
		registryURL         string
		registryUsername    string
		registryPassword    string
	)

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Run:ai on the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			var clusterEnvPtr *config.RunaiEnvType
			if cmd.Flags().Changed("cluster-env") {
				clusterEnvPtr = &clusterEnv
			}
			cfg := &config.UpgradeConfig{
				DevMode:             globals.devMode,
				LogsDir:             globals.logsDir,
				StepTimeout:         globals.stepTimeout,
				ExecutionTimeout:    executionTimeout,
				Env:                 env,
				ClusterEnv:          clusterEnvPtr,
				ControlPlaneVersion: controlPlaneVersion,
				ClusterVersion:      clusterVersion,
				UpgradeControlPlane: upgradeControlPlane,
				UpgradeCluster:      upgradeCluster,
				FQDN:                fqdn,
				ClusterUUID:         clusterUUID,
				CACertPath:          caCertPath,
				AirgapPackagePath:   airgapPackagePath,
				RegistryURL:         registryURL,
				RegistryUsername:    registryUsername,
				RegistryPassword:    registryPassword,
			}

			if err := logging.InitFromDir(cfg.LogsDir); err != nil {
				return err
			}
			logBanner("Upgrade Starting")
			logging.Infof("Upgrade config: %+v", redactUpgradeConfig(cfg))
			if cfg.DevMode {
				logging.Infof("Running in DEVELOPMENT MODE")
			}

			k8sClients, err := initK8sClient("")
			if err != nil {
				return err
			}
			deps := upgrade.ProductionDeps(k8sClients, cfg.ControlPlaneVersion)

			if err := orchestrator.RunUpgrade(cfg, deps); err != nil {
				return err
			}

			logBanner("Upgrade Finished")
			return nil
		},
	}

	cmd.Flags().DurationVar(&executionTimeout, "execution-timeout", 50*time.Minute, "total timeout")
	cmd.Flags().Var(&env, "env", "environment (prod or test)")
	cmd.Flags().Var(&clusterEnv, "cluster-env", "cluster chart environment (defaults to --env)")
	cmd.Flags().StringVar(&controlPlaneVersion, "control-plane-version", ">0.0.0-0", "target control plane version")
	cmd.Flags().StringVar(&clusterVersion, "cluster-version", ">0.0.0-0", "target cluster version")
	cmd.Flags().BoolVar(&upgradeControlPlane, "upgrade-control-plane", true, "upgrade the control plane")
	cmd.Flags().BoolVar(&upgradeCluster, "upgrade-cluster", true, "upgrade the cluster")
	cmd.Flags().StringVar(&fqdn, "fqdn", "", "control plane domain (auto-detected from existing install if omitted)")
	cmd.Flags().StringVar(&clusterUUID, "cluster-uuid", "", "cluster UUID (auto-detected from existing install if omitted)")
	cmd.Flags().StringVar(&caCertPath, "ca-cert-path", "", "path to CA certificate (for self-signed certs)")
	cmd.Flags().StringVar(&airgapPackagePath, "airgap-package", "", "path to airgap tar.gz package (enables airgap upgrade)")
	cmd.Flags().StringVar(&registryURL, "registry-url", "", "container registry URL for airgap images")
	cmd.Flags().StringVar(&registryUsername, "registry-username", "", "container registry username for airgap")
	cmd.Flags().StringVar(&registryPassword, "registry-password", "", "container registry password for airgap")

	return cmd
}
