/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"testing"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestReconciliationPredicate_Create(t *testing.T) {
	predicate := &reconciliationPredicate{}

	t.Run("should always return true", func(t *testing.T) {
		e := event.CreateEvent{
			Object: &runaiv1.RunaiConfig{},
		}
		assert.True(t, predicate.Create(e))
	})
}

func TestReconciliationPredicate_Delete(t *testing.T) {
	predicate := &reconciliationPredicate{}

	t.Run("should always return true", func(t *testing.T) {
		e := event.DeleteEvent{
			Object: &runaiv1.RunaiConfig{},
		}
		assert.True(t, predicate.Delete(e))
	})
}

func TestReconciliationPredicate_Update(t *testing.T) {
	predicate := &reconciliationPredicate{}

	t.Run("RunaiConfig - should trigger reconciliation when generation changes", func(t *testing.T) {
		oldConfig := &runaiv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-config",
				Namespace:  runaiv1.DefaultNamespace,
				Generation: 1,
				Finalizers: []string{runaiv1.Finalizer},
			},
		}

		newConfig := oldConfig.DeepCopy()
		newConfig.Generation = 2

		e := event.UpdateEvent{
			ObjectOld: oldConfig,
			ObjectNew: newConfig,
		}

		assert.True(t, predicate.Update(e), "Should reconcile when generation changes (spec changed)")
	})

	t.Run("RunaiConfig - should trigger reconciliation when finalizer is missing", func(t *testing.T) {
		oldConfig := &runaiv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-config",
				Namespace:  runaiv1.DefaultNamespace,
				Generation: 1,
				Finalizers: []string{runaiv1.Finalizer},
			},
		}

		newConfig := oldConfig.DeepCopy()
		newConfig.Finalizers = []string{} // Finalizer removed

		e := event.UpdateEvent{
			ObjectOld: oldConfig,
			ObjectNew: newConfig,
		}

		assert.True(t, predicate.Update(e), "Should reconcile when finalizer is missing")
	})

	t.Run("RunaiConfig - should NOT trigger reconciliation when only status changes", func(t *testing.T) {
		oldConfig := &runaiv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "test-config",
				Namespace:       runaiv1.DefaultNamespace,
				Generation:      1,
				ResourceVersion: "100",
				Finalizers:      []string{runaiv1.Finalizer},
			},
			Status: runaiv1.RunaiConfigStatus{
				OperandStatuses: map[string]runaiv1.TimedOperandStatus{
					"operand1": {
						OperandStatus: runaiv1.OperandStatus{
							Ready: false,
						},
					},
				},
			},
		}

		newConfig := oldConfig.DeepCopy()
		newConfig.ResourceVersion = "101" // ResourceVersion changed (status update)
		newConfig.Status.OperandStatuses = map[string]runaiv1.TimedOperandStatus{
			"operand1": {
				OperandStatus: runaiv1.OperandStatus{
					Ready: true,
				},
			},
		}
		// Generation stays the same (no spec change)

		e := event.UpdateEvent{
			ObjectOld: oldConfig,
			ObjectNew: newConfig,
		}

		assert.False(t, predicate.Update(e), "Should NOT reconcile when only status/resourceVersion changes")
	})

	t.Run("RunaiConfig - should NOT trigger reconciliation when nothing meaningful changes", func(t *testing.T) {
		oldConfig := &runaiv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "test-config",
				Namespace:       runaiv1.DefaultNamespace,
				Generation:      1,
				ResourceVersion: "100",
				Finalizers:      []string{runaiv1.Finalizer},
			},
		}

		newConfig := oldConfig.DeepCopy()
		newConfig.ResourceVersion = "101"

		e := event.UpdateEvent{
			ObjectOld: oldConfig,
			ObjectNew: newConfig,
		}

		assert.False(t, predicate.Update(e), "Should NOT reconcile when nothing meaningful changes")
	})

	t.Run("RunaiConfig - should trigger reconciliation when both generation changes and finalizer missing", func(t *testing.T) {
		oldConfig := &runaiv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-config",
				Namespace:  runaiv1.DefaultNamespace,
				Generation: 1,
				Finalizers: []string{runaiv1.Finalizer},
			},
		}

		newConfig := oldConfig.DeepCopy()
		newConfig.Generation = 2
		newConfig.Finalizers = []string{}

		e := event.UpdateEvent{
			ObjectOld: oldConfig,
			ObjectNew: newConfig,
		}

		assert.True(t, predicate.Update(e), "Should reconcile when both generation changes and finalizer is missing")
	})

	t.Run("Owned resource - should trigger reconciliation when resource is owned by RunaiConfig", func(t *testing.T) {
		oldDeployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: runaiv1.DefaultNamespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: runaiv1.GroupVersion.String(),
						Kind:       runaiv1.RunaiConfigKind,
						Name:       "test-config",
						Controller: ptr.To(true),
					},
				},
			},
		}

		newDeployment := oldDeployment.DeepCopy()
		newDeployment.Spec.Replicas = ptr.To[int32](3)

		e := event.UpdateEvent{
			ObjectOld: oldDeployment,
			ObjectNew: newDeployment,
		}

		assert.True(t, predicate.Update(e), "Should reconcile when owned resource changes")
	})

	t.Run("Owned resource - should NOT trigger reconciliation when resource is not in default namespace", func(t *testing.T) {
		oldDeployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "other-namespace",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: runaiv1.GroupVersion.String(),
						Kind:       runaiv1.RunaiConfigKind,
						Name:       "test-config",
					},
				},
			},
		}

		newDeployment := oldDeployment.DeepCopy()
		newDeployment.Spec.Replicas = ptr.To[int32](3)

		e := event.UpdateEvent{
			ObjectOld: oldDeployment,
			ObjectNew: newDeployment,
		}

		assert.False(t, predicate.Update(e), "Should NOT reconcile when owned resource is not in default namespace")
	})

	t.Run("Owned resource - should NOT trigger reconciliation when resource is not owned by RunaiConfig", func(t *testing.T) {
		oldDeployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: runaiv1.DefaultNamespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "apps/v1",
						Kind:       "ReplicaSet",
						Name:       "test-replicaset",
					},
				},
			},
		}

		newDeployment := oldDeployment.DeepCopy()
		newDeployment.Spec.Replicas = ptr.To[int32](3)

		e := event.UpdateEvent{
			ObjectOld: oldDeployment,
			ObjectNew: newDeployment,
		}

		assert.False(t, predicate.Update(e), "Should NOT reconcile when resource is not owned by RunaiConfig")
	})

	t.Run("Owned resource - should NOT trigger reconciliation when resource has no owner", func(t *testing.T) {
		oldConfigMap := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-configmap",
				Namespace: runaiv1.DefaultNamespace,
			},
		}

		newConfigMap := oldConfigMap.DeepCopy()
		newConfigMap.Data = map[string]string{"key": "value"}

		e := event.UpdateEvent{
			ObjectOld: oldConfigMap,
			ObjectNew: newConfigMap,
		}

		assert.False(t, predicate.Update(e), "Should NOT reconcile when resource has no owner")
	})

	t.Run("Non-RunaiConfig resource - should NOT trigger reconciliation for unrelated resource", func(t *testing.T) {
		oldNamespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace",
			},
		}

		newNamespace := oldNamespace.DeepCopy()
		newNamespace.Labels = map[string]string{"key": "value"}

		e := event.UpdateEvent{
			ObjectOld: oldNamespace,
			ObjectNew: newNamespace,
		}

		assert.False(t, predicate.Update(e), "Should NOT reconcile for unrelated resources")
	})
}

