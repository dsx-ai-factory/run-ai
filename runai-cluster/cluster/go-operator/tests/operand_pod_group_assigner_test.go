/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pod_group_assigner "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/pod-group-assigner"
	pod_group_assigner_resources "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/pod-group-assigner/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"
	appsv1 "k8s.io/api/apps/v1"
)

var _ = Describe("Pod Group Assigner Operand", func() {
	Context("Resources", func() {
		It("should have resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := pod_group_assigner.Operand(hybridClient, runaiConfig)
			resources := op.Resources()
			Expect(resources).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := pod_group_assigner.Operand(hybridClient, runaiConfig)

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
			op := pod_group_assigner.Operand(hybridClient, runaiConfig)

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
			It("should reconcile a modified deployment", func() {
				hybridClient := kc.New(k8s, k8s, nil)
				depResource := pod_group_assigner_resources.NewDeploymentResource("pod-group-assigner", hybridClient, runaiConfig)
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
