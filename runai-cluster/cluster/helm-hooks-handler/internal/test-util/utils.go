/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test_util

import (
	"context"
	schemeInit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	ra "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	fakeClient "k8s.io/client-go/dynamic/fake"
	"os"
	"sigs.k8s.io/yaml"
)

func AddToDynamicClient(resource schema.GroupVersionResource, obj interface{}, dynamicClient dynamic.Interface) *unstructured.Unstructured {
	data, _ := yaml.Marshal(obj)
	unstructuredData := &unstructured.Unstructured{Object: map[string]interface{}{}}
	_ = yaml.Unmarshal(data, &unstructuredData)
	ctx := context.TODO()
	dynamicResource := dynamicClient.Resource(resource).Namespace(runv1.DefaultNamespace)
	unstructuredData, _ = dynamicResource.Create(ctx, unstructuredData, metav1.CreateOptions{})
	return unstructuredData
}

func AddYamlRunaiConf(dynamicClient dynamic.Interface, yamlFilePath string) *unstructured.Unstructured {
	yamlFile, err := os.ReadFile(yamlFilePath)
	if err != nil {
		panic(err)
	}
	obj := &unstructured.Unstructured{Object: map[string]interface{}{}}
	yaml.Unmarshal(yamlFile, &obj)
	ctx := context.TODO()
	dynamicResource := dynamicClient.Resource(ra.RunaiConfigResource).Namespace(runv1.DefaultNamespace)
	_, _ = dynamicResource.Create(ctx, obj, metav1.CreateOptions{})
	return obj
}

func NewTestDynamicClient(createRunaiConfig bool, createCM bool) dynamic.Interface {
	runaiSchema := schemeInit.Scheme()
	apiextensionv1beta1.AddToScheme(runaiSchema)
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "runai",
		},
	}

	objects := []runtime.Object{&namespace}
	crd := &apiextensionv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "runaiconfigs.run.ai", Namespace: "runai"},
		Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
			Group: runv1.Group,
			Scope: apiextensionv1beta1.NamespaceScoped,
			Names: apiextensionv1beta1.CustomResourceDefinitionNames{
				Plural: "runaiconfigs",
				Kind:   runv1.RunaiConfigKind,
			},
			Versions: []apiextensionv1beta1.CustomResourceDefinitionVersion{
				{Name: runv1.Version, Served: true},
			},
		},
	}
	objects = append(objects, crd)

	client := fakeClient.NewSimpleDynamicClient(runaiSchema, objects...)

	if createRunaiConfig {
		runaiConf := &runv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: runv1.DefaultNamespace,
				Name:      runv1.DefaultName,
			},
		}
		runaiConf.SetDefaultsWhereNeeded()
		AddToDynamicClient(ra.RunaiConfigResource, runaiConf, client)
	}
	if createCM {
		cm := v1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runai-public",
				Namespace: "runai",
			},
			Data: map[string]string{
				"aa": "aa",
				"ab": "ab",
			},
		}
		cmResource := schema.GroupVersionResource{
			Group:    "",
			Version:  runv1.Version,
			Resource: "configmaps",
		}
		AddToDynamicClient(cmResource, cm, client)
	}
	return client
}
