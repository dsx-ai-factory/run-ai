package os

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func RemoveFile(file string) error {
	err := os.Remove(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// IsFileExists returns whether the given path exist
func IsFileExists(path string, isDirectory bool) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err == nil {
		if fileInfo.IsDir() == isDirectory {
			return true, nil
		}
		return false, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, consts.DefaultRunaiDir)
}
func createConfigDirectory(configPath string) error {
	// Check directory exists
	ok, err := IsFileExists(configPath, true)
	if err != nil {
		return fmt.Errorf("error checking configuration file path: %w", err)
	}

	// Create directory if missing
	if !ok {
		zap.L().Debug("configuration file path does not exist, creating it",
			zap.String("path", configPath))
		err = os.MkdirAll(configPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating configuration file path: %w", err)
		}
	}
	return nil
}
func createConfigFile(configPath, configFile string, viperInstance *viper.Viper) error {

	// Check file exists
	configFilePath := filepath.Join(configPath, configFile)
	ok, err := IsFileExists(configFilePath, false)
	if err != nil {
		return fmt.Errorf("error checking %s configuration file: %w", configFile, err)
	}

	// Create file if missing
	if !ok {
		zap.L().Debug(fmt.Sprintf("%s configuration file does not exist, creating it", configFile),
			zap.String("file", configFile))
		err = WriteViperConfigToFile(viperInstance, configFilePath)
		if err != nil {
			return fmt.Errorf("%s configuration file generation failed. %w", configFile, err)
		}
	}
	return nil
}

// CreateConfigFileWithViper creates a config file using the provided viper instance
func CreateConfigFileWithViper(configPath, configFile string, viperInstance *viper.Viper) error {
	err := createConfigDirectory(configPath)
	if err != nil {
		return err
	}
	err = createConfigFile(configPath, configFile, viperInstance)
	if err != nil {
		return err
	}
	return nil
}

// WriteViperConfigToFile writes a viper configuration to file
func WriteViperConfigToFile(viperInstance *viper.Viper, filePath string) error {
	err := viperInstance.SafeWriteConfigAs(filePath)
	if err != nil {
		var unsupportedConfigErr viper.UnsupportedConfigError
		if errors.As(err, &unsupportedConfigErr) {
			return fmt.Errorf("%w. supported extensions: %v", err, viper.SupportedExts)
		}
		return err
	}
	return nil
}
