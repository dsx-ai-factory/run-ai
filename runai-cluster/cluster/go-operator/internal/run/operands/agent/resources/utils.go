/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	v1 "k8s.io/api/core/v1"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

func volumeMountsForHTTPProxy(runaiConfig *runaiv1.RunaiConfig) []v1.VolumeMount {
	global := runaiConfig.Spec.Global
	if global.HTTPProxy.Enabled && global.HTTPProxy.TLSCertSecret != nil {
		return []v1.VolumeMount{
			{
				MountPath: "/opt/runai/tls.crt",
				Name:      global.HTTPProxy.ConfigMapName,
				SubPath:   "tls.crt",
			},
		}
	}

	return nil
}

func envVarsForHTTPProxy(runaiConfig *runaiv1.RunaiConfig) []v1.EnvVar {
	global := runaiConfig.Spec.Global
	envVars := []v1.EnvVar{}
	if global.HTTPProxy.Enabled {
		if global.HTTPProxy.ProxyURL != "" {
			envVars = append(envVars, []v1.EnvVar{
				{
					Name:  "NO_PROXY",
					Value: "$(KUBERNETES_SERVICE_HOST)",
				},
				{
					Name:  "HTTPS_PROXY",
					Value: global.HTTPProxy.ProxyURL,
				},
				{
					Name:  "HTTP_PROXY",
					Value: global.HTTPProxy.ProxyURL,
				},
			}...)
		}

		if global.HTTPProxy.TLSCertSecret != nil {
			envVars = append(envVars, v1.EnvVar{
				Name:  "NODE_EXTRA_CA_CERTS",
				Value: "/opt/runai/tls.crt",
			})
		}

		if global.HTTPProxy.TLSSkipValidation {
			envVars = append(envVars, v1.EnvVar{
				Name:  "NODE_TLS_REJECT_UNAUTHORIZED",
				Value: "0",
			})
		}
	}

	return envVars
}

func volumesForHTTPProxy(runaiConfig *runaiv1.RunaiConfig) []v1.Volume {
	global := runaiConfig.Spec.Global
	if global.HTTPProxy.Enabled && global.HTTPProxy.TLSCertSecret != nil {
		return []v1.Volume{
			{
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: global.HTTPProxy.ConfigMapName,
						},
					},
				},
				Name: global.HTTPProxy.ConfigMapName,
			},
		}
	}

	return nil
}
