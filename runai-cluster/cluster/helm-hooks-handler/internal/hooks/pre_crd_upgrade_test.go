/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package hooks_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	schemeInit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	ra "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	pre_hooks "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks/pre-hooks"
	testu "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/test-util"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var legacyEngineSchedulingShardGVR = schema.GroupVersionResource{
	Group: "engine.run.ai", Version: "v1alpha1", Resource: "schedulingshards",
}

var _ = Describe("PreCrdUpgradeHook", func() {
	var (
		ctx            context.Context
		cancel         context.CancelFunc
		clientsC       *clients.Clients
		params         *hooks.ClusterHookParams
		runaiconfigObj *unstructured.Unstructured
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 500*time.Millisecond)

		scheme := schemeInit.Scheme()
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		dClient := testu.NewTestDynamicClient(false, false)
		engineShards := make([]runtime.Object, 0, 3)
		for _, shardName := range []string{"shard1", "shard2", "shard3"} {
			schedulingShard := &unstructured.Unstructured{}
			schedulingShard.SetAPIVersion("engine.run.ai/v1alpha1")
			schedulingShard.SetKind("SchedulingShard")
			schedulingShard.SetName(shardName)
			engineShards = append(engineShards, schedulingShard)
		}
		remoteDynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{legacyEngineSchedulingShardGVR: "SchedulingShardList"},
			engineShards...,
		)
		clientsC = &clients.Clients{
			LocalClusterClient:         k8sClient,
			LocalClusterDynamicClient:  dClient,
			RemoteClusterClient:        k8sClient,
			RemoteClusterDynamicClient: remoteDynamicClient,
		}
		params = &hooks.ClusterHookParams{
			OperandDownIntervalMillis: 10,
		}
		// Always create the runai namespace
		Expect(k8sClient.Create(ctx, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: runaiv1.DefaultNamespace}})).To(Succeed())

		// create a queue CRD with the not upgraded version
		queueCrd := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "queues.scheduling.run.ai",
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name:    "v2alpha2",
						Storage: true,
					},
					{
						Name:    "v2",
						Storage: false,
					},
				}},
		}
		Expect(clientsC.LocalClusterClient.Create(ctx, queueCrd)).To(Succeed())
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Context("runaiconfig exists", func() {
		BeforeEach(func() {
			// Add a RunaiConfig with nodepool-controller enabled
			runaiconfigObj = testu.AddYamlRunaiConf(clientsC.LocalClusterDynamicClient, "../test-util/test_resources/runaiconfig2_22_nodepool_controller_enabled.yaml")
		})

		It("should succeed when no nodepool-controller pods exist and disable nodepool-controller in runaiconfig", func() {
			hook := pre_hooks.NewPreCrdUpgradeHook(params)
			err := hook.Call(ctx, clientsC)
			Expect(err).ToNot(HaveOccurred())

			dynamicResource := clientsC.LocalClusterDynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
			val, _ := dynamicResource.Get(ctx, runaiconfigObj.GetName(), metav1.GetOptions{})
			spec := val.Object["spec"].(map[string]interface{})
			runaiNPCConf := spec["nodepool-controller"].(map[string]interface{})
			enabled, ok := runaiNPCConf["enabled"].(bool)
			Expect(ok).To(BeTrue())
			Expect(enabled).To(BeFalse(), "nodepool-controller should be disabled after pre-CRD upgrade")
		})

		It("should fail (timeout) when a nodepool-controller pod exists and not enable nodepool-controller in runaiconfig", func() {
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodepool-controller-0",
					Namespace: runaiv1.DefaultNamespace,
					Labels:    map[string]string{"app": "nodepool-controller"},
				},
			}
			Expect(clientsC.LocalClusterClient.Create(ctx, pod)).To(Succeed())
			hook := pre_hooks.NewPreCrdUpgradeHook(params)
			start := time.Now()
			err := hook.Call(ctx, clientsC)
			dur := time.Since(start)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("timeout reached"))
			Expect(dur).To(BeNumerically("<", time.Second)) // Should be fast

			dynamicResource := clientsC.LocalClusterDynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
			val, _ := dynamicResource.Get(ctx, runaiconfigObj.GetName(), metav1.GetOptions{})
			spec := val.Object["spec"].(map[string]interface{})
			runaiNPCConf := spec["nodepool-controller"].(map[string]interface{})
			enabled, ok := runaiNPCConf["enabled"].(bool)
			Expect(ok).To(BeTrue())
			Expect(enabled).To(BeFalse(), "nodepool-controller should be disabled after pre-CRD upgrade, even if timeout")
		})

		It("should fail (timeout) when a runai-scheduler pod exists and not enable nodepool-controller in runaiconfig", func() {
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai-scheduler-0",
					Namespace: runaiv1.DefaultNamespace,
					Labels:    map[string]string{"app": "runai-scheduler"},
				},
			}
			Expect(clientsC.RemoteClusterClient.Create(ctx, pod)).To(Succeed())
			hook := pre_hooks.NewPreCrdUpgradeHook(params)
			start := time.Now()
			err := hook.Call(ctx, clientsC)
			dur := time.Since(start)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("timeout reached"))
			Expect(dur).To(BeNumerically("<", time.Second)) // Should be fast

			dynamicResource := clientsC.LocalClusterDynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
			val, _ := dynamicResource.Get(ctx, runaiconfigObj.GetName(), metav1.GetOptions{})
			spec := val.Object["spec"].(map[string]interface{})
			runaiNPCConf := spec["nodepool-controller"].(map[string]interface{})
			enabled, ok := runaiNPCConf["enabled"].(bool)
			Expect(ok).To(BeTrue())
			Expect(enabled).To(BeFalse(), "nodepool-controller should be disabled after pre-CRD upgrade, even if timeout")
		})

		It("should succeed when only non-nodepool-controller pods exist and disable nodepool-controller in runaiconfig", func() {
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "other-pod",
					Namespace: runaiv1.DefaultNamespace,
					Labels:    map[string]string{"app": "not-a-nodepool-controller"},
				},
			}
			Expect(clientsC.LocalClusterClient.Create(ctx, pod)).To(Succeed())
			hook := pre_hooks.NewPreCrdUpgradeHook(params)
			err := hook.Call(ctx, clientsC)
			Expect(err).ToNot(HaveOccurred())

			dynamicResource := clientsC.LocalClusterDynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
			val, _ := dynamicResource.Get(ctx, runaiconfigObj.GetName(), metav1.GetOptions{})
			spec := val.Object["spec"].(map[string]interface{})
			runaiNPCConf := spec["nodepool-controller"].(map[string]interface{})
			enabled, ok := runaiNPCConf["enabled"].(bool)
			Expect(ok).To(BeTrue())
			Expect(enabled).To(BeFalse(), "nodepool-controller should be disabled after pre-CRD upgrade")
		})
	})

	Context("runaiconfig does not exist", func() {
		It("should succeed when no nodepool-controller pods exist and no runaiconfig", func() {
			hook := pre_hooks.NewPreCrdUpgradeHook(params)
			err := hook.Call(ctx, clientsC)
			Expect(err).ToNot(HaveOccurred())

			// Optionally: check that no RunaiConfig was created
			dynamicResource := clientsC.LocalClusterDynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
			_, err = dynamicResource.Get(ctx, "runai", metav1.GetOptions{})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("queue crd is already upgraded / doesnt exist", func() {
		BeforeEach(func() {
			// Add a RunaiConfig with nodepool-controller enabled
			runaiconfigObj = testu.AddYamlRunaiConf(clientsC.LocalClusterDynamicClient, "../test-util/test_resources/runaiconfig2_22_nodepool_controller_enabled.yaml")

			// create a nodepool-controller pod to simulate the scenario
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodepool-controller-0",
					Namespace: runaiv1.DefaultNamespace,
					Labels:    map[string]string{"app": "nodepool-controller"},
				},
			}
			Expect(clientsC.LocalClusterClient.Create(ctx, pod)).To(Succeed())
		})

		It("queue version already stored v2 - should not perform queue version pre-crd upgrade - and still succeed even though we have nodepool-controller pod", func() {
			// update the queue CRD to have the new version as storage
			queueCrd := &apiextensionsv1.CustomResourceDefinition{}
			Expect(clientsC.RemoteClusterClient.Get(ctx, client.ObjectKey{Name: "queues.scheduling.run.ai"}, queueCrd)).To(Succeed())
			queueCrd.Spec.Versions = []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v2alpha2",
					Storage: false,
				},
				{
					Name:    "v2",
					Storage: true,
				},
			}
			Expect(clientsC.RemoteClusterClient.Update(ctx, queueCrd)).To(Succeed())

			hook := pre_hooks.NewPreCrdUpgradeHook(params)
			err := hook.Call(ctx, clientsC)
			Expect(err).ToNot(HaveOccurred())

			dynamicResource := clientsC.LocalClusterDynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
			val, _ := dynamicResource.Get(ctx, runaiconfigObj.GetName(), metav1.GetOptions{})
			spec := val.Object["spec"].(map[string]interface{})
			runaiNPCConf := spec["nodepool-controller"].(map[string]interface{})
			enabled, ok := runaiNPCConf["enabled"].(bool)
			Expect(ok).To(BeTrue())
			Expect(enabled).To(BeTrue(), "nodepool-controller should be enabled after pre-CRD upgrade because queue CRD is already upgraded")
		})

		It("queue crd doesn't exist - should not perform queue version pre-crd upgrade - and still succeed even though we have nodepool-controller pod", func() {
			queueCrd := &apiextensionsv1.CustomResourceDefinition{}
			Expect(clientsC.RemoteClusterClient.Get(ctx, client.ObjectKey{Name: "queues.scheduling.run.ai"}, queueCrd)).To(Succeed())
			Expect(clientsC.RemoteClusterClient.Delete(ctx, queueCrd)).To(Succeed())

			hook := pre_hooks.NewPreCrdUpgradeHook(params)
			err := hook.Call(ctx, clientsC)
			Expect(err).ToNot(HaveOccurred())

			dynamicResource := clientsC.LocalClusterDynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
			val, _ := dynamicResource.Get(ctx, runaiconfigObj.GetName(), metav1.GetOptions{})
			spec := val.Object["spec"].(map[string]interface{})
			runaiNPCConf := spec["nodepool-controller"].(map[string]interface{})
			enabled, ok := runaiNPCConf["enabled"].(bool)
			Expect(ok).To(BeTrue())
			Expect(enabled).To(BeTrue(), "nodepool-controller should be enabled after pre-CRD upgrade because queue CRD doesn't exist")
		})
	})
})
