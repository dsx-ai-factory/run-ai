/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStandaloneProxyPreservesErrorPageWithoutUsernameLoader(t *testing.T) {
	t.Setenv(workloadconfig.EnvVarOauth2ProxyTemplatesEnabled, "true")
	t.Setenv(workloadconfig.EnvVarOauth2ProxyResources, "")
	runner := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{
		Name: "workload", Namespace: "runai-project",
	}}
	plugin := &Plugin{}
	builders := map[string]func() (*appsv1.Deployment, error){
		"legacy networking": func() (*appsv1.Deployment, error) {
			return plugin.buildAuthProxyDeployment(runner, nil, "http://workload:8888", "workload.example.com", "/")
		},
		"endpoints": func() (*appsv1.Deployment, error) {
			return plugin.buildEndpointAuthProxyDeployment(runner, nil, "proxy", "http://workload:8888", "workload.example.com", "/")
		},
	}
	for name, build := range builders {
		t.Run(name, func(t *testing.T) {
			g := NewWithT(t)
			deployment, err := build()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(deployment.Namespace).To(Equal(runner.Namespace))
			pod := deployment.Spec.Template.Spec
			g.Expect(pod.InitContainers).To(BeEmpty())
			g.Expect(pod.Containers[0].VolumeMounts).To(ContainElement(corev1.VolumeMount{
				Name: oauth2proxy.VolumeName, MountPath: oauth2proxy.TemplatesPath, ReadOnly: true,
			}))
			optional := true
			g.Expect(pod.Volumes).To(ContainElement(corev1.Volume{
				Name: oauth2proxy.VolumeName,
				VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: oauth2proxy.ConfigMapName},
					Items:                []corev1.KeyToPath{{Key: oauth2proxy.TemplateKey, Path: oauth2proxy.TemplateKey}},
					Optional:             &optional,
				}},
			}))
		})
	}
}
