package config

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	TestInt  int
	TestVal1 string            `mapstructure:"test_val1"`
	TestVal2 string            `mapstructure:"test_val2"`
	TestMap  map[string]string `mapstructure:"test_map"`
}

func (t *TestConfig) SetDefaults() {
	t.TestVal1 = "test"
}

type SecondTestConfig struct {
	TestVal3 string `mapstructure:"test_val3" validate:"required"`
}

func (t *SecondTestConfig) SetDefaults() {}

type PointerTestConfig struct {
	TestVal1 *bool `mapstructure:"test_val1"`
}

func (t *PointerTestConfig) SetDefaults() {}

type ReferenceTypesConfig struct {
	Map   map[string]string `mapstructure:"map"`
	Slice []string          `mapstructure:"slice"`
}

func (t *ReferenceTypesConfig) SetDefaults() {}

type ConfigWithUnexported struct {
	Exported   string `mapstructure:"exported"`
	unexported string `mapstructure:"unexported"`
}

func (t *ConfigWithUnexported) SetDefaults() {
	t.unexported = "kept"
}

type MiddleConfig struct {
	Inner    ConfigWithUnexported `mapstructure:"inner"`
	OwnField string               `mapstructure:"own_field"`
}

type OuterWithUnexported struct {
	Middle   MiddleConfig `mapstructure:"middle"`
	TopField string       `mapstructure:"top_field"`
}

func (t *OuterWithUnexported) SetDefaults() {}

type ConfigWithUnexportedMutex struct {
	Exported string `mapstructure:"exported"`
	mu       sync.Mutex
}

func (t *ConfigWithUnexportedMutex) SetDefaults() {}

type ConfigWithUnexportedSecret struct {
	Exported       string `mapstructure:"exported"`
	SecretExported string `mapstructure:"secret_exported" config:"secret"`
	secretPrivate  string `mapstructure:"secret_private" config:"secret"`
}

func (t *ConfigWithUnexportedSecret) SetDefaults() {
	t.SecretExported = "from-defaults"
	t.secretPrivate = "from-defaults-private"
}

type NestedWithUnexported struct {
	Inner ConfigWithUnexportedMutex `mapstructure:"inner"`
	Val   string                    `mapstructure:"val"`
}

func (t *NestedWithUnexported) SetDefaults() {}

func TestLoad_PointerConfig_True(t *testing.T) {
	config := `
test_val1: true
`
	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)
	registry.Register(&PointerTestConfig{})
	result := registry.Get(&PointerTestConfig{}).(*PointerTestConfig)
	assert.Equal(t, true, *result.TestVal1)
}

func TestLoad_PointerConfig_False(t *testing.T) {
	config := `
test_val1: false
`
	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)
	registry.Register(&PointerTestConfig{})
	result := registry.Get(&PointerTestConfig{}).(*PointerTestConfig)
	assert.Equal(t, false, *result.TestVal1)
}

func TestLoad_PointerConfig_Nil(t *testing.T) {
	config := `
test_val2: false
`
	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)
	registry.Register(&PointerTestConfig{})
	result := registry.Get(&PointerTestConfig{}).(*PointerTestConfig)
	assert.Nil(t, result.TestVal1)
}

func TestSetOverrides_AppliesImmediately(t *testing.T) {
	registry := newRegistry()

	registry.Register(&TestConfig{})
	err := registry.SetOverrides(&TestConfig{TestVal1: "override"})
	require.NoError(t, err)

	result := registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, "override", result.TestVal1)
}

func TestSetOverrides_BeforeRegisterIsPendingUntilRegistered(t *testing.T) {
	registry := newRegistry()

	err := registry.SetOverrides(&TestConfig{TestVal1: "override"})
	require.NoError(t, err)

	assert.Nil(t, registry.Get(&TestConfig{}), "config has no value before it is registered")
}

