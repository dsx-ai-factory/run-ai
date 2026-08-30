package diagnostics

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/diagnostics/archive"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/filesystem"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"
)

var (
	instance = config.GetInstance()
)

type ClusterValidatorer interface {
	ValidateConnectivity(ctx context.Context) error
	ValidateClusterIdentity(ctx context.Context) (bool, error)
}

type NamespaceFilterer interface {
	Filter(ctx context.Context) (*FilterResult, error)
}

type FilterResult struct {
	Valid        []string
	NonExisting  []string
	NonPermitted []string
}

type LogsCollectorer interface {
	Collect(ctx context.Context, namespaces []string)
}

type CollectLogsCommand struct {
	clusterValidator      ClusterValidatorer
	logsCollector         LogsCollectorer
	nsFilter              NamespaceFilterer
	namespacesFlagChanged bool
}

func NewCollectLogsCommand(
	clusterValidator ClusterValidatorer,
	logsCollector LogsCollectorer,
	nsFilter NamespaceFilterer,
	namespacesFlagChanged bool,
) *CollectLogsCommand {
	return &CollectLogsCommand{
		clusterValidator:      clusterValidator,
		logsCollector:         logsCollector,
		nsFilter:              nsFilter,
		namespacesFlagChanged: namespacesFlagChanged,
	}
}

func (clc *CollectLogsCommand) Run(ctx context.Context) error {
	if err := clc.validateCluster(ctx); err != nil {
		return err
	}

	namespacesFilterResult, err := clc.nsFilter.Filter(ctx)
	if err != nil {
		return err
	}

	if err = clc.validateFilteredNamespaces(namespacesFilterResult); err != nil {
		return err
	}

	instance.GetLogger().Info(fmt.Sprintf("about to collect logs from namespaces [%s]", strings.Join(namespacesFilterResult.Valid, ",")))
	clc.logsCollector.Collect(ctx, namespacesFilterResult.Valid)
	return nil
}

func (clc *CollectLogsCommand) validateCluster(ctx context.Context) error {
	valid, err := clc.clusterValidator.ValidateClusterIdentity(ctx)
	if err != nil {
		instance.GetLogger().Error(ClusterValidationFailed, zap.Error(err))
		return errors.New(ClusterValidationFailed)
	}

	if !valid {
		instance.GetLogger().Warn(ClusterNotValid)
	}

	if err = clc.clusterValidator.ValidateConnectivity(ctx); err != nil {
		instance.GetLogger().Error(K8SCommunicationFailed, zap.Error(err))
		return errors.New(K8SCommunicationFailed)
	}

	return nil
}

func (clc *CollectLogsCommand) validateFilteredNamespaces(namespacesFilterResult *FilterResult) error {
	if len(namespacesFilterResult.Valid) == 0 {
		return errors.New(NoValidNamespaces)
	}

	if len(namespacesFilterResult.NonPermitted) == 0 && len(namespacesFilterResult.NonExisting) == 0 {
		return nil
	}

	if !clc.namespacesFlagChanged {
		if len(namespacesFilterResult.NonPermitted) > 0 {
			instance.GetLogger().Warn(fmt.Sprintf("%s [%s], will be skipped", NamespacesNotPermitted, strings.Join(namespacesFilterResult.NonPermitted, ",")))
		}
		return nil
	}

	var err error
	if len(namespacesFilterResult.NonExisting) > 0 {
		err = errors.Join(err, fmt.Errorf("%s [%s]", NamespacesNotExist, strings.Join(namespacesFilterResult.NonExisting, ",")))
	}
	if len(namespacesFilterResult.NonPermitted) > 0 {
		err = errors.Join(err, fmt.Errorf("%s [%s]", NamespacesNotPermitted, strings.Join(namespacesFilterResult.NonPermitted, ",")))
	}
	instance.GetLogger().Error("", zap.Error(err))

	return err
}

func FinalizeArchive(tarHandler *archive.TarArchiveHandler, tarFilePath string) {
	if err := compressArchive(tarHandler); err != nil {
		instance.GetLogger().Error("failed to compress archive", zap.String("path", tarFilePath), zap.Error(err))
		return
	}

	instance.GetLogger().Info("archive is compressed", zap.String("path", tarFilePath+".gz"))
}

func CleanupArchive(tarFilePath string) {
	if err := os.Remove(tarFilePath); err != nil {
		instance.GetLogger().Error("failed to cleanup archive file", zap.String("path", tarFilePath), zap.Error(err))
	} else {
		instance.GetLogger().Info("cleaned up incomplete archive file", zap.String("path", tarFilePath))
	}
}

func AppendLoggerOutputToArchive(tarHandler *archive.TarArchiveHandler, loggerOutputFilePath string) {
	instance.GetLogger().Sync() // Flush logger buffers to ensure all logs are written to disk

	if err := tarHandler.AppendFile(loggerOutputFilePath, ArchiveLoggerOutputFileName); err != nil {
		instance.GetLogger().Warn("failed to write CLI log file to archive", zap.Error(err))
		return
	}
	if err := os.Remove(loggerOutputFilePath); err != nil {
		instance.GetLogger().Warn("failed to remove CLI log file", zap.Error(err))
	}
}

func ValidateOutputDir(outputDir string) error {
	return filesystem.IsDirWritable(outputDir)
}

func compressArchive(tarHandler *archive.TarArchiveHandler) error {
	if err := tarHandler.Close(); err != nil {
		return err
	}
	if err := tarHandler.Compress(); err != nil {
		return err
	}
	return nil
}
