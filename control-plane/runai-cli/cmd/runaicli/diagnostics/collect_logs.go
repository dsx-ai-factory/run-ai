package diagnostics

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/diagnostics"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/diagnostics/archive"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/diagnostics/collector"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/diagnostics/filter"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/k8s/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/k8s/validation"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	instance = config.GetInstance()

	allowedNamespaces = []string{
		"runai",
		"runai-reservation",
		"runai-backend",
		"training-operator",
		"knative",
		"gpu-operator",
		"nim-operator",
	}

	outputDirFlag  string
	noPreviousFlag bool
	namespacesFlag = append([]string{}, allowedNamespaces...)

	tarHandler           *archive.TarArchiveHandler
	tarFilePath          string
	loggerOutputFilePath string

	runErr error
)

var collectLogs = &cobra.Command{
	Use:     diagnostics.CollectLogsUse,
	Short:   diagnostics.CollectLogsShortMsg,
	Args:    cobra.NoArgs,
	Aliases: []string{diagnostics.CollectLogsAlias},
	Example: diagnostics.CollectLogsExample,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := diagnostics.ValidateOutputDir(outputDirFlag)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, diagnostics.FailedToValidateOutputDir, err)
		}

		timestamp := time.Now().Format(diagnostics.TimestampFormat)
		tarFilePath = filepath.Join(outputDirFlag, fmt.Sprintf(diagnostics.TarOutputFileNameFormat, timestamp))
		loggerOutputFilePath = filepath.Join(outputDirFlag, fmt.Sprintf(diagnostics.LoggerOutputFileNameFormat, timestamp))

		err = instance.ConfigureFileLogger(loggerOutputFilePath)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, diagnostics.FailedToConfigureFileLogging, err)
		}
		tarHandler, err = archive.NewTarArchiveHandler(tarFilePath)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, diagnostics.FailedToInitializeArchive, err)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var kubeClient *client.Client
		kubeClient, runErr = client.NewClient()
		if runErr != nil {
			return
		}

		clusterValidator := validation.NewClusterValidator(kubeClient)
		logsCollector := collector.NewLogsCollector(kubeClient, tarHandler, outputDirFlag, noPreviousFlag)
		nsFilter := filter.NewNamespacesFilter(kubeClient, namespacesFlag)
		namespacesFlagChanged := cmd.Flags().Changed(diagnostics.FlagNameNamespaces)

		collectLogsCommand := diagnostics.NewCollectLogsCommand(clusterValidator, logsCollector, nsFilter, namespacesFlagChanged)
		runErr = collectLogsCommand.Run(cmd.Context())
	},
	PostRunE: func(cmd *cobra.Command, args []string) error {
		if runErr != nil {
			diagnostics.CleanupArchive(tarFilePath)
			return cmdtxn.CmdErrorPrintAndSet(cmd, diagnostics.FailedToCollectLogs, runErr)
		}

		diagnostics.AppendLoggerOutputToArchive(tarHandler, loggerOutputFilePath)
		diagnostics.FinalizeArchive(tarHandler, tarFilePath)
		return nil
	},
}

func init() {
	Cmd.AddCommand(collectLogs)
	collectLogs.Flags().StringVar(&outputDirFlag, diagnostics.FlagNameOutputDir, diagnostics.FlagDefaultOutputDir, diagnostics.FlagHelpOutputDir)
	collectLogs.Flags().BoolVar(&noPreviousFlag, diagnostics.FlagNameNoPrevious, diagnostics.FlagDefaultNoPrevious, diagnostics.FlagHelpNoPrevious)
	collectLogs.Flags().Var(cli.NewEnumSliceFlag(allowedNamespaces, &namespacesFlag), diagnostics.FlagNameNamespaces, diagnostics.FlagHelpNamespaces)
}
