/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller_utils

import (
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var (
	runaiGvKind = schema.GroupVersionKind{
		Group:   "run.ai",
		Version: "v1",
		Kind:    "RunaiConfig",
	}
	runaiGV = schema.GroupVersion{Group: runaiv1.Group, Version: runaiv1.Version}
)
var _ = Describe("SetControllerReference", func() {
	var (
		testScheme *runtime.Scheme
	)

	BeforeEach(func() {
		testScheme = scheme_init.Scheme()
		// Add necessary types to the scheme
		testScheme.AddKnownTypes(runaiGV, &runaiv1.RunaiConfig{})
		testScheme.AddKnownTypes(schema.GroupVersion{Group: "", Version: "v1"},
			&corev1.ConfigMap{}, &corev1.Secret{}, &corev1.Service{})
	})

	Context("when owner is cluster-scoped", func() {
		It("should set controller reference using controllerutil.SetControllerReference", func() {
			// Create a cluster-scoped owner (no namespace)
			owner := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					UID:  "test-uid",
				},
			}
			owner.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Node",
			})

			// Create a namespaced resource
			resource := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-configmap",
					Namespace: "test-namespace",
				},
			}
			resource.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "ConfigMap",
			})

			// Call SetControllerReference
			err := SetControllerReference(owner, resource, testScheme)

			// Verify it succeeded and set owner reference correctly
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetOwnerReferences()).To(HaveLen(1))
			ownerRef := resource.GetOwnerReferences()[0]
			Expect(ownerRef.Name).To(Equal("test-node"))
			Expect(ownerRef.Kind).To(Equal("Node"))
			Expect(ownerRef.UID).To(Equal(types.UID("test-uid")))
			Expect(*ownerRef.Controller).To(BeTrue())
		})
	})

	Context("when owner and resource are in the same namespace", func() {
		It("should set controller reference using controllerutil.SetControllerReference", func() {
			// Create a namespaced owner
			owner := &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runaiconfig",
					Namespace: "test-namespace",
					UID:       "test-uid",
				},
			}
			owner.SetGroupVersionKind(runaiGvKind)

			// Create a resource in the same namespace
			resource := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-configmap",
					Namespace: "test-namespace",
				},
			}
			resource.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "ConfigMap",
			})

			// Call SetControllerReference
			err := SetControllerReference(owner, resource, testScheme)

			// Verify it succeeded and set owner reference correctly
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetOwnerReferences()).To(HaveLen(1))
			ownerRef := resource.GetOwnerReferences()[0]
			Expect(ownerRef.Name).To(Equal("test-runaiconfig"))
			Expect(ownerRef.Kind).To(Equal("RunaiConfig"))
			Expect(ownerRef.UID).To(Equal(types.UID("test-uid")))
			Expect(*ownerRef.Controller).To(BeTrue())
		})
	})

	Context("when owner is namespaced and resource is in a different namespace", func() {
		It("should set owner annotations instead of controller reference", func() {
			// Create a namespaced owner
			owner := &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runaiconfig",
					Namespace: "owner-namespace",
					UID:       "test-uid",
				},
			}
			owner.SetGroupVersionKind(runaiGvKind)

			// Create a resource in a different namespace
			resource := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-configmap",
					Namespace: "resource-namespace",
				},
			}
			resource.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "ConfigMap",
			})

			// Call SetControllerReference
			err := SetControllerReference(owner, resource, testScheme)

			// Verify it succeeded and set annotations instead of owner references
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetOwnerReferences()).To(BeEmpty())

			annotations := resource.GetAnnotations()
			Expect(annotations).NotTo(BeNil())
			Expect(annotations[runaiv1.NamespacedNameAnnotation]).To(Equal("owner-namespace/test-runaiconfig"))
			Expect(annotations[runaiv1.TypeAnnotation]).To(Equal("RunaiConfig.run.ai"))
		})
	})

	Context("when owner is namespaced and resource is cluster-scoped", func() {
		It("should set owner annotations instead of controller reference", func() {
			// Create a namespaced owner
			owner := &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runaiconfig",
					Namespace: "owner-namespace",
					UID:       "test-uid",
				},
			}
			owner.SetGroupVersionKind(runaiGvKind)

			// Create a cluster-scoped resource (no namespace)
			resource := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
				},
			}
			resource.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Node",
			})

			// Call SetControllerReference
			err := SetControllerReference(owner, resource, testScheme)

			// Verify it succeeded and set annotations instead of owner references
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetOwnerReferences()).To(BeEmpty())

			annotations := resource.GetAnnotations()
			Expect(annotations).NotTo(BeNil())
			Expect(annotations[runaiv1.NamespacedNameAnnotation]).To(Equal("owner-namespace/test-runaiconfig"))
			Expect(annotations[runaiv1.TypeAnnotation]).To(Equal("RunaiConfig.run.ai"))
		})
	})

	Context("when SetOwnerAnnotations fails", func() {
		It("should return the error from SetOwnerAnnotations", func() {
			// Create an owner without name to cause SetOwnerAnnotations to fail
			owner := &runaiv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					// Missing Name which is required by SetOwnerAnnotations
					Namespace: "owner-namespace",
					UID:       "test-uid",
				},
			}
			owner.SetGroupVersionKind(runaiGvKind)

			resource := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-configmap",
					Namespace: "different-namespace",
				},
			}
			resource.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "ConfigMap",
			})

			// Call SetControllerReference
			err := SetControllerReference(owner, resource, testScheme)

			// Should return error from SetOwnerAnnotations
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not have a name"))
		})
	})
	_ = Describe("IsOwnedByMe", func() {
		Context("when object has controller owner reference matching the criteria", func() {
			It("should return true", func() {
				controller := true
				obj := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: "test-namespace",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "run.ai/v1",
								Kind:       "RunaiConfig",
								Name:       "test-config",
								Controller: &controller,
								UID:        "test-uid",
							},
						},
					},
				}

				result := IsOwnedByMe(obj, "run.ai/v1", "RunaiConfig")
				Expect(result).To(BeTrue())
			})
		})

		Context("when object has multiple owner references with matching controller", func() {
			It("should return true when one matches", func() {
				controller := true
				notController := false
				obj := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: "test-namespace",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       "test-deployment",
								Controller: &notController,
								UID:        "uid-1",
							},
							{
								APIVersion: "run.ai/v1",
								Kind:       "RunaiConfig",
								Name:       "test-config",
								Controller: &controller,
								UID:        "uid-2",
							},
						},
					},
				}

				result := IsOwnedByMe(obj, "run.ai/v1", "RunaiConfig")
				Expect(result).To(BeTrue())
			})
		})

		Context("when object has owner reference but controller is false", func() {
			It("should return false", func() {
				controller := false
				obj := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: "test-namespace",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "run.ai/v1",
								Kind:       "RunaiConfig",
								Name:       "test-config",
								Controller: &controller,
								UID:        "test-uid",
							},
						},
					},
				}

				result := IsOwnedByMe(obj, "run.ai/v1", "RunaiConfig")
				Expect(result).To(BeFalse())
			})
		})

		Context("when object has owner reference but controller is nil", func() {
			It("should return false", func() {
				obj := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: "test-namespace",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "run.ai/v1",
								Kind:       "RunaiConfig",
								Name:       "test-config",
								Controller: nil,
								UID:        "test-uid",
							},
						},
					},
				}

				result := IsOwnedByMe(obj, "run.ai/v1", "RunaiConfig")
				Expect(result).To(BeFalse())
			})
		})

		Context("when object has owner reference with wrong APIVersion", func() {
			It("should return false", func() {
				controller := true
				obj := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: "test-namespace",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "run.ai/v2", // Wrong version
								Kind:       "RunaiConfig",
								Name:       "test-config",
								Controller: &controller,
								UID:        "test-uid",
							},
						},
					},
				}

				result := IsOwnedByMe(obj, "run.ai/v1", "RunaiConfig")
				Expect(result).To(BeFalse())
			})
		})

		Context("when object has owner reference with wrong Kind", func() {
			It("should return false", func() {
				controller := true
				obj := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: "test-namespace",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "run.ai/v1",
								Kind:       "DifferentKind", // Wrong kind
								Name:       "test-config",
								Controller: &controller,
								UID:        "test-uid",
							},
						},
					},
				}

				result := IsOwnedByMe(obj, "run.ai/v1", "RunaiConfig")
				Expect(result).To(BeFalse())
			})
		})

		Context("when object has no owner references", func() {
			It("should return false", func() {
				obj := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-configmap",
						Namespace:       "test-namespace",
						OwnerReferences: []metav1.OwnerReference{},
					},
				}

				result := IsOwnedByMe(obj, "run.ai/v1", "RunaiConfig")
				Expect(result).To(BeFalse())
			})
		})

		Context("when object has nil owner references", func() {
			It("should return false", func() {
				obj := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "test-configmap",
						Namespace:       "test-namespace",
						OwnerReferences: nil,
					},
				}

				result := IsOwnedByMe(obj, "run.ai/v1", "RunaiConfig")
				Expect(result).To(BeFalse())
			})
		})

		Context("when object is nil", func() {
			It("should return false", func() {
				result := IsOwnedByMe(nil, "run.ai/v1", "RunaiConfig")
				Expect(result).To(BeFalse())
			})
		})

		Context("when object has only non-controller owner references", func() {
			It("should return false", func() {
				controller := true
				notController := false
				obj := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: "test-namespace",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       "test-deployment",
								Controller: &controller,
								UID:        "uid-1",
							},
							{
								APIVersion: "run.ai/v1",
								Kind:       "RunaiConfig",
								Name:       "test-config",
								Controller: &notController, // Not controller
								UID:        "uid-2",
							},
						},
					},
				}

				result := IsOwnedByMe(obj, "run.ai/v1", "RunaiConfig")
				Expect(result).To(BeFalse())
			})
		})

		Context("when checking for different resource types", func() {
			It("should work with Services", func() {
				controller := true
				obj := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service",
						Namespace: "test-namespace",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       "test-deployment",
								Controller: &controller,
								UID:        "test-uid",
							},
						},
					},
				}

				result := IsOwnedByMe(obj, "apps/v1", "Deployment")
				Expect(result).To(BeTrue())
			})

			It("should work with Secrets", func() {
				controller := true
				obj := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: "test-namespace",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "run.ai/v1",
								Kind:       "RunaiConfig",
								Name:       "test-config",
								Controller: &controller,
								UID:        "test-uid",
							},
						},
					},
				}

				result := IsOwnedByMe(obj, "run.ai/v1", "RunaiConfig")
				Expect(result).To(BeTrue())
			})
		})
	})
})
