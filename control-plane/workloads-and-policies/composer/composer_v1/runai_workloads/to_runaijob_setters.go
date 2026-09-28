/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runai_workloads

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
)

// store a string value into a string field
func stringToStringField(src *string, target any, _ any) *workloadsApi.Error {
	if src != nil {
		targetPtr, ok := target.(**field_types.StringField)
		if !ok {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("stringToStringField: target type %T is not of type **field_types.StringField", target))
		}
		*targetPtr = &field_types.StringField{
			Value: src,
		}
	}
	return nil
}

// store a string value into a string field, in case the string is non-empty
func nonEmptyStringToStringField(src *string, target any, _ any) *workloadsApi.Error {
	if convutils.IsStringHasValue(src) {
		targetPtr, ok := target.(**field_types.StringField)
		if !ok {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("nonEmptyStringToStringField: target type %T is not of type **field_types.StringField", target))
		}
		*targetPtr = &field_types.StringField{
			Value: src,
		}
	}
	return nil
}

// store a string value into a quantity field (such as memory or memoryLimit)
func nonEmptyStringToQuantityField(src *string, target any, _ any) *workloadsApi.Error {
	if convutils.IsStringHasValue(src) {
		targetPtr, ok := target.(**field_types.QuantityField)
		if !ok {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("nonEmptyStringToQuantityField: target type %T is not of type **field_types.QuantityField", target))
		}
		*targetPtr = &field_types.QuantityField{
			Value: src,
		}
	}
	return nil
}

// store an array as a delimited-string in a string field. for example: array of
// authorized users which we save as a delimited string in the cluster. the
// delimiter can be sent as a cookie to this function, default delimiter is blank
func arrayToStringField(arr []string, target any, cookie any /*the delimiter as string*/) *workloadsApi.Error {
	if len(arr) != 0 {
		if cookie == nil {
			cookie = " " // default delimiter
		}
		return stringToStringField(convutils.Ref(strings.Join(arr, cookie.(string))), target, nil)
	}
	return nil
}

// store a boolean flag into a dictionary field (such as labels, annotations, etc.). the key should be sent to the
// function as a cookie, and the function stores the string as the value of that key.
func stringToDictField(src *string, target any, cookie any /*key of the dictionary, as string*/) *workloadsApi.Error {
	if convutils.IsStringHasValue(src) {
		aMap := map[string]*string{
			cookie.(string): src,
		}
		return mapToDictField(aMap, target, nil)
	}

	return nil
}

// store a boolean flag into a boolean field
func boolToBoolField(src *bool, target any, _ any) *workloadsApi.Error {
	if src != nil {
		targetPtr, ok := target.(**field_types.BoolField)
		if !ok {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("boolToBoolField: target type %T is not of type **field_types.BoolField", target))
		}
		*targetPtr = &field_types.BoolField{
			Value: src,
		}
	}
	return nil
}

// store a boolean flag into a dictionary field (such as label). the key should be sent to the
// function as a cookie, and the function stores "true" value in that key.
func boolToDictField(src *bool, target any, cookie any /*key of the dictionary, as string*/) *workloadsApi.Error {
	if convutils.BoolIsTrue(src) {
		aMap := map[string]*string{
			cookie.(string): convutils.Ref("true"),
		}
		return mapToDictField(aMap, target, nil)
	}

	return nil
}

// store an integer value into an integer field.
func intToIntField[T composer.IntType](src *T, target any, _ any) *workloadsApi.Error {
	if src != nil {
		targetPtr, ok := target.(**field_types.IntField)
		if !ok {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("intToIntField: target type %T is not of type **field_types.IntField", target))
		}
		*targetPtr = &field_types.IntField{
			Value: convutils.Ref(int(*src)),
		}
	}
	return nil
}

// store an integer value to an integer field, if the integer value is non-zero
func nonZeroIntToIntField[T composer.IntType](src *T, target any, _ any) *workloadsApi.Error {
	if src != nil && *src != 0 {
		targetPtr, ok := target.(**field_types.IntField)
		if !ok {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("nonZeroIntToIntField: target type %T is not of type **field_types.IntField", target))
		}
		*targetPtr = &field_types.IntField{
			Value: convutils.Ref(int(*src)),
		}
	}
	return nil
}

