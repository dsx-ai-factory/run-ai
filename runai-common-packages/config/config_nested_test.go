package config_test

import (
	"os"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/assert"
)

type NestedConfig struct {
	Inner TestConfig
	Val   string
}

func (t *NestedConfig) SetDefaults() {
	t.Inner = TestConfig{
		TestVal1: "test",
	}
	t.Val = "test"
}

type NestedConfigS struct {
	Inner TestConfig `mapstructure:",squash"`
	Val   string
}

func (t *NestedConfigS) SetDefaults() {
	t.Inner = TestConfig{
		TestVal1: "test",
	}
	t.Val = "test"
}

func TestNestedConfig_Defaults(t *testing.T) {
	config.FlushAll()
	config.Register[*NestedConfig]()
	config := config.Get[*NestedConfig]()

	assert.Equal(t, "test", config.Inner.TestVal1)
	assert.Equal(t, "test", config.Val)
}

func TestNestedConfig_InnerEnv(t *testing.T) {
	config.FlushAll()
	t.Setenv("INNER_TEST_VAL1", "bla")
	config.Register[*NestedConfig]()
	config := config.Get[*NestedConfig]()

	assert.Equal(t, "bla", config.Inner.TestVal1)
	assert.Equal(t, "test", config.Val)
}

func TestMultipleConfigs(t *testing.T) {
	config.FlushAll()
	t.Setenv("INNER_TEST_VAL1", "bla")
	t.Setenv("TEST_VAL1", "bla2")
	config.Register[*NestedConfig]()
	config.Register[*TestConfig]()
	nested := config.Get[*NestedConfig]()
	conf := config.Get[*TestConfig]()

	assert.Equal(t, "bla", nested.Inner.TestVal1)
	assert.Equal(t, "test", nested.Val)
	assert.Equal(t, "bla2", conf.TestVal1)
}

func TestMultipleConfigsSquash(t *testing.T) {
	config.FlushAll()
	t.Setenv("TEST_VAL1", "bla")
	config.Register[*NestedConfigS]()
	config.Register[*TestConfig]()
	nested := config.Get[*NestedConfigS]()

	assert.Equal(t, "bla", nested.Inner.TestVal1)
	assert.Equal(t, "test", nested.Val)
}

func TestMultipleConfigs_WithFile(t *testing.T) {
	config.FlushAll()
	t.Setenv("VAL", "bla2")
	configFile := `
inner:
    test_val1: bla
`

	f, err := createTempFile(configFile)
	assert.Nil(t, err)
	err = config.SetConfigFiles(f.Name())
	assert.Nil(t, err)

	conf, err := config.GetR[*NestedConfig]()
	assert.Nil(t, err)
	assert.Equal(t, "bla2", conf.Val)
	assert.Equal(t, "bla", conf.Inner.TestVal1)

}

func TestNestedConfig_SetWitDefaults(t *testing.T) {
	config.FlushAll()
	c := &NestedConfig{Inner: TestConfig{TestVal2: "bla"}}

	err := config.SetWithDefaults(c)
	assert.Nil(t, err)
	conf := config.Get[*NestedConfig]()
	assert.Equal(t, "test", conf.Inner.TestVal1)
	assert.Equal(t, "bla", conf.Inner.TestVal2)
	assert.Equal(t, "test", conf.Val)
}

func createTempFile(data string) (*os.File, error) {
	tempFile, err := os.CreateTemp("", "temp-config-*.yaml")
	if err != nil {
		return nil, err
	}

	if _, err := tempFile.Write([]byte(data)); err != nil {
		return nil, err
	}
	if err := tempFile.Close(); err != nil {
		return nil, err
	}
	return tempFile, nil
}

type ComposedConfig struct {
	TestConfig `mapstructure:",squash"`
}

func (t *ComposedConfig) SetDefaults() {
	t.TestVal1 = "test"
}

func TestComposedConfig_Defaults(t *testing.T) {
	config.FlushAll()
	config.Register[*ComposedConfig]()
	config := config.Get[*ComposedConfig]()

	assert.Equal(t, "test", config.TestVal1)
}

func TestComposedConfig_InnerEnv(t *testing.T) {
	config.FlushAll()
	t.Setenv("TEST_VAL1", "bla")
	config.Register[*ComposedConfig]()
	config := config.Get[*ComposedConfig]()

	assert.Equal(t, "bla", config.TestVal1)
}
