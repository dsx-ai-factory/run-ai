/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package issue

import (
	"fmt"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftV1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	workloadName = "workload-deployment"
	workloadKind = "Deployment"
)

var (
	workload = &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind: workloadKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: workloadName,
		},
	}
)

// SomeJobResourceGrouper is a custom ResourceGrouper that handles special job types
type SomeJobResourceGrouper struct {
	DefaultResourceGrouper
}

func (rg *SomeJobResourceGrouper) GetObjectResource(resource client.Object) (metav1.GroupVersionKind, error) {
	if resource == nil {
		return metav1.GroupVersionKind{}, fmt.Errorf("object is nil %v", resource)
	}

	var objectResource metav1.GroupVersionKind
	switch resource.(type) {
	case *appsv1.Deployment:
		objectResource.Kind = "Deployment"
	default:
		return rg.DefaultResourceGrouper.GetObjectResource(resource)
	}

	return objectResource, nil
}

func (rg *SomeJobResourceGrouper) GetIssueGroupByGvk(gvk metav1.GroupVersionKind) Group {
	switch gvk.Kind {
	case "Deployment":
		return Compute
	default:
		return rg.DefaultResourceGrouper.GetIssueGroupByGvk(gvk)
	}
}

func TestIssue(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Issue tests")
}

