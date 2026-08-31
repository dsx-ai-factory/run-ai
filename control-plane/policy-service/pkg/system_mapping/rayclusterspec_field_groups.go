/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
)

// BuildSystemRayClusterSpecFg builds the field group for the RayCluster spec body.
//
// This field group is bound to:
//   - ray.io/RayCluster/v1   .spec                  (workloadSpec)
//   - ray.io/RayJob/v1       .spec.rayClusterSpec   (rayClusterSpec)
//   - ray.io/RayService/v1   .spec.rayClusterConfig (rayClusterSpec)
//
// Only RayCluster-shape scalars belong here. Pod-template fields come from
// the shared SystemPodSpec* field groups; per-worker-group scalars come
// from SystemRayWorkerGroupSpecFgName.
func BuildSystemRayClusterSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(rayClusterSpecFgFields())
	return fields, nil
}

func rayClusterSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	keyValueMap := dbmodel.BuiltinFormatterKeyValueMap
	rayClusterSpec := SystemRayClusterSpecFormatterName

	return []SystemFgField{
		{
			Name:      "rayVersion",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "enableInTreeAutoscaling",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
		{
			Name:      "managedBy",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "gcsFaultToleranceOptions",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:      "autoscalerOptions-Image",
			Formatter: keyValue,
			SpecRef:   strRef("autoscalerOptions.image"),
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "autoscalerOptions-ImagePullPolicy",
			Formatter: rayClusterSpec,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "autoscalerOptions-IdleTimeoutSeconds",
			Formatter: keyValue,
			SpecRef:   strRef("autoscalerOptions.idleTimeoutSeconds"),
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "autoscalerOptions-UpscalingMode",
			Formatter: rayClusterSpec,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "autoscalerOptions-Version",
			Formatter: rayClusterSpec,
			Type:      api.MAPPINGFIELDTYPE_STRING,
		},
		{
			Name:      "headServiceAnnotations",
			Formatter: keyValueMap,
			SpecRef:   strRef("headServiceAnnotations"),
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:      "suspend",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_BOOLEAN,
		},
	}
}

// BuildSystemRayWorkerGroupSpecFg builds the field group for one element of
// .workerGroupSpecs[]. The bound selectors anchor at a single worker group
// (e.g. workerGroupSpecByGroupName), so SpecRefs are relative to that item.
//
// Identifier field `groupName` is intentionally omitted — it's used by the
// parameterized selector to address the group, not mutated by policies.
func BuildSystemRayWorkerGroupSpecFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(rayWorkerGroupSpecFgFields())
	return fields, nil
}

func rayWorkerGroupSpecFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	keyValueMap := dbmodel.BuiltinFormatterKeyValueMap

	return []SystemFgField{
		{
			Name:      "replicas",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "minReplicas",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "maxReplicas",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "numOfHosts",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "idleTimeoutSeconds",
			Formatter: keyValue,
			Type:      api.MAPPINGFIELDTYPE_INTEGER,
		},
		{
			Name:      "rayStartParams",
			Formatter: keyValueMap,
			SpecRef:   strRef("rayStartParams"),
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
	}
}