func TestSetOverrides_BeforeRegisterAppliedOnRegister(t *testing.T) {
	registry := newRegistry()

	err := registry.SetOverrides(&TestConfig{TestVal2: "override"})
	require.NoError(t, err)

	require.NoError(t, registry.Register(&TestConfig{}))

	result := registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, "override", result.TestVal2)
	assert.Equal(t, "test", result.TestVal1, "fields without an override keep their default")
}

func TestSetOverrides_BeforeRegisterAppliedOnRegisterRoot(t *testing.T) {
	registry := newRegistry()

	err := registry.SetOverrides(&TestConfig{TestVal2: "override"})
	require.NoError(t, err)

	require.NoError(t, registry.RegisterRoot(&TestConfig{}))

	result := registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, "override", result.TestVal2)
	assert.Equal(t, "test", result.TestVal1)
}

func TestSetOverrides_BeforeRegisterLastCallWins(t *testing.T) {
	registry := newRegistry()

	require.NoError(t, registry.SetOverrides(&TestConfig{TestVal1: "first"}))
	require.NoError(t, registry.SetOverrides(&TestConfig{TestVal2: "second"}))

	require.NoError(t, registry.Register(&TestConfig{}))

	result := registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, "second", result.TestVal2)
	assert.Equal(t, "test", result.TestVal1, "a pending override replaces the previous one instead of merging with it")
}

func TestSetOverrides_BeforeRegisterTakesPriorityOverConfigFile(t *testing.T) {
	config := `
test_val1: from_file
`
	f, err := createTempFile(config)
	require.NoError(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	require.NoError(t, registry.SetConfigFiles(f.Name()))

	require.NoError(t, registry.SetOverrides(&TestConfig{TestVal1: "override"}))
	require.NoError(t, registry.Register(&TestConfig{}))

	result := registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, "override", result.TestVal1)
}

func TestSetOverrides_BeforeRegisterIgnoresUnexportedFields(t *testing.T) {
	registry := newRegistry()

	cfg := &ConfigWithUnexported{Exported: "value"}
	cfg.unexported = "secret"

	require.NoError(t, registry.SetOverrides(cfg))
	require.NoError(t, registry.Register(&ConfigWithUnexported{}))

	result := registry.Get(&ConfigWithUnexported{}).(*ConfigWithUnexported)
	assert.Equal(t, "value", result.Exported, "exported fields are overridden on register")
	assert.Equal(t, "kept", result.unexported, "the unexported field of the override is dropped, so the registered value stays")
}

func TestSetOverrides_IgnoresUnexportedFieldSet(t *testing.T) {
	registry := newRegistry()
	require.NoError(t, registry.Register(&ConfigWithUnexported{}))

	cfg := &ConfigWithUnexported{Exported: "value"}
	cfg.unexported = "secret"

	require.NoError(t, registry.SetOverrides(cfg))

	result := registry.Get(&ConfigWithUnexported{}).(*ConfigWithUnexported)
	assert.Equal(t, "value", result.Exported)
	assert.Empty(t, result.unexported, "unexported fields are dropped when the config is cloned")
}

func TestSetOverrides_SucceedsWhenUnexportedFieldIsZero(t *testing.T) {
	registry := newRegistry()
	require.NoError(t, registry.Register(&ConfigWithUnexported{}))

	result := registry.Get(&ConfigWithUnexported{}).(*ConfigWithUnexported)
	assert.Equal(t, "kept", result.unexported, "unexported fields are left alone by removeSecretValuesRecursive")

	err := registry.SetOverrides(&ConfigWithUnexported{Exported: "value"})
	require.NoError(t, err)

	result = registry.Get(&ConfigWithUnexported{}).(*ConfigWithUnexported)
	assert.Equal(t, "value", result.Exported)
	assert.Empty(t, result.unexported, "unexported fields are dropped when the config is cloned")
}

