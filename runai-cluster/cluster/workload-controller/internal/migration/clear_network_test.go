/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openShiftV1 "github.com/openshift/api/route/v1"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespace1 = "namespace1"
	namespace2 = "namespace2"
)

var _ = Describe("NetworkCleaner", func() {

	Context("NewIngressCleaner", func() {
		Context("On OpenShift", func() {
			It("should return NetworkCleaner with Route list", func() {
				result := NewNetworkCleaner(nil, true)
				Expect(result.getListObj()).To(BeEquivalentTo(getRouteListObjectList()))
			})
		})
		Context("On Kubernetes", func() {
			It("should return NetworkCleaner with Ingress list", func() {
				result := NewNetworkCleaner(nil, false)
				Expect(result.getListObj()).To(BeEquivalentTo(getIngressListObjectList()))
			})
		})
	})
	Context("RequiresWebhooks", func() {
		Context("On OpenShift", func() {
			It("should return false", func() {
				result := NewNetworkCleaner(nil, true)
				Expect(result.RequiresWebhooks()).To(BeFalse())
			})
		})
		Context("On Kubernetes", func() {
			It("should return false", func() {
				result := NewNetworkCleaner(nil, false)
				Expect(result.RequiresWebhooks()).To(BeFalse())
			})
		})
	})
	Context("convertListObjectToList", func() {
		Context("On openShift", func() {
			It("should return a list of Routes", func() {
				routeList := openShiftV1.RouteList{Items: []openShiftV1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "route1"},
					},
				},
				}
				result, err := convertListObjectToList(&routeList)
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(result[0].GetName()).To(Equal("route1"))
			})
		})
		Context("On Kubernetes", func() {
			It("should return a list of Ingresses", func() {
				ingressList := networkingv1.IngressList{Items: []networkingv1.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "ingress1"},
					},
				},
				}
				result, err := convertListObjectToList(&ingressList)
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(result[0].GetName()).To(Equal("ingress1"))
			})
		})
		Context("On missing type", func() {
			It("should return an error", func() {
				_, err := convertListObjectToList(&metav1.List{})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("service list", func() {
			It("should return a list of services", func() {
				serviceList := v1.ServiceList{Items: []v1.Service{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "service1"},
					},
				},
				}
				result, err := convertListObjectToList(&serviceList)
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(result[0].GetName()).To(Equal("service1"))
			})
		})
	})

	Context("isOwnedByWorkload", func() {
		It("should return false when there are no owner references and no backward ownership", func() {
			obj := &networkingv1.Ingress{}
			Expect(isOwnedByWorkload(obj)).To(BeFalse())
		})

		It("should return false when there are multiple owner references", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{{}, {}},
				},
			}
			Expect(isOwnedByWorkload(obj)).To(BeFalse())
		})

		It("should return true when owner is a valid InteractiveWorkload", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{Kind: commonconstants.KindInteractiveWorkload},
					},
				},
			}
			Expect(isOwnedByWorkload(obj)).To(BeTrue())
		})

		It("should return true when owner is a valid TrainingWorkload", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{Kind: commonconstants.KindTrainingWorkload},
					},
				},
			}
			Expect(isOwnedByWorkload(obj)).To(BeTrue())
		})

		It("should return true when owner is a valid DistributedWorkload", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{Kind: commonconstants.KindDistributedWorkload},
					},
				},
			}
			Expect(isOwnedByWorkload(obj)).To(BeTrue())
		})

		It("should return false when owner is a valid InferenceWorkload", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{Kind: commonconstants.KindInferenceWorkload},
					},
				},
			}
			Expect(isOwnedByWorkload(obj)).To(BeFalse())
		})

		It("should return false when owner is not a valid workload", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{Kind: "InvalidWorkload"},
					},
				},
			}
			Expect(isOwnedByWorkload(obj)).To(BeFalse())
		})

		It("should return true when there's backward ownership for Jupyter", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "jupyter-example",
					Labels: map[string]string{
						commonconstants.LabelJupyter: "true",
					},
				},
			}
			Expect(isOwnedByWorkload(obj)).To(BeTrue())
		})

		It("should return true when there's backward ownership for Tensorboard", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tensorboard-example",
					Labels: map[string]string{
						commonconstants.LabelTensorboard: "true",
					},
				},
			}
			Expect(isOwnedByWorkload(obj)).To(BeTrue())
		})

		It("should return true when the name contains prevName", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-" + prevName + "-suffix",
				},
			}
			Expect(isOwnedByWorkload(obj)).To(BeTrue())
		})

		It("should return true when it has the LabelWorkloadKind", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						commonconstants.LabelWorkloadKind: "SomeKind",
					},
				},
			}
			Expect(isOwnedByWorkload(obj)).To(BeTrue())
		})

		It("should return true when it has the LabelWorkloadName", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						commonconstants.LabelWorkloadName: "SomeName",
					},
				},
			}
			Expect(isOwnedByWorkload(obj)).To(BeTrue())
		})
	})

	Context("isBackwardOwnership", func() {
		It("should return false when there are no relevant labels or name patterns", func() {
			obj := &networkingv1.Ingress{}
			Expect(isBackwardOwnership(obj)).To(BeFalse())
		})

		It("should return true when name contains jupyter and jupyter label is present", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "example-jupyter",
					Labels: map[string]string{commonconstants.LabelJupyter: "true"},
				},
			}
			Expect(isBackwardOwnership(obj)).To(BeTrue())
		})

		It("should return true when name contains tensorboard and tensorboard label is present", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "example-tensorboard",
					Labels: map[string]string{commonconstants.LabelTensorboard: "true"},
				},
			}
			Expect(isBackwardOwnership(obj)).To(BeTrue())
		})

		It("should return true when name contains prevName", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-" + prevName,
				},
			}
			Expect(isBackwardOwnership(obj)).To(BeTrue())
		})

		It("should return true when LabelWorkloadKind is present", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{commonconstants.LabelWorkloadKind: "SomeKind"},
				},
			}
			Expect(isBackwardOwnership(obj)).To(BeTrue())
		})

		It("should return true when LabelWorkloadName is present", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{commonconstants.LabelWorkloadName: "SomeName"},
				},
			}
			Expect(isBackwardOwnership(obj)).To(BeTrue())
		})

		It("should return false when name contains jupyter but jupyter label is not present", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-jupyter",
				},
			}
			Expect(isBackwardOwnership(obj)).To(BeFalse())
		})

		It("should return false when name contains tensorboard but tensorboard label is not present", func() {
			obj := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-tensorboard",
				},
			}
			Expect(isBackwardOwnership(obj)).To(BeFalse())
		})
	})

	Context("filterList", func() {
		Context("Kubernetes Ingress", func() {
			It("should correctly filter Ingress objects", func() {
				ingresses := []client.Object{
					&networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "ingress-1",
							OwnerReferences: []metav1.OwnerReference{
								{Kind: commonconstants.KindInteractiveWorkload},
							},
						},
					},
					&networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "ingress-2",
							OwnerReferences: []metav1.OwnerReference{
								{Kind: "InvalidWorkload"},
							},
						},
					},
					&networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "ingress-3-jupyter",
							Labels: map[string]string{commonconstants.LabelJupyter: "true"},
						},
					},
					&networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "ingress-4",
						},
					},
				}

				filteredList := filterList(ingresses)
				Expect(filteredList).To(HaveLen(2))
				Expect(filteredList[0].GetName()).To(Equal("ingress-1"))
				Expect(filteredList[1].GetName()).To(Equal("ingress-3-jupyter"))
			})
		})

		Context("OpenShift Route", func() {
			It("should correctly filter Route objects", func() {
				routes := []client.Object{
					&openShiftV1.Route{
						ObjectMeta: metav1.ObjectMeta{
							Name: "route-1",
							OwnerReferences: []metav1.OwnerReference{
								{Kind: commonconstants.KindTrainingWorkload},
							},
						},
					},
					&openShiftV1.Route{
						ObjectMeta: metav1.ObjectMeta{
							Name: "route-2",
							OwnerReferences: []metav1.OwnerReference{
								{Kind: "InvalidWorkload"},
							},
						},
					},
					&openShiftV1.Route{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "route-3-tensorboard",
							Labels: map[string]string{commonconstants.LabelTensorboard: "true"},
						},
					},
					&openShiftV1.Route{
						ObjectMeta: metav1.ObjectMeta{
							Name: "route-4",
						},
					},
				}

				filteredList := filterList(routes)
				Expect(filteredList).To(HaveLen(2))
				Expect(filteredList[0].GetName()).To(Equal("route-1"))
				Expect(filteredList[1].GetName()).To(Equal("route-3-tensorboard"))
			})
		})

		Context("Kubernetes Service", func() {
			It("should correctly filter Service objects", func() {
				services := []client.Object{
					&v1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service-1",
							OwnerReferences: []metav1.OwnerReference{
								{Kind: commonconstants.KindDistributedWorkload},
							},
						},
					},
					&v1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service-2",
							OwnerReferences: []metav1.OwnerReference{
								{Kind: "InvalidWorkload"},
							},
						},
					},
					&v1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "service-3-jupyter",
							Labels: map[string]string{commonconstants.LabelJupyter: "true"},
						},
					},
					&v1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service-4",
							Labels: map[string]string{
								commonconstants.LabelWorkloadKind: commonconstants.KindInferenceWorkload,
							},
						},
					},
					&v1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "service-5",
						},
					},
				}

				filteredList := filterList(services)
				Expect(filteredList).To(HaveLen(3))
				Expect(filteredList[0].GetName()).To(Equal("service-1"))
				Expect(filteredList[1].GetName()).To(Equal("service-3-jupyter"))
				Expect(filteredList[2].GetName()).To(Equal("service-4"))
			})
		})
	})
	Context("Migration", func() {
		Context("Migration", func() {
			var (
				networkCleaner *NetworkCleaner
				k8sClient      client.Client
				scheme         *runtime.Scheme
			)

			BeforeEach(func() {
				setLimit(1)
			})

			AfterEach(func() {
				setLimit(500)
			})

			scheme = runtime.NewScheme()
			register.InitSchema(scheme)

			Context("On OpenShift", func() {
				BeforeEach(func() {
					k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)
					for _, obj := range createNamespaces() {
						Expect(k8sClient.Create(context.Background(), &obj)).To(BeNil())
					}

					for _, obj := range createRouteNotToDelete() {
						Expect(k8sClient.Create(context.Background(), &obj)).To(BeNil())
					}
					for _, obj := range createRoutesToDelete() {
						Expect(k8sClient.Create(context.Background(), &obj)).To(BeNil())
					}
					for _, obj := range createServiceToDelete() {
						Expect(k8sClient.Create(context.Background(), &obj)).To(BeNil())
					}
					for _, obj := range createServiceNotToDelete() {
						Expect(k8sClient.Create(context.Background(), &obj)).To(BeNil())
					}

					networkCleaner = NewNetworkCleaner(k8sClient, true)
				})

				It("should not delete any route from the non-delete route list", func() {
					err := networkCleaner.Migrate(context.Background())
					Expect(err).To(BeNil())

					for _, obj := range createRouteNotToDelete() {
						route := &openShiftV1.Route{}
						err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, route)
						Expect(err).To(BeNil())
						Expect(route.Name).To(Equal(obj.GetName()))
						Expect(route.Namespace).To(Equal(obj.GetNamespace()))
					}
				})

				It("should delete all the routes from the delete route list", func() {
					err := networkCleaner.Migrate(context.Background())
					Expect(err).To(BeNil())
					for _, obj := range createRoutesToDelete() {
						route := &openShiftV1.Route{}
						err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, route)
						Expect(client.IgnoreNotFound(err)).ToNot(HaveOccurred())
					}
				})

				It("should delete all the services from the delete service list", func() {
					err := networkCleaner.Migrate(context.Background())
					Expect(err).To(BeNil())
					for _, obj := range createServiceToDelete() {
						service := &v1.Service{}
						err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, service)
						Expect(client.IgnoreNotFound(err)).ToNot(HaveOccurred())
					}
				})

				It("should not delete any services from the non-delete service list", func() {
					err := networkCleaner.Migrate(context.Background())
					Expect(err).To(BeNil())
					for _, obj := range createServiceNotToDelete() {
						service := &v1.Service{}
						err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, service)
						Expect(err).To(BeNil())
						Expect(service.Name).To(Equal(obj.GetName()))
						Expect(service.Namespace).To(Equal(obj.GetNamespace()))
					}
				})
			})

			Context("On Kubernetes", func() {
				BeforeEach(func() {
					k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)
					for _, obj := range createIngressNotToDelete() {
						Expect(k8sClient.Create(context.Background(), &obj)).To(BeNil())
					}
					for _, obj := range createIngressToDelete() {
						Expect(k8sClient.Create(context.Background(), &obj)).To(BeNil())
					}
					for _, obj := range createServiceToDelete() {
						Expect(k8sClient.Create(context.Background(), &obj)).To(BeNil())
					}
					for _, obj := range createServiceNotToDelete() {
						Expect(k8sClient.Create(context.Background(), &obj)).To(BeNil())
					}
					for _, obj := range createNamespaces() {
						Expect(k8sClient.Create(context.Background(), &obj)).To(BeNil())
					}
					networkCleaner = NewNetworkCleaner(k8sClient, false)
				})

				It("should not delete any ingress from the non-delete ingress list", func() {
					err := networkCleaner.Migrate(context.Background())
					Expect(err).To(BeNil())

					for _, obj := range createIngressNotToDelete() {
						ingress := &networkingv1.Ingress{}
						err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, ingress)
						Expect(err).To(BeNil())
						Expect(ingress.Name).To(Equal(obj.GetName()))
						Expect(ingress.Namespace).To(Equal(obj.GetNamespace()))
					}
				})

				It("should delete all the ingresses from the delete ingress list", func() {
					err := networkCleaner.Migrate(context.Background())
					Expect(err).To(BeNil())
					for _, obj := range createIngressToDelete() {
						ingress := &networkingv1.Ingress{}
						err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, ingress)
						Expect(client.IgnoreNotFound(err)).ToNot(HaveOccurred())
					}
				})

				It("should delete all the services from the delete service list", func() {
					err := networkCleaner.Migrate(context.Background())
					Expect(err).To(BeNil())
					for _, obj := range createServiceToDelete() {
						service := &v1.Service{}
						err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, service)
						Expect(client.IgnoreNotFound(err)).ToNot(HaveOccurred())
					}
				})

				It("should not delete any services from the non-delete service list", func() {
					err := networkCleaner.Migrate(context.Background())
					Expect(err).To(BeNil())
					for _, obj := range createServiceNotToDelete() {
						service := &v1.Service{}
						err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}, service)
						Expect(err).To(BeNil())
						Expect(service.Name).To(Equal(obj.GetName()))
						Expect(service.Namespace).To(Equal(obj.GetNamespace()))
					}
				})
			})
		})
	})
})

