package upgrade

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	pathUtil "path"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	systems "github.com/run-ai/runai/control-plane/runai-cli/internal/os"
)

type InstallerGetter interface {
	GetInstallerUnix(ctx context.Context) (string, error)
}

type ScriptDownloader struct {
	apiClient InstallerGetter
}

func NewScriptDownloader(apiClient InstallerGetter) *ScriptDownloader {
	return &ScriptDownloader{apiClient: apiClient}
}

func (bd *ScriptDownloader) Download(ctx context.Context, path, fileName string) (string, error) {
	scriptData, err := bd.apiClient.GetInstallerUnix(ctx)
	if err != nil {
		return "", fmt.Errorf(consts.FailedToDownloadInstaller, err)
	}

	fullPath := pathUtil.Join(path, fileName)
	// write the script to the file
	fileObj, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf(consts.FailedToDownloadInstaller, err)
	}
	defer fileObj.Close()
	_, err = fileObj.Write([]byte(scriptData))
	if err != nil {
		return "", fmt.Errorf(consts.FailedToDownloadInstaller, err)
	}
	return fullPath, nil
}

func writeToFile(filePath string, body io.ReadCloser) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, body)
	if err != nil {
		return err
	}
	return nil
}

type BinaryGetter interface {
	GetBinary(ctx context.Context, operationSystem, arc string) (*os.File, error)
}

type BinaryDownloader struct {
	apiClient    BinaryGetter
	osType       systems.Type
	architecture systems.Arch
}

func NewBinaryDownloader(apiClient BinaryGetter, osType systems.Type, architecture systems.Arch) *BinaryDownloader {
	return &BinaryDownloader{apiClient: apiClient, osType: osType, architecture: architecture}
}

func (bd *BinaryDownloader) Download(ctx context.Context, path, fileName string) (string, error) {
	binaryData, err := bd.apiClient.GetBinary(ctx, bd.osType.String(), bd.architecture.String())

	if err != nil {
		return "", fmt.Errorf(consts.FailedToDownloadInstaller, err)
	}
	if binaryData == nil {
		return "", errors.New(consts.FailedDownloadBinary)
	}
	defer binaryData.Close()

	fileFullPath := pathUtil.Join(path, fileName)
	// write the binary to the file
	err = writeToFile(fileFullPath, binaryData)
	if err != nil {
		return "", fmt.Errorf(consts.FailedToDownloadInstaller, err)
	}
	return fileFullPath, nil
}
