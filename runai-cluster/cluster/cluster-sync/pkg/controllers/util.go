/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/util"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func objectKeyFromObject(object client.Object) client.ObjectKey {
	return client.ObjectKey{
		Name:      object.GetName(),
		Namespace: object.GetNamespace(),
	}
}
