/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	inferenceWorkloadControllerUID    = "inference-workload-controller-uid"
	nonInferenceWorkloadControllerUID = "non-inference-workload-controller-uid"
)

var _ = Describe("Migration WebhookHandler test", func() {
	var (
		webhook    *WebhookHandler
		fakeClient client.Client
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		register.InitSchema(scheme)

		workloadControllerServiceAccount := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.InferenceWorkloadControllerOperandName,
				Namespace: constants.NamespaceRunAi,
				UID:       types.UID(inferenceWorkloadControllerUID),
			},
		}
		fakeClient = client2.InitFakeClientWithObjectNoStatus(scheme, workloadControllerServiceAccount)

	})

	Context("Handle", func() {
		Context("when inference workload controller uid is cached", func() {
			BeforeEach(func() {
				webhook, _ = NewWebhookHandler(fakeClient, inferenceWorkloadControllerUID)
			})
			It("should deny non-update operations", func() {
				req := admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Operation: admissionv1.Create,
					},
				}
				response := webhook.Handle(context.Background(), req)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("migration in progress"))
			})

			It("should allow inference workload controller request", func() {
				req := admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Operation: admissionv1.Update,
						UserInfo: authenticationv1.UserInfo{
							UID: inferenceWorkloadControllerUID,
						},
					},
				}
				response := webhook.Handle(context.Background(), req)
				Expect(response.Allowed).To(BeTrue())
				Expect(response.Result.Message).To(Equal("inference workload controller request"))
			})

			It("should deny non-inference workload controller request", func() {
				req := admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Operation: admissionv1.Update,
						UserInfo: authenticationv1.UserInfo{
							UID: nonInferenceWorkloadControllerUID,
						},
					},
				}
				response := webhook.Handle(context.Background(), req)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("migration in progress"))
			})
			It("should deny request with empty uid", func() {
				req := admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Operation: admissionv1.Update,
						UserInfo:  authenticationv1.UserInfo{},
					},
				}
				response := webhook.Handle(context.Background(), req)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("migration in progress"))
			})
		})
		Context("when inference workload controller uid is not cached", func() {
			BeforeEach(func() {
				webhook, _ = NewWebhookHandler(fakeClient, "")
			})
			It("should deny create operations", func() {
				req := admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Operation: admissionv1.Create,
					},
				}
				response := webhook.Handle(context.Background(), req)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("migration in progress"))
			})
			It("should deny delete operations", func() {
				req := admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Operation: admissionv1.Delete,
					},
				}
				response := webhook.Handle(context.Background(), req)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("migration in progress"))
			})

			It("should allow inference workload controller request", func() {
				req := admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Operation: admissionv1.Update,
						UserInfo: authenticationv1.UserInfo{
							UID: inferenceWorkloadControllerUID,
						},
					},
				}
				response := webhook.Handle(context.Background(), req)
				Expect(response.Allowed).To(BeTrue())
				Expect(response.Result.Message).To(Equal("inference workload controller request"))
			})

			It("should deny non-inference workload controller request", func() {
				req := admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Operation: admissionv1.Update,
						UserInfo: authenticationv1.UserInfo{
							UID: nonInferenceWorkloadControllerUID,
						},
					},
				}
				response := webhook.Handle(context.Background(), req)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("migration in progress"))
			})
			It("should deny request with empty uid", func() {
				req := admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Operation: admissionv1.Update,
						UserInfo:  authenticationv1.UserInfo{},
					},
				}
				response := webhook.Handle(context.Background(), req)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("migration in progress"))
			})
		})
	})
})
