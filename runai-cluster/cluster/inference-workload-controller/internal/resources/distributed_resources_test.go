/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

var _ = Describe("DistributedResourcesManager", func() {
	var (
		ctx              context.Context
		manager          *DistributedResourcesManager
		fakeClient       client.Client
		scheme           *runtime.Scheme
		testNamespace    = "test-namespace"
		testResourceName = "test-resource"
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()

		// Add necessary schemes
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(netv1.AddToScheme(scheme)).To(Succeed())
		Expect(routev1.AddToScheme(scheme)).To(Succeed())
		Expect(gatewayv1.Install(scheme)).To(Succeed())
		Expect(lwsv1.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		manager = NewDistributedResourcesManager(fakeClient)
	})

	Describe("NewDistributedResourcesManager", func() {
		It("should create a new instance with the provided client", func() {
			newManager := NewDistributedResourcesManager(fakeClient)
			Expect(newManager).NotTo(BeNil())
			Expect(newManager.k8sClient).To(Equal(fakeClient))
		})
	})

	Describe("GetResources", func() {
		Context("when resources list is empty", func() {
			It("should return all-nil result and empty missing resources", func() {
				result, missingResources, err := manager.GetResources(ctx, []client.Object{})

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Lws).To(BeNil())
				Expect(result.Services).To(BeNil())
				Expect(result.Ingresses).To(BeNil())
				Expect(result.Routes).To(BeNil())
				Expect(result.Pvcs).To(BeNil())
				Expect(missingResources).To(BeEmpty())
			})
		})

		Context("when all resources exist", func() {
			var (
				lws     *lwsv1.LeaderWorkerSet
				service *corev1.Service
				ingress *netv1.Ingress
				route   *routev1.Route
				pvc     *corev1.PersistentVolumeClaim
			)

			BeforeEach(func() {
				// Create test resources
				lws = &lwsv1.LeaderWorkerSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-lws",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindLeaderWorkerSet,
						APIVersion: lwsv1.GroupVersion.String(),
					},
				}

				service = &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-service",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindService,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
				}

				ingress = &netv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-ingress",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindIngress,
						APIVersion: netv1.SchemeGroupVersion.String(),
					},
				}

				route = &routev1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-route",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindRoute,
						APIVersion: routev1.GroupVersion.String(),
					},
				}

				pvc = &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-pvc",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindPvc,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
				}

				// Create resources in fake client
				Expect(createAndRestoreGVK(ctx, fakeClient, lws)).To(Succeed())
				Expect(createAndRestoreGVK(ctx, fakeClient, service)).To(Succeed())
				Expect(createAndRestoreGVK(ctx, fakeClient, ingress)).To(Succeed())
				Expect(createAndRestoreGVK(ctx, fakeClient, route)).To(Succeed())
				Expect(createAndRestoreGVK(ctx, fakeClient, pvc)).To(Succeed())
			})

			It("should return all resources successfully", func() {
				resources := []client.Object{lws, service, ingress, route, pvc}

				result, missingResources, err := manager.GetResources(ctx, resources)

				Expect(err).NotTo(HaveOccurred())
				Expect(missingResources).To(BeEmpty())

				// Verify LeaderWorkerSet
				Expect(result.Lws).NotTo(BeNil())
				Expect(result.Lws.Name).To(Equal(lws.Name))
				Expect(result.Lws.Namespace).To(Equal(lws.Namespace))

				// Verify Services
				Expect(result.Services).To(HaveLen(1))
				Expect(result.Services[0].Name).To(Equal(service.Name))
				Expect(result.Services[0].Namespace).To(Equal(service.Namespace))

				// Verify Ingresses
				Expect(result.Ingresses).To(HaveLen(1))
				Expect(result.Ingresses[0].Name).To(Equal(ingress.Name))
				Expect(result.Ingresses[0].Namespace).To(Equal(ingress.Namespace))

				// Verify Routes
				Expect(result.Routes).To(HaveLen(1))
				Expect(result.Routes[0].Name).To(Equal(route.Name))
				Expect(result.Routes[0].Namespace).To(Equal(route.Namespace))

				// Verify PVCs
				Expect(result.Pvcs).To(HaveLen(1))
				Expect(result.Pvcs[0].Name).To(Equal(pvc.Name))
				Expect(result.Pvcs[0].Namespace).To(Equal(pvc.Namespace))
			})

			It("should handle multiple services correctly", func() {
				service2 := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-service2",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindService,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
				}
				Expect(createAndRestoreGVK(ctx, fakeClient, service2)).To(Succeed())

				resources := []client.Object{service, service2}

				result, missingResources, err := manager.GetResources(ctx, resources)

				Expect(err).NotTo(HaveOccurred())
				Expect(missingResources).To(BeEmpty())
				Expect(result.Services).To(HaveLen(2))

				serviceNames := []string{result.Services[0].Name, result.Services[1].Name}
				Expect(serviceNames).To(ContainElements(service.Name, service2.Name))
			})

			It("should fetch HTTPRoute resources", func() {
				httpRoute := &gatewayv1.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-httproute",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindHTTPRoute,
						APIVersion: gatewayv1.GroupVersion.String(),
					},
				}
				Expect(createAndRestoreGVK(ctx, fakeClient, httpRoute)).To(Succeed())

				resources := []client.Object{httpRoute}
				result, missingResources, err := manager.GetResources(ctx, resources)

				Expect(err).NotTo(HaveOccurred())
				Expect(missingResources).To(BeEmpty())
				Expect(result.HTTPRoutes).To(HaveLen(1))
				Expect(result.HTTPRoutes[0].Name).To(Equal(httpRoute.Name))
				Expect(result.HTTPRoutes[0].Namespace).To(Equal(httpRoute.Namespace))
			})

			It("should handle missing HTTPRoute as missing resource", func() {
				missingHTTPRoute := &gatewayv1.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "missing-httproute",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindHTTPRoute,
						APIVersion: gatewayv1.GroupVersion.String(),
					},
				}

				resources := []client.Object{missingHTTPRoute}
				result, missingResources, err := manager.GetResources(ctx, resources)

				Expect(err).NotTo(HaveOccurred())
				Expect(missingResources).To(HaveLen(1))
				Expect(result.HTTPRoutes).To(BeNil())
			})
		})

		Context("when some resources are missing", func() {
			var existingService *corev1.Service

			BeforeEach(func() {
				existingService = &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-service",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindService,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
				}
				Expect(createAndRestoreGVK(ctx, fakeClient, existingService)).To(Succeed())
			})

			It("should return found resources and list missing resources", func() {
				missingService := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "missing-service",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindService,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
				}

				resources := []client.Object{existingService, missingService}

				result, missingResources, err := manager.GetResources(ctx, resources)

				Expect(err).NotTo(HaveOccurred())
				Expect(missingResources).To(HaveLen(1))
				Expect(missingResources[0]).To(Equal(missingService))
				Expect(result.Services).To(HaveLen(1))
				Expect(result.Services[0].Name).To(Equal(existingService.Name))
			})
		})

		Context("when encountering unsupported resource kinds", func() {
			It("should skip unsupported resources and continue processing", func() {
				service := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-service",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindService,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
				}
				Expect(createAndRestoreGVK(ctx, fakeClient, service)).To(Succeed())

				// Create a mock resource with unsupported kind
				unsupportedResource := &mockUnsupportedResource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unsupported",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind: "UnsupportedKind",
					},
				}

				resources := []client.Object{service, unsupportedResource}

				result, missingResources, err := manager.GetResources(ctx, resources)

				Expect(err).NotTo(HaveOccurred())
				Expect(missingResources).To(BeEmpty()) // Unsupported resources are skipped, not considered missing
				Expect(result.Services).To(HaveLen(1))
				Expect(result.Services[0].Name).To(Equal(service.Name))
			})
		})

		Context("when client returns errors", func() {
			It("should return error when client fails with non-NotFound error", func() {
				// Use a mock client that returns errors
				mockClient := &mockErrorClient{
					shouldReturnError: true,
					errorToReturn:     errors.New("client error"),
				}
				manager = NewDistributedResourcesManager(mockClient)

				service := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-service",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindService,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
				}

				resources := []client.Object{service}

				result, missingResources, err := manager.GetResources(ctx, resources)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("client error"))
				Expect(result.Services).To(BeNil())
				Expect(missingResources).To(BeEmpty())
			})
		})

		Context("when testing missing resources detection", func() {
			It("should correctly identify when no resources are missing", func() {
				service := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-service",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindService,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
				}
				Expect(createAndRestoreGVK(ctx, fakeClient, service)).To(Succeed())

				resources := []client.Object{service}
				result, missingResources, err := manager.GetResources(ctx, resources)

				Expect(err).NotTo(HaveOccurred())
				Expect(missingResources).To(BeEmpty())
				Expect(result.Services).To(HaveLen(1))
			})

			It("should correctly identify when LWS resources are missing", func() {
				// Request lws resource but don't create it
				lws := &lwsv1.LeaderWorkerSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-lws",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindLeaderWorkerSet,
						APIVersion: lwsv1.GroupVersion.String(),
					},
				}

				resources := []client.Object{lws}
				result, missingResources, err := manager.GetResources(ctx, resources)

				Expect(err).NotTo(HaveOccurred())
				Expect(missingResources).To(HaveLen(1))
				Expect(missingResources[0]).To(Equal(lws))
				Expect(result.Lws).To(BeNil())
			})

			It("should correctly identify when non-LWS resources are missing", func() {
				// Request 2 service resources but only create 1
				service1 := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-service1",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindService,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
				}
				service2 := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testResourceName + "-service2",
						Namespace: testNamespace,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       constants.KindService,
						APIVersion: corev1.SchemeGroupVersion.String(),
					},
				}

				// Only create service1
				Expect(createAndRestoreGVK(ctx, fakeClient, service1)).To(Succeed())

				resources := []client.Object{service1, service2}
				result, missingResources, err := manager.GetResources(ctx, resources)

				Expect(err).NotTo(HaveOccurred())
				Expect(missingResources).To(HaveLen(1))
				Expect(missingResources[0]).To(Equal(service2))
				Expect(result.Services).To(HaveLen(1))
			})
		})
	})
})