func createIngressNotToDelete() []networkingv1.Ingress {
	ingresses := []networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress1-not",
				Namespace: namespace1,
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress1-not-release",
				Namespace: namespace1,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "InvalidWorkload",
					},
				},
				Labels: map[string]string{
					commonconstants.LabelApp:     "app",
					commonconstants.LabelRelease: "some-release",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress1-not-release=2",
				Namespace: namespace1,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: commonconstants.KindTrainingWorkload,
					},
				},
				Labels: map[string]string{
					commonconstants.LabelRelease: "some-release",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress3232",
				Namespace: namespace2,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "InvalidWorkload",
					},
				},
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
			},
		},
	}
	return ingresses

}
func createServiceToDelete() []v1.Service {
	services := []v1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service1",
				Namespace: namespace1,
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: commonconstants.KindTrainingWorkload,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service2",
				Namespace: namespace1,
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: commonconstants.KindInteractiveWorkload,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service3",
				Namespace: namespace2,
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: commonconstants.KindInteractiveWorkload,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-exposed-url-2",
				Namespace: namespace1,
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-exposed-url-3",
				Namespace: namespace1,
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-12",
				Namespace: namespace2,
				Labels: map[string]string{
					commonconstants.LabelApp:          "app",
					commonconstants.LabelWorkloadKind: commonconstants.KindInferenceWorkload,
				},
			},
		},
	}
	return services
}
func createServiceNotToDelete() []v1.Service {
	services := []v1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service1-not-delete",
				Namespace: namespace1,
				Labels:    map[string]string{},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service2-not",
				Namespace: namespace1,
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service323",
				Namespace: namespace2,
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "InvalidWorkload",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-url-2",
				Namespace: namespace1,
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-url-3",
				Namespace: namespace1,
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
			},
		},
	}
	return services
}

