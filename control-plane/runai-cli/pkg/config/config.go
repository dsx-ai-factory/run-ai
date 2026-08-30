package config

import (
	"encoding/json"
	"os"

	runaios "github.com/run-ai/runai/control-plane/runai-cli/internal/os"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config/attributes"

	"github.com/spf13/viper"
)

// Config path consts
const (
	SettingsStr         = "settings"
	SettingsInteractive = "settings.interactive"
	SettingsOutputType  = "settings.output_type"
	GlobalUpdateAuto    = "global.update.auto"
	GlobalUpdateShell   = "global.update.shell"

	AuthenticationFile = "authentication.file"
	AuthenticationUrl  = "authentication.url"

	ApiRetry                 = "api.retry"
	ApiTimeoutMs             = "api.timeout_ms"
	ApiStatusTimeoutDuration = "api.cluster_status_timeout_duration"
	ControlPlaneStr          = "control_plane"
	ControlPlaneUrl          = "control_plane.url"

	CacheUpgradeVersionFile = "cache.upgrade_version"
	ClusterSection          = "cluster"
	ClusterUUID             = "cluster.uuid"
	ClusterName             = "cluster.name"
	ClusterDomain           = "cluster.domain"
	ClusterProjectName      = "cluster.project.name"
	ClusterProjectID        = "cluster.project.id"

	ClusterDepartmentName = "cluster.department.name"
	ClusterDepartmentID   = "cluster.department.id"

	CredentialsStr  = "credentials"
	CredentialsType = "credentials.type"

	TransactionSettingsMetrics = "transaction.settings.metrics"
)

// Config default values

const (
	DefaultAutoUpdate         = false
	DefaultAuthenticationFile = "authentication.json"
	DefaultApiRetry           = 3
	DefaultApiTimeoutMs       = 100
	DefaultUpdateCacheFile    = "version.json"
	DefaultUpdateShell        = "bash"
)

type Update struct {
	Auto  bool   `json:"auto" yaml:"auto"`
	Shell string `json:"shell" yaml:"shell"`
}

type Global struct {
	Update Update `json:"update" yaml:"update"`
}

type Authentication struct {
	File string `json:"file" yaml:"file"`
	Url  string `json:"url" yaml:"url" env:"APPLICATION_AUTH_URL"`
}

type Api struct {
	Retry                        int    `json:"retry"`
	TimeoutMs                    int    `json:"timeout_ms" yaml:"timeout_ms"`
	ClusterStatusTimeoutDuration string `json:"cluster_status_timeout_duration"`
}

type Cache struct {
	UpgradeVersion string `json:"upgrade_version" yaml:"upgrade_version"`
}

type Cluster struct {
	Domain string `json:"domain" yaml:"domain"`
	Name   string `json:"name" yaml:"name"`
	Uuid   string `json:"uuid" yaml:"uuid"`
}

type Credentials struct {
	Type   string `json:"type" yaml:"type"`
	Name   string `json:"client_id" yaml:"client_id"`
	Secret string `json:"client_secret" yaml:"client_secret"`
}

func (c *Credentials) UnmarshalJSON(data []byte) error {
	type Alias struct {
		Type         string `json:"type"`
		ClientID     string `json:"client_id"`
		AppName      string `json:"app_name"`
		ClientSecret string `json:"client_secret"`
		AppSecret    string `json:"app_secret"`
	}

	var temp Alias
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	c.Type = temp.Type
	if temp.ClientID != "" {
		c.Name = temp.ClientID
	} else {
		c.Name = temp.AppName
	}
	if temp.ClientSecret != "" {
		c.Secret = temp.ClientSecret
	} else {
		c.Secret = temp.AppSecret
	}

	return nil
}

type CliConfiguration struct {
	Global         Global         `json:"global" yaml:"global"`
	Authentication Authentication `json:"authentication" yaml:"authentication"`
	Api            Api            `json:"api" yaml:"api"`
	Cluster        Cluster        `json:"cluster" yaml:"cluster"`
	Cache          Cache          `json:"cache" yaml:"cache"`
	Credentials    Credentials    `json:"credentials" yaml:"credentials"`
}

var configStringMap = map[string]any{
	GlobalUpdateShell:       DefaultUpdateShell,
	AuthenticationFile:      DefaultAuthenticationFile,
	ApiRetry:                DefaultApiRetry,
	ApiTimeoutMs:            DefaultApiTimeoutMs,
	CacheUpgradeVersionFile: DefaultUpdateCacheFile,
}

func viperInit(configFilePath string) error {
	viper.SetConfigFile(configFilePath)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	defaultConfigSet()
	return err
}

func defaultConfigSet() {
	for key, value := range configStringMap {
		viper.SetDefault(key, value)
	}
}

func WriteConfiguration() error {
	return viper.WriteConfig()
}

func ReadEnvironmentVariable[V string | attributes.CliLogLevel](key string, defaultValue V) V {
	tempEnv, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return V(tempEnv)
}

// CreateConfigFile creates the config file if it does not exist
func CreateConfigFile(configPath string, configFile string) error {
	return runaios.CreateConfigFileWithViper(configPath, configFile, viper.GetViper())
}
