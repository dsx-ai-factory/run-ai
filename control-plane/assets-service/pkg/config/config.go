package config

import (
	"os"
	"sync"
	"time"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

var once sync.Once

type ServerConfig struct {
	ServicePort           int `mapstructure:"service_port"`
	RetryMaxIntervalSec   int `mapstructure:"retry_max_interval_sec"`
	RetryMaxElapsedSec    int `mapstructure:"retry_max_elapsed_sec"`
	DaysLimitForFirstSync int `mapstructure:"days_limit_for_first_sync"`

	CleanupRefreshRateHours int `mapstructure:"cleanup_refresh_rate_hours"`
	CleanupKeepAssetsDays   int `mapstructure:"cleanup_keep_assets_days"`
}

func (sc *ServerConfig) SetDefaults() {
	sc.ServicePort = 8080
	sc.RetryMaxIntervalSec = 10
	sc.RetryMaxElapsedSec = 60
	sc.DaysLimitForFirstSync = 7
	sc.CleanupKeepAssetsDays = 30
	sc.CleanupRefreshRateHours = 24
}

type ServicesConfig struct {
	BackendSchema         string `mapstructure:"db_backend_schema,omitempty"` // schema of "old" backend
	PolicyUrl             string `mapstructure:"policy_service_url"`
	AuthorizationsUrl     string `mapstructure:"auth_service_url"`
	K8sObjectsTrackerUrl  string `mapstructure:"k8s_objects_tracker_url"`
	WorkloadsUrl          string `mapstructure:"workloads_service_url"`
	ClusterServiceUrl     string `mapstructure:"cluster_service_url"`
	DataVolumesServiceUrl string `mapstructure:"data_volumes_service_url"`
	SyncBulkSize          int    `mapstructure:"sync_bulk_size,omitempty"`
	OrgUnitUrl            string `mapstructure:"org_unit_service_url"`
}

func (sc *ServicesConfig) SetDefaults() {
	sc.BackendSchema = "backend"
	sc.SyncBulkSize = 50
}

type SecretConfig struct {
	DataKeyUser            string `mapstructure:"data_key_user"`
	DataKeyPassword        string `mapstructure:"data_key_password"`
	DataKeyAccessKeyId     string `mapstructure:"data_key_access_key_id"`
	DataKeySecretAccessKey string `mapstructure:"data_key_secret_access_key"`
}

func (sc *SecretConfig) SetDefaults() {
	sc.DataKeyUser = "user"
	sc.DataKeyPassword = "password"
	sc.DataKeyAccessKeyId = "accessKeyId"
	sc.DataKeySecretAccessKey = "secretAccessKey"
}

type CacheConfig struct {
	ExpirationSec int `mapstructure:"expiration_sec"`
	CleanupSec    int `mapstructure:"cleanup_sec"`
}

func (cc *CacheConfig) SetDefaults() {
	cc.ExpirationSec = 600
	cc.CleanupSec = 1200
}

// Expiration returns expiration configuration
func (c *CacheConfig) Expiration() time.Duration {
	return time.Duration(c.ExpirationSec) * time.Second
}

type TenantKeysCache struct {
	CacheConfig `mapstructure:",squash"`
}

type ClusterCache struct {
	CacheConfig `mapstructure:",squash"`
}

type ProjectsCache struct {
	CacheConfig     `mapstructure:",squash"`
	MaxRecordSizeKB int `mapstructure:"cache_projects_max_record_size_kb"`
}

func (pc *ProjectsCache) SetDefaults() {
	pc.CacheConfig.SetDefaults()
	pc.MaxRecordSizeKB = 150
}

type OrgUnitCache struct {
	CacheConfig `mapstructure:",squash"`
}

type HuggingFaceConfig struct {
	MaxConcurrentRequests int    `mapstructure:"max_concurrent_hugging_face_requests"`
	ModelInfoUrl          string `mapstructure:"hugging_face_model_info_url"`
	ModelFileUrl          string `mapstructure:"hugging_face_model_file_url"`
	CalculatorJoinUrl     string `mapstructure:"hugging_face_calculator_join_url"`
	CalculatorDataUrl     string `mapstructure:"hugging_face_calculator_data_url"`
}

func (hfc *HuggingFaceConfig) SetDefaults() {
	hfc.MaxConcurrentRequests = 8
	hfc.ModelInfoUrl = "https://huggingface.co/api/models"
	hfc.ModelFileUrl = "https://huggingface.co/<HF_MODEL>/resolve/main/.gitattributes"
	hfc.CalculatorJoinUrl = "https://hf-accelerate-model-memory-usage.hf.space/queue/join"
	hfc.CalculatorDataUrl = "https://hf-accelerate-model-memory-usage.hf.space/queue/data"

}

type AssetsServiceConfig struct {
	LogLevel             string `mapstructure:"log_level"`
	KeyDelimiter         string `mapstructure:"key_delimiter"`
	KubeConfig           string `mapstructure:"kubeconfig"`              // optional k8s config file, for running outside a pod
	K8sClientConfigQPS   int    `mapstructure:"k8s_client_config_qps"`   // specifies the configuration of k8s client for qps
	K8sClientConfigBurst int    `mapstructure:"k8s_client_config_burst"` // specifies the configuration of k8s client for burst
	PodNamespace         string `mapstructure:"pod_namespace"`           // kubernetes namespace for secrets
}

func (asc *AssetsServiceConfig) SetDefaults() {
	asc.LogLevel = "info"
	asc.KeyDelimiter = "_"
	asc.KubeConfig = ""
	asc.K8sClientConfigQPS = 600
	asc.K8sClientConfigBurst = 3000
	asc.PodNamespace = "runai-backend"
}

func registerConfigs() {
	if err := runai_config.Register[*ServerConfig](); err != nil {
		logging.Fatalf("Failed to initialize server config: %v", err)
	}
	if err := runai_config.Register[*ServicesConfig](); err != nil {
		logging.Fatalf("Failed to initialize services config: %v", err)
	}
	if err := runai_config.RegisterRoot[*SecretConfig](); err != nil {
		logging.Fatalf("Failed to initialize secret config: %v", err)
	}
	if err := runai_config.Register[*HuggingFaceConfig](); err != nil {
		logging.Fatalf("Failed to initialize hugging face config: %v", err)
	}
	if err := runai_config.RegisterRoot[*TenantKeysCache](); err != nil {
		logging.Fatalf("Failed to initialize tenant keys cache config: %v", err)
	}
	if err := runai_config.RegisterRoot[*ClusterCache](); err != nil {
		logging.Fatalf("Failed to initialize clusters cache config: %v", err)
	}
	if err := runai_config.RegisterRoot[*ProjectsCache](); err != nil {
		logging.Fatalf("Failed to initialize projects cache config: %v", err)
	}
	if err := runai_config.RegisterRoot[*OrgUnitCache](); err != nil {
		logging.Fatalf("Failed to initialize org unit cache config: %v", err)
	}
	if err := runai_config.Register[*AssetsServiceConfig](); err != nil {
		logging.Fatalf("Failed to initialize assets service config: %v", err)
	}
}

func Init() {
	once.Do(func() {
		logging.Debugf("initializing environment variables")
		registerConfigs()
	})
}

func MockConfig() {
	// Point the secrets folder at an existing dir (no secret files) so config.Init's
	// SetKbSecretsFolder doesn't fail on the missing default /etc/secrets in tests.
	os.Setenv("SECRETS_MOUNT_PATH", ".")

	os.Setenv("POLICY_SERVICE_URL", "localhost")
	os.Setenv("CLUSTER_SERVICE_URL", "localhost")
	os.Setenv("K8S_OBJECTS_TRACKER_URL", "localhost")
	os.Setenv("DB_BACKEND_SCHEMA", "assets")
	os.Setenv("AUTH_SERVICE_URL", "/")
	os.Setenv("HUGGING_FACE_MODEL_INFO_URL", "/mock")
	os.Setenv("PUBSUB_ENABLED", "false")
}