func TestReconciliationPredicate_Update_EdgeCases(t *testing.T) {
	predicate := &reconciliationPredicate{}

	t.Run("RunaiConfig - should handle nil finalizers", func(t *testing.T) {
		oldConfig := &runaiv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-config",
				Namespace:  runaiv1.DefaultNamespace,
				Generation: 1,
				Finalizers: nil,
			},
		}

		newConfig := oldConfig.DeepCopy()

		e := event.UpdateEvent{
			ObjectOld: oldConfig,
			ObjectNew: newConfig,
		}

		assert.True(t, predicate.Update(e), "Should reconcile when finalizer is missing (nil)")
	})

	t.Run("RunaiConfig - should handle empty finalizers slice", func(t *testing.T) {
		oldConfig := &runaiv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-config",
				Namespace:  runaiv1.DefaultNamespace,
				Generation: 1,
				Finalizers: []string{},
			},
		}

		newConfig := oldConfig.DeepCopy()

		e := event.UpdateEvent{
			ObjectOld: oldConfig,
			ObjectNew: newConfig,
		}

		assert.True(t, predicate.Update(e), "Should reconcile when finalizer is missing (empty slice)")
	})

	t.Run("Owned resource - should handle resource with multiple owners", func(t *testing.T) {
		oldService := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: runaiv1.DefaultNamespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "test-deployment",
					},
					{
						APIVersion: runaiv1.GroupVersion.String(),
						Kind:       runaiv1.RunaiConfigKind,
						Name:       "test-config",
						Controller: ptr.To(true),
					},
				},
			},
		}

		newService := oldService.DeepCopy()
		newService.Spec.Type = v1.ServiceTypeLoadBalancer

		e := event.UpdateEvent{
			ObjectOld: oldService,
			ObjectNew: newService,
		}

		assert.True(t, predicate.Update(e), "Should reconcile when resource has RunaiConfig among multiple owners")
	})
}

