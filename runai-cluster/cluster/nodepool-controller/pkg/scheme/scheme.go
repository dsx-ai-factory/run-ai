/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package scheme builds the runtime scheme for nodepool-controller.
//
// Add a group here only when the service actually Gets, Lists or Watches one of
// its kinds — an unregistered kind fails at runtime with
// "no kind is registered for the type ... in scheme".
package scheme

import (
	grovev1alpha1 "github.com/ai-dynamo/grove/operator/api/core/v1alpha1"
	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	kairesourcesv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kaischedulerv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// AddToScheme registers every kind nodepool-controller reads:
//   - core/v1                   Node, Pod
//   - kai.resources/v1alpha1    NodePool, Project, ManagedNodesConfig, Topology
//   - kai/v1                    SchedulingShard
//   - kai/v1alpha1              Topology
//   - scheduling/v2alpha2       PodGroup
//   - monitoring/v1             ServiceMonitor
//   - apiextensions/v1          CustomResourceDefinition
//   - grove/v1alpha1            PodCliqueSet
//   - topology/v1alpha2         NodeResourceTopology
//
// The run.ai Cluster CR is deliberately absent: it is read as unstructured so
// this module does not depend on the proprietary run.ai API types.
func AddToScheme(s *runtime.Scheme) error {
	for _, add := range []func(*runtime.Scheme) error{
		clientgoscheme.AddToScheme,
		kairesourcesv1alpha1.AddToScheme,
		kaiv1.AddToScheme,
		kaischedulerv1alpha1.AddToScheme,
		kaiv2alpha2.AddToScheme,
		monitoringv1.AddToScheme,
		apiextensionsv1.AddToScheme,
		grovev1alpha1.AddToScheme,
		nrtv1alpha2.AddToScheme,
	} {
		if err := add(s); err != nil {
			return err
		}
	}
	return nil
}

// Scheme returns a new scheme with all of the service's kinds registered. It
// panics on failure, which can only happen if a registration function is broken.
func Scheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(AddToScheme(s))
	return s
}
