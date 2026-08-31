/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestResources(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "ManagedNodesConfig")
}

// UseKaiNodePools is a compile-time const, so only one branch of the operand is
// reachable in any given build. The status specs follow the nodepool-utils convention:
// every fixture is written to BOTH CRDs with identical content, so those assertions hold
// whichever value the const has. Only the API group differs between the branches - the
// object name is the same either way - so just the type assertion is selected off the
// const. Delete the legacy fixtures and branches alongside the else branch.
var _ = Describe("ManagedNodesConfig", func() {
	const nodeLabelKey = "example.com/managed"

	var (
		ctx         context.Context
		runaiConfig *runv1.RunaiConfig
		scheme      *runtime.Scheme
	)

	inclusionCriteria := corev1.NodeSelector{
		NodeSelectorTerms: []corev1.NodeSelectorTerm{{
			MatchExpressions: []corev1.NodeSelectorRequirement{{
				Key:      nodeLabelKey,
				Operator: corev1.NodeSelectorOpExists,
			}},
		}},
	}

	// clientWithConditions seeds both CRDs' singletons with the same conditions, so the
	// status result does not depend on which one the const selects.
	clientWithConditions := func(conditions []metav1.Condition) rclient.Client {
		return fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			&runv1alpha1.ManagedNodesConfig{
				ObjectMeta: metav1.ObjectMeta{Name: constants.ManagedNodesConfigName},
				Status:     runv1alpha1.ManagedNodesConfigStatus{Conditions: conditions},
			},
			&kaiv1alpha1.ManagedNodesConfig{
				ObjectMeta: metav1.ObjectMeta{Name: constants.ManagedNodesConfigName},
				Status:     kaiv1alpha1.ManagedNodesConfigStatus{Conditions: conditions},
			},
		).Build()
	}

	appliedCondition := func(status metav1.ConditionStatus, reason string) []metav1.Condition {
		return []metav1.Condition{{
			Type:    appliedConditionType,
			Status:  status,
			Reason:  reason,
			Message: "some message",
		}}
	}

	BeforeEach(func() {
		ctx = context.Background()

		scheme = runtime.NewScheme()
		Expect(runv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(kaiv1alpha1.AddToScheme(scheme)).To(Succeed())

		runaiConfig = &runv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{Name: runv1.DefaultName, Namespace: runv1.DefaultNamespace},
		}
		runaiConfig.SetDefaultsWhereNeeded()
		runaiConfig.Spec.Global.ManagedNodes.InclusionCriteria = &inclusionCriteria
	})

	Describe("rendering", func() {
		It("renders into the CRD group the flag selects", func() {
			obj, err := ManagedNodesConfigForRunaiConfig(ctx, clientWithConditions(nil), runaiConfig)

			Expect(err).ToNot(HaveOccurred())
			if constants.UseKaiNodePools {
				Expect(obj).To(BeAssignableToTypeOf(&kaiv1alpha1.ManagedNodesConfig{}))
			} else {
				Expect(obj).To(BeAssignableToTypeOf(&runv1alpha1.ManagedNodesConfig{}))
			}
		})

		// The name must not change with the group: go-operator pins it for the deployed
		// controller via --managed-nodes-config-name, so a renamed object would be written
		// by the operand and read by nobody.
		It("names the object the same in either group, carrying the inclusion criteria", func() {
			obj, err := ManagedNodesConfigForRunaiConfig(ctx, clientWithConditions(nil), runaiConfig)

			Expect(err).ToNot(HaveOccurred())
			Expect(obj.GetName()).To(Equal(constants.ManagedNodesConfigName))
			switch mnc := obj.(type) {
			case *kaiv1alpha1.ManagedNodesConfig:
				Expect(mnc.Spec.InclusionCriteria).To(Equal(inclusionCriteria))
			case *runv1alpha1.ManagedNodesConfig:
				Expect(mnc.Spec.InclusionCriteria).To(Equal(inclusionCriteria))
			}
		})
	})

	Describe("status", func() {
		It("is not ready when the managed-nodes-config is missing", func() {
			emptyClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			status, err := ManagedNodesConfigStatus(ctx, emptyClient, runaiConfig)

			Expect(err).ToNot(HaveOccurred())
			Expect(status.Ready).To(BeFalse())
			Expect(status.Reasons).To(ContainElement(ContainSubstring("is missing")))
		})

		It("is not ready when the applied condition has not been reported yet", func() {
			status, err := ManagedNodesConfigStatus(ctx, clientWithConditions(nil), runaiConfig)

			Expect(err).ToNot(HaveOccurred())
			Expect(status.Ready).To(BeFalse())
			Expect(status.Reasons).To(ContainElement(ContainSubstring("no applied condition")))
		})

		It("is ready when the config is applied", func() {
			conditions := appliedCondition(metav1.ConditionTrue, string(runv1alpha1.MNCConditionReasonAllNodesIncludedCorrectly))

			status, err := ManagedNodesConfigStatus(ctx, clientWithConditions(conditions), runaiConfig)

			Expect(err).ToNot(HaveOccurred())
			Expect(status.Ready).To(BeTrue())
		})

		// Nodes still waiting to be excluded are the steady state while the controller
		// drains them, so this reason must not hold the operand back.
		It("is ready when the config is not applied only because nodes are still to be excluded", func() {
			conditions := appliedCondition(metav1.ConditionFalse, toBeExcludedNodesReason)

			status, err := ManagedNodesConfigStatus(ctx, clientWithConditions(conditions), runaiConfig)

			Expect(err).ToNot(HaveOccurred())
			Expect(status.Ready).To(BeTrue())
		})

		It("is not ready when the config failed to apply for any other reason", func() {
			conditions := appliedCondition(metav1.ConditionFalse, "SomeOtherReason")

			status, err := ManagedNodesConfigStatus(ctx, clientWithConditions(conditions), runaiConfig)

			Expect(err).ToNot(HaveOccurred())
			Expect(status.Ready).To(BeFalse())
			Expect(status.Reasons).To(ContainElement(ContainSubstring("is not applied: some message")))
		})
	})
})
