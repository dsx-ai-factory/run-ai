/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

var _ = Describe("Legacy workload cleanup gates", func() {
	It("accepts one reconciled and available RunaiConfig readiness marker", func() {
		config := readyRunaiConfig(7)
		Expect(validateReadinessMarker(config)).To(Succeed())
	})

	DescribeTable("rejects a stale or unhealthy readiness marker",
		func(mutate func(*runaiv1.RunaiConfig), errorSubstring string) {
			config := readyRunaiConfig(7)
			mutate(config)
			Expect(validateReadinessMarker(config)).To(MatchError(ContainSubstring(errorSubstring)))
		},
		Entry("stale observed generation", func(config *runaiv1.RunaiConfig) {
			config.Status.ObservedGeneration = 6
		}, "observed generation"),
		Entry("stale reconciled condition generation", func(config *runaiv1.RunaiConfig) {
			config.Status.Conditions[0].ObservedGeneration = 6
		}, "Reconciled condition observed generation"),
		Entry("stale available condition generation", func(config *runaiv1.RunaiConfig) {
			config.Status.Conditions[1].ObservedGeneration = 6
		}, "Available condition observed generation"),
		Entry("missing reconciled condition", func(config *runaiv1.RunaiConfig) {
			config.Status.Conditions = config.Status.Conditions[1:]
		}, "Reconciled"),
		Entry("failed reconciled condition", func(config *runaiv1.RunaiConfig) {
			config.Status.Conditions[0].Status = metav1.ConditionFalse
		}, "Reconciled"),
		Entry("missing available condition", func(config *runaiv1.RunaiConfig) {
			config.Status.Conditions = config.Status.Conditions[:1]
		}, "Available"),
		Entry("failed available condition", func(config *runaiv1.RunaiConfig) {
			config.Status.Conditions[1].Status = metav1.ConditionFalse
		}, "Available"),
	)
})

func readyRunaiConfig(generation int64) *runaiv1.RunaiConfig {
	return &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{Generation: generation},
		Status: runaiv1.RunaiConfigStatus{
			RunaiCustomTypeStatus: runaiv1.RunaiCustomTypeStatus{
				ObservedGeneration: generation,
				Conditions: []metav1.Condition{
					{Type: string(runaiv1.ConditionTypeReconciled), Status: metav1.ConditionTrue, ObservedGeneration: generation},
					{Type: string(runaiv1.ConditionTypeAvailable), Status: metav1.ConditionTrue, ObservedGeneration: generation},
				},
			},
		},
	}
}
