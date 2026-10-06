/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller_utils

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateOrUpdateIfNeeded creates or updates the given resource,
// upon a successful operation, resource is updated
func CreateOrUpdateWithOptionsIfNeeded(ctx context.Context, k8sClient client.Client,
	logger logr.Logger, resource client.Object, createOptions client.CreateOptions, updateOption client.UpdateOptions) error {
	// Creating a new instance of the resource type to allocate
	// new memory for the Get method to populate
	existingResource := reflect.New(reflect.TypeOf(resource).Elem()).Interface().(client.Object)
	found := true

	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
	}, existingResource)
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}

	if found {
		if specsDiffer(existingResource, resource) {
			logger.Info(fmt.Sprintf("updating resource [%s/%s/%s]",
				resource.GetNamespace(),
				resourceStructType(resource),
				resource.GetName()))
			resource.SetResourceVersion(existingResource.GetResourceVersion())
			err = k8sClient.Update(ctx, resource, &updateOption)
			if err != nil {
				return err
			}
		}
	} else {
		logger.Info(fmt.Sprintf("creating resource [%s/%s/%s]",
			resource.GetNamespace(),
			resourceStructType(resource),
			resource.GetName()))
		err = k8sClient.Create(ctx, resource, &createOptions)
		if err != nil {
			return err
		}
	}

	return nil
}

// specsDiffer reports whether the operator should issue an Update by comparing
// only the Spec field using equality.Semantic.DeepEqual. Everything else on
// the object is intentionally ignored because Spec is the only thing the
// operator authoritatively owns and writes via client.Update.
//
// The motivating problem: server-controlled fields (resourceVersion,
// managedFields, status, uid, creationTimestamp) change on every concurrent
// write by other managers — most notably engine-operator's ~1/sec status writes
// to engine-config. A full-object DeepEqual treats those as a diff every
// reconcile, triggers a no-op Update, and the Update races engine-operator's
// next write into a 409 Conflict.
//
// Semantic.DeepEqual (not reflect.DeepEqual) is needed because K8s types like
// resource.Quantity have multiple equivalent internal representations with
// custom Equal() methods that only Semantic honors; engine-config's Spec is
// full of Quantity values, so reflect.DeepEqual would still always diff.
//
// Trade-off worth knowing: writable metadata the operator does set (labels,
// annotations, ownerReferences, finalizers) is also not compared. If any of
// those drift on the live object (manual kubectl edit, third party stripping
// them, etc.) the operator will NOT notice and will NOT restore them on its
// own. If you see operator-managed resources missing the labels / annotations /
// ownerReferences the operator is supposed to set, or carrying finalizers it
// did not add, that is this trade-off surfacing — we should find the root
// cause and re-assess this solution accordingly.
//
// For resource kinds without a Spec field (ConfigMap, Secret, RBAC) we fall
// back to whole-object Semantic.DeepEqual; those kinds lack a fast concurrent
// writer like engine-operator, so the spurious-Update problem is less severe
// there.
func specsDiffer(existing, desired client.Object) bool {
	existingSpec, hasSpec := specField(existing)
	if !hasSpec {
		return !equality.Semantic.DeepEqual(existing, desired)
	}
	desiredSpec, _ := specField(desired)
	return !equality.Semantic.DeepEqual(existingSpec, desiredSpec)
}

// specField returns obj.Spec via reflection, or (nil, false) if the type has no
// Spec field (ConfigMap, Secret, Role, RoleBinding, ServiceAccount, etc.).
func specField(obj client.Object) (interface{}, bool) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, false
	}
	spec := v.FieldByName("Spec")
	if !spec.IsValid() {
		return nil, false
	}
	return spec.Interface(), true
}

func resourceStructType(resource client.Object) string {
	return reflect.TypeOf(resource).Elem().Name()
}
