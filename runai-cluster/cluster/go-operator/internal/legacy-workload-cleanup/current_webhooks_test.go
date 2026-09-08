/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Current webhook configurations", func() {
	It("removes retired resources from the pod-controller mutation rules and keeps the rest", func() {
		webhook := currentPodControllerMutatorWithRetiredResource()
		k8sClient := runnerClient(webhook)

		Expect(stripRetiredResourcesFromCurrentWebhooks(context.Background(), k8sClient)).To(Succeed())

		stored := &admissionv1.MutatingWebhookConfiguration{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: webhook.Name}, stored)).To(Succeed())
		Expect(stored.Webhooks[0].Rules).To(HaveLen(1))
		Expect(stored.Webhooks[0].Rules[0].Resources).To(Equal([]string{"statefulsets", "jobs", "deployments"}))
		Expect(stored.Webhooks[0].Rules[0].APIGroups).To(Equal([]string{"*"}))
		Expect(stored.Webhooks[0].Rules[0].Operations).To(Equal(
			[]admissionv1.OperationType{admissionv1.Create, admissionv1.Update}))
	})

	It("drops a rule that only targeted retired resources", func() {
		webhook := currentPodControllerMutatorWithRetiredResource()
		webhook.Webhooks[0].Rules = append(webhook.Webhooks[0].Rules, admissionv1.RuleWithOperations{
			Operations: []admissionv1.OperationType{admissionv1.Create},
			Rule:       admissionv1.Rule{APIGroups: []string{"run.ai"}, APIVersions: []string{"v1"}, Resources: []string{"runaijobs"}},
		})
		k8sClient := runnerClient(webhook)

		Expect(stripRetiredResourcesFromCurrentWebhooks(context.Background(), k8sClient)).To(Succeed())

		stored := &admissionv1.MutatingWebhookConfiguration{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: webhook.Name}, stored)).To(Succeed())
		Expect(stored.Webhooks[0].Rules).To(HaveLen(1))
		Expect(stored.Webhooks[0].Rules[0].Resources).To(Equal([]string{"statefulsets", "jobs", "deployments"}))
	})

	It("does not update a pod-controller mutation webhook that is already clean", func() {
		webhook := currentPodControllerMutatorWithRetiredResource()
		webhook.Webhooks[0].Rules[0].Resources = []string{"statefulsets", "jobs", "deployments"}
		k8sClient := runnerClient(webhook)
		before := &admissionv1.MutatingWebhookConfiguration{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: webhook.Name}, before)).To(Succeed())

		Expect(stripRetiredResourcesFromCurrentWebhooks(context.Background(), k8sClient)).To(Succeed())

		after := &admissionv1.MutatingWebhookConfiguration{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: webhook.Name}, after)).To(Succeed())
		Expect(after.ResourceVersion).To(Equal(before.ResourceVersion))
	})

	It("succeeds when the pod-controller mutation webhook does not exist", func() {
		Expect(stripRetiredResourcesFromCurrentWebhooks(context.Background(), runnerClient())).To(Succeed())
	})

	It("leaves webhook configurations outside the current allowlist untouched", func() {
		unknown := legacyRetiredWebhookMutator("custom-retired-workload-mutation")
		k8sClient := runnerClient(unknown)

		Expect(stripRetiredResourcesFromCurrentWebhooks(context.Background(), k8sClient)).To(Succeed())

		stored := &admissionv1.MutatingWebhookConfiguration{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: unknown.Name}, stored)).To(Succeed())
		Expect(stored.Webhooks[0].Rules[0].Resources).To(Equal([]string{"distributedinferenceworkloads"}))
	})

	It("lists only the pod-controller mutation webhook as a current webhook", func() {
		Expect(currentWebhookConfigurations).To(HaveLen(1))
		Expect(currentWebhookConfigurations[0].kind).To(Equal("MutatingWebhookConfiguration"))
		Expect(currentWebhookConfigurations[0].names).To(Equal([]string{"runai-pod-controller-mutation"}))
	})
})
