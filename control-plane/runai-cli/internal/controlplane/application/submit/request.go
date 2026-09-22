/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

type CreationOptions struct {
	ClusterId         string
	ProjectId         string
	Name              string
	ChartURL          string
	Values            map[string]interface{}
	CredentialAssetId string
	UserCredential    bool
}

func BuildCreationRequest(options CreationOptions) openapi.AIApplicationCreationRequest {
	spec := openapi.AIApplicationSpec{
		ChartRef: openapi.AIApplicationSpecChartRef{
			ChartURL: &options.ChartURL,
		},
		Values: options.Values,
	}

	if options.CredentialAssetId != "" {
		spec.Auth = &openapi.AIApplicationAuth{AssetId: &options.CredentialAssetId}
		if options.UserCredential {
			spec.Auth.SetUserCredential(true)
		}
	}

	return openapi.AIApplicationCreationRequest{
		ClusterId: options.ClusterId,
		ProjectId: options.ProjectId,
		Name:      options.Name,
		Spec:      spec,
	}
}
