/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer_v2

import (
	"context"
	"slices"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/init_containers"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = DescribeTable("OAuth2 templates follow the cluster capability in init and serving proxies",
	func(version string, supported bool) {
		capabilities, err := checker.NewClusterCapabilitiesCheckerForVersion(version)
		Expect(err).NotTo(HaveOccurred())
		options := composer.FieldsComposerOptions{}
		Expect(SetClusterDependentOptions(context.Background(), &options, capabilities)).To(Succeed())
		Expect(options.OAuth2ProxyTemplatesSupported).To(Equal(supported))

		config := composer.StaticClusterStorageConfig{}
		config.SetExternalAuthUrlEnabled(true)
		config.SetOauth2ProxyImage("registry.example/oauth2-proxy:reported")
		fc := composer.FieldsComposer{
			Options: options, ClusterStorageConfig: config,
			WorkloadName: "workload", ProjectName: "project",
			ItemizedNumbering: map[string]int{}, OauthPortNumber: 4181,
		}
		url := workloadsApi.ExposedUrl{}
		url.SetContainer(8080)
		url.SetUrl("https://workload.example/tool")
		var info *composer.ExposedUrlInfo
		var jupyterPort int32
		initPod := &corev1.PodSpec{}
		Expect(fc.ComposeExposedUrls("url", []workloadsApi.ExposedUrl{url}, initPod,
			UrlsToK8sInitContainers, nil, &jupyterPort, &info, false)()).To(BeNil())
		Expect(info.OAuth2ProxyTemplatesSupported).To(Equal(supported))
		Expect(initPod.InitContainers).To(HaveLen(1))
		Expect(initPod.InitContainers[0].RestartPolicy).NotTo(BeNil())

		sidecarPod := &corev1.PodSpec{Containers: []corev1.Container{{Name: "user"}}}
		Expect(init_containers.BuildOauth2ProxySidecar(sidecarPod, init_containers.ServingProxySpec{
			ContainerPort: 8080, ServingURL: "https://workload.example/", CookiePrefix: "serving",
			WorkloadName: "workload", ProjectName: "project", Config: config,
			ExposedURLInfo: info, OAuth2ProxyTemplatesSupported: options.OAuth2ProxyTemplatesSupported,
		})).To(BeNil())
		Expect(sidecarPod.Containers).To(HaveLen(3))
		Expect(sidecarPod.Containers[0].VolumeMounts).To(BeEmpty())
		Expect(sidecarPod.Containers[1].RestartPolicy).To(BeNil())

		mount := corev1.VolumeMount{Name: oauth2proxy.VolumeName, MountPath: oauth2proxy.TemplatesPath, ReadOnly: true}
		for _, container := range append(initPod.InitContainers, sidecarPod.Containers[1:]...) {
			Expect(container.Image).To(Equal("registry.example/oauth2-proxy:reported"))
			Expect(slices.Contains(container.VolumeMounts, mount)).To(Equal(supported))
		}
		for _, pod := range []*corev1.PodSpec{initPod, sidecarPod} {
			index := slices.IndexFunc(pod.Volumes, func(volume corev1.Volume) bool { return volume.Name == oauth2proxy.VolumeName })
			Expect(index >= 0).To(Equal(supported))
			if supported {
				Expect(pod.Volumes[index].ConfigMap).To(Equal(&corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: oauth2proxy.ConfigMapName},
					Items:                []corev1.KeyToPath{{Key: oauth2proxy.TemplateKey, Path: oauth2proxy.TemplateKey}},
				}))
				Expect(pod.Volumes).To(HaveLen(2), "one shared template volume regardless of proxy count")
			}
		}
	},
	Entry("older cluster keeps embedded templates", "2.99.99", false),
	Entry("capable cluster mounts the provided templates", "3.0.0", true),
)
