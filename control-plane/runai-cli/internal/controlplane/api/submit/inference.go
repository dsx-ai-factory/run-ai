/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/pflag"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

type Inference struct {
	ServingPort                  string
	MinReplicas                  openapi.NullableInt32
	MaxReplicas                  openapi.NullableInt32
	InitialReplicas              openapi.NullableInt32
	ActivationReplicas           openapi.NullableInt32
	ConcurrencyHardLimit         openapi.NullableInt32
	ScaleToZeroRetentionSeconds  openapi.NullableInt32
	ScaleDownDelaySeconds        openapi.NullableInt32
	InitializationTimeoutSeconds openapi.NullableInt32
	RequestTimeoutSeconds        openapi.NullableInt32
	Metric                       string
	MetricThreshold              openapi.NullableInt32
	MetricThresholdPercentage    openapi.NullableFloat32
}

// Serving port parsing keys
const (
	keyContainer              = "container"
	keyAuthorizationType      = "authorization-type"
	keyAuthorizedUsers        = "authorized-users"
	keyAuthorizedGroups       = "authorized-groups"
	keyProtocol               = "protocol"
	keyClusterLocalAccessOnly = "cluster-local-access-only"
	keyExposeExternally       = "expose-externally"
	keyExposedUrl             = "exposed-url"
)

// Serving port parsing separators
const (
	separatorKeyValue           = "="
	separatorAuthorizedEntities = ":"
	separatorPairs              = ","
)

// Serving port parsing error messages
const (
	errInvalidServingPortFormat     = "invalid serving port format: empty key-value pair detected"
	errInvalidServingPortKey        = "invalid key: the key '%s' inside '--serving-port' is not supported"
	errInvalidValueRequiresKey      = "invalid value: the key '%s' inside '--serving-port' requires a value"
	errInvalidValueShouldNotHaveKey = "invalid value: the key '%s' should not have a value"
	errInvalidBooleanValue          = "invalid value: the key '%s' inside '--serving-port' must be 'true' or 'false', got '%s'"
	errInvalidPortNumber            = "invalid value: the port number inside '--serving-port' must be a number, got '%s'"
	errPortOutOfRange               = "invalid value: the port number inside '--serving-port' must be between %s, got %d, %w"
	errPortRange                    = "1 and 65535"
	errInvalidKeyValueFormat        = "invalid serving port format, expected format: key=value or key"
)

var (
	ErrInvalidServingPortRange = errors.New("invalid serving port range")
)

// servingPortKeyValue represents a parsed key-value pair from serving port input
type servingPortKeyValue struct {
	Key   string
	Value string
}

// parseServingPortInput parses the input string and returns key-value pairs
func parseServingPortInput(input string) ([]servingPortKeyValue, error) {
	// Check if input is a plain port number - treat as container=PORT
	if _, err := validateContainerPort(input); err == nil {
		return []servingPortKeyValue{{Key: keyContainer, Value: input}}, nil
	} else if errors.Is(err, ErrInvalidServingPortRange) {
		return nil, err
	}

	// Parse key-value pairs
	rawPairs := strings.Split(input, separatorPairs)
	pairs := make([]servingPortKeyValue, 0, len(rawPairs))

	for _, pair := range rawPairs {
		if pair == "" {
			return nil, errors.New(errInvalidServingPortFormat)
		}
		key, value, err := parseKeyValue(pair)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, servingPortKeyValue{Key: key, Value: value})
	}

	return pairs, nil
}

func parseKeyValue(pair string) (string, string, error) {
	kv := strings.SplitN(pair, separatorKeyValue, 2)
	switch len(kv) {
	case 1:
		return kv[0], "", nil
	case 2:
		return kv[0], kv[1], nil
	default:
		return "", "", errors.New(errInvalidKeyValueFormat)
	}
}

