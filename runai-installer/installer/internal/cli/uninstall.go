package cli

import (
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator"
	"github.com/run-ai/runai/runai-installer/installer/internal/uninstall"
	"github.com/spf13/cobra"
)

func newUninstallCmd() *cobra.Command {
	var executionTimeout time.Duration

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Run:ai from the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &config.UninstallConfig{
				DevMode:          globals.devMode,
				LogsDir:          globals.logsDir,
				StepTimeout:      globals.stepTimeout,
				ExecutionTimeout: executionTimeout,
			}

			if err := logging.InitFromDir(cfg.LogsDir); err != nil {
				return err
			}
			logBanner("Uninstall Starting")
			logging.Infof("Uninstall config: %+v", cfg)
			if cfg.DevMode {
				logging.Infof("Running in DEVELOPMENT MODE")
			}

			deps, err := uninstall.ProductionDeps()
			if err != nil {
				return err
			}

			if err := orchestrator.RunUninstall(cfg, deps); err != nil {
				return err
			}

			logBanner("Uninstall Finished")
			return nil
		},
	}

	cmd.Flags().DurationVar(&executionTimeout, "execution-timeout", 30*time.Minute, "total timeout")

	return cmd
}
