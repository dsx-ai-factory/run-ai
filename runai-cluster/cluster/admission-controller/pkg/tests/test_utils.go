/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"

	kube_client "github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/kube-client"
	runaitypes "github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

type FakeK8sClient struct {
	k8sReader client.Reader
}

func NewClient(k8sReader client.Reader) *FakeK8sClient {
	return &FakeK8sClient{
		k8sReader: k8sReader,
	}
}

func (f *FakeK8sClient) GetNamespace(namespaceName string) (*v1.Namespace, error) {
	namespace := &v1.Namespace{}
	err := f.k8sReader.Get(context.Background(), types.NamespacedName{Name: namespaceName}, namespace)
	if err != nil {
		return nil, err
	}
	return namespace, nil
}

func (f *FakeK8sClient) GetProject(projectName string) (*unstructured.Unstructured, error) {
	project := &unstructured.Unstructured{}
	project.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   kube_client.KaiSchemaGroup,
		Version: kube_client.KaiProjectVersion,
		Kind:    "Project",
	})
	err := f.k8sReader.Get(context.Background(), types.NamespacedName{Name: projectName}, project)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (f *FakeK8sClient) GetNodePool(nodePoolName string) (*unstructured.Unstructured, error) {
	nodePool := &unstructured.Unstructured{}
	nodePool.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   kube_client.KaiSchemaGroup,
		Version: kube_client.NodePoolVersion,
		Kind:    "NodePool",
	})
	err := f.k8sReader.Get(context.Background(), types.NamespacedName{Name: nodePoolName}, nodePool)
	if err != nil {
		return nil, err
	}
	return nodePool, nil
}

func (f *FakeK8sClient) GetNodePoolsList() (*unstructured.UnstructuredList, error) {
	nodePoolList := &unstructured.UnstructuredList{}
	nodePoolList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   kube_client.KaiSchemaGroup,
		Version: kube_client.NodePoolVersion,
		Kind:    "NodePoolList",
	})
	err := f.k8sReader.List(context.Background(), nodePoolList)
	if err != nil {
		return nil, err
	}
	return nodePoolList, nil
}

func (f *FakeK8sClient) GetRunaiConfigMap() (*v1.ConfigMap, error) {
	configMap := &v1.ConfigMap{}
	err := f.k8sReader.Get(context.Background(),
		types.NamespacedName{Name: "default-limit-range", Namespace: runaitypes.RunaiNamespace},
		configMap)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

func (f *FakeK8sClient) GetCachedReader() client.Reader {
	return f.k8sReader
}

func InitScheme() *runtime.Scheme {
	runtimeScheme := runtime.NewScheme()

	nodePoolsGroupVersion := schema.GroupVersion{Group: kube_client.KaiSchemaGroup, Version: kube_client.NodePoolVersion}
	projectsGroupVersion := schema.GroupVersion{Group: kube_client.KaiSchemaGroup, Version: kube_client.KaiProjectVersion}

	f := []func(s *runtime.Scheme) error{
		v1.AddToScheme,
		(&scheme.Builder{GroupVersion: nodePoolsGroupVersion}).AddToScheme,
		(&scheme.Builder{GroupVersion: projectsGroupVersion}).AddToScheme,
	}

	for _, f := range f {
		err := f(runtimeScheme)
		if err != nil {
			panic(fmt.Errorf("could not add types to scheme: %w", err))
		}
	}

	return runtimeScheme
}