// ParseServingPort parses serving port for standard inference
func ParseServingPort(input string) (*openapi.ServingPort, error) {
	pairs, err := parseServingPortInput(input)
	if err != nil {
		return nil, err
	}

	sp := openapi.NewServingPort()
	for _, kv := range pairs {
		if err := assignValueServingPort(sp, kv.Key, kv.Value); err != nil {
			return nil, err
		}
	}
	return sp, nil
}

func assignValueServingPort(sp *openapi.ServingPort, key, value string) error {
	if value == "" && key != keyClusterLocalAccessOnly {
		return fmt.Errorf(errInvalidValueRequiresKey, key)
	} else if value != "" && key == keyClusterLocalAccessOnly {
		return fmt.Errorf(errInvalidValueShouldNotHaveKey, key)
	}

	switch key {
	case keyContainer:
		port, err := validateContainerPort(value)
		if err != nil {
			return err
		}
		sp.Container = *openapi.NewNullableInt32(openapi.PtrInt32(int32(port)))
	case keyAuthorizationType:
		enumVal, err := openapi.NewServingPortAccessAuthorizationTypeEnumFromValue(value)
		if err != nil {
			return err
		}
		sp.AuthorizationType = *openapi.NewNullableServingPortAccessAuthorizationTypeEnum(enumVal)
	case keyAuthorizedUsers:
		sp.AuthorizedUsers = strings.Split(value, separatorAuthorizedEntities)
	case keyAuthorizedGroups:
		sp.AuthorizedGroups = strings.Split(value, separatorAuthorizedEntities)
	case keyProtocol:
		var protocol openapi.NullableServingPortProtocol
		protocol.Set((*openapi.ServingPortProtocol)(&value))
		sp.Protocol = protocol
	case keyClusterLocalAccessOnly:
		sp.ClusterLocalAccessOnly.Set(openapi.PtrBool(true))
	default:
		return fmt.Errorf(errInvalidServingPortKey, key)
	}
	return nil
}

func validateContainerPort(value string) (int, error) {
	port, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf(errInvalidPortNumber, value)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf(errPortOutOfRange, errPortRange, port, ErrInvalidServingPortRange)
	}
	return port, nil
}

func InferenceCommonFlagSet(flagsInference *pflag.FlagSet, specInference *Inference) {
	// ServingPort flags
	flagsInference.StringVar(&specInference.ServingPort, consts.AttributeNameServingPort, "", consts.AttributeUsageServingPort)

	// auto-scaling flags
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.MinReplicas), consts.AttributeNameMinReplicas, consts.AttributeUsageMinReplicas)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.MaxReplicas), consts.AttributeNameMaxReplicas, consts.AttributeUsageMaxReplicas)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.InitialReplicas), consts.AttributeNameInitialReplicas, consts.AttributeUsageInitialReplicas)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.ActivationReplicas), consts.AttributeNameActivationReplicas, consts.AttributeUsageActivationReplicas)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.ConcurrencyHardLimit), consts.AttributeNameConcurrencyHardLimit, consts.AttributeUsageConcurrencyHardLimit)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.ScaleToZeroRetentionSeconds), consts.AttributeNameScaleToZeroRetentionSeconds, consts.AttributeUsageScaleToZeroRetentionSeconds)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.ScaleDownDelaySeconds), consts.AttributeNameScaleDownDelaySeconds, consts.AttributeUsageScaleDownDelaySeconds)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.InitializationTimeoutSeconds), consts.AttributeNameInitializationTimeoutSeconds, consts.AttributeUsageInitializationTimeoutSeconds)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.RequestTimeoutSeconds), consts.AttributeNameRequestTimeoutSeconds, consts.AttributeUsageRequestTimeoutSeconds)
	flagsInference.StringVar(&specInference.Metric, consts.AttributeNameMetric, "", consts.AttributeUsageMetric)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.MetricThreshold), consts.AttributeNameMetricThreshold, consts.AttributeUsageMetricThreshold)
	flagsInference.Var(cli.NewNullableFloat32Flag(&specInference.MetricThresholdPercentage), consts.AttributeNameMetricThresholdPercentage, consts.AttributeUsageMetricThresholdPercentage)
}
