/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kube_client

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	GetNamespace(namespaceName string) (*v1.Namespace, error)
	GetProject(projectName string) (*unstructured.Unstructured, error)
	GetNodePool(nodePoolName string) (*unstructured.Unstructured, error)
	GetNodePoolsList() (*unstructured.UnstructuredList, error)
	GetRunaiConfigMap() (*v1.ConfigMap, error)
	GetCachedReader() client.Reader
}
