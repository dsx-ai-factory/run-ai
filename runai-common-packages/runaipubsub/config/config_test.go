package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) {
	config.FlushAll()
	secretsPath := t.TempDir()
	err := createFileForSecretTesting(secretsPath)
	assert.NoError(t, err)
	t.Setenv("SECRETS_MOUNT_PATH", secretsPath)
}

func TestInitConfig_Defaults(t *testing.T) {
	setup(t)
	t.Setenv("PUBSUB_TYPE", "kafka")
	t.Setenv("PUBSUB_URL", "localhost:12345")
	t.Setenv("SERVICE_NAME", "test-service")

	err := InitConfig(nil)
	assert.NoError(t, err)
	cfg := config.Get[*Pubsub]()

	assert.True(t, *cfg.Enabled)
	assert.True(t, cfg.EnableTLS)
	assert.Equal(t, cfg.SubscribeRetries, 1)
	assert.Equal(t, cfg.PublishRetries, 3)
	assert.Equal(t, cfg.ConnectRetries, 3)
	assert.Equal(t, cfg.SubscriptionInitRetries, 5)
	assert.True(t, cfg.EnableDLQ)
	assert.Equal(t, cfg.WaitForAckSec, 180)
	assert.Equal(t, cfg.PublishRetryInterval, 1)
	assert.False(t, cfg.RetryOnReject)
	assert.Equal(t, cfg.ConnectTimeoutSeconds, 2)
	assert.Equal(t, cfg.CleanerIntervalSeconds, 60)
	assert.Equal(t, cfg.ConnMaxIdleTimeSec, int64(300))
	assert.Equal(t, cfg.ConsumerTimeoutSec, int64(HalfYearSeconds))
	assert.Equal(t, cfg.MaxEntries, int64(500_000))
	assert.Equal(t, cfg.MaxMessageAgeMinutes, int64(60*24*7)) // 7 days
	assert.Equal(t, cfg.Replicas, 3)
	assert.False(t, cfg.EnableAdditionalSubscriber)
}

func TestInitConfig_Override(t *testing.T) {
	setup(t)
	var pstKafka PubSubType = "kafka"
	t.Setenv("PUBSUB_TYPE", "kafka")
	t.Setenv("PUBSUB_URL", "localhost:12345")
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("PUBSUB_ENABLED", "false")

	err := InitConfig(func() {})
	assert.NoError(t, err)
	cfg := config.Get[*Pubsub]()
	assert.Equal(t, pstKafka, cfg.Type)
	assert.Equal(t, "localhost:12345", strings.Split(cfg.Url, ",")[0])
	assert.False(t, *cfg.Enabled)
}

func TestInitConfig_FullConfig(t *testing.T) {
	setup(t)
	t.Setenv("PUBSUB_URL", "localhost:2345")
	t.Setenv("PUBSUB_TYPE", string(PSTNats))
	t.Setenv("SERVICE_NAME", "service")

	err := InitConfig(func() {})
	assert.NoError(t, err)
	cfg := config.Get[*Pubsub]()
	assert.Equal(t, PSTNats, cfg.Type)
	assert.Equal(t, "localhost:2345", cfg.Url)
	assert.Equal(t, "service", cfg.ServiceName)
}

func TestInitConfig_NoURL(t *testing.T) {
	setup(t)
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("PUBSUB_TYPE", string(PSTRedis))
	t.Setenv("PUBSUB_URL", "")
	err := InitConfig(nil)
	assert.EqualError(t, err, "failed to register pubsub config: failed to validate config struct Pubsub, error: Key: 'Pubsub.Url' Error:Field validation for 'Url' failed on the 'required' tag")
}

func TestInitConfig_NoServiceName(t *testing.T) {
	setup(t)
	t.Setenv("PUBSUB_TYPE", string(PSTRedis))
	t.Setenv("PUBSUB_URL", "http://localhost:1234")
	t.Setenv("SERVICE_NAME", "")
	err := InitConfig(nil)
	assert.EqualError(t, err, "failed to register pubsub config: failed to validate config struct Pubsub, error: Key: 'Pubsub.ServiceName' Error:Field validation for 'ServiceName' failed on the 'required' tag")
}

func TestInitConfig_NoType(t *testing.T) {
	setup(t)
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("PUBSUB_URL", "http://localhost:1234")
	t.Setenv("PUBSUB_TYPE", "")
	err := InitConfig(nil)
	assert.EqualError(t, err, "failed to register pubsub config: failed to validate config struct Pubsub, error: Key: 'Pubsub.Type' Error:Field validation for 'Type' failed on the 'required' tag")
}

func TestInitConfig_WithSecrets_FullConfig(t *testing.T) {
	setup(t)
	t.Setenv("PUBSUB_URL", "localhost:2345")
	t.Setenv("PUBSUB_TYPE", string(PSTNats))
	t.Setenv("SERVICE_NAME", "service")
	t.Setenv("PUBSUB_PASSWORD", "password")
	err := InitConfig(func() {})
	assert.NoError(t, err)
	cfg := config.Get[*Pubsub]()
	assert.Equal(t, PSTNats, cfg.Type)
	assert.Equal(t, "localhost:2345", cfg.Url)
	assert.Equal(t, "service", cfg.ServiceName)
	assert.Equal(t, "secret_password", cfg.Password)
}

func TestInitConfig_PasswordOnlyFromSecret(t *testing.T) {
	setup(t)
	t.Setenv("PUBSUB_URL", "localhost:2345")
	t.Setenv("PUBSUB_TYPE", string(PSTNats))
	t.Setenv("SERVICE_NAME", "service")
	err := InitConfig(func() {})
	assert.NoError(t, err)
	cfg := config.Get[*Pubsub]()
	assert.Equal(t, PSTNats, cfg.Type)
	assert.Equal(t, "localhost:2345", cfg.Url)
	assert.Equal(t, "service", cfg.ServiceName)
	assert.Equal(t, "secret_password", cfg.Password)
}

func createFileForSecretTesting(secretsPath string) error {
	fileName := filepath.Join(secretsPath, "PUBSUB_PASSWORD")
	return writeToFile(fileName, "secret_password")
}

func writeToFile(fileName, content string) error {
	tempFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	if _, err := tempFile.WriteString(content); err != nil {
		return err
	}

	return tempFile.Close()
}
