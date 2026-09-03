/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"fmt"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/nim-client/generated"
	"github.com/spf13/pflag"
)

type NimService struct {
	ServingPort string
	Replicas    int32

	Autoscale          bool
	MinReplicas        int32
	MaxReplicas        int32
	Metric             string
	MetricThreshold    int32
	ScaleWindowSeconds int32

	ModelPvc    string
	ModelNewPvc string
	ModelCache  string

	Workers int32

	NgcSecret string

	Category         string
	PriorityClass    string
	Preemptibility   string
	ImagePullSecrets []string

	ReadinessProbe string

	Remove []string
}

const (
	keyPortNim            = "port"
	keyGrpcPortNim        = "grpc-port"
	keyMetricsPortNim     = "metrics-port"
	keyServiceTypeNim     = "service-type"
	keyExposedProtocolNim = "exposed-protocol"
	nimServingPortParser  = "serving-port"
)

func NimConfigurationFlagSet(flags *pflag.FlagSet, spec *NimService) {
	flags.StringVar(&spec.NgcSecret, consts.AttributeNameNgcSecret, "", consts.AttributeUsageNgcSecret)

	flags.StringVar(&spec.ModelPvc, consts.AttributeNameModelPvc, "", consts.AttributeUsageModelPvc)
	flags.StringVar(&spec.ModelNewPvc, consts.AttributeNameModelNewPvc, "", consts.AttributeUsageModelNewPvc)
	flags.StringVar(&spec.ModelCache, consts.AttributeNameModelCache, "", consts.AttributeUsageModelCache)

	flags.StringVar(&spec.ServingPort, consts.AttributeNameServingPort, "", consts.AttributeUsageServingPortNim)
}

func NimScalingFlagSet(flags *pflag.FlagSet, spec *NimService) {
	flags.Int32Var(&spec.Replicas, consts.AttributeNameReplicas, 1, consts.AttributeUsageReplicasNim)

	flags.Int32Var(&spec.MinReplicas, consts.AttributeNameMinReplicas, 1, consts.AttributeUsageMinReplicasNim)
	flags.Int32Var(&spec.MaxReplicas, consts.AttributeNameMaxReplicas, 0, consts.AttributeUsageMaxReplicasNim)
	flags.StringVar(&spec.Metric, consts.AttributeNameMetric, "", consts.AttributeUsageMetricNim)
	flags.Int32Var(&spec.MetricThreshold, consts.AttributeNameMetricThreshold, 0, consts.AttributeUsageMetricThreshold)
	flags.Int32Var(&spec.ScaleWindowSeconds, consts.AttributeNameScaleWindowSeconds, 0, consts.AttributeUsageScaleWindowSeconds)

	flags.Int32Var(&spec.Workers, consts.AttributeNameWorkers, 0, consts.AttributeUsageNimWorkers)
}

func NimProbesFlagSet(flags *pflag.FlagSet, spec *NimService) {
	flags.StringVar(&spec.ReadinessProbe, consts.AttributeNameReadinessProbe, "", consts.AttributeUsageReadinessProbeNim)
}

func ParseReadinessProbe(input string) (*openapi.Probe, error) {
	mp := parser.NewLooseMapParser("readiness-probe", mapset.NewSet[string]("port", "path", "host", "scheme", "initial-delay", "period", "timeout", "success", "failure"))
	am := parser.NewAttributeMapper(input, mp)
	if err := am.Parse(); err != nil {
		return nil, err
	}

	if err := parser.CheckMandatoryFields(am, "port"); err != nil {
		return nil, err
	}

	probe := openapi.NewProbe()

	if portStr := am.Get("port"); portStr != nil {
		port, err := strconv.ParseInt(*portStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid port value: %s", *portStr)
		}

		handler := openapi.NewProbeHandler()
		httpGet := openapi.NewProbeHandlerHttpGet()
		httpGet.SetPort(int32(port))

		if path := am.Get("path"); path != nil {
			httpGet.SetPath(*path)
		}

		if host := am.Get("host"); host != nil {
			httpGet.SetHost(*host)
		}

		if scheme := am.Get("scheme"); scheme != nil {
			httpGet.SetScheme(openapi.ProbeHandlerScheme(*scheme))
		}

		handler.SetHttpGet(*httpGet)
		probe.SetHandler(*handler)
	}

	if initialDelay := am.Get("initial-delay"); initialDelay != nil {
		val, err := strconv.ParseInt(*initialDelay, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid initial-delay value: %s", *initialDelay)
		}
		probe.SetInitialDelaySeconds(int32(val))
	}

	if period := am.Get("period"); period != nil {
		val, err := strconv.ParseInt(*period, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid period value: %s", *period)
		}
		probe.SetPeriodSeconds(int32(val))
	}

	if timeout := am.Get("timeout"); timeout != nil {
		val, err := strconv.ParseInt(*timeout, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout value: %s", *timeout)
		}
		probe.SetTimeoutSeconds(int32(val))
	}

	if success := am.Get("success"); success != nil {
		val, err := strconv.ParseInt(*success, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid success value: %s", *success)
		}
		probe.SetSuccessThreshold(int32(val))
	}

	if failure := am.Get("failure"); failure != nil {
		val, err := strconv.ParseInt(*failure, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid failure value: %s", *failure)
		}
		probe.SetFailureThreshold(int32(val))
	}

	return probe, nil
}

func NimLifecycleFlagSet(flags *pflag.FlagSet, spec *NimService) {
	flags.StringVar(&spec.Category, consts.AttributeNameCategory, "", consts.AttributeUsageCategory)
	flags.StringVar(&spec.PriorityClass, consts.AttributeNamePriorityClass, "", consts.AttributeUsagePriorityClass)
}

