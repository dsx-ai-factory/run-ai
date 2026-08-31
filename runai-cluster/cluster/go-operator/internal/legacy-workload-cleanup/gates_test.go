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
	DescribeTable("evaluates the upgrade boundary",
		func(source string, shouldRun bool, errorSubstring string) {
			run, err := SourceRequiresCleanup(source)
			if errorSubstring != "" {
				Expect(err).To(MatchError(ContainSubstring(errorSubstring)))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(run).To(Equal(shouldRun))
		},
		Entry("requires cleanup for 2.26", "2.26.0", true, ""),
		Entry("does not infer pre-3 provenance from a branch build", "0.0.0-12345.v300-main", false, ""),
		Entry("does not require cleanup for 3.0", "3.0.0", false, ""),
		Entry("does not require cleanup for a 3.0 prerelease", "3.0.0-rc.1", false, ""),
		Entry("fails for a missing source", "", false, "source version is missing"),
		Entry("fails for an invalid source", "not-a-version", false, "parse source version"),
	)

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
