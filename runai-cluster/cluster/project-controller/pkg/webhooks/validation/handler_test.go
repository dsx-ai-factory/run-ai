/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validation_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo" //nolint:staticcheck // dot import for test framework is intentional
	. "github.com/onsi/gomega" //nolint:staticcheck // dot import for test framework is intentional

	admissionv1 "k8s.io/api/admission/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/webhooks/validation"
)

func TestValidationWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	test.RunTest(t, "Validation Webhook Tests", "../../../bin/test/results/validation_webhook_test_results.xml")
}

const (
	existingNodePoolA = "np-a"
	existingNodePoolB = "np-b"
	missingNodePool   = "np-missing"
	existingParent    = "dep-1"
	missingParent     = "dep-missing"
)

var _ = Describe("Validation webhook Handle", func() {
	var (
		ctx       context.Context
		validator *validation.Validator
	)

	BeforeEach(func() {
		ctx = context.Background()

		scheme := runtime.NewScheme()
		Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
		Expect(kaiv1alpha1.AddToScheme(scheme)).To(Succeed())

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			&kaiv1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: existingNodePoolA}},
			&kaiv1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: existingNodePoolB}},
			&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: existingParent}},
		).Build()

		validator = validation.NewValidator(fakeClient)
	})

	queue := func(name, nodepool string) kaiv1alpha1.QueueConfig {
		return kaiv1alpha1.QueueConfig{Name: name, Nodepool: nodepool}
	}

	handleProject := func(spec kaiv1alpha1.ProjectSpec) admission.Response {
		project := &kaiv1alpha1.Project{
			ObjectMeta: metav1.ObjectMeta{Name: "proj-1"},
			Spec:       spec,
		}
		raw, err := json.Marshal(project)
		Expect(err).NotTo(HaveOccurred())
		return validator.Handle(ctx, admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
			Kind: metav1.GroupVersionKind{
				Group:   kaiv1alpha1.GroupVersion.Group,
				Version: kaiv1alpha1.GroupVersion.Version,
				Kind:    "Project",
			},
			Object: runtime.RawExtension{Raw: raw},
		}})
	}

	handleDepartment := func(spec kaiv1alpha1.DepartmentSpec) admission.Response {
		department := &kaiv1alpha1.Department{
			ObjectMeta: metav1.ObjectMeta{Name: "dep-x"},
			Spec:       spec,
		}
		raw, err := json.Marshal(department)
		Expect(err).NotTo(HaveOccurred())
		return validator.Handle(ctx, admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
			Kind: metav1.GroupVersionKind{
				Group:   kaiv1alpha1.GroupVersion.Group,
				Version: kaiv1alpha1.GroupVersion.Version,
				Kind:    "Department",
			},
			Object: runtime.RawExtension{Raw: raw},
		}})
	}

	expectDenied := func(resp admission.Response, substr string) {
		Expect(resp.Allowed).To(BeFalse())
		Expect(resp.Result).NotTo(BeNil())
		Expect(resp.Result.Message).To(ContainSubstring(substr))
	}

	Context("Project", func() {
		It("allows an empty parent with no queues and no default node pools", func() {
			resp := handleProject(kaiv1alpha1.ProjectSpec{})
			Expect(resp.Allowed).To(BeTrue())
		})

		It("allows an existing parent and queues referencing existing node pools", func() {
			resp := handleProject(kaiv1alpha1.ProjectSpec{
				Parent: existingParent,
				Queues: []kaiv1alpha1.QueueConfig{queue("q-a", existingNodePoolA)},
			})
			Expect(resp.Allowed).To(BeTrue())
		})

		It("denies a non-empty parent that does not exist", func() {
			resp := handleProject(kaiv1alpha1.ProjectSpec{Parent: missingParent})
			expectDenied(resp, missingParent)
		})

		It("denies a queue that references a non-existing node pool", func() {
			resp := handleProject(kaiv1alpha1.ProjectSpec{
				Queues: []kaiv1alpha1.QueueConfig{queue("q-missing", missingNodePool)},
			})
			expectDenied(resp, missingNodePool)
		})

		It("denies a queue with an empty node pool", func() {
			resp := handleProject(kaiv1alpha1.ProjectSpec{
				Queues: []kaiv1alpha1.QueueConfig{queue("q-empty", "")},
			})
			expectDenied(resp, "node pool must not be empty")
		})

		It("denies the same node pool being referenced by more than one queue", func() {
			resp := handleProject(kaiv1alpha1.ProjectSpec{
				Queues: []kaiv1alpha1.QueueConfig{
					queue("q-a", existingNodePoolA),
					queue("q-a-dup", existingNodePoolA),
				},
			})
			expectDenied(resp, "is already referenced by another queue")
		})

		It("denies a default node pool that has no queue in the spec", func() {
			resp := handleProject(kaiv1alpha1.ProjectSpec{
				Queues:           []kaiv1alpha1.QueueConfig{queue("q-a", existingNodePoolA)},
				DefaultNodePools: []string{existingNodePoolB},
			})
			expectDenied(resp, existingNodePoolB)
		})

		It("allows a default node pool that is backed by a queue", func() {
			resp := handleProject(kaiv1alpha1.ProjectSpec{
				Queues:           []kaiv1alpha1.QueueConfig{queue("q-a", existingNodePoolA)},
				DefaultNodePools: []string{existingNodePoolA},
			})
			Expect(resp.Allowed).To(BeTrue())
		})

		It("errors when the project object cannot be unmarshaled", func() {
			resp := validator.Handle(ctx, admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Group:   kaiv1alpha1.GroupVersion.Group,
					Version: kaiv1alpha1.GroupVersion.Version,
					Kind:    "Project",
				},
				Object: runtime.RawExtension{Raw: []byte("not-json")},
			}})
			Expect(resp.Allowed).To(BeFalse())
			Expect(resp.Result.Message).To(ContainSubstring("failed to unmarshal"))
		})
	})

	Context("Department", func() {
		It("allows empty queues", func() {
			resp := handleDepartment(kaiv1alpha1.DepartmentSpec{})
			Expect(resp.Allowed).To(BeTrue())
		})

		It("allows queues referencing existing node pools", func() {
			resp := handleDepartment(kaiv1alpha1.DepartmentSpec{
				Queues: []kaiv1alpha1.QueueConfig{queue("q-a", existingNodePoolA)},
			})
			Expect(resp.Allowed).To(BeTrue())
		})

		It("denies a queue that references a non-existing node pool", func() {
			resp := handleDepartment(kaiv1alpha1.DepartmentSpec{
				Queues: []kaiv1alpha1.QueueConfig{queue("q-missing", missingNodePool)},
			})
			expectDenied(resp, missingNodePool)
		})

		It("denies a queue with an empty node pool", func() {
			resp := handleDepartment(kaiv1alpha1.DepartmentSpec{
				Queues: []kaiv1alpha1.QueueConfig{queue("q-empty", "")},
			})
			expectDenied(resp, "node pool must not be empty")
		})

		It("denies the same node pool being referenced by more than one queue", func() {
			resp := handleDepartment(kaiv1alpha1.DepartmentSpec{
				Queues: []kaiv1alpha1.QueueConfig{
					queue("q-a", existingNodePoolA),
					queue("q-a-dup", existingNodePoolA),
				},
			})
			expectDenied(resp, "is already referenced by another queue")
		})

		It("errors when the department object cannot be unmarshaled", func() {
			resp := validator.Handle(ctx, admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Group:   kaiv1alpha1.GroupVersion.Group,
					Version: kaiv1alpha1.GroupVersion.Version,
					Kind:    "Department",
				},
				Object: runtime.RawExtension{Raw: []byte("not-json")},
			}})
			Expect(resp.Allowed).To(BeFalse())
			Expect(resp.Result.Message).To(ContainSubstring("failed to unmarshal"))
		})
	})

	It("denies an unsupported resource kind", func() {
		resp := validator.Handle(ctx, admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
			Kind: metav1.GroupVersionKind{
				Group:   kaiv1alpha1.GroupVersion.Group,
				Version: kaiv1alpha1.GroupVersion.Version,
				Kind:    "Unsupported",
			},
			Object: runtime.RawExtension{Raw: []byte("{}")},
		}})
		Expect(resp.Allowed).To(BeFalse())
		Expect(resp.Result).NotTo(BeNil())
	})

	Context("backend failures", func() {
		It("returns an internal server error when listing node pools fails", func() {
			scheme := runtime.NewScheme()
			Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
			Expect(kaiv1alpha1.AddToScheme(scheme)).To(Succeed())

			failingClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
				List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
					return apierrors.NewServiceUnavailable("api server is down")
				},
			}).Build()
			failingValidator := validation.NewValidator(failingClient)

			project := &kaiv1alpha1.Project{
				ObjectMeta: metav1.ObjectMeta{Name: "proj-1"},
				Spec:       kaiv1alpha1.ProjectSpec{Queues: []kaiv1alpha1.QueueConfig{queue("q-a", existingNodePoolA)}},
			}
			raw, err := json.Marshal(project)
			Expect(err).NotTo(HaveOccurred())

			resp := failingValidator.Handle(ctx, admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Group:   kaiv1alpha1.GroupVersion.Group,
					Version: kaiv1alpha1.GroupVersion.Version,
					Kind:    "Project",
				},
				Object: runtime.RawExtension{Raw: raw},
			}})

			Expect(resp.Allowed).To(BeFalse())
			Expect(resp.Result).NotTo(BeNil())
			Expect(resp.Result.Code).To(Equal(int32(http.StatusInternalServerError)))
		})
	})
})
