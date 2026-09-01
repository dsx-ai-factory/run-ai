/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cachedclient

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllertest"
)

type recordingClient struct {
	client.Client
	got bool
}

type optionalLegacyIndexerCache struct {
	runtimecache.Cache
	indexCalls      int
	trainingIndexed bool
}

func (c *optionalLegacyIndexerCache) GetInformer(
	_ context.Context,
	obj client.Object,
	_ ...runtimecache.InformerGetOption,
) (runtimecache.Informer, error) {
	if _, isTrainingWorkload := obj.(*twv2.TrainingWorkload); isTrainingWorkload {
		return nil, errors.New("TrainingWorkload API is not installed")
	}
	return &controllertest.FakeInformer{Synced: true}, nil
}

func (c *optionalLegacyIndexerCache) IndexField(
	_ context.Context,
	obj client.Object,
	_ string,
	_ client.IndexerFunc,
) error {
	c.indexCalls++
	if _, isTrainingWorkload := obj.(*twv2.TrainingWorkload); isTrainingWorkload {
		c.trainingIndexed = true
	}
	return nil
}

func (c *recordingClient) Get(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error {
	c.got = true
	return apierrors.NewNotFound(schema.GroupResource{Group: runaiv1.GroupVersion.Group, Resource: "runaijobs"}, "missing")
}

func TestCachedClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cached Client Suite")
}

var _ = Describe("legacy workload API lifecycle", func() {
	DescribeTable("treats directly watched legacy workloads as optional",
		func(workloadType SupportedType) {
			Expect(isOptionalInformer(workloadType)).To(BeTrue())
		},
		Entry("training workload", TrainingWorkloadType),
		Entry("interactive workload", InteractiveWorkloadType),
		Entry("distributed workload", DistributedWorkloadType),
		Entry("inference workload", InferenceWorkloadType),
		Entry("distributed inference workload", DistributedInferenceWorkloadType),
	)

	DescribeTable("treats directly watched legacy policies as optional",
		func(policyType SupportedType) {
			Expect(isOptionalInformer(policyType)).To(BeTrue())
		},
		Entry("training policy", TrainingPolicyType),
		Entry("interactive policy", InteractivePolicyType),
		Entry("distributed policy", DistributedPolicyType),
		Entry("inference policy", InferencePolicyType),
	)

	DescribeTable("skips only NoMatch errors for optional informer types",
		func(resourceType SupportedType, informerErr error, expected bool) {
			Expect(shouldSkipInformerError(resourceType, informerErr)).To(Equal(expected))
		},
		Entry("optional API missing", InteractivePolicyType, &meta.NoKindMatchError{
			GroupKind:        schema.GroupVersionKind(InteractivePolicyType).GroupKind(),
			SearchedVersions: []string{schema.GroupVersionKind(InteractivePolicyType).Version},
		}, true),
		Entry("optional API forbidden", InteractivePolicyType, apierrors.NewForbidden(
			schema.GroupResource{
				Group:    schema.GroupVersionKind(InteractivePolicyType).Group,
				Resource: "interactivepolicies",
			}, "", errors.New("denied"),
		), false),
		Entry("required API missing", NodeType, &meta.NoKindMatchError{GroupKind: schema.GroupVersionKind(NodeType).GroupKind(), SearchedVersions: []string{schema.GroupVersionKind(NodeType).Version}}, false),
	)

	It("returns an empty list when an optional legacy API has no informer", func() {
		scheme := runtime.NewScheme()
		Expect(twv2.AddToScheme(scheme)).To(Succeed())
		cachedClient := &ClusterSyncCachedClient{
			scheme:              scheme,
			objectInformers:     map[SupportedType]CachedInformer{},
			supportedGroupKinds: map[schema.GroupKind]bool{},
		}

		workloads := &twv2.TrainingWorkloadList{Items: []twv2.TrainingWorkload{{ObjectMeta: metav1.ObjectMeta{Name: "stale"}}}}
		Expect(cachedClient.List(context.Background(), workloads)).To(Succeed())
		Expect(workloads.Items).To(BeEmpty())
	})

	It("uses the uncached client for optional RunaiJob owner parsing", func() {
		writer := &recordingClient{}
		supportedGroupKinds := getInitialSupportedGroupKinds()
		Expect(supportedGroupKinds).NotTo(HaveKey(schema.GroupKind{
			Group: runaiv1.GroupVersion.Group,
			Kind:  "RunaiJob",
		}))
		cachedClient := &ClusterSyncCachedClient{
			Client:              writer,
			objectInformers:     map[SupportedType]CachedInformer{},
			supportedGroupKinds: supportedGroupKinds,
		}
		owner := &metav1.PartialObjectMetadata{}
		owner.SetGroupVersionKind(runaiv1.GroupVersion.WithKind("RunaiJob"))

		err := cachedClient.Get(context.Background(), types.NamespacedName{Namespace: "runai", Name: "missing"}, owner)

		Expect(apierrors.IsNotFound(err)).To(BeTrue())
		Expect(writer.got).To(BeTrue())
	})

	It("adds legacy workload kinds only when their informers exist", func() {
		supportedGroupKinds := getInitialSupportedGroupKinds()

		Expect(supportedGroupKinds).NotTo(HaveKey(schema.GroupKind{
			Group: twv2.GroupVersion.Group,
			Kind:  "TrainingWorkload",
		}))
	})

	It("registers indexes only for installed legacy workload APIs", func() {
		cachedClient := &optionalLegacyIndexerCache{}

		addRunaiWorkloadIndexers(cachedClient)

		Expect(cachedClient.trainingIndexed).To(BeFalse())
		Expect(cachedClient.indexCalls).To(Equal(4))
	})
})

