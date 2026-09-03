/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/agent"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/agent/resources"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

var _ = Describe("Agent Operand", func() {
	Context("Resources", func() {
		It("should have resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := agent.Operand(hybridClient, runaiConfig)
			Expect(op.Resources()).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := agent.Operand(hybridClient, runaiConfig)
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
			op := agent.Operand(hybridClient, runaiConfig)
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
				deployRes := resources.NewDeploymentResource("agent", hybridClient, runaiConfig)
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

		Context("ServiceMonitor", func() {
			It("should reconcile a modified servicemonitor", func() {
				hybridClient := kc.New(k8s, k8s, nil)
				smRes := resources.NewServiceMonitorResource("agent", hybridClient, runaiConfig)
				svcMonitor, err := smRes.Object(apiCtx)
				ExpectNoErr(err)

				changedJobLabel := "run.ai/test"
				svcMonitor.(*monitoringv1.ServiceMonitor).Spec.JobLabel = changedJobLabel

				EventuallyUpdateResource(k8s, svcMonitor)

				EventuallyWithTimeout(func() string {
					ExpectGetResource(k8s, svcMonitor)
					return svcMonitor.(*monitoringv1.ServiceMonitor).Spec.JobLabel
				}).Should(Equal(changedJobLabel))
			})
		})
	})
})
