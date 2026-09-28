/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helm_deployer

import (
	"context"
	"os"
	"path/filepath"

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

	It("should write TLS secret files to one private dir and remove it on Cleanup", func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

		k8sClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-secret",
					Namespace: "test-namespace",
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{
					corev1.TLSCertKey:       []byte("cert"),
					corev1.TLSPrivateKeyKey: []byte("key"),
					CaCertFile:              []byte("ca"),
				},
			}).
			Build()

		authOpts, err := getAuthOptions(context.Background(), k8sClient, "test-namespace", &runv1alpha1.HelmAuthSpec{
			SecretRef: &runv1alpha1.SecretReference{
				Name: "tls-secret",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		tempDir := filepath.Dir(authOpts.CertFile)
		Expect(filepath.Base(tempDir)).To(HavePrefix(TempAuthDir + "-"))
		Expect(filepath.Dir(authOpts.KeyFile)).To(Equal(tempDir))
		Expect(filepath.Dir(authOpts.CAFile)).To(Equal(tempDir))

		dirInfo, err := os.Stat(tempDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(dirInfo.Mode().Perm()).To(Equal(os.FileMode(0o700)))
		Expect(os.ReadFile(authOpts.KeyFile)).To(Equal([]byte("key")))

		Expect(authOpts.Cleanup()).To(Succeed())
		Expect(tempDir).NotTo(BeADirectory())
		Expect(authOpts.Cleanup()).To(Succeed())
	})

	It("should not remove a CA file that comes from the auth spec", func() {
		caFile := filepath.Join(GinkgoT().TempDir(), "ca.crt")
		Expect(os.WriteFile(caFile, []byte("ca"), 0o600)).To(Succeed())

		authOpts, err := getAuthOptions(context.Background(), nil, "test-namespace", &runv1alpha1.HelmAuthSpec{
			CAFile: caFile,
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(authOpts.Cleanup()).To(Succeed())
		Expect(caFile).To(BeAnExistingFile())
	})

	It("should remove auth files when a dropped install task is cleaned up", func() {
		authOpts := &AuthOptions{}
		certFile, err := authOpts.writeTempFile(PrefixTlsCert, []byte("cert"))
		Expect(err).NotTo(HaveOccurred())

		var task Task = &TaskInstall{InstallParams: &InstallParams{authOpts: authOpts}}
		Expect(task.Cleanup()).To(Succeed())
		Expect(filepath.Dir(certFile)).NotTo(BeADirectory())
	})

	It("should allow Cleanup on nil auth options", func() {
		var authOpts *AuthOptions
		Expect(authOpts.Cleanup()).To(Succeed())
	})
})
