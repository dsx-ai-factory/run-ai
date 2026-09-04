/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	project_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/project-controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/project-controller/resources"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Project Controller Operand", func() {
	Context("Resources", func() {
		It("should have resources", func() {
			resources, err := project_controller.Operand().ResourcesForRunaiConfig(apiCtx,
				kc.New(k8s, k8s, nil), runaiConfig)
			ExpectNoErr(err)
			Expect(resources).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			resources, err := project_controller.Operand().ResourcesForRunaiConfig(apiCtx,
				kc.New(k8s, k8s, nil), runaiConfig)
			ExpectNoErr(err)

			for _, resource := range resources {
				ExpectGetResource(k8s, resource.Object)
			}
		})

		It("should re-create deleted resources", func() {
			resources, err := project_controller.Operand().ResourcesForRunaiConfig(apiCtx,
				kc.New(k8s, k8s, nil), runaiConfig)
			ExpectNoErr(err)

			for _, resource := range resources {
				DeleteAndPollUntilDeleted(k8s, resource.Object)
			}

			EventuallyWithTimeout(func() bool {
				reCreated := true
				for _, resource := range resources {
					if !ResourceExists(k8s, resource.Object) {
						reCreated = false
					}
				}

				return reCreated
			}).Should(BeTrue())
		})

		Context("Deployment", func() {
			It("should reconcile a modified deployment", func() {
				deployment, err := resources.DeploymentForRunaiConfig(apiCtx,
					kc.New(k8s, k8s, nil), runaiConfig)
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

		Context("ConfigMap", func() {
			It("should reconcile a modified configmap", func() {
				cm, err := resources.ConfigMapForRunaiConfig(apiCtx,
					kc.New(k8s, k8s, nil), runaiConfig)
				ExpectNoErr(err)

				cm.(*v1.ConfigMap).Data = map[string]string{}

				EventuallyUpdateResource(k8s, cm)

				EventuallyWithTimeout(func() bool {
					ExpectGetResource(k8s, cm)
					return len(cm.(*v1.ConfigMap).Data) != 0
				}).Should(BeTrue())
			})
		})

		Context("ServiceMonitor", func() {
			It("should reconcile a modified service monitor", func() {
				sm, err := resources.ServiceMonitorForRunaiConfig(apiCtx,
					kc.New(k8s, k8s, nil), runaiConfig)
				ExpectNoErr(err)

				sm.(*monitoringv1.ServiceMonitor).Spec.JobLabel = "test"

				EventuallyUpdateResource(k8s, sm)

				EventuallyWithTimeout(func() bool {
					ExpectGetResource(k8s, sm)
					return sm.(*monitoringv1.ServiceMonitor).Spec.JobLabel != "test"
				}).Should(BeTrue())
			})
		})

		Context("Service", func() {
			It("should reconcile a modified service", func() {
				svc, err := resources.ServiceForRunaiConfig(apiCtx,
					kc.New(k8s, k8s, nil), runaiConfig)
				ExpectNoErr(err)

				svc.(*v1.Service).Spec.Selector = map[string]string{}

				EventuallyUpdateResource(k8s, svc)

				EventuallyWithTimeout(func() bool {
					ExpectGetResource(k8s, svc)
					return len(svc.(*v1.Service).Spec.Selector) != 0
				}).Should(BeTrue())
			})
		})
	})
})
