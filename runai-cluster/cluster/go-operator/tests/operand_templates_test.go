/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	templates "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/templates"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Templates Operand", func() {
	Context("Resources", func() {
		It("should have resources", func() {
			resources := templates.Operand(k8s, *runaiConfig).Resources()
			Expect(resources).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			resources := templates.Operand(k8s, *runaiConfig).Resources()

			for _, resource := range utils.GetEnabled(&resources) {
				obj, err := resource.Object(apiCtx)
				ExpectNoErr(err)
				ExpectGetResource(k8s, obj)
			}
		})

		It("should re-create deleted resources", func() {
			resources := templates.Operand(k8s, *runaiConfig).Resources()
			resources = utils.GetEnabled(&resources)

			for _, resource := range resources {
				obj, err := resource.Object(apiCtx)
				ExpectNoErr(err)
				DeleteAndPollUntilDeleted(k8s, obj)
			}

			EventuallyWithTimeout(func() bool {
				reCreated := true
				for _, resource := range resources {
					obj, err := resource.Object(apiCtx)
					ExpectNoErr(err)
					if !ResourceExists(k8s, obj) {
						reCreated = false
					}
				}

				return reCreated
			}).Should(BeTrue())
		})
	})

	Context("Monitoring namespace", func() {
		BeforeEach(func() {
			runaiConfig.Spec.Global.HTTPProxy.Enabled = true
			runaiConfig.Spec.Global.HTTPProxy.TLSCertSecret = &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "test-tls-secret"},
				Key:                  "tls.crt",
			}
			EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)
			PollForTrueCondition(k8s, runaiv1.ConditionTypeDeployed, runaiConfig)
		})
	})
})
