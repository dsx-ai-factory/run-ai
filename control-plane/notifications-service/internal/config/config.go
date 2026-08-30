package config

import (
	"log"
	"os"
	"reflect"

	"github.com/spf13/viper"
)

func ShouldRootTakeDefaults(v *viper.Viper, key string) bool {
	return v.IsSet(key)
}

// SetDefault - if the field is set by viper (config) and it is an empty map,struct or a pointer, we set the default
func SetDefault(v *viper.Viper, key string, value interface{}) {
	if v.IsSet(key) && !IsConfigFieldEmpty(v.Get(key)) {
		return
	}
	v.Set(key, value)
}

func IsConfigFieldEmpty(config interface{}) bool {
	if config == nil {
		return true
	}

	switch reflect.ValueOf(config).Kind() {
	case reflect.Map:
		return len(config.(map[string]interface{})) == 0
	case reflect.Pointer, reflect.Slice:
		return reflect.ValueOf(config).IsNil()
	case reflect.Struct:
		return reflect.ValueOf(config).IsZero()
	}

	return config == reflect.Zero(reflect.TypeOf(config)).Interface()
}

func CreateTempConfigFile(content string) *os.File {
	file, err := os.CreateTemp("", "service-config*.yaml")
	if err != nil {
		log.Fatal(err)
	}
	file.WriteString(content)
	file.Close()
	return file
}