// store a float value to a number field
func floatToNumberField(src *float64, target any, _ any) *workloadsApi.Error {
	if src != nil {
		targetPtr, ok := target.(**field_types.NumberField)
		if !ok {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("floatToNumberField: target type %T is not of type **ield_types.NumberField", target))
		}
		*targetPtr = &field_types.NumberField{
			Value: convutils.Ref(fmt.Sprintf("%v", *src)),
		}
	}
	return nil
}

// store a float value to a number field, if the value is non-zero.
func nonZeroFloatToNumberField(src *float64, target any, _ any) *workloadsApi.Error {
	if src != nil && *src != 0 {
		targetPtr, ok := target.(**field_types.NumberField)
		if !ok {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("nonZeroFloatToNumberField: target type %T is not of type **field_types.NumberField", target))
		}
		*targetPtr = &field_types.NumberField{
			Value: convutils.Ref(fmt.Sprintf("%v", *src)),
		}
	}
	return nil
}

// store a map into a dictionary field, like annotations or labels. the keys of the map
// will be used as the keys of the dictionary. a nil value indicates excluded field,
// otherwise value of the map will be stores in the value of the dictionary.
func mapToDictField(dict map[string]*string, target any, _ any) *workloadsApi.Error {
	if len(dict) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.DictionaryField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToDictField: target type %T is not of type **field_types.DictionaryField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.DictionaryField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.DictionaryItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for key, value := range dict {
		if value == nil {
			targetMap[key] = field_types.DictionaryItemAndRules{
				Deleted: convutils.Ref(true),
			}
		} else {
			targetMap[key] = field_types.DictionaryItemAndRules{
				Value: convutils.Ref(fmt.Sprintf("%v", *value)),
			}
		}
	}

	return nil
}

// store a map into a capabilities field. an empty value in the map indicates an excluded key.
// otherwise, each key-value pair of the map will be stores as an item in the dictionary.
func mapToCapabilitiesField(dict map[string]*workloadsApi.Capability, target any, _ any) *workloadsApi.Error {
	if len(dict) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.CapabilitiesField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToCapabilitiesField: target type %T is not of type **field_types.CapabilitiesField", target))
	}

	if *targetPtr == nil {
		*targetPtr = &field_types.CapabilitiesField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.CapabilitiesItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for key, value := range dict {
		if value == nil {
			targetMap[key] = field_types.CapabilitiesItemAndRules{
				Deleted: convutils.Ref(true),
			}
		} else {
			targetMap[key] = field_types.CapabilitiesItemAndRules{
				Value: &field_types.CapabilitiesItemValue{
					Capability: composer.EnumToString(value),
				},
			}
		}
	}

	return nil
}

// store array of environment variables info, collected from env variables compose function,
// to an environment variables field.
func envVarsToDictField(envVarsInfo []composer.EnvVarInfo, target any, cookie any) *workloadsApi.Error {
	result := make(map[string]*string)

	for _, envVarInfo := range envVarsInfo {
		if envVarInfo.Deleted {
			result[envVarInfo.Name] = nil
			continue
		}

		switch envVarInfo.ValueType {
		case composer.EnvValueText:
			result[envVarInfo.Name] = &envVarInfo.ValueFrom
		case composer.EnvValueSecret:
			result[envVarInfo.Name] = convutils.Ref(fmt.Sprintf("SECRET:%s,%s", envVarInfo.ValueFrom, envVarInfo.Key))
		case composer.EnvValueConfigMap:
			result[envVarInfo.Name] = convutils.Ref(fmt.Sprintf("CONFIGMAP:%s,%s", envVarInfo.ValueFrom, envVarInfo.Key))
		case composer.EnvValuePodFieldRef:
			result[envVarInfo.Name] = convutils.Ref(fmt.Sprintf("PODFIELDREF:%s,%s", envVarInfo.ValueFrom, envVarInfo.Key))
		case composer.EnvValueUserCredentials:
			result[envVarInfo.Name] = &envVarInfo.ValueFrom
		}
	}

	return mapToDictField(result, target, cookie)
}