func TestSetOverrides_IgnoresNestedInNestedUnexportedFieldSet(t *testing.T) {
	registry := newRegistry()
	require.NoError(t, registry.Register(&OuterWithUnexported{}))

	inner := ConfigWithUnexported{Exported: "val"}
	inner.unexported = "hidden"
	override := &OuterWithUnexported{
		Middle: MiddleConfig{Inner: inner},
	}

	require.NoError(t, registry.SetOverrides(override))

	result := registry.Get(&OuterWithUnexported{}).(*OuterWithUnexported)
	assert.Equal(t, "val", result.Middle.Inner.Exported)
	assert.Empty(t, result.Middle.Inner.unexported, "nested unexported fields are dropped too")
}

func TestSetOverrides_NestedInNestedExportedFieldsMerged(t *testing.T) {
	registry := newRegistry()
	require.NoError(t, registry.Register(&OuterWithUnexported{}))

	err := registry.SetOverrides(&OuterWithUnexported{
		TopField: "top",
		Middle: MiddleConfig{
			OwnField: "mid",
			Inner:    ConfigWithUnexported{Exported: "inner"},
		},
	})
	require.NoError(t, err)

	result := registry.Get(&OuterWithUnexported{}).(*OuterWithUnexported)
	assert.Equal(t, "top", result.TopField)
	assert.Equal(t, "mid", result.Middle.OwnField)
	assert.Equal(t, "inner", result.Middle.Inner.Exported)
}

func TestSetOverrides_DoesNotAliasCallerMapsAndSlices(t *testing.T) {
	registry := newRegistry()
	require.NoError(t, registry.Register(&ReferenceTypesConfig{}))

	callerMap := map[string]string{"k": "v"}
	callerSlice := []string{"a"}
	require.NoError(t, registry.SetOverrides(&ReferenceTypesConfig{Map: callerMap, Slice: callerSlice}))

	callerMap["k"] = "mutated_by_caller"
	callerSlice[0] = "mutated_by_caller"

	result := registry.Get(&ReferenceTypesConfig{}).(*ReferenceTypesConfig)
	assert.Equal(t, "v", result.Map["k"], "the registry must not share the caller's map")
	assert.Equal(t, "a", result.Slice[0], "the registry must not share the caller's slice")
}

func TestSetOverrides_CloneDoesNotAliasPreviousMapsAndSlices(t *testing.T) {
	registry := newRegistry()

	callerMap := map[string]string{"k": "v"}
	callerSlice := []string{"a"}
	require.NoError(t, registry.Set(&ReferenceTypesConfig{Map: callerMap, Slice: callerSlice}))

	require.NoError(t, registry.SetOverrides(&ReferenceTypesConfig{}))

	callerMap["k"] = "mutated_by_caller"
	callerSlice[0] = "mutated_by_caller"

	result := registry.Get(&ReferenceTypesConfig{}).(*ReferenceTypesConfig)
	assert.Equal(t, "v", result.Map["k"], "the clone must not share the map of the item it was cloned from")
	assert.Equal(t, "a", result.Slice[0], "the clone must not share the slice of the item it was cloned from")
}

func TestSetOverrides_SurvivesAutoReload(t *testing.T) {
	config := `
testint: 1
`
	f, err := createTempFile(config)
	require.NoError(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.NoError(t, err)

	registry.Register(&TestConfig{})
	err = registry.SetOverrides(&TestConfig{TestVal2: "override"})
	require.NoError(t, err)

	change := make(chan struct{}, 1)
	registry.OnConfigChange(func() { change <- struct{}{} })

	time.Sleep(500 * time.Millisecond)
	rewriteFile(t, f.Name(), `
testint: 11
`)
	select {
	case <-change:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the config change notification")
	}

	result := registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, 11, result.TestInt)
	assert.Equal(t, "override", result.TestVal2)
}

