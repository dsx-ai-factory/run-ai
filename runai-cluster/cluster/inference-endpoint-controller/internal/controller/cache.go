/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/inference-endpoint-controller/internal/render"
)

// ChildCacheOptions limits the manager cache to the objects this controller
// renders. Every child carries render.EndpointLabel, and without a selector the
// controller would hold every Deployment, Service, ConfigMap, ServiceAccount, Role,
// RoleBinding and NetworkPolicy of the cluster in memory, which can exceed the
// operand's memory limit on a large cluster. Children are only read through the
// API reader or written with server-side apply, so objects without the label are
// never needed from the cache.
func ChildCacheOptions() cache.Options {
	requirement, err := labels.NewRequirement(render.EndpointLabel, selection.Exists, nil)
	runtime.Must(err)
	byEndpoint := cache.ByObject{Label: labels.NewSelector().Add(*requirement)}

	pool := &unstructured.Unstructured{}
	pool.SetGroupVersionKind(poolGVK)
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(routeGVK)

	byObject := map[client.Object]cache.ByObject{}
	for _, obj := range []client.Object{
		pool, route,
		&appsv1.Deployment{}, &corev1.Service{}, &corev1.ConfigMap{}, &corev1.ServiceAccount{},
		&rbacv1.Role{}, &rbacv1.RoleBinding{}, &networkingv1.NetworkPolicy{},
	} {
		byObject[obj] = byEndpoint
	}
	return cache.Options{ByObject: byObject}
}