// store gpu sizing information, as collected by the compose gpu function, to various resource allocation fields
func gpuSizingInfoFieldsSetter(src *composer.GpuSizingInfo, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*field_groups.ResourceAllocationFields)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("gpuSizingInfoFieldsSetter: target type %T is not of type *field_groups.ResourceAllocationFields", target))
	}

	if src.NumDevices != nil {
		_ = intToIntField(convutils.Ref(int64(*src.NumDevices)), &targetPtr.GpuDevices, nil)
	}

	_ = floatToNumberField(src.PortionRequest, &targetPtr.Gpu, nil)
	_ = floatToNumberField(src.PortionLimit, &targetPtr.GpuLimit, nil)
	_ = nonEmptyStringToQuantityField(src.MemoryRequest, &targetPtr.GpuMemory, nil)
	_ = nonEmptyStringToQuantityField(src.MemoryLimit, &targetPtr.GpuMemoryLimit, nil)

	return nil
}

func extendedResourcesFieldSetter(src map[string]*workloadsApi.ExtendedResource, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	extendedResourcesItems := make(map[string]field_types.ExtendedResourcesItemAndRules)

	for _, key := range composer.SortedMapKeys(src) {
		instance := src[key]

		if convutils.BoolIsTrue(instance.Exclude.Get()) {
			extendedResourcesItems[key] = field_types.ExtendedResourcesItemAndRules{
				Deleted: convutils.Ref(true),
			}
		} else {
			extendedResourcesItems[key] = field_types.ExtendedResourcesItemAndRules{
				Value: &field_types.ExtendedResourcesItemValue{
					Name:     instance.Resource.Get(),
					Quantity: instance.Quantity.Get(),
				},
			}
		}
	}

	targetPtr, ok := target.(**field_types.ExtendedResourcesField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("extendedResourcesFieldSetter: target type %T is not of type **field_types.ExtendedResourcesField", target))
	}

	*targetPtr = &field_types.ExtendedResourcesField{
		Value: &extendedResourcesItems,
	}

	return nil
}

func autoScalingSetter(src *workloadsApi.AutoScaling, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	targetPtr, ok := target.(*field_groups.AutoScalingFields)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("autoScalingSetter: target type %T is not of type *field_groups.AutoScalingFields", target))
	}

	_ = intToIntField(src.MinReplicas.Get(), &targetPtr.MinScale, nil)
	_ = intToIntField(src.MaxReplicas.Get(), &targetPtr.MaxScale, nil)

	_ = intToIntField(src.InitialReplicas.Get(), &targetPtr.InitialScale, nil)
	_ = intToIntField(src.ActivationReplicas.Get(), &targetPtr.ActivationScale, nil)

	if convutils.HasValue(src.Metric) {
		metricAsString := *src.Metric.Get()

		if isBuiltinMetric(metricAsString) {
			_ = stringToStringField(&metricAsString, &targetPtr.Metric, nil)
		} else {
			_ = stringToStringField(&metricAsString, &targetPtr.MetricName, nil)
		}
	}

	_ = intToIntField(src.MetricThreshold.Get(), &targetPtr.Target, nil)

	if convutils.HasValue(src.MetricThresholdPercentage) {
		metricThresholdPercentage := float64(*src.MetricThresholdPercentage.Get())
		_ = floatToNumberField(&metricThresholdPercentage, &targetPtr.TargetUtilizationPercentage, nil)
	}

	_ = intToIntField(src.ConcurrencyHardLimit.Get(), &targetPtr.ConcurrencyHardLimit, nil)
	_ = intToIntField(src.ScaleToZeroRetentionSeconds.Get(), &targetPtr.ScaleToZeroRetentionSeconds, nil)
	_ = intToIntField(src.ScaleDownDelaySeconds.Get(), &targetPtr.ScaleDownDelaySeconds, nil)

	return nil
}

func isBuiltinMetric(metric string) bool {
	return metric == composer.AutoScalingMetric_THROUGHPUT ||
		metric == composer.AutoScalingMetric_CONCURRENCY ||
		metric == composer.AutoScalingMetric_LATENCY
}