func TestSetOverrides_OverrideTakesPriorityOverConfigFile(t *testing.T) {
	config := `
test_val1: from_file
`
	f, err := createTempFile(config)
	require.NoError(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.NoError(t, err)

	registry.Register(&TestConfig{})
	err = registry.SetOverrides(&TestConfig{TestVal1: "override"})
	require.NoError(t, err)

	result := registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, "override", result.TestVal1)
}

func TestSetOverrides_WipedOnSet(t *testing.T) {
	registry := newRegistry()

	registry.Register(&TestConfig{})
	err := registry.SetOverrides(&TestConfig{TestVal1: "override"})
	require.NoError(t, err)

	registry.Set(&TestConfig{TestVal1: "replaced"})

	result := registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, "replaced", result.TestVal1)
	assert.Equal(t, "", result.TestVal2)

	// overrides gone: a reload should not re-apply them
	registry.reloadItem(reflect.TypeOf(TestConfig{}))
	result = registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, "test", result.TestVal1) // SetDefaults value, not override
}

func TestSetOverrides_WipedOnSetWithDefaults(t *testing.T) {
	registry := newRegistry()

	registry.Register(&TestConfig{})
	err := registry.SetOverrides(&TestConfig{TestVal1: "override"})
	require.NoError(t, err)

	registry.SetWithDefaults(&TestConfig{TestVal2: "replaced"})

	result := registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, "test", result.TestVal1) // default, not override
	assert.Equal(t, "replaced", result.TestVal2)
}

func TestSet_SetsTheStruct(t *testing.T) {
	registry := newRegistry()

	testConf := TestConfig{TestVal1: "blala"}
	registry.Set(&testConf)
	result := registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, "blala", result.TestVal1)
}

func TestSet_SetsWithDefaults(t *testing.T) {
	registry := newRegistry()

	testConf := TestConfig{TestVal2: "bla"}
	registry.SetWithDefaults(&testConf)
	result := registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, "test", result.TestVal1)
	assert.Equal(t, "bla", result.TestVal2)
}

func TestSetConfigFiles_Loads(t *testing.T) {
	config := `
test_val1: bla
`
	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)

	registry.Register(&TestConfig{})
	result := registry.Get(&TestConfig{}).(*TestConfig)

	assert.Equal(t, "bla", result.TestVal1)
}

func TestSetConfigFiles_LoadsAndEnvOverrides(t *testing.T) {
	config := `
test_val1: bla
test_val2: bla
`
	t.Setenv("TEST_VAL2", "overrides")

	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)

	registry.Register(&TestConfig{})
	result := registry.Get(&TestConfig{}).(*TestConfig)

	assert.Equal(t, "bla", result.TestVal1)
	assert.Equal(t, "overrides", result.TestVal2)
}

func TestSetConfigFiles_LoadsInt(t *testing.T) {
	config := `
testint: 1
`

	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)

	registry.Register(&TestConfig{})
	result := registry.Get(&TestConfig{}).(*TestConfig)

	assert.Equal(t, 1, result.TestInt)
}

func TestSetConfigFiles_LoadsIntWithRoot(t *testing.T) {
	config := `
test:
  testint: 1
`
	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)

	registry.RegisterRoot(&TestConfig{})
	result := registry.Get(&TestConfig{}).(*TestConfig)

	assert.Equal(t, 1, result.TestInt)
}

func TestSetConfigFiles_RegisterRootReloads(t *testing.T) {
	config := `
test:
  testint: 1
`
	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)

	change := make(chan struct{}, 1)
	registry.OnConfigChange(func() {
		change <- struct{}{}
	})

	registry.RegisterRoot(&TestConfig{})
	result := registry.Get(&TestConfig{}).(*TestConfig)

	assert.Equal(t, 1, result.TestInt)

	time.Sleep(500 * time.Millisecond) // allow time for the watch go routine to start
	config_change := `
test:
  testint: 11
`
	rewriteFile(t, f.Name(), config_change)

	<-change

	result = registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, 11, result.TestInt)
}

