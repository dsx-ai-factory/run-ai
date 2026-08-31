/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package fake

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fakeClient struct {
	writer          client.Client
	cachedReader    client.Reader
	uncachedReader  client.Reader
	discoveryClient discovery.DiscoveryInterface
}

func New(cachedClient client.Client, uncachedReader client.Reader, discoveryClient discovery.DiscoveryInterface) *fakeClient {
	return &fakeClient{
		writer:          cachedClient,
		cachedReader:    cachedClient.(client.Reader),
		uncachedReader:  uncachedReader,
		discoveryClient: discoveryClient,
	}
}

func (u *fakeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return u.cachedReader.Get(ctx, key, obj, opts...)
}

func (u *fakeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	listOpts := &client.ListOptions{}

	for _, opt := range opts {
		opt.ApplyToList(listOpts)
	}

	return u.cachedReader.List(ctx, list, opts...)
}

// Create saves the object obj in the Kubernetes cluster.
func (u *fakeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return u.writer.Create(ctx, obj, opts...)
}

// Delete deletes the given obj from Kubernetes cluster.
func (u *fakeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return u.writer.Delete(ctx, obj, opts...)
}

// Update updates the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (u *fakeClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return u.writer.Update(ctx, obj, opts...)
}

// Patch patches the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (u *fakeClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return u.writer.Patch(ctx, obj, patch, opts...)
}

// Apply applies the given apply configuration to the Kubernetes cluster.
func (u *fakeClient) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.ApplyOption) error {
	return u.writer.Apply(ctx, obj, opts...)
}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (u *fakeClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return u.writer.DeleteAllOf(ctx, obj, opts...)
}

func (u *fakeClient) Status() client.StatusWriter {
	return u.writer.Status()
}

// Scheme returns the scheme this client is using.
func (u *fakeClient) Scheme() *runtime.Scheme {
	return u.writer.Scheme()
}

// RESTMapper returns the rest this client is using.
func (u *fakeClient) RESTMapper() meta.RESTMapper {
	return u.writer.RESTMapper()
}

func (u *fakeClient) SubResource(subResource string) client.SubResourceClient {
	return u.writer.SubResource(subResource)
}

func (u *fakeClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return u.writer.GroupVersionKindFor(obj)
}

func (u *fakeClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return u.writer.IsObjectNamespaced(obj)
}

// Uncached returns a client that does not cache any resources.
// All operations are performed on the API server directly.
func (u *fakeClient) Uncached() client.Client {
	return &fakeClient{
		writer:         u.writer,
		cachedReader:   u.cachedReader,
		uncachedReader: u.uncachedReader,
	}
}

func (u *fakeClient) NewDiscoveryClient() (discovery.DiscoveryInterface, error) {
	return u.discoveryClient, nil
}
