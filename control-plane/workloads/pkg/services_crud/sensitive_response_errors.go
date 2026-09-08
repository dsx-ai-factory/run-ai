/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"encoding/json"
	"strings"
)

func sensitiveEnvironmentVariableValuesFromMap(value interface{}) []string {
	var result []string
	switch typed := value.(type) {
	case map[string]interface{}:
		if name, ok := typed["name"].(string); ok {
			if isSensitivePlaintextEnvironmentVariable(name) {
				if literal, ok := typed["value"].(string); ok {
					result = append(result, literal)
				}
			}
		}
		for _, child := range typed {
			result = append(result, sensitiveEnvironmentVariableValuesFromMap(child)...)
		}
	case []interface{}:
		for _, child := range typed {
			result = append(result, sensitiveEnvironmentVariableValuesFromMap(child)...)
		}
	}
	return result
}

func redactSensitiveValuesFromResponse[T any](response *T, values []string) *T {
	if response == nil || len(values) == 0 {
		return response
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return response
	}
	var structuredResponse interface{}
	if err := json.Unmarshal(responseBytes, &structuredResponse); err != nil {
		return response
	}
	structuredResponse = redactSensitiveStrings(structuredResponse, values)
	redactedBytes, err := json.Marshal(structuredResponse)
	if err != nil {
		return response
	}
	var result T
	if err := json.Unmarshal(redactedBytes, &result); err != nil {
		return response
	}
	return &result
}

func redactSensitiveStrings(value interface{}, sensitiveValues []string) interface{} {
	switch typed := value.(type) {
	case string:
		redacted := typed
		for _, sensitiveValue := range sensitiveValues {
			if sensitiveValue != "" {
				redacted = strings.ReplaceAll(redacted, sensitiveValue, "[REDACTED]")
			}
		}
		return redacted
	case map[string]interface{}:
		for key, child := range typed {
			typed[key] = redactSensitiveStrings(child, sensitiveValues)
		}
		return typed
	case []interface{}:
		for index, child := range typed {
			typed[index] = redactSensitiveStrings(child, sensitiveValues)
		}
		return typed
	default:
		return typed
	}
}
