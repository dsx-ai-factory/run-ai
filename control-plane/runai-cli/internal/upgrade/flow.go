package upgrade

import (
	"context"
	"fmt"
	"os"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	oscli "github.com/run-ai/runai/control-plane/runai-cli/internal/os"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/progressbar/spinner"
)

const (
	upgradingSpinnerMsg = "Upgrading RunAI CLI ..."
	successMsg          = "Upgrade completed successfully"
)

var instance = config.GetInstance()

type Downloader interface {
	Download(ctx context.Context, path, fileName string) (string, error)
}

type Installer interface {
	Install(ctx context.Context, path, execShell string) (string, error)
}

type BasicFlow struct {
	downloader Downloader
	installer  Installer
	fileName   string
	cacheFile  string
}

func NewBasicFlow(downloader Downloader, installer Installer, cacheFile, fileName string) *BasicFlow {
	return &BasicFlow{
		downloader: downloader,
		installer:  installer,
		fileName:   fileName,
		cacheFile:  cacheFile,
	}
}

func (bf *BasicFlow) Upgrade(ctx context.Context, execShell string) (string, error) {
	logger := config.GetInstance().GetLogger()
	tempDir, err := os.MkdirTemp("", "cli")
	if err != nil {
		return "", fmt.Errorf(consts.FailedToCreateTmpDir, err)
	}
	defer func(path string) {
		err = os.RemoveAll(path)
		if err != nil {
			logger.Debug(fmt.Sprintf(consts.FailedToRemoveTmpDir, path))
		}
	}(tempDir)

	spinnerUx := spinner.NewSpinner(upgradingSpinnerMsg)
	go spinnerUx.Run()

	filePath, err := bf.downloader.Download(ctx, tempDir, bf.fileName)
	if err != nil {
		spinnerUx.Stop()
		return "", err
	}
	output, err := bf.installer.Install(ctx, filePath, execShell)
	spinnerUx.Stop()
	if err != nil {
		return "", err
	}
	logger.Debug(output)

	err = bf.postUpgrade()
	if err != nil {
		return "", fmt.Errorf(consts.FailedToClearCacheFile, err)
	}

	return successMsg, nil
}

func (bf *BasicFlow) postUpgrade() error {
	logger := instance.GetLogger()
	logger.Debug("clearing cache")

	return oscli.RemoveFile(bf.cacheFile)
}
