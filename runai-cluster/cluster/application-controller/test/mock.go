/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	toolscache "k8s.io/client-go/tools/cache"
	crcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MockCache implements crcache.Cache and delegates Get/List to the provided client
// while exposing MockInformers for event handler registration

type MockCache struct {
	Cli                  client.Client
	Informers            []*MockInformer
	GetInformerErr       error
	WaitForSyncResult    bool
	GetInformerCalled    bool
	GetInformerCallCount int
	GetInformerWithKind  string
}

func (m *MockCache) GetInformer(ctx context.Context, obj client.Object, _ ...crcache.InformerGetOption) (crcache.Informer, error) {
	m.GetInformerCalled = true
	m.GetInformerCallCount++

	var kind string
	switch obj.(type) {
	case *corev1.Secret:
		kind = "Secret"
	default:
		kind = "Unknown"
	}
	m.GetInformerWithKind = kind

	if m.GetInformerErr != nil {
		return nil, m.GetInformerErr
	}

	// Create a new informer for each source type
	informer := &MockInformer{}
	m.Informers = append(m.Informers, informer)

	return informer, nil
}

func (m *MockCache) GetInformerForKind(context.Context, schema.GroupVersionKind, ...crcache.InformerGetOption) (crcache.Informer, error) {
	if len(m.Informers) > 0 {
		return m.Informers[0], nil
	}
	return nil, nil
}
func (m *MockCache) RemoveInformer(context.Context, client.Object) error { return nil }
func (m *MockCache) Start(context.Context) error                         { return nil }
func (m *MockCache) WaitForCacheSync(context.Context) bool               { return m.WaitForSyncResult }
func (m *MockCache) IndexField(context.Context, client.Object, string, client.IndexerFunc) error {
	return nil
}

// Get Delegate Reader methods to fake client
func (m *MockCache) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return m.Cli.Get(ctx, key, obj, opts...)
}

func (m *MockCache) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return m.Cli.List(ctx, list, opts...)
}

// MockInformer implements crcache.Informer with minimal behavior to capture handlers
type MockInformer struct {
	Handler toolscache.ResourceEventHandler
}

func (mi *MockInformer) AddEventHandlerWithOptions(h toolscache.ResourceEventHandler, _ toolscache.HandlerOptions) (toolscache.ResourceEventHandlerRegistration, error) {
	mi.Handler = h
	return nil, nil
}

func (mi *MockInformer) AddEventHandler(h toolscache.ResourceEventHandler) (toolscache.ResourceEventHandlerRegistration, error) {
	mi.Handler = h
	return nil, nil
}

func (mi *MockInformer) AddEventHandlerWithResyncPeriod(h toolscache.ResourceEventHandler, _ time.Duration) (toolscache.ResourceEventHandlerRegistration, error) {
	mi.Handler = h
	return nil, nil
}

func (mi *MockInformer) RemoveEventHandler(toolscache.ResourceEventHandlerRegistration) error {
	return nil
}
func (mi *MockInformer) AddIndexers(toolscache.Indexers) error { return nil }
func (mi *MockInformer) HasSynced() bool                       { return true }
func (mi *MockInformer) IsStopped() bool                       { return false }

// MockObjectResolver implements resolver.ObjectResolver for tests
type MockObjectResolver struct{}

func (m *MockObjectResolver) GetSecretChartName(_ string, _ client.Object) (string, error) {
	return "chart", nil
}

func (m *MockObjectResolver) GetLegacyProject(ctx context.Context, c client.Client, obj client.Object) (*v2.Project, error) {
	// Return a project if it exists in the fake client
	namespace := corev1.Namespace{}
	err := c.Get(ctx, types.NamespacedName{Name: obj.GetNamespace()}, &namespace)
	if err != nil {
		return nil, err
	}

	projectName, hasRunaiQueueLabel := namespace.Labels[constants.RunaiQueueLabel]
	if !hasRunaiQueueLabel || projectName == "" {
		return nil, fmt.Errorf("namespace does not have queue label")
	}

	project := &v2.Project{}
	err = c.Get(ctx, types.NamespacedName{Name: projectName}, project)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", projectName, err)
	}
	return project, nil
}

func (m *MockObjectResolver) GetProject(ctx context.Context, c client.Client, obj client.Object) (*kaiv1alpha1.Project, error) {
	// Return a project if it exists in the fake client
	namespace := corev1.Namespace{}
	err := c.Get(ctx, types.NamespacedName{Name: obj.GetNamespace()}, &namespace)
	if err != nil {
		return nil, err
	}

	projectName, hasRunaiQueueLabel := namespace.Labels[constants.RunaiQueueLabel]
	if !hasRunaiQueueLabel || projectName == "" {
		return nil, fmt.Errorf("namespace does not have queue label")
	}

	project := &kaiv1alpha1.Project{}
	err = c.Get(ctx, types.NamespacedName{Name: projectName}, project)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", projectName, err)
	}
	return project, nil
}

func (m *MockObjectResolver) CreateObjectOwnerReference(c client.Client, obj client.Object) (*metav1.OwnerReference, error) {
	gvk, err := apiutil.GVKForObject(obj, c.Scheme())
	if err != nil {
		return nil, err
	}
	return &metav1.OwnerReference{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Name:       obj.GetName(),
		UID:        obj.GetUID(),
	}, nil
}
