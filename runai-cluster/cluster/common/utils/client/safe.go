/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//		fakeClientConcurrentSafe wrapper to client.Client with enforced lock
//	 do not use this in production, only for testing
type fakeClientConcurrentSafe struct {
	innerClient client.Client
	mu          sync.Mutex
}

func (f *fakeClientConcurrentSafe) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerClient.Get(ctx, key, obj, opts...)
}

func (f *fakeClientConcurrentSafe) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerClient.List(ctx, list, opts...)
}

func (f *fakeClientConcurrentSafe) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerClient.Create(ctx, obj, opts...)
}

func (f *fakeClientConcurrentSafe) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerClient.Delete(ctx, obj, opts...)
}

func (f *fakeClientConcurrentSafe) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerClient.Update(ctx, obj, opts...)
}

func (f *fakeClientConcurrentSafe) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerClient.Patch(ctx, obj, patch, opts...)
}

func (f *fakeClientConcurrentSafe) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerClient.DeleteAllOf(ctx, obj, opts...)
}

func (f *fakeClientConcurrentSafe) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.ApplyOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerClient.Apply(ctx, obj, opts...)
}

func (f *fakeClientConcurrentSafe) Status() client.SubResourceWriter {
	f.mu.Lock()
	defer f.mu.Unlock()
	return NewFakeSubResourceWriterConcurrentSafe(f.innerClient.Status())
}

func (f *fakeClientConcurrentSafe) SubResource(subResource string) client.SubResourceClient {
	f.mu.Lock()
	defer f.mu.Unlock()
	return NewFakeSubResourceClientConcurrentSafe(f.innerClient.SubResource(subResource))
}

func (f *fakeClientConcurrentSafe) Scheme() *runtime.Scheme {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerClient.Scheme()
}

func (f *fakeClientConcurrentSafe) RESTMapper() meta.RESTMapper {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerClient.RESTMapper()
}

func (f *fakeClientConcurrentSafe) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerClient.GroupVersionKindFor(obj)
}

func (f *fakeClientConcurrentSafe) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerClient.IsObjectNamespaced(obj)
}

func NewFakeClientConcurrentSafe(innerClient client.Client) client.Client {
	return &fakeClientConcurrentSafe{innerClient: innerClient}
}

type fakeSubResourceWriterConcurrentSafe struct {
	mu               sync.Mutex
	innerSubResource client.SubResourceWriter
}

func (f *fakeSubResourceWriterConcurrentSafe) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerSubResource.Create(ctx, obj, subResource, opts...)
}

func (f *fakeSubResourceWriterConcurrentSafe) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerSubResource.Update(ctx, obj, opts...)
}

func (f *fakeSubResourceWriterConcurrentSafe) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerSubResource.Patch(ctx, obj, patch, opts...)
}

func (f *fakeSubResourceWriterConcurrentSafe) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.SubResourceApplyOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerSubResource.Apply(ctx, obj, opts...)
}

type fakeSubResourceClientConcurrentSafe struct {
	fakeSubResourceWriterConcurrentSafe
	innerSubResourceReader client.SubResourceReader
}

func (f *fakeSubResourceClientConcurrentSafe) Get(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceGetOption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.innerSubResourceReader.Get(ctx, obj, subResource, opts...)

}

func NewFakeSubResourceClientConcurrentSafe(innerSubResourceClient client.SubResourceClient) client.SubResourceClient {
	return &fakeSubResourceClientConcurrentSafe{
		fakeSubResourceWriterConcurrentSafe: fakeSubResourceWriterConcurrentSafe{innerSubResource: innerSubResourceClient},
		innerSubResourceReader:              innerSubResourceClient,
	}
}

func NewFakeSubResourceWriterConcurrentSafe(innerSubResourceWriter client.SubResourceWriter) client.SubResourceWriter {
	return &fakeSubResourceWriterConcurrentSafe{innerSubResource: innerSubResourceWriter}
}
