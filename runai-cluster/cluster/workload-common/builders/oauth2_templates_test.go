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
	for _, tc := range []struct {
		name    string
		value   string
		enabled bool
	}{
		{name: "unset"},
		{name: "disabled", value: "false"},
		{name: "enabled", value: "true", enabled: true},
		{name: "case insensitive", value: "TRUE", enabled: true},
		{name: "config map name is not enablement", value: "other-templates"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(workloadconfig.EnvVarOauth2ProxyTemplatesEnabled, tc.value)
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
			if !tc.enabled {
				require.Nil(t, templateMount)
				require.Nil(t, templateVolume)
				return
			}
			require.NotNil(t, templateMount)
			require.Equal(t, oauth2proxy.TemplatesPath, templateMount.MountPath)
			require.True(t, templateMount.ReadOnly)
			require.NotNil(t, templateVolume)
			require.NotNil(t, templateVolume.ConfigMap)
			require.Equal(t, oauth2proxy.ConfigMapName, templateVolume.ConfigMap.Name)
			require.Equal(t, []corev1.KeyToPath{{Key: "error.html", Path: "error.html"}}, templateVolume.ConfigMap.Items)
			require.NotNil(t, templateVolume.ConfigMap.Optional)
			require.True(t, *templateVolume.ConfigMap.Optional, "missing templates must not block the workload")
		})
	}
}
