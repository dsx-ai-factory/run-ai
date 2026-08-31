/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Tolerations tests", func() {
	var (
		ctx          context.Context
		sharedFields *field_groups.WorkloadFields
	)

	BeforeEach(func() {
		// Setup common test variables
		ctx = context.Background()
		sharedFields = &field_groups.WorkloadFields{
			ContainerFields: field_groups.ContainerFields{
				Tolerations: &flds.TolerationsField{
					Value: &map[string]flds.TolerationsItemAndRules{},
				},
			},
		}
	})

	// Helper function to add a toleration to the sharedFields
	addToleration := func(nick string, key string, value string, operator string, effect string, seconds *int) {
		(*sharedFields.ContainerFields.Tolerations.Value)[nick] = flds.TolerationsItemAndRules{
			Value: &flds.TolerationsItemValue{
				Key:      &key,
				Value:    &value,
				Operator: &operator,
				Effect:   &effect,
				Seconds:  seconds,
			},
		}
	}

	Context("Basic functionality", func() {
		It("should return empty when no tolerations are specified", func() {
			tolerations, err := GetTolerations(ctx, &sharedFields.ContainerFields)
			Expect(err).NotTo(HaveOccurred())
			Expect(tolerations).To(BeEmpty())
		})

		It("should create a basic toleration with default operator", func() {
			key := "key1"
			effect := string(corev1.TaintEffectNoSchedule)
			addToleration("main", key, "", "", effect, nil)

			tolerations, err := GetTolerations(ctx, &sharedFields.ContainerFields)
			Expect(err).NotTo(HaveOccurred())
			Expect(tolerations).To(HaveLen(1))
			Expect(tolerations[0].Key).To(Equal(key))
			Expect(tolerations[0].Operator).To(BeEmpty())
			Expect(tolerations[0].Effect).To(Equal(corev1.TaintEffect(effect)))
			Expect(tolerations[0].Value).To(Equal(""))
			Expect(tolerations[0].TolerationSeconds).To(BeNil())
		})

		It("should create a toleration with Equals operator and value", func() {
			key := "key1"
			value := "value1"
			operator := string(corev1.TolerationOpEqual)
			effect := string(corev1.TaintEffectNoSchedule)
			addToleration("main", key, value, operator, effect, nil)

			tolerations, err := GetTolerations(ctx, &sharedFields.ContainerFields)
			Expect(err).NotTo(HaveOccurred())
			Expect(tolerations).To(HaveLen(1))
			Expect(tolerations[0].Key).To(Equal(key))
			Expect(tolerations[0].Value).To(Equal(value))
			Expect(tolerations[0].Operator).To(Equal(corev1.TolerationOperator(operator)))
			Expect(tolerations[0].Effect).To(Equal(corev1.TaintEffect(effect)))
		})

		It("should handle TolerationSeconds field", func() {
			key := "key1"
			seconds := 300
			addToleration("main", key, "", "", "", &seconds)

			tolerations, err := GetTolerations(ctx, &sharedFields.ContainerFields)
			Expect(err).NotTo(HaveOccurred())
			Expect(tolerations).To(HaveLen(1))
			Expect(tolerations[0].TolerationSeconds).NotTo(BeNil())
			Expect(*tolerations[0].TolerationSeconds).To(Equal(int64(seconds)))
		})
	})

	Context("Multiple tolerations", func() {
		It("should handle multiple tolerations correctly", func() {
			addToleration("first", "key1", "", string(corev1.TolerationOpExists), string(corev1.TaintEffectNoSchedule), nil)
			addToleration("second", "key2", "value2", string(corev1.TolerationOpEqual), string(corev1.TaintEffectNoExecute), nil)

			tolerations, err := GetTolerations(ctx, &sharedFields.ContainerFields)
			Expect(err).NotTo(HaveOccurred())
			Expect(tolerations).To(HaveLen(2))

			// Check both tolerations are included with correct values
			keys := []string{tolerations[0].Key, tolerations[1].Key}
			Expect(keys).To(ConsistOf("key1", "key2"))

			// Find and verify each toleration
			for _, toleration := range tolerations {
				if toleration.Key == "key1" {
					Expect(toleration.Operator).To(Equal(corev1.TolerationOpExists))
					Expect(toleration.Effect).To(Equal(corev1.TaintEffectNoSchedule))
					Expect(toleration.Value).To(Equal(""))
				} else if toleration.Key == "key2" {
					Expect(toleration.Operator).To(Equal(corev1.TolerationOpEqual))
					Expect(toleration.Effect).To(Equal(corev1.TaintEffectNoExecute))
					Expect(toleration.Value).To(Equal("value2"))
				}
			}
		})
	})

	Context("Validation and error handling", func() {
		It("should return error when operator is Exists but value is not empty", func() {
			key := "key1"
			value := "non-empty-value"
			operator := string(corev1.TolerationOpExists)
			addToleration("main", key, value, operator, "", nil)

			tolerations, err := GetTolerations(ctx, &sharedFields.ContainerFields)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("tolerationValue must be empty when operator is Exists"))
			Expect(tolerations).To(BeNil())
		})

		It("should skip tolerations with nil value", func() {
			// Add a valid toleration
			addToleration("valid", "key1", "", "", "", nil)

			// Add a nil toleration
			(*sharedFields.ContainerFields.Tolerations.Value)["nil"] = flds.TolerationsItemAndRules{
				Value: nil,
			}

			tolerations, err := GetTolerations(ctx, &sharedFields.ContainerFields)
			Expect(err).NotTo(HaveOccurred())
			Expect(tolerations).To(HaveLen(1))
			Expect(tolerations[0].Key).To(Equal("key1"))
		})
	})

	Context("Default values", func() {
		It("should use default values when fields are not specified", func() {
			// Create toleration with minimal fields (just a key)
			key := "key1"
			nick := "minimal"
			(*sharedFields.ContainerFields.Tolerations.Value)[nick] = flds.TolerationsItemAndRules{
				Value: &flds.TolerationsItemValue{
					Key: &key,
				},
			}

			tolerations, err := GetTolerations(ctx, &sharedFields.ContainerFields)
			Expect(err).NotTo(HaveOccurred())
			Expect(tolerations).To(HaveLen(1))
			Expect(tolerations[0].Key).To(Equal(key))
			Expect(tolerations[0].Operator).To(Equal(corev1.TolerationOperator(defaultOperator)))
			Expect(tolerations[0].Value).To(Equal(""))
			Expect(tolerations[0].Effect).To(Equal(corev1.TaintEffect("")))
			Expect(tolerations[0].TolerationSeconds).To(BeNil())
		})

		It("should use empty key when key is not specified", func() {
			value := "value1"
			operator := string(corev1.TolerationOpEqual)
			nick := "no-key"
			(*sharedFields.ContainerFields.Tolerations.Value)[nick] = flds.TolerationsItemAndRules{
				Value: &flds.TolerationsItemValue{
					Value:    &value,
					Operator: &operator,
				},
			}

			tolerations, err := GetTolerations(ctx, &sharedFields.ContainerFields)
			Expect(err).NotTo(HaveOccurred())
			Expect(tolerations).To(HaveLen(1))
			Expect(tolerations[0].Key).To(Equal(""))
			Expect(tolerations[0].Value).To(Equal(value))
			Expect(tolerations[0].Operator).To(Equal(corev1.TolerationOperator(operator)))
		})
	})
})
