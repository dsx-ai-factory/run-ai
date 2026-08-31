/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	init_ca "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/init-ca"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/init-ca/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Init CA Operand", func() {
	AfterEach(func() {
		runaiConfig.Spec.PodGroupAssigner.Enabled = ptr.To[bool](true)
		EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)
	})

	Context("Resources", func() {
		It("should have resources", func() {
			resources, err := init_ca.Operand().ResourcesForRunaiConfig(apiCtx,
				kc.New(k8s, k8s, nil), runaiConfig)
			ExpectNoErr(err)
			Expect(resources).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			resources, err := init_ca.Operand().ResourcesForRunaiConfig(apiCtx,
				kc.New(k8s, k8s, nil), runaiConfig)
			ExpectNoErr(err)

			for _, resource := range resources {
				ExpectGetResource(k8s, resource.Object)
			}
		})

		It("should re-create deleted resources", func() {
			resources, err := init_ca.Operand().ResourcesForRunaiConfig(apiCtx,
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

		It("should have pod-group-assigner webhook in configmap if needed", func() {
			runaiConfig.Spec.PodGroupAssigner.Enabled = ptr.To[bool](true)
			EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)

			EventuallyWithTimeout(func() bool {
				cm, err := resources.ConfigMapForRunaiConfig(apiCtx, kc.New(k8s, k8s, nil), runaiConfig)
				ExpectNoErr(err)
				Expect(cm).NotTo(BeNil())

				configMap := &corev1.ConfigMap{}
				err = k8s.Get(apiCtx, NamespacedNameFor(cm), configMap)
				ExpectNoErr(err)
				Expect(configMap).NotTo(BeNil())

				return strings.Contains(configMap.Data["webhooks.yaml"], "pod-group-assigner")
			}).Should(BeTrue())
		})

		Context("Should have the enable-workloads-ownership-protection webhook in configmap if needed", func() {
			It("webhook enabled", func() {
				runaiConfig.Spec.Global.EnableWorkloadOwnershipProtection = ptr.To[bool](true)
				EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)

				EventuallyWithTimeout(func() bool {
					cm, err := resources.ConfigMapForRunaiConfig(apiCtx, kc.New(k8s, k8s, nil), runaiConfig)
					ExpectNoErr(err)
					Expect(cm).NotTo(BeNil())

					configMap := &corev1.ConfigMap{}
					err = k8s.Get(apiCtx, NamespacedNameFor(cm), configMap)
					ExpectNoErr(err)
					Expect(configMap).NotTo(BeNil())

					return strings.Contains(configMap.Data["webhooks.yaml"], "runai-workload-ownership-protection-validation")
				}).Should(BeTrue())
			})
			It("webhook disabled", func() {
				runaiConfig.Spec.Global.EnableWorkloadOwnershipProtection = ptr.To[bool](false)
				EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)

				EventuallyWithTimeout(func() bool {
					cm, err := resources.ConfigMapForRunaiConfig(apiCtx, kc.New(k8s, k8s, nil), runaiConfig)
					ExpectNoErr(err)
					Expect(cm).NotTo(BeNil())

					configMap := &corev1.ConfigMap{}
					err = k8s.Get(apiCtx, NamespacedNameFor(cm), configMap)
					ExpectNoErr(err)
					Expect(configMap).NotTo(BeNil())

					return !strings.Contains(configMap.Data["webhooks.yaml"], "runai-workload-ownership-protection-validation")
				}).Should(BeTrue())
			})
		})
	})
})