func TestSetConfigFiles_RegisterReloads(t *testing.T) {
	config := `
testint: 1
`
	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)

	change := make(chan struct{}, 1)
	registry.OnConfigChange(func() {
		change <- struct{}{}
	})

	registry.Register(&TestConfig{})
	result := registry.Get(&TestConfig{}).(*TestConfig)

	assert.Equal(t, 1, result.TestInt)

	time.Sleep(500 * time.Millisecond) // allow time for the watch go routine to start
	config_change := `
testint: 11
`
	rewriteFile(t, f.Name(), config_change)

	<-change

	result = registry.Get(&TestConfig{}).(*TestConfig)
	assert.Equal(t, 11, result.TestInt)
}

func TestOnConfigChange_CallsCallback(t *testing.T) {
	config := `
test:
  testint: 1
`
	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)

	change := make(chan struct{}, 1)
	registry.OnConfigChange(func() {
		change <- struct{}{}
	})

	registry.RegisterRoot(&TestConfig{})
	result := registry.Get(&TestConfig{}).(*TestConfig)

	assert.Equal(t, 1, result.TestInt)

	time.Sleep(500 * time.Millisecond) // allow time for the watch go routine to start
	config_change := `
test:
  testint: 11
`
	rewriteFile(t, f.Name(), config_change)

	<-change
	assert.Len(t, change, 0)
}

func TestWatchConfigFiles_NotifiesChangeForMultipleFiles(t *testing.T) {
	f1, err := createTempFile(`
test:
  testint: 1
`)
	assert.Nil(t, err)
	defer os.Remove(f1.Name())

	f2, err := createTempFile(`
second_test:
  test_val3: bla
`)
	assert.Nil(t, err)
	defer os.Remove(f2.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f1.Name(), f2.Name())
	require.Nil(t, err)

	change := make(chan struct{}, 1)
	registry.OnConfigChange(func() { change <- struct{}{} })

	registry.RegisterRoot(&TestConfig{})
	assert.Equal(t, 1, registry.Get(&TestConfig{}).(*TestConfig).TestInt)

	registry.RegisterRoot(&SecondTestConfig{})
	assert.Equal(t, "bla", registry.Get(&SecondTestConfig{}).(*SecondTestConfig).TestVal3)

	time.Sleep(500 * time.Millisecond)

	rewriteFile(t, f1.Name(), `
test:
  testint: 11
`)
	<-change
	assert.Equal(t, 11, registry.Get(&TestConfig{}).(*TestConfig).TestInt)

	rewriteFile(t, f2.Name(), `
second_test:
  test_val3: bla1
`)
	<-change
	assert.Equal(t, "bla1", registry.Get(&SecondTestConfig{}).(*SecondTestConfig).TestVal3)
	assert.Len(t, change, 0)
}

func TestWatchConfigFiles_WatcherResetsWhenSetFilesCalledAgain(t *testing.T) {
	f1, err := createTempFile(`
test:
  testint: 1
`)
	assert.Nil(t, err)
	defer os.Remove(f1.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f1.Name())
	require.Nil(t, err)

	calledCount := 0
	registry.OnConfigChange(func() { calledCount++ })

	registry.RegisterRoot(&TestConfig{})
	assert.Equal(t, 1, registry.Get(&TestConfig{}).(*TestConfig).TestInt)

	f2, err := createTempFile(`
test:
  testint: 11
`)
	assert.Nil(t, err)
	defer os.Remove(f2.Name())

	err = registry.SetConfigFiles(f2.Name())
	require.Nil(t, err)
	registry.RegisterRoot(&TestConfig{})

	assert.Equal(t, 11, registry.Get(&TestConfig{}).(*TestConfig).TestInt)

	rewriteFile(t, f1.Name(), `
test:
  testint: 2
`)

	assert.Equal(t, 0, calledCount)
}

