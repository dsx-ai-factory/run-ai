package config

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"os"
	"path/filepath"

	"go.uber.org/zap"
)

var (
	expectedDefaultCliConfig = CliConfiguration{
		Authentication: Authentication{
			File: DefaultAuthenticationFile,
		},
		Api: Api{
			Retry:     DefaultApiRetry,
			TimeoutMs: DefaultApiTimeoutMs,
		},
		Cache: Cache{
			UpgradeVersion: DefaultUpdateCacheFile,
		},
		Global: Global{
			Update: Update{
				Auto:  DefaultAutoUpdate,
				Shell: DefaultUpdateShell,
			},
		},
	}

	expectedEmptyCliConfig = CliConfiguration{}
)

var _ = Describe("CLI Configuration Test", func() {
	const (
		TestFilePath    = "internal/tests/config_samples/"
		EmptyConfigFile = "empty_config.json"
		EmptyJsonFile   = "empty_json.json"
		EmptyYamlFile   = "empty_yaml.yml"
	)

	var (
		err              error
		cliConfiguration *CliConfiguration
		homePath         string
	)

	BeforeEach(func() {
		homePath, err = os.Getwd()
		if err != nil {
			Fail("failed to get pwd")
		}
		homePath = strings.SplitAfter(homePath, "runai-cli/")[0]
		homePath += TestFilePath
	})

	Context("Get Empty Configuration", func() {
		It("configuration should be nil", func() {
			Expect(cliConfiguration).To(BeNil())
		})
	})

	Context("Configuration validation", func() {
		It("Empty file", func() {
			err = GetInstance().Initializing(Options{Debug: true, ConfigPath: homePath, ConfigFile: EmptyConfigFile})
			cliConfiguration = GetInstance().GetConfiguration()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("While parsing config: unexpected end of JSON input"))
			Expect(cliConfiguration).To(Equal(&expectedEmptyCliConfig))
		})

		It("Empty JSON file - config with default values", func() {
			err = GetInstance().Initializing(Options{Debug: true, ConfigPath: homePath, ConfigFile: EmptyJsonFile})
			cliConfiguration = GetInstance().GetConfiguration()
			Expect(err).To(BeNil())
			Expect(cliConfiguration).ToNot(BeNil())
			Expect(cliConfiguration).To(Equal(&expectedDefaultCliConfig))
			Expect(cliConfiguration.Global.Update.Auto).To(Equal(DefaultAutoUpdate))
			Expect(cliConfiguration.Global.Update.Shell).To(Equal(DefaultUpdateShell))
		})

		It("configuration with path to a yml", func() {
			err = GetInstance().Initializing(Options{Debug: true, ConfigPath: homePath, ConfigFile: EmptyYamlFile})
			Expect(err).To(BeNil())
			cliConfiguration = GetInstance().GetConfiguration()
			Expect(cliConfiguration).To(Equal(&expectedDefaultCliConfig))
			Expect(cliConfiguration.Global.Update.Auto).To(Equal(DefaultAutoUpdate))
			Expect(cliConfiguration.Global.Update.Shell).To(Equal(DefaultUpdateShell))
		})

		It("should return default auto-update value when not configured", func() {
			instance := GetInstance()
			err = instance.Initializing(Options{Debug: true, ConfigPath: homePath, ConfigFile: EmptyJsonFile})
			Expect(err).To(BeNil())

			autoUpdateValue := instance.GetAutoUpdate()
			Expect(autoUpdateValue).To(Equal(DefaultAutoUpdate))

			config := instance.GetConfiguration()
			Expect(config.Global.Update.Auto).To(Equal(DefaultAutoUpdate))
			Expect(config.Global.Update.Shell).To(Equal(DefaultUpdateShell))
		})

	})
})

var _ = Describe("CreateConfigFile", func() {
	var (
		tempDir    string
		configPath string
		configFile string
		err        error
	)

	BeforeEach(func() {
		instance := GetInstance()
		instance.SetLogger(zap.NewNop())
		// Create a temporary directory for testing
		tempDir, err = os.MkdirTemp("", "config-test-")
		Expect(err).NotTo(HaveOccurred())

		configPath = filepath.Join(tempDir, "config")
		configFile = "config.yaml"
	})

	AfterEach(func() {
		// Clean up after each test
		err = os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the config path does not exist", func() {
		It("should create the path and config file", func() {
			err = CreateConfigFile(configPath, configFile)
			Expect(err).NotTo(HaveOccurred())

			// Verify path was created
			pathInfo, err := os.Stat(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(pathInfo.IsDir()).To(BeTrue())

			// Verify config file was created
			fileInfo, err := os.Stat(filepath.Join(configPath, configFile))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.IsDir()).To(BeFalse())
		})
	})

	Context("when the config path exists but file doesn't", func() {
		BeforeEach(func() {
			err = os.MkdirAll(configPath, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create only the config file", func() {
			err = CreateConfigFile(configPath, configFile)
			Expect(err).NotTo(HaveOccurred())

			// Verify config file was created
			fileInfo, err := os.Stat(filepath.Join(configPath, configFile))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.IsDir()).To(BeFalse())
		})
	})

	Context("when both path and file exist", func() {
		BeforeEach(func() {
			// Create path and file
			err = os.MkdirAll(configPath, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			filePath := filepath.Join(configPath, configFile)
			err = os.WriteFile(filePath, []byte("existing content"), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should not modify existing file", func() {
			existingContent, err := os.ReadFile(filepath.Join(configPath, configFile))
			Expect(err).NotTo(HaveOccurred())

			err = CreateConfigFile(configPath, configFile)
			Expect(err).NotTo(HaveOccurred())

			newContent, err := os.ReadFile(filepath.Join(configPath, configFile))
			Expect(err).NotTo(HaveOccurred())
			Expect(newContent).To(Equal(existingContent))
		})
	})

	Context("when path creation fails", func() {
		BeforeEach(func() {
			// Create a file where the path should be to cause failure
			err = os.WriteFile(configPath, []byte(""), 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return an error", func() {
			err = CreateConfigFile(configPath, configFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error creating configuration file path"))
		})

		It("should return an error", func() {
			err = CreateConfigFile("/no/path", configFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error creating configuration file path"))
		})
	})
})
