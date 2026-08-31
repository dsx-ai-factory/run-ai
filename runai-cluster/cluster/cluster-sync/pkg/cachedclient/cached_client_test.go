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
