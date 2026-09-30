/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestOAuth2ProxyTemplateVolume(t *testing.T) {
	for _, configMapName := range []string{"", oauth2proxy.ConfigMapName} {
		t.Run("configMap="+configMapName, func(t *testing.T) {
			t.Setenv(workloadconfig.EnvVarOauth2ProxyTemplatesConfigMap, configMapName)
			t.Setenv(workloadconfig.EnvVarOauth2ProxyResources, "")
			container, err := BuildOAuth2ProxyContainer(OAuth2ProxyConfig{
				ContainerName: "proxy", UpstreamURL: "http://workload:8888",
				ExternalHost: "workload.example.com", Port: 4180,
			})
			require.NoError(t, err)
			require.Contains(t, container.Args, "--custom-templates-dir="+oauth2proxy.TemplatesPath)

			var templateMount *corev1.VolumeMount
			for i := range container.VolumeMounts {
				if container.VolumeMounts[i].Name == oauth2proxy.VolumeName {
					templateMount = &container.VolumeMounts[i]
				}
			}
			var templateVolume *corev1.Volume
			volumes := ComposeAuthProxyVolumes()
			for i := range volumes {
				if volumes[i].Name == oauth2proxy.VolumeName {
					templateVolume = &volumes[i]
				}
			}
			if configMapName == "" {
				require.Nil(t, templateMount, "legacy image still supplies its embedded page")
				require.Nil(t, templateVolume)
				return
			}
			require.NotNil(t, templateMount)
			require.Equal(t, oauth2proxy.TemplatesPath, templateMount.MountPath)
			require.True(t, templateMount.ReadOnly)
			require.NotNil(t, templateVolume)
			require.NotNil(t, templateVolume.ConfigMap)
			require.Equal(t, configMapName, templateVolume.ConfigMap.Name)
			require.Equal(t, []corev1.KeyToPath{{Key: "error.html", Path: "error.html"}}, templateVolume.ConfigMap.Items)
			require.Nil(t, templateVolume.ConfigMap.Optional, "wait for the preserved page rather than silently use the upstream page")
		})
	}
}
