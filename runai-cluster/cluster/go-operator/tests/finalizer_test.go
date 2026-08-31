/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"

	ccontroller "github.com/run-ai/runai/runai-cluster/cluster/go-operator/controllers/run"
	controller "github.com/run-ai/runai/runai-cluster/cluster/go-operator/controllers/run"
	cutils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

var _ = Describe("Finalizer", func() {
	When("RunaiConfig is active", func() {
		It("should have a finalizer", func() {
			EventuallyWithTimeout(func() bool {
				runaiConfig = GetTestingRunaiConfig(k8s)
				return cutils.HasFinalizer(runaiConfig, runaiv1.Finalizer)
			}).Should(BeTrue())
		})

		It("should reinstate finalizer if removed", func() {
			runaiConfig.SetFinalizers(nil)
			EventuallyUpdateResource(k8s, runaiConfig)

			EventuallyWithTimeout(func() bool {
				runaiConfig = GetTestingRunaiConfig(k8s)
				return cutils.HasFinalizer(runaiConfig, runaiv1.Finalizer)
			}).Should(BeTrue())
		})
	})

	When("RunaiConfig is pending deletion", func() {
		It("should remove all resources, and only then allow deletion", func() {
			PollForTrueCondition(k8s, runaiv1.ConditionTypeDeployed, runaiConfig)

			ExpectDeleteResource(k8s, runaiConfig)

			EventuallyWithTimeout(func() bool {
				hybridClient := kc.New(k8s, k8s, nil)
				managedOperands := ccontroller.NewManagedOperands(apiCtx, hybridClient, hybridClient, runaiConfig)
				finalized := true
				resources, err := controller.ResourcesForOperands(apiCtx,
					managedOperands.GetRemoteClusterEnabled())
				ExpectNoErr(err)

				res := append(resources.CreateOrUpdate, resources.Delete...)

				for _, resource := range res {
					obj, err := resource.Object(apiCtx)
					ExpectNoErr(err)
					err = k8s.Get(apiCtx,
						NamespacedNameFor(obj), obj)
					if err != nil {
						Expect(errors.IsNotFound(err)).To(BeTrue())
					} else {
						if policy, found := obj.GetAnnotations()[runaiv1.DeletionPolicyAnnotationKey]; found {
							if policy == string(runaiv1.DeletionPolicyRetain) {
								continue
							}
						}
						finalized = false
						break
					}
				}

				return finalized
			}).Should(BeTrue())

			PollUntilDeleted(k8s, runaiConfig)

			runaiConfig = CreateTestingRunaiConfig(k8s)
		})
	})
})
