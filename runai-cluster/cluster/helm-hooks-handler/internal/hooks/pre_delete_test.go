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
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	ra "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/resource-actions"
	pre_hooks "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks/pre-hooks"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("PreDelete Hook", func() {
	var (
		resources *clients.Clients
	)

	BeforeEach(func() {
		resources = &clients.Clients{
			LocalClusterClient: fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build(),
		}
	})

	It("should delete cluster", func() {
		cluster := &runv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: runv1.ClusterName,
			},
		}
		Expect(resources.LocalClusterClient.Create(context.Background(), cluster)).To(Succeed())
		cluster, err := ra.HandleClusterCR(context.Background(), resources.LocalClusterClient, false)
		Expect(err).To(BeNil())
		resources.LocalClusterClient.List(context.Background(), &runv1.RunaiConfigList{})
		hook := pre_hooks.NewPreDeleteHook()
		Expect(hook.Call(context.Background(), resources)).To(Succeed())
		Expect(resources.LocalClusterClient.Get(context.Background(), rclient.ObjectKeyFromObject(cluster), cluster)).ToNot(Succeed())
	})

	It("should wait for RunaiConfig to be removed before finishing", func() {
		cluster := &runv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: runv1.ClusterName,
			},
		}
		Expect(resources.LocalClusterClient.Create(context.Background(), cluster)).To(Succeed())
		cluster, err := ra.HandleClusterCR(context.Background(), resources.LocalClusterClient, false)
		Expect(err).To(BeNil())
		Expect(resources.LocalClusterClient.Get(context.Background(), rclient.ObjectKeyFromObject(cluster), cluster)).To(Succeed())

		finalizer := "run.ai/finalize"
		runaiconfig := &runv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:       runv1.DefaultName,
				Namespace:  runv1.DefaultNamespace,
				Finalizers: []string{finalizer},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: cluster.APIVersion,
						Kind:       cluster.Kind,
						Name:       cluster.Name,
						UID:        cluster.UID,
					},
				},
			},
		}
		// Create the RunaiConfig in the fake client
		Expect(resources.LocalClusterClient.Create(context.Background(), runaiconfig)).To(Succeed())

		hook := pre_hooks.NewPreDeleteHook()
		ch := make(chan error, 1)
		go func() {
			ch <- hook.Call(context.Background(), resources)
		}()

		// Wait 3 seconds before simulating finalizer removal and deletion
		time.Sleep(3 * time.Second)
		// Make sure the hook is not finished yet and the RunaiConfig is still there
		Expect(resources.LocalClusterClient.Get(context.Background(), rclient.ObjectKeyFromObject(runaiconfig), runaiconfig)).To(Succeed())
		Expect(ch).ShouldNot(Receive())

		// Remove the finalizer and update the resource
		runaiconfig.Finalizers = nil
		Expect(resources.LocalClusterClient.Update(context.Background(), runaiconfig)).To(Succeed())
		// Call Delete again to simulate Kubernetes GC after finalizer removal
		Expect(resources.LocalClusterClient.Delete(context.Background(), runaiconfig)).To(Succeed())

		// Now the resource should be deleted
		Expect(resources.LocalClusterClient.Get(context.Background(), rclient.ObjectKeyFromObject(runaiconfig), runaiconfig)).ToNot(Succeed())

		// The hook should only finish after the finalizer is removed and the resource is deleted
		Eventually(ch, 15*time.Second).Should(Receive(BeNil()))
		Expect(resources.LocalClusterClient.Get(context.Background(), rclient.ObjectKeyFromObject(runaiconfig), runaiconfig)).ToNot(Succeed())
	})
})
