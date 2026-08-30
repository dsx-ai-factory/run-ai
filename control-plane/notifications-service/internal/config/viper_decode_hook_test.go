package config

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Password   string `mapstructure:"password"`
	TypeBool   bool   `mapstructure:"type_bool"`
	TypeInt    int    `mapstructure:"type_int"`
	TypeInt8   int8   `mapstructure:"type_int8"`
	TypeInt64  int64  `mapstructure:"type_int64"`
	TypeUInt   uint   `mapstructure:"type_uint"`
	TwoVars    string `mapstructure:"two_vars"`
	TwoIntVars int    `mapstructure:"two_int_vars"`
}

func TestResolvinEnvVars(t *testing.T) {
	configContent := []byte(`
password: ${my_secret}
type_bool: ${my_bool}
type_int: ${my_int}
type_int8: ${my_int8}
type_int64: ${my_int64}
type_uint: ${my_uint}
two_vars: ${my_uint}any${my_secret}
two_int_vars: ${my_int}3${my_int8}
`)

	unmarshalledConf := &TestStruct{}
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigType("yaml")
	assert.NoError(t, v.ReadConfig(bytes.NewBuffer(configContent)))

	os.Setenv("my_secret", "my_resolved_secret")
	os.Setenv("my_bool", "true")
	os.Setenv("my_int", "111")
	os.Setenv("my_int8", "111")
	os.Setenv("my_int64", "111")
	os.Setenv("my_uint", "111")
	os.Setenv("empty_int", "")

	assert.NoError(t, v.Unmarshal(unmarshalledConf, CustomViperDecoderHook()))
	assert.Equal(t, "my_resolved_secret", unmarshalledConf.Password)
	assert.Equal(t, true, unmarshalledConf.TypeBool)
	assert.Equal(t, 111, unmarshalledConf.TypeInt)
	assert.Equal(t, int8(111), unmarshalledConf.TypeInt8)
	assert.Equal(t, int64(111), unmarshalledConf.TypeInt64)
	assert.Equal(t, uint(111), unmarshalledConf.TypeUInt)
	assert.Equal(t, "111anymy_resolved_secret", unmarshalledConf.TwoVars)
	assert.Equal(t, 1113111, unmarshalledConf.TwoIntVars)
}

func TestResolvingEnvVars_Special(t *testing.T) {
	configContent := []byte(`
password: ${my_secret}
`)

	unmarshalledConf := &TestStruct{}
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigType("yaml")
	assert.NoError(t, v.ReadConfig(bytes.NewBuffer(configContent)))

	os.Setenv("my_secret", "2t$4@pW")

	assert.NoError(t, v.Unmarshal(unmarshalledConf, CustomViperDecoderHook()))
	assert.Equal(t, "2t$4@pW", unmarshalledConf.Password)
}

func TestResolvingSecrets_EnvVarNotSet(t *testing.T) {
	configContent := "${int_not_set}"

	_, err := resolveEnvVars(reflect.String, reflect.Int64, configContent)
	assert.ErrorContains(t, err, "Env var is not found")
	assert.Error(t, err)
}

func TestResolvingSecrets_EnvVarIsEmpty(t *testing.T) {
	configContent := "${empty_int}"
	os.Setenv("empty_int", "")

	result, err := resolveEnvVars(reflect.String, reflect.Int64, configContent)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), result)
}

func TestResolvingSecrets_UnsupportedConversionType(t *testing.T) {
	configContent := "${unsupported}"
	os.Setenv("unsupported", "any")

	_, err := resolveEnvVars(reflect.String, reflect.Array, configContent)
	assert.ErrorContains(t, err, "is not supported")
}

type TestStructFileVars struct {
	Password string `mapstructure:"password"`
}

func TestResolveSecretFromFile(t *testing.T) {
	configContent := []byte(`
password: $__file{my_secret}
`)
	file, err := os.Create("./my_secret")
	file.WriteString("my_resolved_secret")
	assert.NoError(t, err)
	defer os.Remove(file.Name())
	unmarshalledConf := &TestStructFileVars{}
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigType("yaml")
	assert.NoError(t, v.ReadConfig(bytes.NewBuffer(configContent)))

	assert.NoError(t, v.Unmarshal(unmarshalledConf, CustomViperDecoderHook()))
	assert.Equal(t, "my_resolved_secret", unmarshalledConf.Password)
}

func TestResolveSecretFromFile_NotExistent(t *testing.T) {
	configContent := []byte(`
password: $__file{my_secret}
`)
	unmarshalledConf := &TestStructFileVars{}
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigType("yaml")
	assert.NoError(t, v.ReadConfig(bytes.NewBuffer(configContent)))

	assert.Error(t, v.Unmarshal(unmarshalledConf, CustomViperDecoderHook()))
	assert.Equal(t, "", unmarshalledConf.Password)
}
