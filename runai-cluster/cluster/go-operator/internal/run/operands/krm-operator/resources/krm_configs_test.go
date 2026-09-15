/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

func defaultRunaiConfig() *runaiv1.RunaiConfig {
	runaiConfig := &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "runai"},
	}
	runaiConfig.SetDefaultsWhereNeeded()
	return runaiConfig
}

func krmConfigFor(runaiConfig *runaiv1.RunaiConfig) *kaiv1alpha1.KRMConfig {
	Expect(kaiv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	k8sClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()

	resource, err := KRMConfigForRunaiConfig(context.TODO(), k8sClient, runaiConfig)
	Expect(err).NotTo(HaveOccurred())
	Expect(resource).NotTo(BeNil())

	return resource.(*kaiv1alpha1.KRMConfig)
}

var _ = Describe("KRMConfigForRunaiConfig", func() {
	It("names the singleton and takes the namespace from the RunaiConfig", func() {
		krmConfig := krmConfigFor(defaultRunaiConfig())

		Expect(krmConfig.GetName()).To(Equal(kaiv1alpha1.KRMConfigSingletonName))
		Expect(krmConfig.Spec.Namespace).To(Equal("runai"))
	})

	It("pins the vocabulary the scheduler was installed with", func() {
		global := krmConfigFor(defaultRunaiConfig()).Spec.Global
		Expect(global).NotTo(BeNil())

		Expect(global.SchedulerName).To(Equal(ptr.To(clustercommonconstants.RunaiScheduler)))
		Expect(global.QueueLabelKey).To(Equal(ptr.To(clustercommonconstants.RunaiQueueLabel)))
		Expect(global.NodePoolLabelKey).To(Equal(ptr.To(clustercommonconstants.NodePoolNameLabel)))
		// Deliberately the same key as QueueLabelKey.
		Expect(global.NamespaceProjectLabelKey).To(Equal(ptr.To(clustercommonconstants.RunaiQueueLabel)))
		Expect(global.ProjectLabelKey).To(Equal(ptr.To(clustercommonconstants.RunaiProjectLabel)))
		Expect(global.FinalizerDomain).To(Equal(ptr.To(clustercommonconstants.RunaiOrgDomain)))
		Expect(global.EnforceSchedulerAnnotationKey).
			To(Equal(ptr.To(clustercommonconstants.EnforceSchedulerAnnotationName)))
	})

	It("leaves leader election to KRM", func() {
		Expect(krmConfigFor(defaultRunaiConfig()).Spec.Global.LeaderElection).To(BeNil())
	})

	It("asks for both service monitors", func() {
		serviceMonitor := krmConfigFor(defaultRunaiConfig()).Spec.Global.ServiceMonitor
		Expect(serviceMonitor).NotTo(BeNil())
		Expect(serviceMonitor.Enabled).To(Equal(ptr.To(true)))
		Expect(serviceMonitor.Accounting).To(Equal(ptr.To(true)))
	})

	Context("pod-group-assigner", func() {
		It("turns the Pod webhook off, since runai's admission-controller serves it", func() {
			podGroupAssigner := krmConfigFor(defaultRunaiConfig()).Spec.PodGroupAssigner
			Expect(podGroupAssigner).NotTo(BeNil())
			Expect(podGroupAssigner.Webhooks).NotTo(BeNil())
			Expect(podGroupAssigner.Webhooks.EnablePodWebhook).To(Equal(ptr.To(false)))
		})

		It("carries the client rate on the service, not as extra args", func() {
			podGroupAssigner := krmConfigFor(defaultRunaiConfig()).Spec.PodGroupAssigner
			Expect(podGroupAssigner.ExtraArgs).To(BeEmpty())
			Expect(podGroupAssigner.Service.K8sClientConfig).NotTo(BeNil())
			// KAI would otherwise default these to 20/100.
			Expect(podGroupAssigner.Service.K8sClientConfig.QPS).To(Equal(ptr.To(50)))
			Expect(podGroupAssigner.Service.K8sClientConfig.Burst).To(Equal(ptr.To(300)))
		})

		It("leaves image and disruption budget to KRM", func() {
			podGroupAssigner := krmConfigFor(defaultRunaiConfig()).Spec.PodGroupAssigner
			Expect(podGroupAssigner.Service).NotTo(BeNil())
			Expect(podGroupAssigner.Service.Image).To(BeNil())
			Expect(podGroupAssigner.Service.PodDisruptionBudget).To(BeNil())
		})

		It("carries the webhook ports and replica count", func() {
			podGroupAssigner := krmConfigFor(defaultRunaiConfig()).Spec.PodGroupAssigner
			Expect(podGroupAssigner.ControllerService).NotTo(BeNil())
			Expect(podGroupAssigner.ControllerService.Webhook).NotTo(BeNil())
			Expect(podGroupAssigner.ControllerService.Webhook.Port).To(Equal(ptr.To(int32(443))))
			Expect(podGroupAssigner.ControllerService.Webhook.TargetPort).To(Equal(ptr.To(int32(8443))))
			Expect(podGroupAssigner.Replicas).NotTo(BeNil())
		})

		It("carries resources through", func() {
			runaiConfig := defaultRunaiConfig()
			podGroupAssigner := krmConfigFor(runaiConfig).Spec.PodGroupAssigner
			Expect(podGroupAssigner.Service.Resources).NotTo(BeNil())
		})

		It("preserves an overridden client rate", func() {
			runaiConfig := defaultRunaiConfig()
			runaiConfig.Spec.PodGroupAssigner.K8sClientConfig.QPS = ptr.To(17)

			podGroupAssigner := krmConfigFor(runaiConfig).Spec.PodGroupAssigner
			Expect(podGroupAssigner.Service.K8sClientConfig.QPS).To(Equal(ptr.To(17)))
		})

		It("carries the enablement toggle through", func() {
			runaiConfig := defaultRunaiConfig()
			runaiConfig.Spec.PodGroupAssigner.Enabled = ptr.To(false)

			podGroupAssigner := krmConfigFor(runaiConfig).Spec.PodGroupAssigner
			Expect(podGroupAssigner.Service.Enabled).To(Equal(ptr.To(false)))
		})
	})
})
