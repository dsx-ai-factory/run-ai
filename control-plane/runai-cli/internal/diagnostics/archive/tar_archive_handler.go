package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"sync"

	"errors"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"
)

type TarArchiveHandler struct {
	file        *os.File
	tarWriter   *tar.Writer
	appendMu    *sync.Mutex
	tarFilePath string
	closed      bool
}

func NewTarArchiveHandler(tarFilePath string) (*TarArchiveHandler, error) {
	file, err := os.Create(tarFilePath)
	if err != nil {
		return nil, err
	}

	tarWriter := tar.NewWriter(file)

	return &TarArchiveHandler{
		file:        file,
		tarWriter:   tarWriter,
		appendMu:    &sync.Mutex{},
		tarFilePath: tarFilePath,
	}, nil
}

func (t *TarArchiveHandler) AppendFile(srcPath, archivePath string) error {
	if t.closed {
		return fmt.Errorf("tar archive is closed")
	}

	t.appendMu.Lock()
	defer t.appendMu.Unlock()

	info, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", srcPath, err)
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("failed to create tar header: %w", err)
	}

	header.Name = archivePath

	// Tar entry = write header and then file data (must be in this order)
	if err := t.tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open file for tar: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(t.tarWriter, file); err != nil {
		return fmt.Errorf("failed to copy file into tar: %w", err)
	}

	err = t.flush()
	if err != nil {
		config.GetInstance().GetLogger().Debug("failed to flush tar file", zap.Error(err))
	}

	return nil
}

func (t *TarArchiveHandler) Close() error {
	if t.closed {
		return nil
	}
	t.closed = true

	var errs []error

	if err := t.tarWriter.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close tar writer: %w", err))
	}

	if err := t.file.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close file: %w", err))
	}

	return errors.Join(errs...)
}

func (t *TarArchiveHandler) Compress() error {
	if !t.closed {
		return fmt.Errorf("tar archive need to be closed before compressing")
	}

	tarFile, err := os.Open(t.tarFilePath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	gzFile, err := os.Create(t.tarFilePath + ".gz")
	if err != nil {
		return err
	}
	defer gzFile.Close()

	gzipWriter := gzip.NewWriter(gzFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, tarFile)
	if err != nil {
		return err
	}

	err = os.Remove(t.tarFilePath)
	if err != nil {
		config.GetInstance().GetLogger().Debug("failed to remove tar file", zap.Error(err))
	}

	return nil
}

func (t *TarArchiveHandler) flush() error {
	if t.closed {
		return fmt.Errorf("tar archive is closed")
	}

	if err := t.tarWriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush tar writer: %w", err)
	}
	if err := t.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}
	return nil
}
