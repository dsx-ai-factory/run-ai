/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var logUtils = logger.Log.WithName("utils.naming")

// NamespacedName compose a namespaced name of a resource, mainly for debugging etc
func NamespacedName(objMeta metav1.ObjectMeta) string {
	if objMeta.Namespace != "" {
		return objMeta.Namespace + ":" + objMeta.Name
	} else {
		return objMeta.Name
	}
}

// ProjectNameOfNamespace derive the project of a given namespace. returns empty string if the namepsace
// is not associated with a project
func ProjectNameOfNamespace(ctx context.Context, k8sClient client.Client, namespace string) (string, error) {

	var result corev1.Namespace

	err := k8sClient.Get(ctx, client.ObjectKey{
		Name: namespace,
	}, &result)
	if err != nil {
		return "", err
	}

	return result.Labels[constants.RunaiQueueLabel], nil
}

func ProjectOfNamespace(ctx context.Context, k8sClient client.Client, namespace string) (*v2.Project, error) {
	projectName, err := ProjectNameOfNamespace(ctx, k8sClient, namespace)
	if err != nil {
		logUtils.Error(err, "failed to get namespace info", "namespace", namespace)
		return nil, err
	}

	return GetProject(ctx, k8sClient, projectName)
}

func GetProject(ctx context.Context, k8sClient client.Client, name string) (*v2.Project, error) {
	project := &v2.Project{}
	projectKey := types.NamespacedName{
		Name: name,
	}

	err := k8sClient.Get(ctx, projectKey, project)
	if err != nil {
		logUtils.Error(err, "Failed getting project", "name", name)
		return nil, err
	}

	return project, nil
}
