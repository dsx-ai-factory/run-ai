/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"sync"

	. "github.com/onsi/ginkgo/v2"
	ccontroller "github.com/run-ai/runai/runai-cluster/cluster/go-operator/controllers/run"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

var _ = Describe("Conditions", func() {
	Context("Deployed Condition", func() {
		When("reconciliation completed", func() {
			It("should report the Deployed condition for a valid config", func() {
				PollForTrueCondition(k8s, runaiv1.ConditionTypeDeployed, runaiConfig)
			})

			It("should report not deployed condition for an invalid config", func() {
				runaiConfig.Spec.ClusterSync.Image.PullPolicy = "invalid"
				EventuallyUpdateResource(k8s, runaiConfig)
				PollForFalseCondition(k8s, runaiv1.ConditionTypeDeployed, runaiConfig)
			})
		})
	})

	Context("DependenciesFulfilled Condition", func() {
		It("should report the DependenciesFulfilled condition", func() {
			PollForCondition(k8s, runaiv1.ConditionDependenciesFulfilled, runaiConfig)
		})

		It("should report a True DependenciesFulfilled condition", func() {
			PollForTrueCondition(k8s, runaiv1.ConditionDependenciesFulfilled, runaiConfig)
		})
	})

	PContext("Available Condition", func() {
		When("a resource is not ready", func() {
			It("should not report Available condition", func() {
				PollForTrueCondition(k8s, runaiv1.ConditionTypeDeployed, runaiConfig)

				hybridClient := kc.New(k8s, k8s, nil)
				managedOperands := ccontroller.NewManagedOperands(apiCtx, hybridClient, hybridClient, runaiConfig)
				resources, err := ccontroller.ResourcesForOperands(apiCtx, managedOperands.GetRemoteClusterEnabled())
				ExpectNoErr(err)

				enabledResources := resources.CreateOrUpdate

				testedResource, err := enabledResources[0].Object(apiCtx)
				ExpectNoErr(err)

				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()

					PollUntilDeleted(k8s, testedResource)
					PollForFalseCondition(k8s, runaiv1.ConditionTypeAvailable, runaiConfig)
				}()

				// Delete an arbitrary resource and wait for the controller to report
				// Available = false
				ExpectDeleteResource(k8s, testedResource)

				wg.Wait()
			})
		})

		When("all resources are ready", func() {
			// This test should be implemented after all operands were integrated
			PIt("should report the Available condition")
		})
	})
})
