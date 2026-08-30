package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	runaios "github.com/run-ai/runai/control-plane/runai-cli/internal/os"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction/config"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/credentials"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	"github.com/spf13/viper"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

type InteractiveMode int

const (
	InteractiveUnset InteractiveMode = iota
	InteractiveEnabled
	InteractiveDisabled
)

var privilegedKeyName = []string{"token"}

const cacheDirectoryName = "cache"

type Instance struct {
	options Options
	token   string
	txn     *transaction.Txn
}

type Options struct {
	ConfigPath  string
	ConfigFile  string
	Interactive InteractiveMode
	Debug       bool
	Verbose     bool
	Quiet       bool
}

var (
	singleInstance *Instance
	lock           = &sync.Mutex{}
	logger         *zap.Logger
)

func GetInstance() *Instance {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleInstance == nil {
			zap.L().Debug("creating single instance now.")
			singleInstance = &Instance{}
		}
	}
	return singleInstance
}

func (li *Instance) Initializing(options Options) error {
	lock.Lock()
	defer lock.Unlock()
	err := li.InitLogger(options.Debug)
	if err != nil {
		return err
	}

	// check the environment variable if the flag does not set
	if options.ConfigFile == "" || options.ConfigFile == consts.DefaultConfigFile {
		options.ConfigFile = ReadEnvironmentVariable(consts.EnvRunaiCliConfigFile, consts.DefaultConfigFile)
	}

	if options.ConfigPath == "" {
		options.ConfigPath = ReadEnvironmentVariable(consts.EnvRunaiCliConfigPath, runaios.GetDefaultConfigPath())

		// can be empty if GetDefaultConfigPath() failed
		if options.ConfigPath == "" {
			return fmt.Errorf("unable to get default path, please set a config path --%s=<path>", consts.ConfigPath)
		}
	}

	li.txn = transaction.NewTxn()
	return li.initConfiguration(options)
}

func (li *Instance) InitLogger(debug bool) error {
	var (
		err       error
		zapConfig zap.Config
	)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	if debug {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.DisableStacktrace = true
	} else {
		zapConfig = zap.NewProductionConfig()
		zapConfig.Level.SetLevel(zap.FatalLevel)
		zapConfig.DisableStacktrace = true
	}
	zapConfig.Encoding = "console"
	zapConfig.EncoderConfig = encoderCfg

	logger, err = zapConfig.Build()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}

func (li *Instance) ConfigureFileLogger(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewConsoleEncoder(encoderCfg)
	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(file), zap.DebugLevel)

	currentLogger := zap.L()
	logger = zap.New(zapcore.NewTee(currentLogger.Core(), fileCore))

	zap.ReplaceGlobals(logger)
	return nil
}

func (li *Instance) GetLogger() *zap.Logger {
	return logger
}

func (li *Instance) SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func (li *Instance) initConfiguration(options Options) error {
	logger.Debug("using config", zap.String("path", options.ConfigPath), zap.String("file", options.ConfigFile))

	configFilePath := filepath.Join(options.ConfigPath, options.ConfigFile)
	err := CreateConfigFile(options.ConfigPath, options.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to check config path %s: %w", options.ConfigPath, err)
	}
	options.ConfigPath = filepath.Clean(options.ConfigPath) + string(filepath.Separator)
	err = viperInit(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file not found. %w \ncan be created with `runai config generate --file %s`",
				err, filepath.Join(options.ConfigPath, options.ConfigFile))
		}
		return err
	}

	if options.Interactive == InteractiveUnset {
		var isInteractive bool
		if viper.IsSet(SettingsInteractive) {
			isInteractive = viper.GetBool(SettingsInteractive)
		} else {
			isInteractive, _ = strconv.ParseBool(consts.DefaultInteractive)
		}
		if isInteractive {
			options.Interactive = InteractiveEnabled
		} else {
			options.Interactive = InteractiveDisabled
		}
	}

	li.options = options
	return nil
}

func (li *Instance) Reload() {
	viper.Reset()
	_ = viper.ReadInConfig()
}

func (li *Instance) GetConfiguration() *CliConfiguration {
	configMap := viper.AllSettings()
	encodedConfig, _ := json.MarshalIndent(configMap, "", " ")
	localConfig := CliConfiguration{}
	err := json.Unmarshal(encodedConfig, &localConfig)
	if err != nil {
		return nil
	}
	return &localConfig
}

func (li *Instance) GetConfigValue(key string) string {
	return viper.GetString(key)
}

func (li *Instance) GetConfigValueWithDefault(key string, defaultValue string) string {
	if viper.IsSet(key) {
		return viper.GetString(key)
	}
	return defaultValue
}
func (li *Instance) GetAuthenticationPath() string {
	configPath := li.options.ConfigPath
	fileName := viper.GetString(AuthenticationFile)
	return filepath.Join(configPath, fileName)
}

func (li *Instance) GetCacheDirPath() string {
	baseDirPath := li.options.ConfigPath
	return filepath.Join(baseDirPath, cacheDirectoryName)

}

