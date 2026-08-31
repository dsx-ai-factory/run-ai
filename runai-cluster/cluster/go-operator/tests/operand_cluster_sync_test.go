/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cluster_sync "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/cluster-sync"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/cluster-sync/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"
	appsv1 "k8s.io/api/apps/v1"
)

var _ = Describe("ClusterSync Operand", func() {
	Context("Resources", func() {
		It("should have resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := cluster_sync.Operand(hybridClient, runaiConfig)
			Expect(op.Resources()).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := cluster_sync.Operand(hybridClient, runaiConfig)
			for _, res := range op.Resources() {
				obj, err := res.Object(apiCtx)
				ExpectNoErr(err)
				if obj != nil {
					ExpectGetResource(k8s, obj)
				}
			}
		})

		It("should re-create deleted resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := cluster_sync.Operand(hybridClient, runaiConfig)
			for _, res := range op.Resources() {
				obj, err := res.Object(apiCtx)
				ExpectNoErr(err)
				if obj != nil && res.Enabled() {
					DeleteAndPollUntilDeleted(k8s, obj)
				}
			}

			EventuallyWithTimeout(func() bool {
				reCreated := true
				for _, res := range op.Resources() {
					obj, _ := res.Object(apiCtx)
					if obj != nil && res.Enabled() {
						if !ResourceExists(k8s, obj) {
							reCreated = false
						}
					}
				}
				return reCreated
			}).Should(BeTrue())
		})

		Context("Deployment", func() {
			It("should reconcile a modified deployment", func() {
				hybridClient := kc.New(k8s, k8s, nil)
				deployRes := resources.NewDeploymentResource("cluster-sync", hybridClient, runaiConfig)
				deployment, err := deployRes.Object(apiCtx)
				ExpectNoErr(err)

				changedScheduler := "run.ai/test"
				deployment.(*appsv1.Deployment).Spec.
					Template.Spec.SchedulerName = changedScheduler

				EventuallyUpdateResource(k8s, deployment)

				EventuallyWithTimeout(func() bool {
					ExpectGetResource(k8s, deployment)
					return deployment.(*appsv1.Deployment).Spec.
						Template.Spec.SchedulerName == changedScheduler
				}).Should(BeFalse())
			})
		})
	})
})
