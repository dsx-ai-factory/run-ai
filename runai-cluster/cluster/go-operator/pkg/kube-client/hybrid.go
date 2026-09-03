/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kube_client

import (
	"context"
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type hybridClient struct {
	writer                client.Client
	cachedReader          client.Reader
	uncachedReader        client.Reader
	discoveryClientConfig *rest.Config
}

var _ client.Client = &hybridClient{}

// New returns a hybrid client.
// A hybrid client is a client that caches all resoureces in the specified namespace, but does not
// cache resources in the cluster-scope or in different namespcaes.
// The motivation to use a hybrid client is to reduce the RBAC footprint of the operator
// since cached clients require hard coded RBAC permissions that cannot be modified.
// This is a temporary solution until a fully cached client will be implemented.
// discoveryClientConfig [optional] is used to create a discovery client that can be used to discover resources in the cluster.
// if discoveryClientConfig is nil, the discovery client will not be created.

func New(cachedClient client.Client, uncachedReader client.Reader,
	discoveryClientConfig *rest.Config) *hybridClient {
	return &hybridClient{
		writer:                cachedClient,
		cachedReader:          cachedClient.(client.Reader),
		uncachedReader:        uncachedReader,
		discoveryClientConfig: discoveryClientConfig,
	}
}

// Leases are read uncached: a cached Get would start a Lease informer needing
// list/watch RBAC, but only get/delete are granted (orphan-Lease cleanup).
func readUncached(obj runtime.Object) bool {
	switch obj.(type) {
	case *coordinationv1.Lease, *coordinationv1.LeaseList:
		return true
	default:
		return false
	}
}

func (u *hybridClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if readUncached(obj) {
		return u.uncachedReader.Get(ctx, key, obj, opts...)
	}
	return u.cachedReader.Get(ctx, key, obj, opts...)
}

func (u *hybridClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	listOpts := &client.ListOptions{}

	for _, opt := range opts {
		opt.ApplyToList(listOpts)
	}

	if listOpts.Limit != 0 {
		return fmt.Errorf("pagination is not supported for cached clients")
	}

	if readUncached(list) {
		return u.uncachedReader.List(ctx, list, opts...)
	}

	return u.cachedReader.List(ctx, list, opts...)
}

// Create saves the object obj in the Kubernetes cluster.
func (u *hybridClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return u.writer.Create(ctx, obj, opts...)
}

// Delete deletes the given obj from Kubernetes cluster.
func (u *hybridClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if len(obj.GetAnnotations()) > 0 &&
		obj.GetAnnotations()[runaiv1.DeletionPolicyAnnotationKey] == string(runaiv1.DeletionPolicyRetain) {
		logger := log.FromContext(ctx)
		logger.Info(fmt.Sprintf("resource has %s:%s, skipping deletion",
			runaiv1.DeletionPolicyAnnotationKey,
			runaiv1.DeletionPolicyRetain))

		return nil
	}

	return u.writer.Delete(ctx, obj, opts...)
}

// Update updates the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (u *hybridClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if len(obj.GetAnnotations()) > 0 &&
		obj.GetAnnotations()[runaiv1.UpdatePolicyAnnotationKey] == string(runaiv1.UpdatePolicySkip) {
		logger := log.FromContext(ctx)
		logger.Info(fmt.Sprintf("resource has %s:%s, skipping update",
			runaiv1.UpdatePolicyAnnotationKey,
			runaiv1.UpdatePolicySkip))

		return nil
	}
	return u.writer.Update(ctx, obj, opts...)
}

// Patch patches the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (u *hybridClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return u.writer.Patch(ctx, obj, patch, opts...)
}

// Apply applies the given apply configuration to the Kubernetes cluster.
func (u *hybridClient) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.ApplyOption) error {
	return u.writer.Apply(ctx, obj, opts...)
}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (u *hybridClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return u.writer.DeleteAllOf(ctx, obj, opts...)
}

func (u *hybridClient) Status() client.StatusWriter {
	return u.writer.Status()
}

// Scheme returns the scheme this client is using.
func (u *hybridClient) Scheme() *runtime.Scheme {
	return u.writer.Scheme()
}

// RESTMapper returns the rest this client is using.
func (u *hybridClient) RESTMapper() meta.RESTMapper {
	return u.writer.RESTMapper()
}

func (u *hybridClient) SubResource(subResource string) client.SubResourceClient {
	return u.writer.SubResource(subResource)
}

func (u *hybridClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return u.writer.GroupVersionKindFor(obj)
}

func (u *hybridClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return u.writer.IsObjectNamespaced(obj)
}

// Uncached returns a client that does not cache any resources.
// All operations are performed on the API server directly.
func (u *hybridClient) Uncached() client.Client {
	return &hybridClient{
		writer:         u.writer,
		cachedReader:   u.uncachedReader,
		uncachedReader: u.uncachedReader,
	}
}

func (u *hybridClient) NewDiscoveryClient() (discovery.DiscoveryInterface, error) {
	if u.discoveryClientConfig == nil {
		return nil, fmt.Errorf("discovery client config is nil")
	}
	return discovery.NewDiscoveryClientForConfig(u.discoveryClientConfig)
}