func (li *Instance) GetVersionCachePath() string {
	baseCacheDirPath := li.GetCacheDirPath()
	version := viper.GetString(CacheUpgradeVersionFile)
	return filepath.Join(baseCacheDirPath, version)
}

func (li *Instance) GetOptions() Options { return li.options }
func (li *Instance) SetCluster(uuid, name, domain string) error {
	err := li.unset(ClusterSection)
	if err != nil {
		return err
	}
	viper.Set(ClusterUUID, uuid)
	viper.Set(ClusterName, name)
	viper.Set(ClusterDomain, domain)
	return WriteConfiguration()
}
func (li *Instance) ClearCluster() error {
	err := li.unset(ClusterSection)
	if err != nil {
		return err
	}
	return WriteConfiguration()
}
func (li *Instance) GetToken() string {
	return li.token
}

func (li *Instance) SetToken(token string) {
	li.token = token
}

func (li *Instance) SetCredentials(authType credentials.Type) error {
	err := li.unset(CredentialsStr)
	if err != nil {
		return err
	}
	stringList, err := stringMapper(authType)
	if err != nil {
		return err
	}

	viper.Set(CredentialsType, authType.GetType())
	for key, value := range stringList {
		viper.Set(fmt.Sprintf("%s.%s", CredentialsStr, key), value)
	}
	return WriteConfiguration()
}

func stringMapper(authType credentials.Type) (map[string]string, error) {
	jsonStr, err := json.Marshal(authType)
	if err != nil {
		return nil, err
	}
	var mapper map[string]string
	err = json.Unmarshal(jsonStr, &mapper)
	return mapper, err
}

func (li *Instance) ClearCredentials() error {
	err := li.unset(CredentialsStr)
	if err != nil {
		return err
	}
	return WriteConfiguration()
}

func (li *Instance) LoadCredentials() (credentials.Type, error) {
	credType := viper.GetString(CredentialsType)
	credential, err := credentials.GetCredentialTypeFactory(credentials.LoginType(credType))
	if err != nil {
		return nil, err
	}
	viperCred := viper.Get(CredentialsStr)
	jsonStr, err := json.Marshal(viperCred)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonStr, &credential)
	return credential, err
}

func (li *Instance) SetInteractive(enabled bool) error {
	viper.Set(SettingsInteractive, enabled)
	return WriteConfiguration()
}

func (li *Instance) SetCPURL(url string) error {
	viper.Set(ControlPlaneUrl, url)
	return WriteConfiguration()
}
func (li *Instance) SetAuthURL(url string) error {
	viper.Set(AuthenticationUrl, url)
	return WriteConfiguration()
}

func (li *Instance) SetStatusDuration(dur string) error {
	viper.Set(ApiStatusTimeoutDuration, dur)
	return WriteConfiguration()
}

func (li *Instance) SetProject(project string, id int32) error {
	viper.Set(ClusterProjectName, project)
	viper.Set(ClusterProjectID, id)
	return WriteConfiguration()
}

func (li *Instance) SetDepartment(department string, id int32) error {
	viper.Set(ClusterDepartmentName, department)
	viper.Set(ClusterDepartmentID, id)
	return WriteConfiguration()
}

func (li *Instance) GetTxnConfig() (*config.Txn, error) {
	ts := viper.Get(TransactionSettingsMetrics)
	marshalTs, err := json.Marshal(ts)
	if err != nil {
		return nil, err
	}
	txnConfig := config.Txn{}
	err = json.Unmarshal(marshalTs, &txnConfig)
	if err != nil {
		return nil, err
	}
	return &txnConfig, nil
}

func (li *Instance) SetTxnConfig(txnConfi *config.Txn) error {
	viper.Set(TransactionSettingsMetrics, txnConfi)
	return viper.WriteConfig()
}

func (li *Instance) GetTxn() *transaction.Txn {
	return li.txn
}

func (li *Instance) unset(key string) error {
	configMap := viper.AllSettings()
	for _, cleared := range privilegedKeyName {
		delete(configMap, cleared)
	}
	delete(configMap, key)
	encodedConfig, _ := json.MarshalIndent(configMap, "", " ")
	err := viper.ReadConfig(bytes.NewReader(encodedConfig))
	if err != nil {
		return err
	}
	err = viper.WriteConfig()
	return err
}

func (li *Instance) SetAutoUpdate(enabled bool) error {
	viper.Set(GlobalUpdateAuto, enabled)
	return WriteConfiguration()
}

func (li *Instance) GetAutoUpdate() bool {
	if !viper.IsSet(GlobalUpdateAuto) {
		return DefaultAutoUpdate
	}
	return viper.GetBool(GlobalUpdateAuto)
}

func (li *Instance) IsAutoUpdateInit() bool {
	return viper.IsSet(GlobalUpdateAuto)
}

func (li *Instance) SetUpdateShell(shell string) error {
	viper.Set(GlobalUpdateShell, shell)
	return WriteConfiguration()
}

func (li *Instance) GetUpdateShell() string {
	if !viper.IsSet(GlobalUpdateShell) {
		return DefaultUpdateShell
	}
	return viper.GetString(GlobalUpdateShell)
}
