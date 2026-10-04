/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("cluster-api resources utils", func() {
	Context("getTlsSecretCertAndKey", func() {
		It("should return tls secret cert and key", func() {
			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret",
					Namespace: "namespace",
				},
				Data: map[string][]byte{
					"tls.crt": []byte("cert"),
					"tls.key": []byte("key"),
				},
			}
			k8sReader := fake.NewClientBuilder().WithObjects(secret).Build()

			cert, key, err := GetTlsSecretCertAndKey(context.Background(), k8sReader, "secret", "namespace")
			Expect(err).NotTo(HaveOccurred())
			Expect(cert).To(Equal("cert"))
			Expect(key).To(Equal("key"))
		})

		It("should return error if secret not found", func() {
			k8sReader := fake.NewClientBuilder().Build()
			cert, key, err := GetTlsSecretCertAndKey(context.Background(), k8sReader, "secret", "namespace")
			Expect(err).To(HaveOccurred())
			Expect(cert).To(BeEmpty())
			Expect(key).To(BeEmpty())
		})
	})
})
