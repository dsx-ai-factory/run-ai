/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

// BriefEnvironment represents a single environment in brief format
type BriefEnvironment struct {
	Name          string `json:"name"`
	Scope         string `json:"scope"`
	Image         string `json:"image"`
	Tools         string `json:"tools"`
	WorkloadTypes string `json:"workloadTypes"`
}

// EnvironmentsAlignedRow represents a single environment row with selected columns
type EnvironmentsAlignedRow struct {
	Name             string                                 `json:"Name,omitempty" yaml:"Name,omitempty"`
	Scope            string                                 `json:"Scope,omitempty" yaml:"Scope,omitempty"`
	Image            string                                 `json:"Image,omitempty" yaml:"Image,omitempty"`
	Tools            string                                 `json:"Tools,omitempty" yaml:"Tools,omitempty"`
	WorkloadTypes    string                                 `json:"WorkloadTypes,omitempty" yaml:"WorkloadTypes,omitempty"`
	RawConnections   []openapi.Connection                   `json:"-"`
	RawWorkloadTypes openapi.NullableWorkloadSupportedTypes `json:"-"`
}

// EnvironmentsAlignedData represents the aligned format for JSON/YAML output
type EnvironmentsAlignedData struct {
	Environments []EnvironmentsAlignedRow `json:"environments" yaml:"environments"`
}

// Data holds the raw response data
type Data struct {
	Environments []openapi.EnvironmentAsset `json:"environments" yaml:"environments"`
}

// BriefData represents the brief (non-raw) format
type BriefData struct {
	Environments []BriefEnvironment `json:"environments" yaml:"environments"`
}
