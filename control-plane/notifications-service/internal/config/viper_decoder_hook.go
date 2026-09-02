/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

func CustomViperDecoderHook() viper.DecoderConfigOption {
	return viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(envVarDecoderHook(), fileVarDecoderHook()))
}

var envVarRegex = regexp.MustCompile(`\$\{([^\{\}]+)\}`)

func envVarDecoderHook() mapstructure.DecodeHookFuncType {
	return func(
		fromType reflect.Type,
		toType reflect.Type,
		data interface{},
	) (interface{}, error) {
		return resolveEnvVars(fromType.Kind(), toType.Kind(), data)
	}
}

func resolveEnvVars(fromType, toType reflect.Kind, data interface{}) (interface{}, error) {
	if fromType != reflect.String {
		return data, nil
	}
	strBytes, err := resolveEnvVar(data.(string))
	if err != nil {
		return data, err
	}
	return convertResultToExpectedType(strBytes, toType)
}

// resolveEnvVar searches a field data for env vars (such as '${env_var}'), and replaces the var name with the actual var value
// there may be few of them.
func resolveEnvVar(data string) (res string, err error) {
	matches := envVarRegex.FindAllStringSubmatch(data, -1)
	for _, match := range matches {
		if len(match) != 2 {
			log.Error().Msgf("expected for a sub-match for %q", data)
			continue
		}

		envVarName := match[1]
		resolvedEnvVar, found := os.LookupEnv(envVarName)
		if !found {
			log.Error().Msgf("failed resolving an env var field %q. Env var is not found", envVarName)
			return data, fmt.Errorf("failed resolving an env var field %q. Env var is not found", envVarName)
		}

		data = strings.ReplaceAll(data, match[0], resolvedEnvVar)
		log.Debug().Msgf("succeeded resolving an env var %q", envVarName)
	}

	return data, nil
}

var fileVarRegex = regexp.MustCompile(`\$__file\{([^\{\}]+)\}`)

func fileVarDecoderHook() mapstructure.DecodeHookFuncType {
	return func(
		fromType reflect.Type,
		toType reflect.Type,
		data interface{},
	) (interface{}, error) {
		return resolveFileVars(fromType.Kind(), toType.Kind(), data)
	}
}

func resolveFileVars(fromType, toType reflect.Kind, data interface{}) (interface{}, error) {
	if fromType != reflect.String {
		return data, nil
	}
	strBytes, err := resolveFileVar(data.(string))
	if err != nil {
		return data, err
	}
	return convertResultToExpectedType(strBytes, toType)
}

// resolveFileVar searches a field data for file vars (such as '$__file{env_var}'), and replaces the var name with the actual var value
// there may be few of them.
func resolveFileVar(data string) (res string, err error) {
	matches := fileVarRegex.FindAllStringSubmatch(data, -1)

	for _, match := range matches {
		if len(match) != 2 {
			log.Error().Msgf("expected for a sub-match for %q", data)
			continue
		}

		fileVarName := match[1]
		resolvedFileVar, err := os.ReadFile(fileVarName)
		if err != nil {
			log.Error().Err(err).Msgf("failed resolving a file var %q", fileVarName)
			return data, fmt.Errorf("failed resolving a file var %q", fileVarName)
		}

		data = strings.ReplaceAll(data, match[0], string(resolvedFileVar))
		log.Debug().Msgf("succeeded resolving an file var %q", fileVarName)
	}

	return data, nil
}

func convertResultToExpectedType(val string, toType reflect.Kind) (interface{}, error) {
	switch toType {
	case reflect.String:
		return val, nil
	case reflect.Slice:
		return cast.ToStringSliceE(val)
	case reflect.Bool:
		return cast.ToBoolE(val)
	case reflect.Int:
		return cast.ToIntE(val)
	case reflect.Int8:
		return cast.ToIntE(val)
	case reflect.Uint:
		return cast.ToUintE(val)
	case reflect.Int64:
		return cast.ToInt64E(val)
	default:
		return nil, fmt.Errorf("type %s is not supported. 'conversionMap' should be updated", toType.String())
	}
}
