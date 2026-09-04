/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgocache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// isRelevantResourceScope checks if a secret/pvc is original or replicated
// we need to know it because we do not synchronize replicates.
// it is done so by matching the namespace with the scope labels
func isRelevantResourceScope(cachedReader client.Reader, objectMeta *metav1.ObjectMeta) bool {
	if util.IsClusterOrDeptScopedResource(objectMeta) {
		// cluster/department scope - are relevant only in runai namespace
		return objectMeta.Namespace == runaiv1.DefaultNamespace
	}

	// project scope - relevant only in project namespace
	return util.IsRunaiManagedNamespace(objectMeta.Namespace, cachedReader)
}

func isUserCredential(objectMeta metav1.ObjectMeta) bool {
	// we do not sync the user-credential secrets created in the project namespaces
	if objectMeta.Namespace != runaiv1.DefaultNamespace {
		return false
	}

	userCredential, found := objectMeta.Labels[commonconstants.UserCredentialsLabel]
	if !found {
		return false
	}

	// user credentials are relevant only in runai namespace, we don't want to handle the ones created in the project namespaces
	return userCredential == util.TrueAsString && objectMeta.Namespace == runaiv1.DefaultNamespace
}

func genericControllerErrorHandler(err error, operation constants.OperationName, resource constants.ClusterResource, object client.Object) {
	resourceName := ""
	if object != nil {
		resourceName = fmt.Sprintf("%s|%s", object.GetNamespace(), object.GetName())
	}
	if err != nil {
		zap.S().Errorf("Had an error while trying to %s %s <%s>: %v \n", operation, resource, resourceName, err)
		metrics.GetMetricsExporter().AddErrorCounter(err, resource)
	}
}

// filteredObject unwraps a delete tombstone for FilteringResourceEventHandler.FilterFunc,
// which runs before DeleteFunc and would otherwise reject or panic on one.
func filteredObject(obj any) any {
	if tombstone, ok := obj.(clientgocache.DeletedFinalStateUnknown); ok {
		return tombstone.Obj
	}
	return obj
}

// deletedObject unwraps the object of a delete event. The informer delivers a
// DeletedFinalStateUnknown tombstone instead of the object when it missed the deletion.
func deletedObject[T client.Object](obj any) (T, bool) {
	if typed, ok := obj.(T); ok {
		return typed, true
	}

	var zero T
	tombstone, ok := obj.(clientgocache.DeletedFinalStateUnknown)
	if !ok {
		zap.S().Errorf("Got a delete event for an unexpected object of type %T, wanted %T", obj, zero)
		return zero, false
	}

	typed, ok := tombstone.Obj.(T)
	if !ok {
		zap.S().Errorf("Got a delete tombstone <%s> wrapping an unexpected object of type %T, wanted %T",
			tombstone.Key, tombstone.Obj, zero)
		return zero, false
	}
	return typed, true
}

func objectKeyFromObject(object client.Object) client.ObjectKey {
	return client.ObjectKey{
		Name:      object.GetName(),
		Namespace: object.GetNamespace(),
	}
}
