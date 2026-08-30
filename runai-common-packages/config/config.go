package config

import (
	"reflect"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	globalRegistry = newRegistry()
	once           sync.Once
)

type Configer interface {
	SetDefaults()
}

func Init(conf *Config) error {
	var err error

	once.Do(func() {
		if conf == nil {
			return
		}

		if err = validator.New().Struct(conf); err != nil {
			return
		}

		globalRegistry = newRegistryWithConfig(conf)
	})
	return err
}

// SetKbSecretsFolder - sets the folder for mounted secrets in k8s format
func SetKbSecretsFolder(folderPath string) error {
	return globalRegistry.SetKbSecretsFolder(folderPath)
}

// SetConfigFiles - sets the config files to load values from to your registered configs
func SetConfigFiles(paths ...string) error {
	return globalRegistry.SetConfigFiles(paths...)
}

// SetSecretConfigFile - sets the config file for secrets, fields tagged with config:"secret"
func SetSecretConfigFile(path string) error {
	return globalRegistry.SetSecretConfigFile(path)
}

// RegisterRoot - load this struct values from config file and env vars provided and adds it to the registry includes the element name as root
func RegisterRoot[T Configer]() error {
	c := initializeConfigType[T]()

	return globalRegistry.RegisterRoot(c)
}

// Register - load this struct values from config file and env vars provided and adds it to the registry
func Register[T Configer]() error {
	c := initializeConfigType[T]()

	return globalRegistry.Register(c)
}

// Get - Retrieves a registered config, if not registered, returns empty struct (not nil)
func Get[T Configer]() T {
	config := initializeConfigType[T]()

	result := globalRegistry.Get(config)
	if result == nil {
		return config
	}
	return result.(T)
}

// GetR - Retrieves a config, registers it if not already registered
func GetR[T Configer]() (T, error) {
	c := initializeConfigType[T]()

	result := globalRegistry.Get(c)
	if result != nil {
		return result.(T), nil
	}

	err := globalRegistry.Register(c)
	if err != nil {
		return c, err
	}

	result = globalRegistry.Get(c)
	if result != nil {
		return result.(T), nil
	}

	return c, nil
}

// GetRootR - Retrieves a config, registers root if not already registered
func GetRootR[T Configer]() (T, error) {
	c := initializeConfigType[T]()

	result := globalRegistry.Get(c)
	if result != nil {
		return result.(T), nil
	}

	err := globalRegistry.RegisterRoot(c)
	if err != nil {
		return c, err
	}

	result = globalRegistry.Get(c)
	if result != nil {
		return result.(T), nil
	}

	return c, nil
}

// Set - add the provided struct to the registry as is
func Set(config Configer) error {
	return globalRegistry.Set(config)
}

func SetOverrides(c Configer) error {
	return globalRegistry.SetOverrides(c)
}

// Set - add the provided struct to the registry and applies the default values to it before
func SetWithDefaults(config Configer) error {
	return globalRegistry.SetWithDefaults(config)
}

// FlushAll - flushes all the registry
func FlushAll() {
	globalRegistry.FlushAll()
}

func OnConfigChange(callback func()) {
	globalRegistry.OnConfigChange(callback)
}

func initializeConfigType[T Configer]() T {
	var c T

	// c is a pointer to the struct, get the type it points to
	cType := reflect.TypeOf(c).Elem()

	// initialize a new instance of cType, then gets it as an interface{} and cast it to T
	newValue := reflect.New(cType).Interface().(T)

	return newValue
}
