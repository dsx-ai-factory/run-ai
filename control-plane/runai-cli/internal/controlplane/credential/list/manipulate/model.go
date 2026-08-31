/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

type BriefCredential struct {
	Name    string `json:"name"`
	K8sName string `json:"k8sName"`
	Type    string `json:"type"`
	Scope   string `json:"scope"`
	Status  string `json:"status"`
}

type CredentialsAlignedRow struct {
	Name    string `json:"Name,omitempty" yaml:"Name,omitempty"`
	K8sName string `json:"K8sName,omitempty" yaml:"K8sName,omitempty"`
	Type    string `json:"Type,omitempty" yaml:"Type,omitempty"`
	Scope   string `json:"Scope,omitempty" yaml:"Scope,omitempty"`
	Status  string `json:"Status,omitempty" yaml:"Status,omitempty"`
}

type CredentialsAlignedData struct {
	Credentials []CredentialsAlignedRow `json:"credentials" yaml:"credentials"`
}

type Data struct {
	Credentials []openapi.CredentialsListResponseEntry `json:"credentials" yaml:"credentials"`
}

type BriefData struct {
	Credentials []BriefCredential `json:"credentials" yaml:"credentials"`
}
