/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package getters

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetNamespaceByProject returns the namespace that has the runai/queue label of the project.
// the runai/queue label of the namespaces.
func GetNamespaceByProject(cxt context.Context, project string, kubeclient client.Client) (namespace *corev1.Namespace, err error) {
	labelCondition := map[string]string{
		constants.RunaiQueueLabel: project,
	}

	namespaceList := corev1.NamespaceList{}
	err = kubeclient.List(cxt, &namespaceList, client.MatchingLabels(labelCondition))
	if err != nil {
		return nil, err
	}

	if len(namespaceList.Items) == 0 {
		err = apierrors.NewNotFound(corev1.Resource("namespace"),
			"no namespace found to have runai/queue label with the name of the project")
		return nil, err
	}
	if len(namespaceList.Items) != 1 {
		err = fmt.Errorf(
			"multiple namespaces found to have runai/queue label with the name of the project (%s, %s, ...)",
			namespaceList.Items[0].Name, namespaceList.Items[1].Name)
		return nil, err
	}

	return &namespaceList.Items[0], nil
}

// GetNamespaceByKaiProjectObj returns the namespace of a kai.resources project.
func GetNamespaceByKaiProjectObj(ctx context.Context, project *kaiorg.Project, kubeclient client.Client) (*corev1.Namespace, error) {
	if project.Status.Namespace != "" {
		// just to validate the namespace does exist
		namespace, err := getNamespace(ctx, project.Status.Namespace, kubeclient)
		if err != nil {
			return nil, err
		}
		return namespace, nil
	}
	return GetNamespaceByProject(ctx, project.Name, kubeclient)
}

func getNamespacesForKaiProjects(ctx context.Context, projects []kaiorg.Project, kubeclient client.Client) (map[string]bool, error) {
	var err error
	result := map[string]bool{}
	for i := range projects {
		project := projects[i]
		namespace, innerErr := GetNamespaceByKaiProjectObj(ctx, &project, kubeclient)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
			continue
		}
		result[namespace.Name] = true
	}

	return result, err
}

func getNamespace(ctx context.Context, namespaceName string, kubeclient client.Client) (*corev1.Namespace, error) {
	namespace := &corev1.Namespace{}
	err := kubeclient.Get(ctx, client.ObjectKey{Name: namespaceName}, namespace)
	if err != nil {
		return nil, err
	}
	return namespace, nil
}