func createRoutesToDelete() []openShiftV1.Route {
	routes := []openShiftV1.Route{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "route1",
				Namespace: namespace1,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: commonconstants.KindTrainingWorkload,
					},
				},
				Labels: map[string]string{
					commonconstants.LabelApp:     "app",
					commonconstants.LabelRelease: "some-release",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "route2",
				Namespace: namespace1,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: commonconstants.KindInteractiveWorkload,
					},
				},
				Labels: map[string]string{
					commonconstants.LabelRelease: "some-release",
					commonconstants.LabelApp:     "app",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "route32",
				Namespace: namespace1,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: commonconstants.KindInferenceWorkload,
					},
				},
				Labels: map[string]string{
					commonconstants.LabelApp:     "app",
					commonconstants.LabelRelease: "some-release",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "route322",
				Namespace: namespace1,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: commonconstants.KindDistributedWorkload,
					},
				},
				Labels: map[string]string{
					commonconstants.LabelApp:     "app",
					commonconstants.LabelRelease: "some-release",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "route3",
				Namespace: namespace2,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: commonconstants.KindInteractiveWorkload,
					},
				},
				Labels: map[string]string{
					commonconstants.LabelApp:     "app",
					commonconstants.LabelRelease: "some-release",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "route-exposed-url-2",
				Namespace: namespace1,
				Labels: map[string]string{
					commonconstants.LabelApp:     "app",
					commonconstants.LabelRelease: "some-release",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "route-url-3",
				Namespace: namespace1,
				Labels: map[string]string{
					commonconstants.LabelApp:          "app",
					commonconstants.LabelWorkloadKind: commonconstants.KindInferenceWorkload,
				},
			},
		},
	}
	return routes
}

