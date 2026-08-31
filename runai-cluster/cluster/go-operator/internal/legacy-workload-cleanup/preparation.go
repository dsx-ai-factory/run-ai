/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"context"
	"encoding/json"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
)

func prepareWorkload(ctx context.Context, dynamicClient dynamic.Interface, descriptor resourceDescriptor,
	workload *unstructured.Unstructured,
) error {
	if !descriptor.RemoveWorkloadFinalizer {
		return nil
	}
	return removeObjectFinalizer(ctx, dynamicClient, descriptor.GVR, workload, legacyWorkloadFinalizer)
}

func removeObjectFinalizer(ctx context.Context, dynamicClient dynamic.Interface, gvr schema.GroupVersionResource,
	object *unstructured.Unstructured, allowlistedFinalizer string,
) error {
	current := object.DeepCopy()
	refresh := false
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if refresh {
			latest, getErr := dynamicClient.Resource(gvr).Namespace(object.GetNamespace()).Get(
				ctx, object.GetName(), metav1.GetOptions{})
			if apierrors.IsNotFound(getErr) {
				return nil
			}
			if getErr != nil {
				return getErr
			}
			if latest.GetUID() != object.GetUID() {
				return fmt.Errorf("%s %s/%s UID changed from %q to %q",
					object.GetKind(), object.GetNamespace(), object.GetName(), object.GetUID(), latest.GetUID())
			}
			current = latest
		}
		remaining, changed := withoutFinalizer(current.GetFinalizers(), allowlistedFinalizer)
		if !changed {
			return nil
		}
		operations := []map[string]any{
			{"op": "test", "path": "/metadata/uid", "value": string(current.GetUID())},
			{"op": "test", "path": "/metadata/resourceVersion", "value": current.GetResourceVersion()},
			{"op": "replace", "path": "/metadata/finalizers", "value": remaining},
		}
		patch, marshalErr := json.Marshal(operations)
		if marshalErr != nil {
			return marshalErr
		}
		_, patchErr := dynamicClient.Resource(gvr).Namespace(object.GetNamespace()).Patch(
			ctx, object.GetName(), types.JSONPatchType, patch, metav1.PatchOptions{})
		if apierrors.IsNotFound(patchErr) {
			return nil
		}
		refresh = true
		return patchErr
	})
	if err != nil {
		return fmt.Errorf("remove allowlisted finalizer %q from %s %s/%s: %w",
			allowlistedFinalizer, object.GetKind(), object.GetNamespace(), object.GetName(), err)
	}
	return nil
}
