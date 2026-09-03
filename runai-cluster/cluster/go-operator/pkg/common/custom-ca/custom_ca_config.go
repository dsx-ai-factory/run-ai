/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package custom_ca

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	v1 "k8s.io/api/core/v1"
	"sort"
	"strconv"
)

const (
	EnvVarPrefix    = "CUSTOM_CA" // this one is for the control plane
	EnvVarPrefixGit = "CUSTOM_CA_GIT"
	EnvVarPrefixS3  = "CUSTOM_CA_S3"
)

// ConfigEnvVars returns the environment variables that are necessary to set up a custom CA.
// We currently have 3 sets of customCA configurations, one for our control plane, the other
// for GIT and finally for S3.
func ConfigEnvVars(customCAs map[string]*common.CustomCA) []v1.EnvVar {
	envVars := make([]v1.EnvVar, 0)

	keys := make([]string, 0, len(customCAs))

	for k := range customCAs {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, envVarPrefix := range keys {
		customCA := customCAs[envVarPrefix]

		if customCA != nil && customCA.Enabled != nil {
			envVars = append(envVars, v1.EnvVar{Name: envVarPrefix + "_ENABLED", Value: strconv.FormatBool(*customCA.Enabled)})
			// Using a custom-custom CA config, generate the environment variables
			if customCA.Secret != nil && customCA.Secret.Optional != nil && *customCA.Secret.Optional == true {
				envVars = append(envVars, v1.EnvVar{Name: envVarPrefix + "_KEY", Value: customCA.Secret.Key})
				envVars = append(envVars, v1.EnvVar{Name: envVarPrefix + "_SECRET", Value: customCA.Secret.Name})
			}
		}
	}

	return envVars
}