func nodeAffinitySetter(src [][]corev1.NodeSelectorRequirement, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	nodeAffinity := field_types.NodeSelector{
		NodeSelectorTerms: make([]field_types.NodeSelectorTerm, 0, len(src)),
	}

	for _, item := range src {
		nodeAffinity.NodeSelectorTerms = append(nodeAffinity.NodeSelectorTerms, field_types.NodeSelectorTerm{
			MatchExpressions: item,
		})
	}

	targetPtr, ok := target.(**field_types.NodeSelector)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("nodeAffinitySetter: target type %T is not of type **field_types.NodeSelector", target))
	}
	*targetPtr = &nodeAffinity

	return nil
}

func mapToTolerationSetter(src map[string]*workloadsApi.Toleration, target any, _ any) *workloadsApi.Error {
	if len(src) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.TolerationsField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToTolerationSetter: target type %T is not of type **field_types.TolerationsField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.TolerationsField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.TolerationsItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for _, key := range composer.SortedMapKeys(src) {
		instance := src[key]

		if instance == nil {
			targetMap[key] = field_types.TolerationsItemAndRules{
				Deleted: convutils.Ref(true),
			}
			continue
		}

		item := field_types.TolerationsItemValue{
			Key:   instance.Key.Get(),
			Value: instance.Value.Get(),
		}

		if convutils.HasValue(instance.Effect) {
			if *instance.Effect.Get() == workloadsApi.TOLERATIONEFFECT_ANY {
				item.Effect = nil
			} else {
				item.Effect = convutils.Ref(string(instance.GetEffect()))
			}
		}

		if convutils.HasValue(instance.Operator) {
			item.Operator = composer.EnumToString(instance.Operator.Get())
		}

		if convutils.HasValue(instance.Seconds) {
			item.Seconds = convutils.Ref(int(*instance.Seconds.Get()))
		}

		targetMap[key] = field_types.TolerationsItemAndRules{
			Value: &item,
		}
	}

	return nil
}

func probeSetter(src *workloadsApi.Probe, target any, _ any) *workloadsApi.Error {
	httpGetHandler := &workloadsApi.ProbeHandlerHttpGet{}
	if convutils.HasValue(src.Handler) && src.Handler.Get().HttpGet != nil {
		httpGetHandler = src.Handler.Get().HttpGet
	}

	var probe field_types.ProbeMembers

	_ = nonZeroIntToIntField(src.InitialDelaySeconds.Get(), &probe.InitialDelaySeconds, nil)
	_ = nonZeroIntToIntField(src.PeriodSeconds.Get(), &probe.PeriodSeconds, nil)
	_ = nonZeroIntToIntField(src.TimeoutSeconds.Get(), &probe.TimeoutSeconds, nil)
	_ = nonZeroIntToIntField(src.SuccessThreshold.Get(), &probe.SuccessThreshold, nil)
	_ = nonZeroIntToIntField(src.FailureThreshold.Get(), &probe.FailureThreshold, nil)
	_ = nonEmptyStringToStringField(httpGetHandler.Path.Get(), &probe.HttpGetPath, nil)
	_ = nonEmptyStringToStringField(httpGetHandler.Host.Get(), &probe.HttpGetHost, nil)
	_ = nonEmptyStringToStringField(composer.EnumToString(httpGetHandler.Scheme.Get()), &probe.HttpGetScheme, nil)

	port := int(convutils.Deref(httpGetHandler.Port.Get(), 0))
	probe.HttpGetPort = field_types.IntField{
		Value: convutils.Ref(port),
	}

	targetPtr, ok := target.(**field_types.ProbeField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("probeSetter: target type %T is not of type **field_types.ProbeField", target))
	}
	*targetPtr = &field_types.ProbeField{
		Values: &probe,
	}

	return nil
}

func mapToExposedUrls(info *composer.ExposedUrlInfo, target any, _ any) *workloadsApi.Error {
	if info == nil || len(info.Items) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.ExposedUrlsField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToExposedUrls: target type %T is not of type **field_types.ExposedUrlsField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.ExposedUrlsField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.ExposedUrlItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for _, key := range composer.SortedMapKeys(info.Items) {
		item := info.Items[key]

		if item == nil {
			targetMap[key] = field_types.ExposedUrlItemAndRules{
				Deleted: convutils.Ref(true),
			}
			continue
		}

		urlItem := field_types.ExposedUrlItemValue{
			ContainerPort:    item.ContainerPort,
			AuthorizedUsers:  item.AuthorizedUsers,
			AuthorizedGroups: item.AuthorizedGroups,
		}

		if item.Url != "" {
			urlItem.Url = convutils.Ref(item.Url)
			urlItem.CustomUrl = convutils.Ref(true)
		}

		targetMap[key] = field_types.ExposedUrlItemAndRules{
			Value: &urlItem,
		}
	}

	return nil
}

