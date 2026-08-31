/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"time"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/nodepool_controller"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/operands"

	runai_scheduler "github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/operands/runai-scheduler"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/operands/runai-scheduler/resources"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Runai Scheduler Operand", Ordered, func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		stopper chan struct{}

		nodePool *v1alpha1.NodePool
		testCase *TestCase

		npc        *nodepool_controller.NodePoolController
		fakeClient client.WithWatch

		nodePoolControllerParams *common.NodePoolControllerParams
	)

	BeforeAll(func() {
		nodePoolName := "node-pool-operand-tests"

		testCase = &TestCase{
			Name: "Runai Scheduler Operand Tests",
			NodePools: []TestNodePool{
				{
					Name:       nodePoolName,
					LabelKey:   "nodePool0Key",
					LabelValue: "nodePool0Value",
					PlacementStrategy: PlacementStrategy{
						Gpu:       "binpack",
						GpuDevice: "binpack",
						Cpu:       "spread",
					},
				},
			},
			Nodes: map[string]TestNode{
				"kind-worker": {
					Name: "kind-worker",
					Labels: map[string]string{"nodePool0Key": "nodePool0Value",
						config.Get().NodePoolNameLabel: nodePoolName},
				},
			},
			ExpectedNodePoolsState: map[string]ExpectedNodePoolState{
				nodePoolName: {
					NodeNames: []string{"kind-worker"},
					Status: v1alpha1.NodePoolStatus{
						Phase: v1alpha1.NodePoolReady,
					},
				},
			},
		}

		stopper = make(chan struct{}, 1)
		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(func() {
			close(stopper)
			cancel()
		})

		fakeClient = preTestSetup(ctx, stopper, testCase, scheme)
		nodePool = getNodePoolFromClient(nodePoolName, fakeClient)

		nodePoolControllerParams = &common.NodePoolControllerParams{}
		npc = nodepool_controller.NewNodePoolController(fakeClient, scheme, nodePoolControllerParams)
		npc.SetServiceMonitorEnabled(true)
		reconcileAllNodePools(ctx, testCase.NodePools, npc, nil)
	})

	Context("Resources", func() {
		It("should have resources", func() {
			reconcileAllNodePools(ctx, testCase.NodePools, npc, nil)

			resources, err := runai_scheduler.Operand(nodePool.Name, true).ResourcesForNodePool(ctx,
				fakeClient, nodePool, nodePoolControllerParams)
			ExpectNoErr(err)
			Expect(resources).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			resources, err := runai_scheduler.Operand(nodePool.Name, true).ResourcesForNodePool(ctx,
				fakeClient, nodePool, nodePoolControllerParams)
			ExpectNoErr(err)

			reconcileAllNodePools(ctx, testCase.NodePools, npc, nil)

			Eventually(func(g Gomega) {
				for _, resource := range resources {
					obj := resource.Object
					err := fakeClient.Get(ctx, types.NamespacedName{
						Name:      obj.GetName(),
						Namespace: obj.GetNamespace(),
					}, obj)
					g.Expect(err).NotTo(HaveOccurred())
				}
			}, validateTestTimeout, validateTestInterval).Should(Succeed())
		})

		It("should re-create deleted resources", func() {
			var resources []operands.ResourceOld
			var err error
			Eventually(func() error {
				resources, err = runai_scheduler.Operand(nodePool.Name, true).ResourcesForNodePool(ctx,
					fakeClient, nodePool, nodePoolControllerParams)
				return err
			}, validateTestTimeout, validateTestInterval).Should(Succeed())

			Eventually(func(g Gomega) {
				for _, resource := range resources {
					deleteAndPollUntilDeleted(g, fakeClient, resource.Object, false)
				}
			}, validateTestTimeout, validateTestInterval).Should(Succeed())

			reconcileAllNodePools(ctx, testCase.NodePools, npc, nil)

			Eventually(func() bool {
				reCreated := true
				for _, resource := range resources {
					if !ResourceExists(fakeClient, resource.Object) {
						reCreated = false
					}
				}

				return reCreated
			}, validateTestTimeout, validateTestInterval).Should(BeTrue())
		})

		Context("SchedulingShard", func() {
			It("should reconcile a modified scheduling shard", func() {
				shard, err := resources.SchedulingShardForNodePool(ctx,
					fakeClient, nodePool, nodePoolControllerParams, nodePool.Name)
				Expect(err).NotTo(HaveOccurred())

				shard.(*kaiv1.SchedulingShard).Spec.Args["v"] = "6"

				EventuallyUpdateResource(fakeClient, shard)

				reconcileAllNodePools(ctx, testCase.NodePools, npc, nil)

				Eventually(func() bool {
					ExpectGetResource(fakeClient, shard)
					args := shard.(*kaiv1.SchedulingShard).Spec.Args
					return args["v"] == "6"
				}, validateTestTimeout, validateTestInterval).Should(BeFalse())
			})
		})

	})
})

func deleteAndPollUntilDeleted(g Gomega, k8sClient client.Client, resource client.Object, immediately bool) {
	By(fmt.Sprintf("waiting for %s/%s/%s to be deleted...",
		resource.GetNamespace(), resource.GetName(), resource.GetObjectKind().GroupVersionKind().String()))

	deleteOptions := &client.DeleteOptions{}
	var gracePeriodSeconds int64 = 0
	if immediately {
		deleteOptions.GracePeriodSeconds = &gracePeriodSeconds
	}

	deleteFn := func() error {
		err := k8sClient.Delete(context.Background(), resource, deleteOptions)
		if err == nil || errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	g.ExpectWithOffset(1, deleteFn()).To(Succeed())

	actual := func() bool {
		err := k8sClient.Get(context.Background(), types.NamespacedName{
			Name:      resource.GetName(),
			Namespace: resource.GetNamespace(),
		}, resource)
		return errors.IsNotFound(err)
	}

	g.EventuallyWithOffset(1, actual, 60*time.Second).Should(BeTrue())

	By(fmt.Sprintf("succesfully deleted resource %s/%s/%s",
		resource.GetNamespace(), resource.GetName(), resource.GetObjectKind().GroupVersionKind().String()))
}
