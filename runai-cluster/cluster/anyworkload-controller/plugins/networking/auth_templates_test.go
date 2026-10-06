/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = DescribeTable("OAuth2 proxy template volume", func(value string, enabled bool) {
	GinkgoT().Setenv(workloadconfig.EnvVarOauth2ProxyTemplatesEnabled, value)
	GinkgoT().Setenv(workloadconfig.EnvVarOauth2ProxyResources, "")
	container, err := BuildOAuth2ProxyContainer(OAuth2ProxyConfig{
		ContainerName: "proxy", UpstreamURL: "http://workload:8888",
		ExternalHost: "workload.example.com", Port: 4180,
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(container.Args).To(ContainElement("--custom-templates-dir=" + oauth2proxy.TemplatesPath))

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
	if !enabled {
		Expect(templateMount).To(BeNil())
		Expect(templateVolume).To(BeNil())
		return
	}
	Expect(templateMount).NotTo(BeNil())
	Expect(templateMount.MountPath).To(Equal(oauth2proxy.TemplatesPath))
	Expect(templateMount.ReadOnly).To(BeTrue())
	Expect(templateVolume).NotTo(BeNil())
	Expect(templateVolume.ConfigMap).NotTo(BeNil())
	Expect(templateVolume.ConfigMap.Name).To(Equal(oauth2proxy.ConfigMapName))
	Expect(templateVolume.ConfigMap.Items).To(Equal([]corev1.KeyToPath{{Key: "error.html", Path: "error.html"}}))
	Expect(templateVolume.ConfigMap.Optional).NotTo(BeNil())
	Expect(*templateVolume.ConfigMap.Optional).To(BeTrue(), "missing templates must not block the workload")
},
	Entry("unset", "", false),
	Entry("disabled", "false", false),
	Entry("enabled", "true", true),
	Entry("case insensitive", "TRUE", true),
	Entry("config map name is not enablement", "other-templates", false),
)