func mapToPortField(info *composer.PortsInfo, target any, _ any) *workloadsApi.Error {
	if info == nil || len(info.Items) == 0 {
		return nil
	}

	targetPtr, ok := target.(**field_types.PortsField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("mapToPort: target type %T is not of type **field_types.PortsField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.PortsField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.PortItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	for _, key := range composer.SortedMapKeys(info.Items) {
		instance := info.Items[key]

		if instance == nil {
			targetMap[key] = field_types.PortItemAndRules{
				Deleted: convutils.Ref(true),
			}
			continue
		}

		item := field_types.PortItemValue{
			ServiceType: (*string)(instance.ServiceType.Get()),
		}

		if convutils.HasValue(instance.Container) {
			item.ContainerPort = convutils.Ref(int(convutils.Deref(instance.Container.Get(), 0)))
		}

		if convutils.HasValue(instance.External) {
			item.HostPort = convutils.Ref(int(convutils.Deref(instance.External.Get(), 0)))
		} else {
			item.AutoGenHostPort = convutils.Ref(true)
		}

		targetMap[key] = field_types.PortItemAndRules{
			Value: &item,
		}
	}

	return nil
}

func servingPortToPortField(src *composer.ServingPortInfo, target any, _ any) *workloadsApi.Error {
	if src == nil || src.PortInfo == nil {
		return nil
	}

	targetPtr, ok := target.(**field_types.PortsField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("servingPort: target type %T is not of type **field_types.PortsField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.PortsField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.PortItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	item := &field_types.PortItemValue{
		ServiceType: src.ServiceType,
	}

	if convutils.HasValue(src.PortInfo.Container) {
		containerPort := convutils.Deref(src.PortInfo.Container.Get(), 0)
		item.ContainerPort = convutils.Ref(int(containerPort))
	}

	if convutils.HasValue(src.PortInfo.Protocol) {
		item.Protocol = convutils.Ref(*src.PortInfo.Protocol.Get())
	}

	targetMap[src.ItemId] = field_types.PortItemAndRules{
		Value: item,
	}

	return nil
}

func servingPortToServingPortField(src *composer.ServingPortInfo, target any, _ any) *workloadsApi.Error {
	if src == nil || src.PortInfo == nil {
		return nil
	}

	targetPtr, ok := target.(**field_types.ServingPortsField)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("servingPortToServingPortField: target type %T is not of type **field_types.ServingPortsField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.ServingPortsField{}
	}
	if (*targetPtr).Value == nil {
		(*targetPtr).Value = &map[string]field_types.ServingPortItemAndRules{}
	}

	targetMap := *((*targetPtr).Value)

	item := &field_types.ServingPortItemValue{
		ExposedUrl:       src.ExposedUrl,
		ExposeExternally: src.ExposedExternally,
	}

	if convutils.HasValue(src.PortInfo.Container) {
		containerPort := convutils.Deref(src.PortInfo.Container.Get(), 0)
		item.Port = convutils.Ref(int(containerPort))
	}

	if convutils.HasValue(src.PortInfo.Protocol) {
		item.Protocol = convutils.Ref(*src.PortInfo.Protocol.Get())
		if *item.Protocol != "http" {
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("servingPort: unsupported protocol '%s', only 'http' is supported", *item.Protocol))
		}
	}

	item.AuthorizationType = src.PortInfo.AuthorizationType.Get()

	if len(src.AuthInfo.AuthorizedUsers) != 0 {
		item.AuthorizedUsers = convutils.Ref(strings.Join(src.AuthInfo.AuthorizedUsers, ","))
	}

	if len(src.AuthInfo.AuthorizedGroups) != 0 {
		item.AuthorizedGroups = convutils.Ref(strings.Join(src.AuthInfo.AuthorizedGroups, ","))
	}

	targetMap[src.ItemId] = field_types.ServingPortItemAndRules{
		Value: item,
	}

	return nil
}

func intToRunPolicyBackoffLimit(src *int64, target any, _ any) *workloadsApi.Error {
	if convutils.Deref(src, 0) != 0 {
		targetPtr, err := ensureRunPolicyField(target)
		if err != nil {
			return err
		}
		return nonZeroIntToIntField(src, &((*targetPtr).Values.BackoffLimit), nil)
	}
	return nil
}

func intToSlotsPerWorker(src *int32, target any, _ any) *workloadsApi.Error {
	if convutils.Deref(src, 0) != 0 {
		targetPtr, err := ensureMpiFields(target)
		if err != nil {
			return err
		}
		return nonZeroIntToIntField(src, &((*targetPtr).Values.SlotsPerWorker), nil)
	}
	return nil
}

func intToMinReplicas(src *int32, target any, _ any) *workloadsApi.Error {
	if convutils.Deref(src, 0) != 0 {
		targetPtr, err := ensurePytorchFields(target)
		if err != nil {
			return err
		}
		return nonZeroIntToIntField(src, &((*targetPtr).Values.MinReplicas), nil)
	}
	return nil
}

func intToMaxReplicas(src *int32, target any, _ any) *workloadsApi.Error {
	if convutils.Deref(src, 0) != 0 {
		targetPtr, err := ensurePytorchFields(target)
		if err != nil {
			return err
		}
		return nonZeroIntToIntField(src, &((*targetPtr).Values.MaxReplicas), nil)
	}
	return nil
}

// make sure mpi field is initialized, otherwise initialize it. this is a utility
// function for mpi related fields
func ensureMpiFields(target any) (*field_types.MpiField, *workloadsApi.Error) {
	targetPtr, ok := target.(**field_types.MpiField)
	if !ok {
		return nil, workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("ensureMpiFields: target type %T is not of type **field_types.MpiField", target))
	}
	if *targetPtr == nil {
		*targetPtr = &field_types.MpiField{
			Values: &field_types.MpiMembers{},
		}
	}

	return *targetPtr, nil
}