func createRouteNotToDelete() []openShiftV1.Route {
	routes := []openShiftV1.Route{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "route1-not",
				Namespace: namespace1,
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "route32323",
				Namespace: namespace2,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: "InvalidWorkload",
					},
				},
				Labels: map[string]string{
					commonconstants.LabelApp:     "app",
					commonconstants.LabelRelease: "some-release",
				},
			},
		},
	}
	return routes

}

func createIngressToDelete() []networkingv1.Ingress {
	ingresses := []networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress1",
				Namespace: namespace1,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: commonconstants.KindTrainingWorkload,
					},
				},
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress2",
				Namespace: namespace1,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: commonconstants.KindInteractiveWorkload,
					},
				},
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress3-exposed-url-2",
				Namespace: namespace1,
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress3",
				Namespace: namespace2,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: commonconstants.KindInteractiveWorkload,
					},
				},
				Labels: map[string]string{
					commonconstants.LabelApp: "app",
				},
			},
		},
	}
	return ingresses
}

func createNamespaces() []v1.Namespace {
	namespaces := []v1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace1,
				Labels: map[string]string{
					commonconstants.RunaiQueueLabel: "project1",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace2,
				Labels: map[string]string{
					commonconstants.RunaiQueueLabel: "project2",
				},
			},
		},
	}
	return namespaces
}
