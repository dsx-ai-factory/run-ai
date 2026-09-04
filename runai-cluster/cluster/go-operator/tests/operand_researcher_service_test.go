/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	researcher_service "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/researcher-service"
	researcher_service_resources "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/researcher-service/resources"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Researcher Service Operand", func() {
	Context("Resources", func() {
		It("should have resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := researcher_service.Operand(hybridClient, runaiConfig)
			resources := op.Resources()
			Expect(resources).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := researcher_service.Operand(hybridClient, runaiConfig)

			for _, res := range op.Resources() {
				if !res.Enabled() {
					continue
				}
				obj, err := res.Object(apiCtx)
				ExpectNoErr(err)
				ExpectGetResource(k8s, obj)
			}
		})

		It("should re-create deleted resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := researcher_service.Operand(hybridClient, runaiConfig)

			for _, res := range op.Resources() {
				if !res.Enabled() {
					continue
				}
				obj, err := res.Object(apiCtx)
				ExpectNoErr(err)
				DeleteAndPollUntilDeleted(k8s, obj)
			}

			EventuallyWithTimeout(func() bool {
				reCreated := true
				for _, res := range op.Resources() {
					if !res.Enabled() {
						continue
					}
					obj, err := res.Object(apiCtx)
					ExpectNoErr(err)
					if !ResourceExists(k8s, obj) {
						reCreated = false
					}
				}

				return reCreated
			}).Should(BeTrue())
		})

		Context("Deployment", func() {
			It("should reconcile a modified deployment spec but save editing object meta", func() {
				hybridClient := kc.New(k8s, k8s, nil)
				depResource := researcher_service_resources.NewDeploymentResource("researcher-service", hybridClient, runaiConfig)
				deployment, err := depResource.Object(apiCtx)
				ExpectNoErr(err)

				deployment.(*appsv1.Deployment).ObjectMeta.Labels = map[string]string{
					"new/label": "newLabel",
				}

				changedScheduler := "run.ai/test"

				deployment.(*appsv1.Deployment).Spec.
					Template.Spec.SchedulerName = changedScheduler

				EventuallyUpdateResource(k8s, deployment)

				EventuallyWithTimeout(func() bool {
					ExpectGetResource(k8s, deployment)
					dep := deployment.(*appsv1.Deployment)
					return dep.Spec.Template.Spec.SchedulerName == changedScheduler
				}).Should(BeFalse())

				EventuallyWithTimeout(func() bool {
					ExpectGetResource(k8s, deployment)
					dep := deployment.(*appsv1.Deployment)
					return dep.ObjectMeta.Labels["new/label"] == "newLabel"
				}).Should(BeTrue())

			})
		})

		Context("Ingress", func() {
			Context("without ingress configured", func() {
				It("should not create ingress", func() {
					hybridClient := kc.New(k8s, k8s, nil)
					ingressRes := researcher_service_resources.NewIngressResource("researcher-service", hybridClient, runaiConfig)
					Expect(ingressRes.Enabled()).To(BeFalse())

					status, err := researcher_service_resources.IngressStatus(apiCtx, kc.New(k8s, k8s, nil), runaiConfig)
					ExpectNoErr(err)
					Expect(status.Ready).To(BeTrue())
				})
			})

			Context("with ingress configured", func() {
				It("should create ingress", func() {
					runaiConfig.Spec.Internal.Global.ClusterURL = "https://test.com"
					runaiConfig.Spec.Internal.ResearcherService.Ingress.Enabled = ptr.To[bool](true)
					EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)
					WaitForDeployedCondition(kc.New(k8s, k8s, nil), runaiConfig)

					hybridClient := kc.New(k8s, k8s, nil)
					ingressRes := researcher_service_resources.NewIngressResource("researcher-service", hybridClient, runaiConfig)
					Expect(ingressRes.Enabled()).To(BeTrue())

					obj, err := ingressRes.Object(apiCtx)
					ExpectNoErr(err)
					Expect(obj).ShouldNot(BeNil())

					EventuallyWithTimeout(func() bool {
						status, err := researcher_service_resources.IngressStatus(apiCtx, kc.New(k8s, k8s, nil), runaiConfig)
						ExpectNoErr(err)

						return status.Ready
					}).Should(BeTrue())
				})
			})
		})
	})
})