// make sure pytorch field is initialized, otherwise initialize it. this is a utility
// function for pytorch related fields
func ensurePytorchFields(target any) (*field_types.PyTorchField, *workloadsApi.Error) {
	targetPtr, ok := target.(**field_types.PyTorchField)
	if !ok {
		return nil, workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("ensurePytorchFields: target type %T is not of type **field_types.PyTorchField", target))
	}

	if *targetPtr == nil {
		*targetPtr = &field_types.PyTorchField{
			Values: &field_types.PyTorchMembers{},
		}
	}

	return *targetPtr, nil
}

// make sure run-policy field is initialized, otherwise initialize it. this is a utility
// function for run-policy related fields
func ensureRunPolicyField(target any) (*field_types.RunPolicyField, *workloadsApi.Error) {
	targetPtr, ok := target.(**field_types.RunPolicyField)
	if !ok {
		return nil, workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("ensureRunPolicyField: target type %T is not of type **field_types.RunPolicyField", target))
	}

	if *targetPtr == nil {
		*targetPtr = &field_types.RunPolicyField{
			Values: &field_types.RunPolicyMembers{},
		}
	}

	return *targetPtr, nil
}

func stringToRunPolicyCleanPod(src *string, target any, _ any) *workloadsApi.Error {
	if convutils.Deref(src, "") != "" {
		targetPtr, err := ensureRunPolicyField(target)
		if err != nil {
			return err
		}
		return nonEmptyStringToStringField(src, &((*targetPtr).Values.CleanPodPolicy), nil)
	}
	return nil
}

func stringToLauncherCreationPolicy(src *string, target any, _ any) *workloadsApi.Error {
	if convutils.Deref(src, "") != "" {
		targetPtr, err := ensureMpiFields(target)
		if err != nil {
			return err
		}
		return nonEmptyStringToStringField(src, &((*targetPtr).Values.LauncherCreationPolicy), nil)
	}
	return nil
}

func stringToSshAuthMountPath(src *string, target any, _ any) *workloadsApi.Error {
	if convutils.Deref(src, "") != "" {
		targetPtr, err := ensureMpiFields(target)
		if err != nil {
			return err
		}
		return nonEmptyStringToStringField(src, &((*targetPtr).Values.SshAuthMountPath), nil)
	}
	return nil
}
