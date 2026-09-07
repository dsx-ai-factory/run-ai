/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
)

const (
	huggingFaceTokenEnvironmentVariable = "HF_TOKEN"
	ngcAPIKeyEnvironmentVariable        = "NGC_API_KEY"
)

func isSensitivePlaintextEnvironmentVariable(name string) bool {
	return name == huggingFaceTokenEnvironmentVariable || name == ngcAPIKeyEnvironmentVariable
}

// redactSensitiveEnvironmentVariables keeps the variable entry while omitting its literal value.
// The returned slice is independent from the input so response sanitization cannot alter persisted data.
func redactSensitiveEnvironmentVariables(environmentVariables *[]api.EnvironmentVariableOfAsset) *[]api.EnvironmentVariableOfAsset {
	if environmentVariables == nil {
		return nil
	}

	redacted := make([]api.EnvironmentVariableOfAsset, len(*environmentVariables))
	copy(redacted, *environmentVariables)
	for i := range redacted {
		if redacted[i].Name == nil {
			continue
		}
		if isSensitivePlaintextEnvironmentVariable(*redacted[i].Name) {
			redacted[i].Value = nil
		}
	}
	return &redacted
}

func redactInferenceSpecificRunParamsSensitiveValues(specificRunParams *api.InferenceSpecificRunParams) *api.InferenceSpecificRunParams {
	if specificRunParams == nil {
		return nil
	}
	redacted := *specificRunParams
	redacted.EnvironmentVariables = redactSensitiveEnvironmentVariables(specificRunParams.EnvironmentVariables)
	return &redacted
}

func redactWorkspaceSpecificRunParamsSensitiveValues(specificRunParams *api.WorkspaceSpecificRunParams) *api.WorkspaceSpecificRunParams {
	if specificRunParams == nil {
		return nil
	}
	redacted := *specificRunParams
	redacted.EnvironmentVariables = redactSensitiveEnvironmentVariables(specificRunParams.EnvironmentVariables)
	return &redacted
}

func redactTrainingSpecificRunParamsSensitiveValues(specificRunParams *api.TrainingSpecificRunParams) *api.TrainingSpecificRunParams {
	if specificRunParams == nil {
		return nil
	}
	redacted := *specificRunParams
	redacted.EnvironmentVariables = redactSensitiveEnvironmentVariables(specificRunParams.EnvironmentVariables)
	return &redacted
}

func redactGenericWorkloadSensitiveValues(workload *dbmodel.GenericWorkload) {
	if workload == nil {
		return
	}

	response := *workload
	response.Spec = workload.Spec
	if workload.Spec.SpecificEnv != nil {
		specificEnv := *workload.Spec.SpecificEnv
		specificEnv.EnvironmentVariables = redactSensitiveEnvironmentVariables(specificEnv.EnvironmentVariables)
		response.Spec.SpecificEnv = &specificEnv
	}
	if workload.Spec.Distributed != nil {
		distributed := *workload.Spec.Distributed
		response.Spec.Distributed = &distributed
		if workload.Spec.Distributed.Master != nil {
			master := *workload.Spec.Distributed.Master
			distributed.Master = &master
			if workload.Spec.Distributed.Master.SpecificEnv != nil {
				masterSpecificEnv := *workload.Spec.Distributed.Master.SpecificEnv
				masterSpecificEnv.EnvironmentVariables = redactSensitiveEnvironmentVariables(masterSpecificEnv.EnvironmentVariables)
				master.SpecificEnv = &masterSpecificEnv
			}
		}
	}
	response.Workload = copyAndRedactSensitiveValuesInUnstructuredMap(workload.Workload)
	*workload = response
}

func copyAndRedactSensitiveValuesInUnstructuredMap(source map[string]interface{}) map[string]interface{} {
	return copyAndRedactSensitiveMapInUnstructuredObject(source, false)
}

func copyAndRedactSensitiveMapInUnstructuredObject(source map[string]interface{}, environmentVariableEntry bool) map[string]interface{} {
	if source == nil {
		return nil
	}

	redacted := make(map[string]interface{}, len(source))
	for key, child := range source {
		redacted[key] = copyAndRedactSensitiveValueInUnstructuredObject(
			child,
			key == "env" || key == "environmentVariables",
		)
	}
	if environmentVariableEntry {
		name, ok := redacted["name"].(string)
		if !ok {
			return redacted
		}
		if isSensitivePlaintextEnvironmentVariable(name) {
			delete(redacted, "value")
		}
	}
	redactSensitiveEnvironmentDictionary(redacted)
	return redacted
}

func redactSensitiveEnvironmentDictionary(source map[string]interface{}) {
	environment, ok := source["environment"].(map[string]interface{})
	if !ok {
		return
	}
	items, ok := environment["items"].(map[string]interface{})
	if !ok {
		return
	}
	for name, item := range items {
		if !isSensitivePlaintextEnvironmentVariable(name) {
			continue
		}
		if itemMap, ok := item.(map[string]interface{}); ok {
			delete(itemMap, "value")
		}
	}
}

func copyAndRedactSensitiveValueInUnstructuredObject(value interface{}, environmentVariableEntry bool) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return copyAndRedactSensitiveMapInUnstructuredObject(typed, environmentVariableEntry)
	case []interface{}:
		redacted := make([]interface{}, len(typed))
		for i, child := range typed {
			redacted[i] = copyAndRedactSensitiveValueInUnstructuredObject(child, environmentVariableEntry)
		}
		return redacted
	default:
		return typed
	}
}
