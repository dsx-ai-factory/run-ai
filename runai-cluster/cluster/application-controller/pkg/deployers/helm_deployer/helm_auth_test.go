/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helm_deployer

import (
	"context"

	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Helm auth", func() {
	It("should map NGC API key secret to Helm basic auth", func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

		k8sClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ngc-secret",
					Namespace: "test-namespace",
				},
				Data: map[string][]byte{
					NgcAPIKey: []byte("ngc-api-key"),
				},
			}).
			Build()

		authOpts, err := getAuthOptions(context.Background(), k8sClient, "test-namespace", &runv1alpha1.HelmAuthSpec{
			SecretRef: &runv1alpha1.SecretReference{
				Name: "ngc-secret",
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(authOpts.Username).To(Equal(NgcBasicAuthUser))
		Expect(authOpts.Password).To(Equal("ngc-api-key"))
	})
})