var _ = Describe("Issue Test Suite", func() {
	Context("Issue Struct", func() {
		Context("Internal Issue", func() {
			It("Internal no error", func() {
				pvc := corev1.PersistentVolumeClaim{}
				iss := New(&pvc, nil, false, "", nil)
				Expect(iss.IsCritical()).To(Equal(false))
				Expect(iss.Message()).To(Equal("[PersistentVolumeClaim:] "))
			})

			It("Internal no error with description", func() {
				pvc := corev1.PersistentVolumeClaim{}
				iss := New(&pvc, nil, false, "some desc", nil)
				Expect(iss.Message()).To(Equal("[PersistentVolumeClaim:] some desc"))
			})

			It("Internal Recoverable Issue With Internal Error and no description", func() {
				inErr := NewInternalError(fmt.Errorf("internal"))
				iss := New(workload, inErr, false, "", nil)
				Expect(iss.IsCritical()).To(Equal(false))
				Expect(iss.Message()).To(Equal(fmt.Sprintf("[%s:%s] %s", workloadKind, workloadName, "internal")))
				Expect(iss.GetGroup()).To(Equal(Internal))
			})

			It("Internal critical Issue With fmt Error and description", func() {
				fmtErr := fmt.Errorf("just an error")
				iss := New(workload, fmtErr, true, "my description", nil)
				Expect(iss.IsCritical()).To(Equal(true))
				Expect(iss.Message()).To(Equal(fmt.Sprintf("[%s:%s] %s: %s", workloadKind, workloadName, "my description", "just an error")))
				Expect(iss.GetGroup()).To(Equal(Internal))
			})
		})

		Context("Resource Issue", func() {
			It("Resource no error", func() {
				pvc := corev1.PersistentVolumeClaim{}
				iss := New(&pvc, nil, false, "", nil)
				Expect(iss.IsCritical()).To(Equal(false))
				Expect(iss.Message()).To(Equal("[PersistentVolumeClaim:] "))
			})

			It("Resource empty client", func() {
				binding := corev1.Binding{}
				iss := New(&binding, fmt.Errorf("err"), false, "", nil)
				Expect(iss.Name()).To(Equal(""))
				Expect(iss.Kind()).To(Equal(""))
				Expect(iss.IsCritical()).To(Equal(false))
				Expect(iss.Message()).To(Equal("[:] err"))
				Expect(iss.GetGroup()).To(Equal(Internal))
			})

			It("Pvc Recoverable Issue With Internal Error and no description", func() {
				pvc := corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvc01",
					},
				}
				inErr := NewInternalError(fmt.Errorf("internal"))
				iss := New(&pvc, inErr, false, "", nil)
				Expect(iss.Name()).To(Equal("pvc01"))
				Expect(iss.Kind()).To(Equal(constants.KindPvc))
				Expect(iss.IsCritical()).To(Equal(false))
				Expect(iss.Message()).To(Equal("[PersistentVolumeClaim:pvc01] internal"))
				Expect(iss.GetGroup()).To(Equal(Storage))
			})

			It("Ingress critical Issue With fmt Error and description", func() {
				ingress := networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ingress01",
					},
				}
				fmtErr := fmt.Errorf("just an error")
				iss := New(&ingress, fmtErr, true, "my description", nil)
				Expect(iss.Name()).To(Equal("ingress01"))
				Expect(iss.Kind()).To(Equal(constants.KindIngress))
				Expect(iss.IsCritical()).To(Equal(true))
				Expect(iss.Message()).To(Equal("[Ingress:ingress01] my description: just an error"))
				Expect(iss.GetGroup()).To(Equal(Network))
			})

			It("Service Recoverable Issue With Validation Error", func() {
				service := corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "service01",
					},
				}
				valErr := NewError("code", "msg")
				iss := New(&service, valErr, false, "", nil)
				Expect(iss.Name()).To(Equal("service01"))
				Expect(iss.Kind()).To(Equal(constants.KindService))
				Expect(iss.IsCritical()).To(Equal(false))
				Expect(iss.Message()).To(Equal("[Service:service01] msg"))
				Expect(iss.GetGroup()).To(Equal(Network))
			})
		})

		Context("Issue List", func() {
			It("Empty List", func() {
				issues := issueEmptyList()
				Expect(issues).To(BeNil())
			})

			It("Mixed List", func() {
				issues := issueMixedList()
				Expect(issues).ToNot(BeNil())
				Expect(len(issues)).To(Equal(2))
				Expect(issues[0].IsCritical()).To(Equal(false))
				Expect(issues[0].Message()).To(Equal(fmt.Sprintf("[%s:%s] core desc: core", workloadKind, workloadName)))
				Expect(issues[0].Error()).To(Equal("core"))
				Expect(issues[0].GetGroup()).To(Equal(Internal))
				Expect(issues[1].IsCritical()).To(Equal(true))
				Expect(issues[1].Message()).To(Equal("[Service:res01] res desc: res"))
				Expect(issues[1].Error()).To(Equal("res"))
				Expect(issues[1].GetGroup()).To(Equal(Network))
			})
		})
	})

	Context("Testing New function with different ResourceGrouper implementations", func() {
		It("should create a new issue with DefaultResourceGrouper", func() {
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "service01",
				},
			}

			iss := New(service, fmt.Errorf("test error"), false, "test description", &DefaultResourceGrouper{})

			Expect(iss.Name()).To(Equal("service01"))
			Expect(iss.Kind()).To(Equal(constants.KindService))
			Expect(iss.GetGroup()).To(Equal(Network))
			Expect(iss.IsCritical()).To(BeFalse())
			Expect(iss.Message()).To(ContainSubstring("test description: test error"))
		})

		It("should create a new issue with nil ResourceGrouper (should use default)", func() {
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "service02",
				},
			}

			iss := New(service, fmt.Errorf("test error"), true, "", nil)

			Expect(iss.Name()).To(Equal("service02"))
			Expect(iss.Kind()).To(Equal(constants.KindService))
			Expect(iss.GetGroup()).To(Equal(Network))
			Expect(iss.IsCritical()).To(BeTrue())
		})

		It("should create a new issue with a custom ResourceGrouper for custom job types", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "deployment-01",
				},
			}

			iss := New(deployment, fmt.Errorf("test error"), false, "deployment issue", &SomeJobResourceGrouper{})

			Expect(iss.Name()).To(Equal("deployment-01"))
			Expect(iss.Kind()).To(Equal("Deployment"))
			Expect(iss.GetGroup()).To(Equal(Compute))
		})
	})

	Context("Testing ResourceGrouper interface implementations", func() {
		It("should return correct GroupVersionKind using DefaultResourceGrouper", func() {
			rg := &DefaultResourceGrouper{}

			// Test with various resource types
			ingress := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ingress"},
			}
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "test-pvc"},
			}

			ingressGvk, err := rg.GetObjectResource(ingress)
			Expect(err).NotTo(HaveOccurred())
			Expect(ingressGvk.Kind).To(Equal(constants.KindIngress))

			pvcGvk, err := rg.GetObjectResource(pvc)
			Expect(err).NotTo(HaveOccurred())
			Expect(pvcGvk.Kind).To(Equal(constants.KindPvc))
		})

		It("should return correct Group using DefaultResourceGrouper", func() {
			rg := &DefaultResourceGrouper{}

			// Test with various resource types
			ingress := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ingress"},
			}
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "test-pvc"},
			}

			Expect(rg.GetIssueGroupByResource(ingress)).To(Equal(Network))
			Expect(rg.GetIssueGroupByResource(pvc)).To(Equal(Storage))
		})

		It("should return correct GVK and Group using DefaultResourceGrouper", func() {
			testCases := []struct {
				name          string
				resource      client.Object
				expectedGvk   *metav1.GroupVersionKind
				expectedGroup Group
			}{
				{
					name: "Network group for Ingress",
					resource: &networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{Name: "test-ingress"},
					},
					expectedGvk:   &metav1.GroupVersionKind{Kind: constants.KindIngress},
					expectedGroup: Network,
				},
			{
				name: "Network group for Route",
				resource: &openshiftV1.Route{
					ObjectMeta: metav1.ObjectMeta{Name: "test-route"},
				},
				expectedGvk:   &metav1.GroupVersionKind{Kind: constants.KindRoute},
				expectedGroup: Network,
			},
			{
				name: "Network group for HTTPRoute",
				resource: &gatewayv1.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{Name: "test-httproute"},
				},
				expectedGvk:   &metav1.GroupVersionKind{Kind: constants.KindHTTPRoute},
				expectedGroup: Network,
			},
			{
				name: "Storage group for PVC",
					resource: &corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{Name: "test-pvc"},
					},
					expectedGvk:   &metav1.GroupVersionKind{Kind: constants.KindPvc},
					expectedGroup: Storage,
				},
				{
					name: "Internal group for unknown kind",
					resource: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "test-pod"},
					},
					expectedGvk:   nil,
					expectedGroup: "",
				},
			}

			rg := &DefaultResourceGrouper{}

			for _, tc := range testCases {
				tc := tc // capture range variable
				By(fmt.Sprintf("Running test case: %s", tc.name))
				gotGvk, err := rg.GetObjectResource(tc.resource)
				if tc.expectedGvk == nil {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(gotGvk).To(Equal(*tc.expectedGvk))
				}
				gotGroup := rg.GetIssueGroupByResource(tc.resource)
				Expect(gotGroup).To(Equal(tc.expectedGroup))
			}
		})

		It("should return correct GroupVersionKind using a custom ResourceGrouper", func() {
			rg := &SomeJobResourceGrouper{}

			// Test with ML job resource types
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-deployment",
				},
			}

			// Regular resource should use default implementation
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "test-pvc"},
			}

			deploymentGvk, err := rg.GetObjectResource(deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploymentGvk.Kind).To(Equal("Deployment"))

			pvcGvk, err := rg.GetObjectResource(pvc)
			Expect(err).NotTo(HaveOccurred())
			Expect(pvcGvk.Kind).To(Equal(constants.KindPvc))
		})

		It("should return correct Group using a custom ResourceGrouper", func() {
			rg := &SomeJobResourceGrouper{}

			// Direct GVK testing
			deploymentGvk := metav1.GroupVersionKind{Kind: "Deployment"}
			ingressGvk := metav1.GroupVersionKind{Kind: constants.KindIngress}

			Expect(rg.GetIssueGroupByGvk(deploymentGvk)).To(Equal(Compute))
			Expect(rg.GetIssueGroupByGvk(ingressGvk)).To(Equal(Network))
		})
	})
})

func issueEmptyList() []Issue {
	return nil
}

func issueMixedList() []Issue {
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "res01",
		},
	}

	return []Issue{
		*New(workload, fmt.Errorf("core"), false, "core desc", nil),
		*New(&service, fmt.Errorf("res"), true, "res desc", nil),
	}
}

type (
	Error interface {
		Error() string
		Code() string
		Message() string
		SetMessage(string) Error
	}

	ErrorObject struct {
		code    string
		message string
	}

	Errors map[string]error

	InternalError interface {
		error
		InternalError() error
	}

	internalError struct {
		error
	}
)

func NewInternalError(err error) InternalError {
	return internalError{error: err}
}
func (e internalError) InternalError() error {
	return e.error
}

func NewError(code, message string) Error {
	return ErrorObject{
		code:    code,
		message: message,
	}
}

func (e ErrorObject) Error() string {
	return e.message
}

func (e ErrorObject) Code() string {
	return e.code
}

func (e ErrorObject) Message() string {
	return e.message
}

func (e ErrorObject) SetMessage(message string) Error {
	e.message = message
	return e
}