func TestReconciliationPredicate_Update_AnnotatedResources(t *testing.T) {
	predicate := &reconciliationPredicate{}

	for name, resource := range admissionPolicyObjects() {
		t.Run(name+" - owned by default RunaiConfig", func(t *testing.T) {
			resource.SetAnnotations(defaultRunaiConfigOwnerAnnotations())

			assert.True(t, predicate.Update(event.UpdateEvent{
				ObjectOld: resource.DeepCopyObject().(client.Object),
				ObjectNew: resource,
			}))
		})

		t.Run(name+" - without owner annotations", func(t *testing.T) {
			resource.SetAnnotations(nil)

			assert.False(t, predicate.Update(event.UpdateEvent{
				ObjectOld: resource.DeepCopyObject().(client.Object),
				ObjectNew: resource,
			}))
		})

		t.Run(name+" - owned by a different RunaiConfig", func(t *testing.T) {
			annotations := defaultRunaiConfigOwnerAnnotations()
			annotations[runaiv1.NamespacedNameAnnotation] = runaiv1.DefaultNamespace + "/other"
			resource.SetAnnotations(annotations)

			assert.False(t, predicate.Update(event.UpdateEvent{
				ObjectOld: resource.DeepCopyObject().(client.Object),
				ObjectNew: resource,
			}))
		})
	}

	t.Run("other annotation-owned resource types trigger reconciliation", func(t *testing.T) {
		configMap := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: defaultRunaiConfigOwnerAnnotations(),
			},
		}

		assert.True(t, predicate.Update(event.UpdateEvent{
			ObjectOld: configMap.DeepCopy(),
			ObjectNew: configMap,
		}))
	})
}

func TestMapResourceToRunaiConfigEvent(t *testing.T) {
	for name, resource := range admissionPolicyObjects() {
		t.Run(name+" - owned by default RunaiConfig", func(t *testing.T) {
			resource.SetAnnotations(defaultRunaiConfigOwnerAnnotations())

			requests := mapResourceToRunaiConfigEvent(t.Context(), resource)

			require.Len(t, requests, 1)
			assert.Equal(t, runaiv1.DefaultName, requests[0].Name)
			assert.Equal(t, runaiv1.DefaultNamespace, requests[0].Namespace)
		})

		t.Run(name+" - unmanaged", func(t *testing.T) {
			resource.SetAnnotations(nil)

			requests := mapResourceToRunaiConfigEvent(t.Context(), resource)

			assert.Empty(t, requests)
		})
	}
}

func admissionPolicyObjects() map[string]client.Object {
	return map[string]client.Object{
		"ValidatingAdmissionPolicy":        &admissionv1.ValidatingAdmissionPolicy{},
		"ValidatingAdmissionPolicyBinding": &admissionv1.ValidatingAdmissionPolicyBinding{},
	}
}

func defaultRunaiConfigOwnerAnnotations() map[string]string {
	ownerType := metav1.GroupKind{
		Group: runaiv1.GroupVersion.Group,
		Kind:  runaiv1.RunaiConfigKind,
	}
	return map[string]string{
		runaiv1.NamespacedNameAnnotation: runaiv1.DefaultNamespace + "/" + runaiv1.DefaultName,
		runaiv1.TypeAnnotation:           ownerType.String(),
	}
}