func TestSetConfigFiles_LoadsIntWithRootAndEnv(t *testing.T) {
	config := `
test:
  testint: 1
`
	t.Setenv("TEST_TESTINT", "2")
	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)

	registry.RegisterRoot(&TestConfig{})
	result := registry.Get(&TestConfig{}).(*TestConfig)

	assert.Equal(t, 2, result.TestInt)
}

func TestSetConfigFiles_WithoutRootInEnvWontOverride(t *testing.T) {
	config := `
test:
  testint: 1
`
	t.Setenv("TESTINT", "2")
	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)

	registry.RegisterRoot(&TestConfig{})
	result := registry.Get(&TestConfig{}).(*TestConfig)

	assert.Equal(t, 1, result.TestInt)
}

func TestSetConfigFiles_FilePathTakenFromEnv(t *testing.T) {
	config := `
test_val1: bla
`
	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	t.Setenv(envConfigFilePath, f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles()
	require.Nil(t, err)

	registry.Register(&TestConfig{})
	result := registry.Get(&TestConfig{}).(*TestConfig)

	assert.Equal(t, "bla", result.TestVal1)
}

func TestSetConfigFiles_WithoutConfig(t *testing.T) {
	registry := newRegistry()
	err := registry.SetConfigFiles()
	require.Equal(t, err.Error(), "no config file paths provided")
}

func TestRegister_ValidationFails(t *testing.T) {
	registry := newRegistry()
	err := registry.Register(&SecondTestConfig{})
	require.Contains(t, err.Error(), "failed to validate config struct SecondTestConfig")
}

func TestRegister_ValidationSucceeds(t *testing.T) {
	f, err := createTempFile(`
test_val3: bla
`)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)

	err = registry.Register(&SecondTestConfig{})
	require.NoError(t, err)
}

func TestRegisterRoot_ValidationFails(t *testing.T) {
	registry := newRegistry()
	err := registry.RegisterRoot(&SecondTestConfig{})
	require.Contains(t, err.Error(), "failed to validate config struct SecondTestConfig")
}

func TestRegisterRoot_ValidationSucceeds(t *testing.T) {
	f, err := createTempFile(`
second_test:
  test_val3: bla
`)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	err = registry.SetConfigFiles(f.Name())
	require.Nil(t, err)

	err = registry.RegisterRoot(&SecondTestConfig{})
	require.NoError(t, err)
}

func TestReloadItem_ConcurrentWithRegistryMutations(t *testing.T) {
	registry := newRegistry()
	require.NoError(t, registry.Register(&TestConfig{}))

	field := reflect.TypeOf(TestConfig{})
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(4)
		go func() { defer wg.Done(); registry.reloadItem(field) }()
		go func() { defer wg.Done(); _ = registry.SetOverrides(&TestConfig{TestVal2: "override"}) }()
		go func() { defer wg.Done(); _ = registry.Set(&TestConfig{TestVal1: "replaced"}) }()
		go func() {
			defer wg.Done()
			if registry.Get(&TestConfig{}) == nil {
				t.Error("config vanished from the registry during reload")
			}
		}()
	}

	wg.Wait()
	assert.NotNil(t, registry.Get(&TestConfig{}))
}

func TestRegister_UnexportedFieldsDoNotPanic(t *testing.T) {
	// Without field.IsExported(), removeSecretValuesRecursive panics on
	// Addr()/SetZero of unexported fields (e.g. sync.Mutex, secretPrivate).
	tests := map[string]struct {
		conf     Configer
		register func(*registry, Configer) error
		get      Configer
	}{
		"Register mutex": {
			conf:     &ConfigWithUnexportedMutex{},
			register: func(r *registry, c Configer) error { return r.Register(c) },
			get:      &ConfigWithUnexportedMutex{},
		},
		"RegisterRoot mutex": {
			conf:     &ConfigWithUnexportedMutex{},
			register: func(r *registry, c Configer) error { return r.RegisterRoot(c) },
			get:      &ConfigWithUnexportedMutex{},
		},
		"Register nested mutex": {
			conf:     &NestedWithUnexported{},
			register: func(r *registry, c Configer) error { return r.Register(c) },
			get:      &NestedWithUnexported{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			registry := newRegistry()
			require.NotPanics(t, func() {
				require.NoError(t, tt.register(registry, tt.conf))
			})
			assert.NotNil(t, registry.Get(tt.get))
		})
	}
}

