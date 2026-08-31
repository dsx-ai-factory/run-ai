/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admission_controller "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/admission-controller"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/admission-controller/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Admission Controller Operand", func() {
	Context("Resources", func() {
		It("should have resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			resources, err := admission_controller.Operand().ResourcesForRunaiConfig(apiCtx,
				hybridClient, runaiConfig)
			ExpectNoErr(err)
			Expect(resources).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			resources, err := admission_controller.Operand().ResourcesForRunaiConfig(apiCtx,
				hybridClient, runaiConfig)
			ExpectNoErr(err)

			for _, resource := range resources {
				ExpectGetResource(k8s, resource.Object)
			}
		})

		It("should re-create deleted resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			resources, err := admission_controller.Operand().ResourcesForRunaiConfig(apiCtx,
				hybridClient, runaiConfig)
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
				hybridClient := kc.New(k8s, k8s, nil)
				deployment, err := resources.DeploymentForRunaiConfig(apiCtx,
					hybridClient, runaiConfig)
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

		Context("Service", func() {
			It("should reconcile a modified service", func() {
				hybridClient := kc.New(k8s, k8s, nil)
				svc, err := resources.ServiceForRunaiConfig(apiCtx,
					hybridClient, runaiConfig)
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
