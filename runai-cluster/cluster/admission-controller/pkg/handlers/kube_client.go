/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"fmt"
	"sync"

	kube_client "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/kube-client"

	v1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	resourcev1beta1 "k8s.io/api/resource/v1beta1"
	resourcev1beta2 "k8s.io/api/resource/v1beta2"
	schedulingv1 "k8s.io/api/scheduling/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var once sync.Once

var k8srestconfig *rest.Config

func getClientConfig() *rest.Config {
	once.Do(func() {
		// creates the in-cluster config
		var err error
		k8srestconfig, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	})
	return k8srestconfig
}

func initScheme() *runtime.Scheme {
	runtimeScheme := runtime.NewScheme()

	nodePoolsGroupVersion := schema.GroupVersion{Group: kube_client.KaiSchemaGroup, Version: kube_client.NodePoolVersion}
	projectsGroupVersion := schema.GroupVersion{Group: kube_client.KaiSchemaGroup, Version: kube_client.KaiProjectVersion}

	addToSchemeFuncs := []func(s *runtime.Scheme) error{
		v1.AddToScheme,
		resourcev1.AddToScheme,
		resourcev1beta1.AddToScheme,
		resourcev1beta2.AddToScheme,
		schedulingv1.AddToScheme,
		(&scheme.Builder{GroupVersion: nodePoolsGroupVersion}).AddToScheme,
		(&scheme.Builder{GroupVersion: projectsGroupVersion}).AddToScheme,
	}

	for _, addToSchemeFunc := range addToSchemeFuncs {
		err := addToSchemeFunc(runtimeScheme)
		if err != nil {
			panic(fmt.Errorf("could not add types to scheme: %w", err))
		}
	}

	return runtimeScheme
}