func ParseNimServingPort(input string) (*openapi.NimServiceServingPort, error) {
	normalized := input
	if !strings.Contains(input, "=") {
		if _, err := validateContainerPort(input); err == nil {
			normalized = keyPortNim + "=" + input
		}
	}

	allowed := mapset.NewSet[string](
		keyPortNim, keyContainer, keyGrpcPortNim, keyMetricsPortNim,
		keyServiceTypeNim, keyExposedProtocolNim, keyExposeExternally, keyExposedUrl,
	)
	mp := parser.NewLooseMapParser(nimServingPortParser, allowed)
	am := parser.NewAttributeMapper(normalized, mp)
	if err := am.Parse(); err != nil {
		return nil, err
	}

	sp := openapi.NewNimServiceServingPort()

	if v := am.Get(keyPortNim); v != nil {
		if *v == "" {
			return nil, fmt.Errorf(errInvalidValueRequiresKey, keyPortNim)
		}
		port, err := validateContainerPort(*v)
		if err != nil {
			return nil, err
		}
		p := int32(port)
		sp.Port.Set(&p)
	}
	if v := am.Get(keyContainer); v != nil {
		if *v == "" {
			return nil, fmt.Errorf(errInvalidValueRequiresKey, keyContainer)
		}
		port, err := validateContainerPort(*v)
		if err != nil {
			return nil, err
		}
		p := int32(port)
		sp.Port.Set(&p)
	}
	if v := am.Get(keyExposeExternally); v != nil {
		if *v != "" {
			return nil, fmt.Errorf(errInvalidValueShouldNotHaveKey, keyExposeExternally)
		}
		sp.SetExposeExternally(true)
	}
	if v := am.Get(keyExposedUrl); v != nil {
		if *v == "" {
			return nil, fmt.Errorf(errInvalidValueRequiresKey, keyExposedUrl)
		}
		sp.SetExposedUrl(*v)
	}
	if v := am.Get(keyExposedProtocolNim); v != nil {
		if *v == "" {
			return nil, fmt.Errorf(errInvalidValueRequiresKey, keyExposedProtocolNim)
		}
		sp.SetExposedProtocol(openapi.ServingPortExposedProtocol(*v))
	}
	if v := am.Get(keyGrpcPortNim); v != nil {
		if *v == "" {
			return nil, fmt.Errorf(errInvalidValueRequiresKey, keyGrpcPortNim)
		}
		port, err := validateContainerPort(*v)
		if err != nil {
			return nil, err
		}
		sp.SetGrpcPort(int32(port))
	}
	if v := am.Get(keyMetricsPortNim); v != nil {
		if *v == "" {
			return nil, fmt.Errorf(errInvalidValueRequiresKey, keyMetricsPortNim)
		}
		port, err := validateContainerPort(*v)
		if err != nil {
			return nil, err
		}
		sp.SetMetricsPort(int32(port))
	}
	if v := am.Get(keyServiceTypeNim); v != nil {
		if *v == "" {
			return nil, fmt.Errorf(errInvalidValueRequiresKey, keyServiceTypeNim)
		}
		sp.SetServiceType(openapi.ServingPortServiceType(*v))
	}

	return sp, nil
}

func NimConfigurationUpdateFlagSet(flags *pflag.FlagSet, spec *NimService) {
	flags.StringVar(&spec.NgcSecret, consts.AttributeNameNgcSecret, "", consts.AttributeUsageNgcSecret)

	flags.StringVar(&spec.ModelPvc, consts.AttributeNameModelPvc, "", consts.AttributeUsageModelPvc)
	flags.StringVar(&spec.ModelNewPvc, consts.AttributeNameModelNewPvc, "", consts.AttributeUsageModelNewPvc)
	flags.StringVar(&spec.ModelCache, consts.AttributeNameModelCache, "", consts.AttributeUsageModelCache)

	flags.StringVar(&spec.ServingPort, consts.AttributeNameServingPort, "", consts.AttributeUsageServingPortNim)
}

func NimScalingUpdateFlagSet(flags *pflag.FlagSet, spec *NimService) {
	flags.Int32Var(&spec.Replicas, consts.AttributeNameReplicas, 0, consts.AttributeUsageReplicasNimUpdate)

	flags.Int32Var(&spec.MinReplicas, consts.AttributeNameMinReplicas, 0, consts.AttributeUsageMinReplicasNim)
	flags.Int32Var(&spec.MaxReplicas, consts.AttributeNameMaxReplicas, 0, consts.AttributeUsageMaxReplicasNim)
	flags.StringVar(&spec.Metric, consts.AttributeNameMetric, "", consts.AttributeUsageMetricNim)
	flags.Int32Var(&spec.MetricThreshold, consts.AttributeNameMetricThreshold, 0, consts.AttributeUsageMetricThreshold)
	flags.Int32Var(&spec.ScaleWindowSeconds, consts.AttributeNameScaleWindowSeconds, 0, consts.AttributeUsageScaleWindowSeconds)

	flags.Int32Var(&spec.Workers, consts.AttributeNameWorkers, 0, consts.AttributeUsageNimWorkers)
}

func NimLifecycleUpdateFlagSet(flags *pflag.FlagSet, spec *Lifecycle) {
	flags.Var(cli.NewNullablePreemptibilityFlag(&spec.Preemptibility),
		consts.AttributeNamePreemptibility,
		consts.AttributeUsagePreemptibility)
}

func NimSecurityUpdateFlagSet(flags *pflag.FlagSet, spec *Security) {
	flags.Int64Var(&spec.RunAsUid, consts.AttributeNameRunAsUid, 0, consts.AttributeUsageRunAsUid)
	flags.Int64Var(&spec.RunAsGid, consts.AttributeNameRunAsGid, 0, consts.AttributeUsageRunAsGid)
}
