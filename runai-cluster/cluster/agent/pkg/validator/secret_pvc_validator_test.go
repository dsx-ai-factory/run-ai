/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"net/http"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"

	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	ProjectControllerUsername = "system:serviceaccount:runai:runai-project-controller"
)

var _ = Describe("Test Secret And PVC Protection Webhook", Ordered, func() {
	var (
		k8sClient                       client.Client
		projectControllerServiceAccount *corev1.ServiceAccount
		objects                         []runtime.Object

		reqProjectControl admission.Request
		reqOther          admission.Request

		operations         = []admissionv1.Operation{admissionv1.Create, admissionv1.Update, admissionv1.Delete}
		secretPvcValidator *SecretPvcValidator
		serviceCounter     = 0
	)

	BeforeAll(func() {
		projectControllerServiceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: runaiv1.DefaultNamespace,
				Name:      "runai-project-controller",
			},
		}
	})

	BeforeEach(func() {
		reqProjectControl = admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Name: "runai-project-controller-" + strconv.Itoa(serviceCounter),
				UserInfo: authenticationv1.UserInfo{
					Username: ProjectControllerUsername,
				},
			},
		}

		reqOther = admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Name: "runai-project-controller-" + strconv.Itoa(serviceCounter),
				UserInfo: authenticationv1.UserInfo{
					Username: "master:of:orion",
				},
			},
		}

		serviceCounter++
		objects = []runtime.Object{projectControllerServiceAccount}
	})

	Context("Test SecretPvcValidator with service account", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			secretPvcValidator = NewSecretPvcValidator(k8sClient, []string{})
		})

		It("Requester is runai-project-controller - allow", func() {
			for _, operation := range operations {
				reqProjectControl.Operation = operation
				res := secretPvcValidator.Handle(context.Background(), reqProjectControl)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester is not runai-agent or accessrule-controller - don't allow", func() {
			for _, operation := range operations {
				reqOther.Operation = operation
				res := secretPvcValidator.Handle(context.Background(), reqOther)

				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(res.Result.Reason).To(Not(BeEmpty()))
				Expect(string(res.Result.Message)).To(ContainSubstring("username master:of:orion is not permitted to do any action on  with the run.ai/asset label"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			}
		})
	})

	Context("Test SecretPvcValidator with allowedServiceAccounts", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			secretPvcValidator = NewSecretPvcValidator(k8sClient, []string{"test-service", "kube-system"})
		})

		It("Requester username contains allowed pattern - allow", func() {
			reqOther.UserInfo.Username = "serviceaccount:test-namespace:test-service-account"

			for _, operation := range operations {
				reqOther.Operation = operation
				res := secretPvcValidator.Handle(context.Background(), reqOther)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester username contains second allowed pattern - allow", func() {
			reqOther.UserInfo.Username = "serviceaccount:kube-system:coredns"

			for _, operation := range operations {
				reqOther.Operation = operation
				res := secretPvcValidator.Handle(context.Background(), reqOther)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester username doesn't contain any allowed pattern - deny", func() {
			reqOther.UserInfo.Username = "serviceaccount:default:unauthorized-service"

			for _, operation := range operations {
				reqOther.Operation = operation
				res := secretPvcValidator.Handle(context.Background(), reqOther)

				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(res.Result.Reason).To(Not(BeEmpty()))
				Expect(string(res.Result.Message)).To(ContainSubstring("username serviceaccount:default:unauthorized-service is not permitted to do any action on  with the run.ai/asset label"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			}
		})
	})

	Context("Test SecretPvcValidator with empty allowedServiceAccounts", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			secretPvcValidator = NewSecretPvcValidator(k8sClient, []string{})
		})

		It("Should fall back to system:serviceaccount check when allowedServiceAccounts is empty", func() {
			reqOther.UserInfo.Username = "system:serviceaccount:test:test-service"

			for _, operation := range operations {
				reqOther.Operation = operation
				res := secretPvcValidator.Handle(context.Background(), reqOther)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})
	})
})
