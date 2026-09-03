/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package personalcreds

import (
	"context"
	"fmt"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Status", func() {
	var (
		ctx       context.Context
		scheme    *runtime.Scheme
		k8sClient client.Client
		state     *reconcile.State
		plugin    *Plugin
		status    *dromosv1alpha1.PluginStatus
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = newReconcileScheme()
		k8sClient = newReconcileClient(scheme)
		state = reconcile.New()
		status = &dromosv1alpha1.PluginStatus{}
		plugin = NewPlugin(k8sClient, testRunaiNS)
	})

	It("is Disabled when the annotation is absent", func() {
		runner := testRunner.DeepCopy()
		runner.Annotations = nil

		Expect(plugin.Status(ctx, runner, k8sClient, status, state)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
	})

	It("is Ready when the copy step raised no critical issue", func() {
		runner := testRunner.DeepCopy()

		Expect(plugin.Status(ctx, runner, k8sClient, status, state)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
		cond := findCondition(status, CredentialsReadyConditionType)
		Expect(cond).NotTo(BeNil())
		Expect(cond.Status).To(Equal(metav1.ConditionTrue))
		Expect(cond.Reason).To(Equal(ReasonReady))
	})

	It("is Failed and creds-attributed when a critical plugin issue was raised during the copy", func() {
		runner := testRunner.DeepCopy()

		// Simulate the copy step recording a critical Storage issue (e.g. the
		// source secret was missing).
		Expect(state.AddPluginIssue(ctx, plugin, runner, fmt.Errorf("source secret not found"), true, "", issue.Storage)).To(HaveOccurred())

		Expect(plugin.Status(ctx, runner, k8sClient, status, state)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseFailed))
		cond := findCondition(status, CredentialsReadyConditionType)
		Expect(cond).NotTo(BeNil())
		Expect(cond.Status).To(Equal(metav1.ConditionFalse))
		Expect(cond.Reason).To(Equal(ReasonFailed))
	})

	It("ignores non-critical plugin issues", func() {
		runner := testRunner.DeepCopy()

		Expect(state.AddPluginIssue(ctx, plugin, runner, fmt.Errorf("transient"), false, "", issue.Storage)).To(HaveOccurred())

		Expect(plugin.Status(ctx, runner, k8sClient, status, state)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
	})
})

func findCondition(status *dromosv1alpha1.PluginStatus, condType string) *metav1.Condition {
	for i := range status.Conditions {
		if status.Conditions[i].Type == condType {
			return &status.Conditions[i]
		}
	}
	return nil
}
