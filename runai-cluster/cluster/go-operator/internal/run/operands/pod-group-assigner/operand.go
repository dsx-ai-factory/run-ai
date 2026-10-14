/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_group_assigner

import (
	"context"
	"fmt"
	"sync"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const name = "pod-group-assigner"

//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;list;watch;delete
//+kubebuilder:rbac:namespace=runai,groups="",resources=services,verbs=get;list;watch;delete

type operand struct {
	resources []operands.Resource
}

func (o *operand) Name() string { return name }

func (o *operand) Resources() []operands.Resource { return o.resources }

func (o *operand) Status(context.Context) (operands.Status, error) {
	return operands.ReadyStatus(), nil
}

// Enabled is always false: KRM supplies the pod-group-assigner now, under the
// same name, so this operand only deletes what the Run:ai-built one left.
func (o *operand) Enabled() bool { return false }

// Operand deletes the 2.x pod-group-assigner Deployment and Service until KRM
// controls them. KRM adopts a same-named object by updating it in place, and on
// OpenShift the adopted Service never gets the serving-cert annotation, so the
// webhook secret is never minted and the pod never starts. Deleting the
// leftovers makes KRM create both from scratch.
func Operand(kubeClient kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	objectMeta := metav1.ObjectMeta{Name: name, Namespace: runaiConfig.GetNamespace()}

	return &operand{
		resources: []operands.Resource{
			&unadoptedResource{reader: kubeClient, build: func() client.Object {
				return &appsv1.Deployment{ObjectMeta: objectMeta}
			}},
			&unadoptedResource{reader: kubeClient, build: func() client.Object {
				return &corev1.Service{ObjectMeta: objectMeta}
			}},
		},
	}
}

// unadoptedResource is a delete-only Resource that yields its object only while
// KRM does not control it. A nil object keeps it out of the reconcile's delete
// list, so KRM's own Deployment and Service are never touched.
type unadoptedResource struct {
	reader client.Reader
	build  func() client.Object

	// The reconcile calls Object again when deleting and expects the answer it
	// got when building the delete list.
	once   sync.Once
	object client.Object
	err    error
}

func (r *unadoptedResource) OperandName() string { return name }

func (r *unadoptedResource) Object(ctx context.Context) (client.Object, error) {
	r.once.Do(func() { r.object, r.err = r.resolve(ctx) })
	return r.object, r.err
}

func (r *unadoptedResource) resolve(ctx context.Context) (client.Object, error) {
	existing := r.build()
	if err := r.reader.Get(ctx, client.ObjectKeyFromObject(existing), existing); err != nil {
		if errors.IsNotFound(err) || meta.IsNoMatchError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get %T %s/%s: %w", existing, existing.GetNamespace(), existing.GetName(), err)
	}
	if controlledByKRM(existing) {
		return nil, nil
	}
	return r.build(), nil
}

func (r *unadoptedResource) Status(context.Context) error { return nil }

func (r *unadoptedResource) Enabled() bool { return false }

func controlledByKRM(object client.Object) bool {
	owner := metav1.GetControllerOf(object)
	if owner == nil {
		return false
	}
	gv, err := schema.ParseGroupVersion(owner.APIVersion)
	return err == nil && gv.Group == kaiv1alpha1.GroupVersion.Group
}
