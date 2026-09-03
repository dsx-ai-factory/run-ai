/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Init CA deployment", func() {
	It("enables leader election without legacy controller flags", func() {
		Expect(handleDeploymentArgs()).To(Equal([]string{"--leader-elect"}))
	})

	securityTests := []struct {
		name          string
		flavor        runv1internal.KubernetesFlavor
		wantRunAsUser *int64
	}{
		{
			name:          "Kubernetes",
			flavor:        runv1internal.KubernetesFlavorVanilla,
			wantRunAsUser: ptr.To[int64](10000),
		},
		{
			name:   "OpenShift",
			flavor: runv1internal.KubernetesFlavorOpenShift,
		},
	}

	for _, tt := range securityTests {
		tt := tt
		It("hardens the init-ca container on "+tt.name, func() {
			scheme := runtime.NewScheme()
			Expect(appsv1.AddToScheme(scheme)).To(Succeed())
			Expect(runv1.AddToScheme(scheme)).To(Succeed())
			reader := fake.NewClientBuilder().WithScheme(scheme).Build()
			runaiConfig := &runv1.RunaiConfig{}
			runaiConfig.SetDefaultsWhereNeeded()
			runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor = tt.flavor

			resource, err := DeploymentForRunaiConfig(context.Background(), reader, runaiConfig)

			Expect(err).NotTo(HaveOccurred())
			deployment := resource.(*appsv1.Deployment)
			Expect(deployment.Spec.Template.Spec.AutomountServiceAccountToken).To(Equal(ptr.To(true)))
			Expect(deployment.Spec.Template.Spec.Containers[0].SecurityContext).To(Equal(&corev1.SecurityContext{
				AllowPrivilegeEscalation: ptr.To(false),
				Capabilities: &corev1.Capabilities{
					Drop: []corev1.Capability{"ALL"},
				},
				ReadOnlyRootFilesystem: ptr.To(true),
				RunAsNonRoot:           ptr.To(true),
				RunAsUser:              tt.wantRunAsUser,
				SeccompProfile: &corev1.SeccompProfile{
					Type: corev1.SeccompProfileTypeRuntimeDefault,
				},
			}))
		})
	}
})