// createAndRestoreGVK creates obj via the fake client and restores its TypeMeta
// afterwards. The newer controller-runtime fake client strips TypeMeta on Create,
// but these tests reuse the same object in the GetResources input slice, where the
// production code dispatches on GroupVersionKind().Kind. Restoring the GVK keeps the
// reused object usable as test input without touching production behavior.
func createAndRestoreGVK(ctx context.Context, c client.Client, obj client.Object) error {
	gvk := obj.GetObjectKind().GroupVersionKind()
	if err := c.Create(ctx, obj); err != nil {
		return err
	}
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return nil
}

// Mock implementations for testing

type mockUnsupportedResource struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (m *mockUnsupportedResource) DeepCopyObject() runtime.Object {
	return &mockUnsupportedResource{
		TypeMeta:   m.TypeMeta,
		ObjectMeta: *m.ObjectMeta.DeepCopy(),
	}
}

type mockErrorClient struct {
	shouldReturnError bool
	errorToReturn     error
}

func (m *mockErrorClient) Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	return apierrors.NewNotFound(schema.GroupResource{}, key.Name)
}

func (m *mockErrorClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return fmt.Errorf("not implemented")
}

func (m *mockErrorClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return fmt.Errorf("not implemented")
}

func (m *mockErrorClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return fmt.Errorf("not implemented")
}

func (m *mockErrorClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return fmt.Errorf("not implemented")
}

func (m *mockErrorClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return fmt.Errorf("not implemented")
}

func (m *mockErrorClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return fmt.Errorf("not implemented")
}

func (m *mockErrorClient) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.ApplyOption) error {
	return fmt.Errorf("not implemented")
}

func (m *mockErrorClient) Status() client.StatusWriter {
	return nil
}

func (m *mockErrorClient) Scheme() *runtime.Scheme {
	return runtime.NewScheme()
}

func (m *mockErrorClient) RESTMapper() meta.RESTMapper {
	return nil
}

func (m *mockErrorClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, fmt.Errorf("not implemented")
}

func (m *mockErrorClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (m *mockErrorClient) SubResource(subResource string) client.SubResourceClient {
	return nil
}
