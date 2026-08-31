/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initca

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/common/cert-utils/pkg/certificates"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

type signerCall struct {
	serviceName string
	secretName  string
	namespace   string
}

type recordingSigner struct {
	calls  []signerCall
	onSign func()
}

func (s *recordingSigner) Sign(_ context.Context, serviceName, secretName, namespace string, _ bool) (*certificates.CertificateDetails, error) {
	if s.onSign != nil {
		s.onSign()
	}
	s.calls = append(s.calls, signerCall{serviceName: serviceName, secretName: secretName, namespace: namespace})
	return certificates.NewCertificateDetails([]byte("ca-bundle")), nil
}

func TestInitCA(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Init CA Suite")
}

var _ = Describe("Legacy controller certificates", func() {
	It("signs and keeps the Secret when the controller is enabled", func() {
		client := fake.NewSimpleClientset(
			&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: workloadControllerSecretName, Namespace: namespace}},
		)
		signer := &recordingSigner{}

		details, err := reconcileLegacyControllerCertificate(
			context.Background(), client, signer, true,
			workloadControllerServiceName, workloadControllerSecretName,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(details).NotTo(BeNil())
		Expect(details.CaBundle).To(Equal([]byte("ca-bundle")))
		Expect(signer.calls).To(HaveLen(1))
		_, err = client.CoreV1().Secrets(namespace).Get(context.Background(), workloadControllerSecretName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		for _, action := range client.Actions() {
			Expect(action.GetResource().Resource).NotTo(BeElementOf("deployments", "pods"))
		}
	})

	It("deletes the Secret without signing or restarting when the controller is disabled", func() {
		client := fake.NewSimpleClientset(
			&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: workloadControllerSecretName, Namespace: namespace}},
		)
		signer := &recordingSigner{}

		details, err := reconcileLegacyControllerCertificate(
			context.Background(), client, signer, false,
			workloadControllerServiceName, workloadControllerSecretName,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(details).To(BeNil())
		Expect(signer.calls).To(BeEmpty())
		_, err = client.CoreV1().Secrets(namespace).Get(context.Background(), workloadControllerSecretName, metav1.GetOptions{})
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
		for _, action := range client.Actions() {
			Expect(action.GetResource().Resource).NotTo(BeElementOf("deployments", "pods"))
		}
	})

	It("accepts an already missing Secret", func() {
		client := fake.NewSimpleClientset()
		_, err := reconcileLegacyControllerCertificate(
			context.Background(), client, &recordingSigner{}, false,
			inferenceWorkloadControllerServiceName, inferenceWorkloadControllerSecretName,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns Secret deletion errors", func() {
		client := fake.NewSimpleClientset(
			&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: inferenceWorkloadControllerSecretName, Namespace: namespace}},
		)
		wantErr := errors.New("delete failed")
		client.PrependReactor("delete", "secrets", func(ktesting.Action) (bool, runtime.Object, error) {
			return true, nil, wantErr
		})

		_, err := reconcileLegacyControllerCertificate(
			context.Background(), client, &recordingSigner{}, false,
			inferenceWorkloadControllerServiceName, inferenceWorkloadControllerSecretName,
		)
		Expect(err).To(MatchError(wantErr))
	})

	It("signs all certificates before restarting deployments", func() {
		client := fake.NewSimpleClientset()
		signer := &recordingSigner{}
		signer.onSign = func() {
			for _, action := range client.Actions() {
				Expect(action.GetResource().Resource).NotTo(BeElementOf("deployments", "pods"))
			}
		}

		_, err := signCertificates(context.Background(), client, signer, Options{
			WorkloadControllerEnabled:          true,
			InferenceWorkloadControllerEnabled: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(signer.calls).To(HaveLen(9))
	})

	for _, tc := range []struct {
		name                   string
		options                Options
		deletedSecret          string
		retainedSecret         string
		unexpectedSignedSecret string
	}{
		{
			name:                   "reconciles a disabled workload controller independently",
			options:                Options{InferenceWorkloadControllerEnabled: true},
			deletedSecret:          workloadControllerSecretName,
			retainedSecret:         inferenceWorkloadControllerSecretName,
			unexpectedSignedSecret: workloadControllerSecretName,
		},
		{
			name:                   "reconciles a disabled inference controller independently",
			options:                Options{WorkloadControllerEnabled: true},
			deletedSecret:          inferenceWorkloadControllerSecretName,
			retainedSecret:         workloadControllerSecretName,
			unexpectedSignedSecret: inferenceWorkloadControllerSecretName,
		},
	} {
		tc := tc
		It(tc.name, func() {
			client := fake.NewSimpleClientset(
				&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: workloadControllerSecretName, Namespace: namespace}},
				&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: inferenceWorkloadControllerSecretName, Namespace: namespace}},
			)
			signer := &recordingSigner{}

			_, err := signCertificates(context.Background(), client, signer, tc.options)
			Expect(err).NotTo(HaveOccurred())
			_, err = client.CoreV1().Secrets(namespace).Get(context.Background(), tc.deletedSecret, metav1.GetOptions{})
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
			_, err = client.CoreV1().Secrets(namespace).Get(context.Background(), tc.retainedSecret, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			for _, call := range signer.calls {
				Expect(call.secretName).NotTo(Equal(tc.unexpectedSignedSecret))
			}
		})
	}
})