const serverResourceVersion = "server-assigned"

// mutatingWriterClient mimics the real writer client, which decodes the API-server response
// into the object it was handed.
type mutatingWriterClient struct {
	client.Client
	written client.Object
}

func (c *mutatingWriterClient) record(obj client.Object) error {
	obj.SetResourceVersion(serverResourceVersion)
	obj.SetManagedFields(nil)
	c.written = obj
	return nil
}

func (c *mutatingWriterClient) Patch(_ context.Context, obj client.Object, _ client.Patch,
	_ ...client.PatchOption) error {
	return c.record(obj)
}

func (c *mutatingWriterClient) Update(_ context.Context, obj client.Object,
	_ ...client.UpdateOption) error {
	return c.record(obj)
}

func (c *mutatingWriterClient) Status() client.SubResourceWriter {
	return &mutatingSubResourceWriter{owner: c}
}

type mutatingSubResourceWriter struct {
	client.SubResourceWriter
	owner *mutatingWriterClient
}

func (w *mutatingSubResourceWriter) Patch(_ context.Context, obj client.Object, _ client.Patch,
	_ ...client.SubResourcePatchOption) error {
	return w.owner.record(obj)
}

func (w *mutatingSubResourceWriter) Update(_ context.Context, obj client.Object,
	_ ...client.SubResourceUpdateOption) error {
	return w.owner.record(obj)
}

var _ = Describe("writes against informer-owned objects", func() {
	var (
		writer       *mutatingWriterClient
		cachedClient *ClusterSyncCachedClient
		workload     *twv2.TrainingWorkload
	)

	BeforeEach(func() {
		writer = &mutatingWriterClient{}
		cachedClient = &ClusterSyncCachedClient{Client: writer}
		workload = &twv2.TrainingWorkload{ObjectMeta: metav1.ObjectMeta{
			Namespace:       "runai",
			Name:            "cached",
			ResourceVersion: "1",
			Labels:          map[string]string{"project": "team-a"},
			Annotations:     map[string]string{"runai/last-sync": "2026-07-28T00:00:00Z"},
			Finalizers:      []string{"runai.io/external-workload"},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "v1", Kind: "Pod", Name: "owner", UID: "owner-uid",
			}},
			// two entries on purpose: RUN-41759 panicked indexing [1] of this slice mid-DeepCopy
			ManagedFields: []metav1.ManagedFieldsEntry{
				{Manager: "cluster-sync", Operation: metav1.ManagedFieldsOperationUpdate},
				{Manager: "workload-controller", Operation: metav1.ManagedFieldsOperationApply},
			},
		}}
	})

	DescribeTable("never let the API-server response be decoded into the caller's object",
		func(write func(context.Context, client.Object) error) {
			snapshot := workload.DeepCopy()

			Expect(write(context.Background(), workload)).To(Succeed())

			Expect(workload).To(Equal(snapshot))
			Expect(writer.written.GetName()).To(Equal("cached"))
			Expect(writer.written.GetResourceVersion()).To(Equal(serverResourceVersion))
			Expect(writer.written.GetManagedFields()).To(BeEmpty())
		},
		Entry("Patch", func(ctx context.Context, obj client.Object) error {
			return cachedClient.Patch(ctx, obj, client.RawPatch(types.MergePatchType, []byte("{}")))
		}),
		Entry("Update", func(ctx context.Context, obj client.Object) error {
			return cachedClient.Update(ctx, obj)
		}),
		Entry("Status().Patch", func(ctx context.Context, obj client.Object) error {
			return cachedClient.Status().Patch(ctx, obj, client.RawPatch(types.MergePatchType, []byte("{}")))
		}),
		Entry("Status().Update", func(ctx context.Context, obj client.Object) error {
			return cachedClient.Status().Update(ctx, obj)
		}),
	)
})
