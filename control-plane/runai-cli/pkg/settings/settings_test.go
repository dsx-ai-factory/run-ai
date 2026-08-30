package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/viper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Settings Tests")
}

// resetInstances resets singleton instances to avoid state pollution between tests
func resetInstances() {
	// Reset settings instance
	lock.Lock()
	singleInstance = nil
	settingsViper = viper.New()
	lock.Unlock()
}

var _ = Describe("Settings Configuration Test", func() {
	const (
		TestFilePath  = "../../internal/tests/config_samples/"
		EmptyJsonFile = "empty_json.json"
	)

	var (
		err      error
		homePath string
	)

	BeforeEach(func() {
		homePath, err = os.Getwd()
		if err != nil {
			Fail("failed to get pwd")
		}
		homePath = strings.SplitAfter(homePath, "runai-cli/")[0]
		homePath += TestFilePath

		resetInstances()
	})

	AfterEach(func() {
		resetInstances()
	})

	Context("Output Configuration validation", func() {
		It("should return default list format align value when not configured", func() {
			configInstance := config.GetInstance()
			err = configInstance.Initializing(config.Options{ConfigPath: homePath, ConfigFile: EmptyJsonFile})
			Expect(err).NotTo(HaveOccurred())

			instance := GetInstance()
			err = instance.Initializing(homePath)
			Expect(err).NotTo(HaveOccurred())

			configuration := instance.GetConfiguration()
			Expect(configuration.Output.Lists.FormatAlign).To(Equal(DefaultListFormatAlign))
			Expect(configuration.Output.Lists.FormatAlign).To(BeFalse())
		})

		It("should handle upgrade scenario - config without lists section", func() {
			tempDir, err := os.MkdirTemp("", "settings-test-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			srcPath := filepath.Join(homePath, EmptyJsonFile)
			destPath := filepath.Join(tempDir, "test_settings.json")

			srcData, err := os.ReadFile(srcPath)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(destPath, srcData, 0644)
			Expect(err).NotTo(HaveOccurred())

			configInstance := config.GetInstance()
			err = configInstance.Initializing(config.Options{ConfigPath: tempDir, ConfigFile: "test_settings.json"})
			Expect(err).NotTo(HaveOccurred())

			instance := GetInstance()
			err = instance.Initializing(tempDir)
			Expect(err).NotTo(HaveOccurred())

			// Test default value
			listFormatAlign := instance.GetListFormatAlign()
			Expect(listFormatAlign).To(Equal(DefaultListFormatAlign))
			Expect(listFormatAlign).To(BeFalse())

			// Test setting value
			err = instance.SetListFormatAlign(true)
			Expect(err).NotTo(HaveOccurred())

			listFormatAlign = instance.GetListFormatAlign()
			Expect(listFormatAlign).To(BeTrue())
		})

		It("should return default empty slice when workload columns not configured", func() {
			configInstance := config.GetInstance()
			err = configInstance.Initializing(config.Options{ConfigPath: homePath, ConfigFile: EmptyJsonFile})
			Expect(err).NotTo(HaveOccurred())

			instance := GetInstance()
			err = instance.Initializing(homePath)
			Expect(err).NotTo(HaveOccurred())

			columns := instance.GetWorkloadColumns()
			Expect(columns).To(Equal([]string{}))
			Expect(len(columns)).To(Equal(0))
		})

		It("should set and get workload columns", func() {
			tempDir, err := os.MkdirTemp("", "settings-test-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			srcPath := filepath.Join(homePath, EmptyJsonFile)
			destPath := filepath.Join(tempDir, "test_settings.json")

			srcData, err := os.ReadFile(srcPath)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(destPath, srcData, 0644)
			Expect(err).NotTo(HaveOccurred())

			configInstance := config.GetInstance()
			err = configInstance.Initializing(config.Options{ConfigPath: tempDir, ConfigFile: "test_settings.json"})
			Expect(err).NotTo(HaveOccurred())

			instance := GetInstance()
			err = instance.Initializing(tempDir)
			Expect(err).NotTo(HaveOccurred())

			// Test default empty value
			columns := instance.GetWorkloadColumns()
			Expect(columns).To(Equal([]string{}))

			// Test setting columns
			testColumns := []string{"name", "status", "age", "node"}
			err = instance.SetWorkloadColumns(testColumns)
			Expect(err).NotTo(HaveOccurred())

			// Test getting columns
			retrievedColumns := instance.GetWorkloadColumns()
			Expect(retrievedColumns).To(Equal(testColumns))
		})

		It("should handle empty workload columns list", func() {
			tempDir, err := os.MkdirTemp("", "settings-test-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			srcPath := filepath.Join(homePath, EmptyJsonFile)
			destPath := filepath.Join(tempDir, "test_settings.json")

			srcData, err := os.ReadFile(srcPath)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(destPath, srcData, 0644)
			Expect(err).NotTo(HaveOccurred())

			configInstance := config.GetInstance()
			err = configInstance.Initializing(config.Options{ConfigPath: tempDir, ConfigFile: "test_settings.json"})
			Expect(err).NotTo(HaveOccurred())

			instance := GetInstance()
			err = instance.Initializing(tempDir)
			Expect(err).NotTo(HaveOccurred())

			// Set empty columns list
			emptyColumns := []string{}
			err = instance.SetWorkloadColumns(emptyColumns)
			Expect(err).NotTo(HaveOccurred())

			// Verify empty list is returned
			retrievedColumns := instance.GetWorkloadColumns()
			Expect(retrievedColumns).To(Equal([]string{}))
			Expect(len(retrievedColumns)).To(Equal(0))
		})

		It("should update workload columns when changed", func() {
			tempDir, err := os.MkdirTemp("", "settings-test-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tempDir)

			srcPath := filepath.Join(homePath, EmptyJsonFile)
			destPath := filepath.Join(tempDir, "test_settings.json")

			srcData, err := os.ReadFile(srcPath)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(destPath, srcData, 0644)
			Expect(err).NotTo(HaveOccurred())

			configInstance := config.GetInstance()
			err = configInstance.Initializing(config.Options{ConfigPath: tempDir, ConfigFile: "test_settings.json"})
			Expect(err).NotTo(HaveOccurred())

			instance := GetInstance()
			err = instance.Initializing(tempDir)
			Expect(err).NotTo(HaveOccurred())

			// Set initial columns
			initialColumns := []string{"name", "status"}
			err = instance.SetWorkloadColumns(initialColumns)
			Expect(err).NotTo(HaveOccurred())

			retrievedColumns := instance.GetWorkloadColumns()
			Expect(retrievedColumns).To(Equal(initialColumns))

			// Update with different columns
			updatedColumns := []string{"name", "status", "age", "node", "gpu"}
			err = instance.SetWorkloadColumns(updatedColumns)
			Expect(err).NotTo(HaveOccurred())

			retrievedColumns = instance.GetWorkloadColumns()
			Expect(retrievedColumns).To(Equal(updatedColumns))
		})
	})
})

var _ = Describe("CreateSettingsConfigFile", func() {
	var (
		tempDir    string
		configPath string
		configFile string
		err        error
	)

	BeforeEach(func() {
		tempDir, err = os.MkdirTemp("", "settings-config-test-")
		Expect(err).NotTo(HaveOccurred())

		configPath = filepath.Join(tempDir, "config")
		configFile = "config.json"

		resetInstances()
	})

	AfterEach(func() {
		err = os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
		resetInstances()
	})

	Context("when the settings config path does not exist", func() {
		It("should create the path and settings config file", func() {
			configInstance := config.GetInstance()
			err = configInstance.Initializing(config.Options{ConfigPath: configPath, ConfigFile: configFile})
			Expect(err).NotTo(HaveOccurred())

			err = CreateSettingsConfigFile(configPath, "settings.json")
			Expect(err).NotTo(HaveOccurred())

			// Verify path was created
			pathInfo, err := os.Stat(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(pathInfo.IsDir()).To(BeTrue())

			// Verify settings config file was created
			settingsPath := filepath.Join(configPath, "settings.json")
			fileInfo, err := os.Stat(settingsPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.IsDir()).To(BeFalse())
		})
	})

	Context("when the settings config path exists but file doesn't", func() {
		BeforeEach(func() {
			err = os.MkdirAll(configPath, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create only the settings config file", func() {
			configInstance := config.GetInstance()
			err = configInstance.Initializing(config.Options{ConfigPath: configPath, ConfigFile: configFile})
			Expect(err).NotTo(HaveOccurred())

			err = CreateSettingsConfigFile(configPath, "settings.json")
			Expect(err).NotTo(HaveOccurred())

			settingsPath := filepath.Join(configPath, "settings.json")
			fileInfo, err := os.Stat(settingsPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.IsDir()).To(BeFalse())
		})
	})

	Context("when both path and settings file exist", func() {
		BeforeEach(func() {
			err = os.MkdirAll(configPath, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			settingsPath := filepath.Join(configPath, "settings.json")
			err = os.WriteFile(settingsPath, []byte("existing settings content"), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should not modify existing settings file", func() {
			configInstance := config.GetInstance()
			err = configInstance.Initializing(config.Options{ConfigPath: configPath, ConfigFile: configFile})
			Expect(err).NotTo(HaveOccurred())

			settingsPath := filepath.Join(configPath, "settings.json")
			existingContent, err := os.ReadFile(settingsPath)
			Expect(err).NotTo(HaveOccurred())

			err = CreateSettingsConfigFile(configPath, "settings.json")
			Expect(err).NotTo(HaveOccurred())

			newContent, err := os.ReadFile(settingsPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(newContent).To(Equal(existingContent))
		})
	})
})
