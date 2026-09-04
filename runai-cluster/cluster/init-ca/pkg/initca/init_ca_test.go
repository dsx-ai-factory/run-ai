/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package initca

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cert-utils/pkg/certificates"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/fake"
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

var _ = Describe("Controller certificates", func() {
	It("signs all certificates before restarting deployments", func() {
		client := fake.NewSimpleClientset()
		signer := &recordingSigner{}
		signer.onSign = func() {
			for _, action := range client.Actions() {
				Expect(action.GetResource().Resource).NotTo(BeElementOf("deployments", "pods"))
			}
		}

		_, err := signCertificates(context.Background(), client, signer)
		Expect(err).NotTo(HaveOccurred())
		Expect(signer.calls).To(HaveLen(7))
	})
})
