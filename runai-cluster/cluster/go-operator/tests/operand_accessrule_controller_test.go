/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	accessrule_controller "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/accessrule-controller"
	nodepool_controller_resources "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/nodepool-controller/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"

	appsv1 "k8s.io/api/apps/v1"
)

var _ = Describe("Access Rule Controller Operand", func() {
	var (
		operand = accessrule_controller.Operand()
	)

	BeforeEach(func() {
		runaiConfig.Spec.AccessRuleController.Enabled = ptr.To[bool](true)
		EventuallyUpdateResource(k8s, runaiConfig)
		PollForTrueCondition(k8s, runaiv1.ConditionTypeDeployed, runaiConfig)
	})

	AfterEach(func() {
		runaiConfig.Spec.AccessRuleController.Enabled = ptr.To[bool](false)
		EventuallyUpdateResource(k8s, runaiConfig)
		PollForTrueCondition(k8s, runaiv1.ConditionTypeDeployed, runaiConfig)
	})

	Context("Resources", func() {
		It("should have resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			resources, err := operand.ResourcesForRunaiConfig(apiCtx,
				hybridClient, runaiConfig)
			ExpectNoErr(err)
			Expect(resources).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			resources, err := operand.ResourcesForRunaiConfig(apiCtx,
				hybridClient, runaiConfig)
			ExpectNoErr(err)

			for _, resource := range resources {
				ExpectGetResource(k8s, resource.Object)
			}
		})

		It("should re-create deleted resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			resources, err := operand.ResourcesForRunaiConfig(apiCtx,
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
				depResource := nodepool_controller_resources.NewDeploymentResource("nodepool-controller", hybridClient, runaiConfig)
				deployment, err := depResource.Object(apiCtx)
				ExpectNoErr(err)

				badServiceAccountName := "bad-service-account-name"

				deployment.(*appsv1.Deployment).Spec.
					Template.Spec.ServiceAccountName = badServiceAccountName

				EventuallyUpdateResource(k8s, deployment)

				EventuallyWithTimeout(func() bool {
					ExpectGetResource(k8s, deployment)
					return deployment.(*appsv1.Deployment).Spec.
						Template.Spec.ServiceAccountName == badServiceAccountName
				}).Should(BeFalse())
			})
		})
	})
})
