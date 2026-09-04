/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runaiv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Deployment", func() {
	BeforeEach(func() {
		GinkgoT().Setenv("IMG_THIRD_PARTY_VALKEY_REGISTRY", "docker.io")
		GinkgoT().Setenv("IMG_THIRD_PARTY_VALKEY_REPOSITORY", "valkey/valkey")
		GinkgoT().Setenv("IMG_THIRD_PARTY_VALKEY_TAG", "9.0.4")
	})

	buildDeployment := func() *appsv1.Deployment {
		runaiConfig := utils.GetDefaultRunaiConfig()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()
		resource := NewDeploymentResource("cluster-redis", fakeClient, runaiConfig)
		obj, err := resource.Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		return obj.(*appsv1.Deployment)
	}

	It("should use the official valkey image from RunaiConfig defaults", func() {
		deployment := buildDeployment()
		Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal("docker.io/valkey/valkey:9.0.4"))
	})

	It("should start valkey-server with the mounted config file", func() {
		deployment := buildDeployment()
		container := deployment.Spec.Template.Spec.Containers[0]
		Expect(container.Command).To(Equal([]string{"valkey-server"}))
		Expect(container.Args).To(Equal([]string{"/usr/local/etc/valkey/valkey.conf"}))
	})

	It("should mount valkey config from the cluster-redis secret", func() {
		deployment := buildDeployment()
		container := deployment.Spec.Template.Spec.Containers[0]
		Expect(container.Env).To(BeEmpty())
		Expect(container.VolumeMounts).To(ConsistOf(
			v1.VolumeMount{Name: "valkey-config", MountPath: "/usr/local/etc/valkey"},
			v1.VolumeMount{Name: "valkey-data", MountPath: "/data"},
		))
		Expect(deployment.Spec.Template.Spec.Volumes).To(ConsistOf(
			v1.Volume{
				Name: "valkey-config",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: DeploymentName,
						Items:      []v1.KeyToPath{{Key: valkeyConfFileName, Path: valkeyConfFileName}},
					},
				},
			},
			v1.Volume{
				Name: "valkey-data",
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		))
	})

	It("should harden the pod and container security context", func() {
		deployment := buildDeployment()
		Expect(deployment.Spec.Template.Spec.AutomountServiceAccountToken).To(Equal(ptr.To(false)))
		container := deployment.Spec.Template.Spec.Containers[0]
		Expect(container.SecurityContext).To(Equal(&v1.SecurityContext{
			AllowPrivilegeEscalation: ptr.To(false),
			RunAsNonRoot:             ptr.To(true),
			RunAsUser:                ptr.To(valkeyImageUID),
			ReadOnlyRootFilesystem:   ptr.To(true),
		}))
	})

	It("should omit runAsUser on OpenShift", func() {
		runaiConfig := utils.GetDefaultRunaiConfig()
		runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor = runaiv1internal.KubernetesFlavorOpenShift
		fakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()
		resource := NewDeploymentResource("cluster-redis", fakeClient, runaiConfig)
		obj, err := resource.Object(context.Background())
		Expect(err).NotTo(HaveOccurred())
		deployment := obj.(*appsv1.Deployment)
		Expect(deployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser).To(BeNil())
	})
})
