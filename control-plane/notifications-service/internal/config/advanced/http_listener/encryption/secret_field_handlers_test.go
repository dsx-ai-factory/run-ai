package encryption

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

type SmtpConfig struct {
	Type     string `mapstructure:"type" json:"type"`
	Subject  string `mapstructure:"subject" json:"subject"`
	SmtpHost string `mapstructure:"smtp_host" json:"smtp_host"`
	SmtpPort int    `mapstructure:"smtp_port" json:"smtp_port"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password" secret:"true"`
	AuthType string `mapstructure:"auth_type" json:"auth_type"`
}

var (
	encryption         = NewEncrypterDecrypter(encryptionKey)
	encPass, _         = encryption.Encrypt("abc!23")
	secretFieldHandler = NewSecretFieldHandler(encryption, &SmtpConfig{})
)

func getFromToConfig() (datatypes.JSON, map[string]any) {
	fromConfig := datatypes.JSON(fmt.Sprintf(`{
  "type": "email",
  "user": "user",
  "password": "%s"
}
`, encPass))

	toConfig := map[string]any{
		"subject": "some_subject",
	}
	return fromConfig, toConfig
}

func TestMergeSecrets_PasswordNotInNewConfig(t *testing.T) {
	fromConfig, toConfig := getFromToConfig()
	mergedConfig, err := secretFieldHandler.MergeSecrets(fromConfig, toConfig)

	require.NoError(t, err)
	require.Equal(t, encPass, mergedConfig["password"])
	require.Equal(t, "some_subject", mergedConfig["subject"])
}

func TestMergeSecrets_PasswordInNewConfig(t *testing.T) {
	fromConfig, toConfig := getFromToConfig()
	toConfig["password"] = "updated_password"

	mergedConfig, err := secretFieldHandler.MergeSecrets(fromConfig, toConfig)

	require.NoError(t, err)
	assert.Equal(t, "updated_password", mergedConfig["password"])
	assert.Equal(t, "some_subject", mergedConfig["subject"])
}

func TestRemoveSecrets_PasswordInConfig(t *testing.T) {
	from, _ := getFromToConfig()
	shouldBeWithoutPassword, err := secretFieldHandler.RemoveSecrets(from)

	require.NoError(t, err)
	require.NotContains(t, shouldBeWithoutPassword, "password")
	require.Equal(t, "user", shouldBeWithoutPassword["user"])
}

func TestDecrypt_PasswordInConfig(t *testing.T) {
	from, _ := getFromToConfig()
	decrypted, err := secretFieldHandler.DecryptSecrets(from)
	require.NoError(t, err)
	d := map[string]interface{}{}
	_ = json.Unmarshal(decrypted, &d)
	assert.Equal(t, "abc!23", d["password"])
}

func TestEncrypt_PasswordInNewConfig(t *testing.T) {
	from, _ := getFromToConfig()
	configWithRawPwd := strings.Replace(string(from), encPass, "some-password", -1)
	encrypted, _ := secretFieldHandler.EncryptSecrets(datatypes.JSON(configWithRawPwd))
	d := map[string]interface{}{}
	_ = json.Unmarshal(encrypted, &d)
	decrypted, err := encryption.Decrypt(d["password"].(string))
	require.NoError(t, err)
	assert.Equal(t, "some-password", decrypted)
}
