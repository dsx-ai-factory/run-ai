/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

// BriefDatasource represents a single datasource in brief format
type BriefDatasource struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Scope string `json:"scope"`
}

// DatasourcesAlignedRow represents a single datasource row with selected columns
type DatasourcesAlignedRow struct {
	Name  string `json:"Name,omitempty" yaml:"Name,omitempty"`
	Type  string `json:"Type,omitempty" yaml:"Type,omitempty"`
	Scope string `json:"Scope,omitempty" yaml:"Scope,omitempty"`
}

// DatasourcesAlignedData represents the aligned format for JSON/YAML output
type DatasourcesAlignedData struct {
	Datasources []DatasourcesAlignedRow `json:"datasources" yaml:"datasources"`
}

// Data holds the raw response data
type Data struct {
	Datasources []openapi.DatasourceListResponseEntry `json:"datasources" yaml:"datasources"`
}

// BriefData represents the brief (non-raw) format
type BriefData struct {
	Datasources []BriefDatasource `json:"datasources" yaml:"datasources"`
}
