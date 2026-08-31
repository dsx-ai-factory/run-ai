/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package certificates_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"

	"github.com/run-ai/runai/runai-cluster/common/cert-utils/pkg/certificates"
	"k8s.io/apimachinery/pkg/api/errors"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmv1fake "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned/fake"
	cmv1 "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("DeleteOldCertificates", func() {
	var (
		ctx        context.Context
		cmv1Client cmv1.CertmanagerV1Interface
		namespace  string
	)

	BeforeEach(func() {
		ctx = context.Background()
		cmv1Client = cmv1fake.NewSimpleClientset().CertmanagerV1()
		namespace = "test-ns"
	})

	It("should return nil when no old certificates found", func() {
		err := certificates.DeleteOldCertificates(ctx, cmv1Client, namespace)
		Expect(err).To(BeNil())
	})

	It("should delete all old certificates", func() {
		oldCrt1 := &certmanagerv1.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-crt-1",
				Namespace: namespace,
				Labels: map[string]string{
					"app": "runai",
				},
			},
		}
		oldCrt2 := &certmanagerv1.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-crt-2",
				Namespace: namespace,
				Labels: map[string]string{
					"app": "runai",
				},
			},
		}

		_, err := cmv1Client.Certificates(namespace).Create(ctx, oldCrt1, metav1.CreateOptions{})
		Expect(err).To(BeNil())

		err = certificates.DeleteOldCertificates(ctx, cmv1Client, namespace)
		Expect(err).To(BeNil())

		_, err = cmv1Client.Certificates(namespace).Get(ctx, oldCrt1.Name, metav1.GetOptions{})
		Expect(err).To(Equal(errors.NewNotFound(certmanagerv1.Resource("certificates"), oldCrt1.Name)))

		_, err = cmv1Client.Certificates(namespace).Get(ctx, oldCrt2.Name, metav1.GetOptions{})
		Expect(err).To(Equal(errors.NewNotFound(certmanagerv1.Resource("certificates"), oldCrt2.Name)))
	})
})