func TestRegister_UnexportedSecretTagIsSkippedExportedSecretIsCleared(t *testing.T) {
	t.Setenv("SECRET_EXPORTED", "from-env")
	t.Setenv("SECRET_PRIVATE", "from-env-private")
	t.Setenv("EXPORTED", "ok")

	registry := newRegistry()
	require.NotPanics(t, func() {
		require.NoError(t, registry.Register(&ConfigWithUnexportedSecret{}))
	}, `without IsExported this panics on SetZero of the unexported secret field`)

	result := registry.Get(&ConfigWithUnexportedSecret{}).(*ConfigWithUnexportedSecret)
	assert.Equal(t, "ok", result.Exported)
	assert.Empty(t, result.SecretExported, "exported secrets are still cleared")
	assert.Equal(t, "from-defaults-private", result.secretPrivate, "unexported secrets are skipped, not cleared")
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

func TestGetEnvOrDefault_ReturnsEnvVarValue(t *testing.T) {
	t.Setenv("TEST_VAR", "test")

	val := getEnvOrDefault("TEST_VAR", "default")
	assert.Equal(t, val, "test")
}

func TestGetEnvOrDefault_ReturnsDefaultValue(t *testing.T) {
	val := getEnvOrDefault("TEST_VAR", "default")
	assert.Equal(t, val, "default")
}

func TestIsDir_ExistingFolder(t *testing.T) {
	baseDir, err := os.MkdirTemp("", "test_isdir_*")
	assert.NoError(t, err)
	defer os.RemoveAll(baseDir)

	folderPath := filepath.Join(baseDir, "test_folder")
	err = os.Mkdir(folderPath, 0755)
	assert.NoError(t, err)

	isDirectory, err := isDir(baseDir, "test_folder")
	assert.True(t, isDirectory)
	assert.NoError(t, err)
}

func TestIsDir_ExistingFile(t *testing.T) {
	baseDir, err := os.MkdirTemp("", "test_isdir_*")
	assert.NoError(t, err)
	defer os.RemoveAll(baseDir)

	filePath := filepath.Join(baseDir, "test_file")
	err = os.WriteFile(filePath, []byte("content"), 0644)
	assert.NoError(t, err)

	isDirectory, err := isDir(baseDir, "test_file")
	assert.False(t, isDirectory)
	assert.NoError(t, err)
}

func TestIsDir_NonExistingEntry(t *testing.T) {
	isDirectory, err := isDir(os.TempDir(), "non_existent_123")
	assert.False(t, isDirectory)
	assert.Error(t, err)
}

func TestTrimConfigSuffixFromPascal(t *testing.T) {
	assert.Equal(t, "test", trimConfigSuffixFromPascal("test_config"))
	assert.Equal(t, "test", trimConfigSuffixFromPascal("test_cfg"))
	assert.Equal(t, "test", trimConfigSuffixFromPascal("test_conf"))
	assert.Equal(t, "test", trimConfigSuffixFromPascal("test_configuration"))
	assert.Equal(t, "test_configuration", trimConfigSuffixFromPascal("test_configuration_config"))

}

func TestSetConfigFiles_WithCustomDelimiter(t *testing.T) {
	t.Setenv(envConfigKeyDelimiter, ";")
	t.Setenv("TEST_VAL1", "bla")

	config := `
test_map:
  1.1.1: value1
  1.2.1: value2
`
	f, err := createTempFile(config)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()

	err = registry.SetConfigFiles(f.Name())
	require.NoError(t, err)

	registry.Register(&TestConfig{})
	result := registry.Get(&TestConfig{}).(*TestConfig)

	assert.Equal(t, result.TestVal1, "bla")
	assert.Equal(t, result.TestMap["1.1.1"], "value1")
	assert.Equal(t, result.TestMap["1.2.1"], "value2")
}

func rewriteFile(t *testing.T, filename, config_change string) {
	f, err := os.OpenFile(filename, os.O_RDWR, 0)
	require.NoError(t, err)
	_, err = f.WriteAt([]byte(config_change), 0)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)
}

