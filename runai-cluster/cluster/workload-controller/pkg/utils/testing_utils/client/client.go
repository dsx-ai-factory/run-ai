/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	safeFake "github.com/run-ai/runai/runai-cluster/cluster/common/utils/client"
	schedulingv1 "k8s.io/api/scheduling/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func InitFakeClientWithObjectNoStatus(scheme *runtime.Scheme, objects ...client.Object) client.Client {
	return InitFakeClientWithObjectWithStatus(scheme, nil, objects...)
}

func InitFakeClientWithObjectWithStatus(scheme *runtime.Scheme, statusObjects []client.Object, objects ...client.Object) client.Client {
	objects = addBasicPriorityClassesToFakeClient(scheme, objects)

	// The controller-runtime fake client strips TypeMeta (apiVersion/kind) from objects on
	// Create/Get/List/Update. Production code (and test rule matching) dispatches on the
	// object's Kind via GetObjectKind().GroupVersionKind().Kind, so we backfill the GVK from
	// the scheme after each call to keep the Kind populated, matching pre-upgrade behaviour.
	testClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).
		WithStatusSubresource(statusObjects...).
		WithInterceptorFuncs(gvkBackfillInterceptor(scheme)).Build()

	safeClient := safeFake.NewFakeClientConcurrentSafe(testClient)

	// Wrap with Event UID index wrapper
	return &FakeClientWithEventUIDIndex{
		Client: safeClient,
	}
}

func addBasicPriorityClassesToFakeClient(scheme *runtime.Scheme, objects []client.Object) []client.Object {
	utilruntime.Must(schedulingv1.AddToScheme(scheme))

	priorityClasses := []client.Object{
		&schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.InteractivePreemptiblePriorityClass,
			},
			Value: 75,
		},
		&schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.InteractivePriorityClass,
			},
			Value: 100,
		},
		&schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "low",
			},
			Value: 50, // doesn't matter for the purpose of the test
		},
		&schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "high",
			},
			Value: 125,
		},
		&schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "medium",
			},
			Value: 75,
		},
	}

	return append(objects, priorityClasses...)
}

// backfillGVK restores the object's GVK from the scheme if it was stripped (empty Kind).
func backfillGVK(scheme *runtime.Scheme, obj runtime.Object) {
	if obj == nil {
		return
	}
	if obj.GetObjectKind().GroupVersionKind().Kind != "" {
		return
	}
	if gvks, _, err := scheme.ObjectKinds(obj); err == nil && len(gvks) > 0 {
		obj.GetObjectKind().SetGroupVersionKind(gvks[0])
	}
}

// gvkBackfillInterceptor restores TypeMeta stripped by the fake client after Create/Update/Get/List.
func gvkBackfillInterceptor(scheme *runtime.Scheme) interceptor.Funcs {
	return interceptor.Funcs{
		Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			err := c.Create(ctx, obj, opts...)
			backfillGVK(scheme, obj)
			return err
		},
		Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
			err := c.Update(ctx, obj, opts...)
			backfillGVK(scheme, obj)
			return err
		},
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			err := c.Get(ctx, key, obj, opts...)
			backfillGVK(scheme, obj)
			return err
		},
		List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
			err := c.List(ctx, list, opts...)
			if err == nil {
				_ = meta.EachListItem(list, func(o runtime.Object) error {
					backfillGVK(scheme, o)
					return nil
				})
			}
			return err
		},
	}
}
