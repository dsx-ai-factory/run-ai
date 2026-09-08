/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
)

const (
	huggingFaceTokenEnvironmentVariable = "HF_TOKEN"
	ngcAPIKeyEnvironmentVariable        = "NGC_API_KEY"
)

func isSensitivePlaintextEnvironmentVariable(name string) bool {
	return name == huggingFaceTokenEnvironmentVariable || name == ngcAPIKeyEnvironmentVariable
}

func redactWorkloadV2SensitiveValues(workload *api.WorkloadV2) {
	if workload == nil {
		return
	}

	workload.RawManifest = copyAndRedactSensitiveValues(workload.RawManifest)
	workload.WorkloadManifest = copyAndRedactSensitiveValues(workload.WorkloadManifest)
}

func redactWorkloadV2GetResponseSensitiveValues(response *api.WorkloadV2GetResponse) {
	if response == nil {
		return
	}
	redactWorkloadV2SensitiveValues(&response.Workload)
}

func redactNimServiceResponseSensitiveValues(response *api.NimServiceResponse) {
	if response == nil || response.Spec.EnvironmentVariables == nil {
		return
	}

	redacted := make([]api.EnvironmentVariable, len(*response.Spec.EnvironmentVariables))
	copy(redacted, *response.Spec.EnvironmentVariables)
	for i := range redacted {
		if redacted[i].Name == nil {
			continue
		}
		if isSensitivePlaintextEnvironmentVariable(*redacted[i].Name) {
			redacted[i].Value = nil
		}
	}
	response.Spec.EnvironmentVariables = &redacted
}

func restoreRedactedNimSensitiveValues(updateSpec *api.NimServiceSpec, storedSpec api.NimServiceSpec) error {
	if updateSpec == nil || updateSpec.EnvironmentVariables == nil {
		return nil
	}

	for i := range *updateSpec.EnvironmentVariables {
		entry := &(*updateSpec.EnvironmentVariables)[i]
		if !isRedactedSensitiveEnvironmentVariable(*entry) {
			continue
		}
		storedValue, found := storedSensitiveEnvironmentVariableValue(*entry.Name, storedSpec.EnvironmentVariables)
		if !found {
			return fmt.Errorf("%s requires a value or credential reference", *entry.Name)
		}
		entry.Value = &storedValue
	}

	return nil
}

func isRedactedSensitiveEnvironmentVariable(environmentVariable api.EnvironmentVariable) bool {
	if environmentVariable.Name == nil {
		return false
	}
	if !isSensitivePlaintextEnvironmentVariable(*environmentVariable.Name) {
		return false
	}
	return environmentVariable.Value == nil &&
		environmentVariable.Secret == nil &&
		environmentVariable.UserCredential == nil &&
		environmentVariable.ConfigMap == nil &&
		environmentVariable.PodFieldRef == nil &&
		(environmentVariable.Exclude == nil || !*environmentVariable.Exclude)
}

func storedSensitiveEnvironmentVariableValue(name string, environmentVariables *[]api.EnvironmentVariable) (string, bool) {
	if environmentVariables == nil {
		return "", false
	}
	for _, environmentVariable := range *environmentVariables {
		if environmentVariable.Name != nil && *environmentVariable.Name == name && environmentVariable.Value != nil {
			return *environmentVariable.Value, true
		}
	}
	return "", false
}

// copyAndRedactSensitiveValues builds a response-only copy of an unstructured manifest.
// This prevents public response sanitization from changing the persisted raw/effective manifests.
func copyAndRedactSensitiveValues(source map[string]interface{}) map[string]interface{} {
	return copyAndRedactSensitiveMap(source, false)
}

func copyAndRedactSensitiveMap(source map[string]interface{}, environmentVariableEntry bool) map[string]interface{} {
	if source == nil {
		return nil
	}

	redacted := make(map[string]interface{}, len(source))
	for key, child := range source {
		redacted[key] = copyAndRedactSensitiveValue(child, key == "env" || key == "environmentVariables")
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
	return redacted
}

func copyAndRedactSensitiveValue(source interface{}, environmentVariableEntry bool) interface{} {
	switch typed := source.(type) {
	case map[string]interface{}:
		return copyAndRedactSensitiveMap(typed, environmentVariableEntry)
	case []interface{}:
		redacted := make([]interface{}, len(typed))
		for i, child := range typed {
			redacted[i] = copyAndRedactSensitiveValue(child, environmentVariableEntry)
		}
		return redacted
	default:
		return typed
	}
}

func restoreRedactedSensitiveManifestValues(update map[string]interface{}, stored map[string]interface{}) error {
	return restoreRedactedSensitiveManifestValue(update, stored, false)
}

func restoreRedactedSensitiveManifestValue(update interface{}, stored interface{}, environmentVariableEntry bool) error {
	switch typedUpdate := update.(type) {
	case map[string]interface{}:
		typedStored, _ := stored.(map[string]interface{})
		if environmentVariableEntry && isNameOnlySensitiveManifestEnvironmentVariable(typedUpdate) {
			storedValue, found := storedLiteralManifestEnvironmentVariableValue(typedUpdate["name"].(string), typedStored)
			if !found {
				return fmt.Errorf("%s requires a value or credential reference", typedUpdate["name"])
			}
			typedUpdate["value"] = storedValue
		}
		for key, child := range typedUpdate {
			var storedChild interface{}
			if typedStored != nil {
				storedChild = typedStored[key]
			}
			if err := restoreRedactedSensitiveManifestValue(child, storedChild, key == "env" || key == "environmentVariables"); err != nil {
				return err
			}
		}
	case []interface{}:
		storedSlice, _ := stored.([]interface{})
		for i, child := range typedUpdate {
			var storedChild interface{}
			if named := findStoredManifestEntryByName(child, storedSlice); named != nil {
				storedChild = named
			} else if i < len(storedSlice) {
				storedChild = storedSlice[i]
			}
			if err := restoreRedactedSensitiveManifestValue(child, storedChild, environmentVariableEntry); err != nil {
				return err
			}
		}
	}
	return nil
}

func isNameOnlySensitiveManifestEnvironmentVariable(environmentVariable map[string]interface{}) bool {
	name, ok := environmentVariable["name"].(string)
	if !ok {
		return false
	}
	if !isSensitivePlaintextEnvironmentVariable(name) {
		return false
	}
	_, hasValue := environmentVariable["value"]
	_, hasValueFrom := environmentVariable["valueFrom"]
	return !hasValue && !hasValueFrom
}

func storedLiteralManifestEnvironmentVariableValue(name string, stored map[string]interface{}) (string, bool) {
	if stored == nil || stored["name"] != name {
		return "", false
	}
	value, found := stored["value"].(string)
	return value, found
}

func findStoredManifestEntryByName(update interface{}, stored []interface{}) interface{} {
	updateMap, ok := update.(map[string]interface{})
	if !ok {
		return nil
	}
	name, ok := updateMap["name"].(string)
	if !ok {
		return nil
	}
	for _, candidate := range stored {
		candidateMap, ok := candidate.(map[string]interface{})
		if ok && candidateMap["name"] == name {
			return candidateMap
		}
	}
	return nil
}