type secretConfig struct {
	APIToken string `mapstructure:"api_token" config:"secret"`
}

func (*secretConfig) SetDefaults() {}

func TestKubernetesMountedSecretReload(t *testing.T) {
	FlushAll()
	t.Cleanup(FlushAll)

	secretsDir := t.TempDir()
	writeKubernetesSecretVersion(t, secretsDir, "..2026_07_27_01", "initial-token", true)

	t.Setenv("SECRETS_MOUNT_PATH", secretsDir)
	require.NoError(t, SetKbSecretsFolder(""))
	require.NoError(t, Register[*secretConfig]())
	require.Equal(t, "initial-token", Get[*secretConfig]().APIToken)

	changed := make(chan struct{}, 1)
	OnConfigChange(func() {
		select {
		case changed <- struct{}{}:
		default:
		}
	})

	writeKubernetesSecretVersion(t, secretsDir, "..2026_07_27_02", "rotated-token", false)

	select {
	case <-changed:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for mounted secret reload")
	}

	require.Equal(t, "rotated-token", Get[*secretConfig]().APIToken)
}

func TestSetKbSecretsFolderTwiceReplacesWatcher(t *testing.T) {
	FlushAll()
	t.Cleanup(FlushAll)

	firstSecretsDir := t.TempDir()
	secondSecretsDir := t.TempDir()
	writeKubernetesSecretVersion(t, firstSecretsDir, "..2026_07_27_01", "first-token", true)
	writeKubernetesSecretVersion(t, secondSecretsDir, "..2026_07_27_01", "second-token", true)

	require.NoError(t, SetKbSecretsFolder(firstSecretsDir))
	require.NoError(t, SetKbSecretsFolder(secondSecretsDir))
	require.NoError(t, Register[*secretConfig]())
	require.Equal(t, "second-token", Get[*secretConfig]().APIToken)

	changed := make(chan struct{}, 1)
	OnConfigChange(func() {
		select {
		case changed <- struct{}{}:
		default:
		}
	})

	writeKubernetesSecretVersion(t, firstSecretsDir, "..2026_07_27_02", "stale-token", false)
	select {
	case <-changed:
		t.Fatal("the replaced secrets folder triggered a config reload")
	case <-time.After(250 * time.Millisecond):
	}
	require.Equal(t, "second-token", Get[*secretConfig]().APIToken)

	writeKubernetesSecretVersion(t, secondSecretsDir, "..2026_07_27_02", "new-token", false)
	select {
	case <-changed:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the current secrets folder to reload")
	}
	require.Equal(t, "new-token", Get[*secretConfig]().APIToken)
}

func writeKubernetesSecretVersion(t *testing.T, secretsDir, version, token string, initial bool) {
	t.Helper()

	versionDir := filepath.Join(secretsDir, version)
	require.NoError(t, os.Mkdir(versionDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(versionDir, "API_TOKEN"), []byte(token), 0o600))

	if initial {
		require.NoError(t, os.Symlink(version, filepath.Join(secretsDir, "..data")))
		require.NoError(t, os.Symlink(filepath.Join("..data", "API_TOKEN"), filepath.Join(secretsDir, "API_TOKEN")))
		return
	}

	temporaryLink := filepath.Join(secretsDir, "..data_tmp")
	require.NoError(t, os.Symlink(version, temporaryLink))
	require.NoError(t, os.Rename(temporaryLink, filepath.Join(secretsDir, "..data")))
}
