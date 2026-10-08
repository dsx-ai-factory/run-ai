/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
)

const EnvVarPrefixCARunAi = "CUSTOM_CA"

func ComposeCAVolume(envVarPrefix string) corev1.Volume {
	return corev1.Volume{
		Name: config.GetCustomCASecret(envVarPrefix),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				DefaultMode: utils.Ref(int32(420)),
				Items: []corev1.KeyToPath{
					{Key: config.GetCustomCAKey(envVarPrefix), Path: config.GetCustomCAKey(envVarPrefix)},
				},
				Optional:   utils.Ref(true),
				SecretName: config.GetCustomCASecret(envVarPrefix),
			},
		},
	}
}

func ComposeCAVolumeMount(envVarPrefix string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      config.GetCustomCASecret(envVarPrefix),
		MountPath: fmt.Sprintf("/etc/ssl/certs/%s", config.GetCustomCAKey(envVarPrefix)),
		SubPath:   config.GetCustomCAKey(envVarPrefix),
	}
}

func ComposeAuthVolume() corev1.Volume {
	return corev1.Volume{
		Name: constants.ExternalUrlAuthVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

func ComposeAuthProxyVolumes() []corev1.Volume {
	volumes := []corev1.Volume{ComposeAuthVolume()}
	if config.OAuth2ProxyTemplatesEnabled() {
		volumes = append(volumes, corev1.Volume{
			Name: oauth2proxy.VolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: oauth2proxy.ConfigMapName},
					Items:                []corev1.KeyToPath{{Key: oauth2proxy.TemplateKey, Path: oauth2proxy.TemplateKey}},
					Optional:             utils.Ref(true),
				},
			},
		})
	}

	if config.IsCustomCAEnabled(EnvVarPrefixCARunAi) {
		volumes = append(volumes, ComposeCAVolume(EnvVarPrefixCARunAi))
	}

	return volumes
}
