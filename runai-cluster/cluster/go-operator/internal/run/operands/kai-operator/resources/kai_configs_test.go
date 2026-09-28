/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	"github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/numa_placement_exporter"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	admission_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/admission-controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/binder"
	runaicommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/engine"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
)

const controlPlaneImagePullSecretName = "runai-reg-creds"

func defaultRunaiConfig() *runaiv1.RunaiConfig {
	runaiConfig := &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "runai"},
	}
	runaiConfig.SetDefaultsWhereNeeded()
	return runaiConfig
}

var _ = Describe("KAIConfigForRunaiConfig", func() {
	It("passes sanity with default runai config and populates all components", func() {
		Expect(kaiv1.AddToScheme(scheme.Scheme)).To(Succeed())
		k8sClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()

		resource, err := KAIConfigForRunaiConfig(context.TODO(), k8sClient, defaultRunaiConfig())
		Expect(err).NotTo(HaveOccurred())
		Expect(resource).NotTo(BeNil())

		kaiConfig := resource.(*kaiv1.Config)
		Expect(kaiConfig.GetName()).To(Equal(kaiConfigCRName))
		Expect(kaiConfig.Spec.Namespace).To(Equal("runai"))

		Expect(kaiConfig.Spec.Global).NotTo(BeNil())
		Expect(kaiConfig.Spec.Binder).NotTo(BeNil())
		Expect(kaiConfig.Spec.PodGrouper).NotTo(BeNil())
		Expect(kaiConfig.Spec.QueueController).NotTo(BeNil())
		Expect(kaiConfig.Spec.PodGroupController).NotTo(BeNil())
		Expect(kaiConfig.Spec.Scheduler).NotTo(BeNil())
		Expect(kaiConfig.Spec.NodeScaleAdjuster).NotTo(BeNil())
		Expect(kaiConfig.Spec.NodeScaleAdjuster.Service.Enabled).To(Equal(ptr.To(false)))
		Expect(kaiConfig.Spec.NumaPlacementExporter).To(BeNil())
		Expect(kaiConfig.Spec.Admission).NotTo(BeNil())
		Expect(kaiConfig.Spec.Prometheus).NotTo(BeNil())

		// RunaiConfig defaulting must not materialize KAI image overrides.
		Expect(kaiConfig.Spec.Admission.Service.Image).To(BeNil())
		Expect(kaiConfig.Spec.Scheduler.Service.Image).To(BeNil())
		Expect(kaiConfig.Spec.Binder.Service.Image).To(BeNil())
		Expect(kaiConfig.Spec.Binder.ResourceReservation.Image).To(BeNil())
		Expect(kaiConfig.Spec.PodGrouper.Service.Image).To(BeNil())
		Expect(kaiConfig.Spec.PodGroupController.Service.Image).To(BeNil())
		Expect(kaiConfig.Spec.QueueController.Service.Image).To(BeNil())
		Expect(kaiConfig.Spec.NodeScaleAdjuster.Service.Image).To(BeNil())
		Expect(kaiConfig.Spec.NodeScaleAdjuster.Args.ScalingPodImage).To(BeNil())
	})

	It("ignores deprecated Engine and Core compatibility data", func() {
		runaiConfig := defaultRunaiConfig()
		expected := buildKAIConfig(runaiConfig)

		runaiConfig.Spec.Engine = &engine.Engine{Enabled: ptr.To(true)}
		//nolint:staticcheck // Core is deprecated by design; this test exists to prove it is ignored.
		runaiConfig.Spec.Global.Core = &global.Core{}
		runaiConfig.Spec.AdmissionController.Args = &admission_controller.Args{
			RunAIFractionalMinAllocationEnforcementBytes: 1234,
		}
		runaiConfig.Spec.Binder.Args = &binder.Args{}
		runaiConfig.Spec.Binder.Args.RunAIFractionalMinAllocationEnforcementBytes = ptr.To(5678)
		runaiConfig.Spec.Binder.Args.NodeLevelScheduler = ptr.To(true)

		actual := buildKAIConfig(runaiConfig)
		Expect(actual.Spec).To(Equal(expected.Spec))
	})

	It("preserves explicit node scale adjuster enablement", func() {
		for _, enabled := range []bool{false, true} {
			runaiConfig := defaultRunaiConfig()
			runaiConfig.Spec.NodeScaleAdjuster.Enabled = ptr.To(enabled)

			service := buildKAIConfig(runaiConfig).Spec.NodeScaleAdjuster.Service
			Expect(service.Enabled).To(Equal(ptr.To(enabled)))
		}
	})

	It("preserves NUMA placement exporter configuration", func() {
		runaiConfig := defaultRunaiConfig()
		runaiConfig.Spec.NumaPlacementExporter = &numa_placement_exporter.NumaPlacementExporter{}
		runaiConfig.Spec.NumaPlacementExporter.NodeSelector = map[string]string{"accelerator": "gpu"}
		config := buildKAIConfig(runaiConfig)
		Expect(config.Spec.NumaPlacementExporter.NodeSelector).To(Equal(map[string]string{"accelerator": "gpu"}))
	})

	It("clears existing NUMA placement exporter configuration when unset", func() {
		Expect(kaiv1.AddToScheme(scheme.Scheme)).To(Succeed())
		existing := &kaiv1.Config{ObjectMeta: metav1.ObjectMeta{Name: kaiConfigCRName}}
		existing.Spec.NumaPlacementExporter = &numa_placement_exporter.NumaPlacementExporter{
			NodeSelector: map[string]string{"accelerator": "gpu"},
		}
		k8sClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(existing).Build()

		resource, err := KAIConfigForRunaiConfig(context.TODO(), k8sClient, defaultRunaiConfig())
		Expect(err).NotTo(HaveOccurred())
		Expect(resource.(*kaiv1.Config).Spec.NumaPlacementExporter).To(BeNil())
	})

	It("maps explicit custom scheduler-stack image overrides to KAI exactly", func() {
		runaiConfig := defaultRunaiConfig()
		custom := func(name string) *runaicommon.Image {
			return &runaicommon.Image{
				Registry: ptr.To("registry.example.com/team"),
				Name:     ptr.To(name), Tag: ptr.To("kai-compatible"), PullPolicy: v1.PullNever,
			}
		}
		runaiConfig.Spec.RunAIScheduler.Image = custom("scheduler-custom")
		runaiConfig.Spec.Binder.Image = custom("binder-custom")
		runaiConfig.Spec.Binder.ResourceReservation = &binder.ResourceReservation{Image: custom("reservation-custom")}
		runaiConfig.Spec.PodGrouper.Image = custom("podgrouper-custom")
		runaiConfig.Spec.PodGroupController.Image = custom("podgroupcontroller-custom")
		runaiConfig.Spec.QueueController.Image = custom("queuecontroller-custom")
		runaiConfig.Spec.NodeScaleAdjuster.Image = custom("nodescaleadjuster-custom")
		runaiConfig.Spec.NodeScaleAdjuster.ScalingPodImage = custom("scalingpod-custom")
		// Re-run the same defaulting path used by reconciliation after images
		// have been provided by the user.
		runaiConfig.SetDefaultsWhereNeeded()

		config := buildKAIConfig(runaiConfig)
		images := map[string]*kaicommon.Image{
			"scheduler-custom":          config.Spec.Scheduler.Service.Image,
			"binder-custom":             config.Spec.Binder.Service.Image,
			"reservation-custom":        config.Spec.Binder.ResourceReservation.Image,
			"podgrouper-custom":         config.Spec.PodGrouper.Service.Image,
			"podgroupcontroller-custom": config.Spec.PodGroupController.Service.Image,
			"queuecontroller-custom":    config.Spec.QueueController.Service.Image,
			"nodescaleadjuster-custom":  config.Spec.NodeScaleAdjuster.Service.Image,
			"scalingpod-custom":         config.Spec.NodeScaleAdjuster.Args.ScalingPodImage,
		}
		for name, image := range images {
			Expect(image).NotTo(BeNil())
			Expect(image.Repository).To(Equal(ptr.To("registry.example.com/team")))
			Expect(image.Name).To(Equal(ptr.To(name)))
			Expect(image.Tag).To(Equal(ptr.To("kai-compatible")))
			Expect(image.PullPolicy).To(Equal(ptr.To(v1.PullNever)))
		}
	})

	It("preserves partial image overrides without filling missing fields", func() {
		runaiConfig := defaultRunaiConfig()
		runaiConfig.Spec.RunAIScheduler.Image = &runaicommon.Image{Tag: ptr.To("custom-tag")}
		runaiConfig.SetDefaultsWhereNeeded()

		image := buildKAIConfig(runaiConfig).Spec.Scheduler.Service.Image
		Expect(image).NotTo(BeNil())
		Expect(image.Repository).To(BeNil())
		Expect(image.Name).To(BeNil())
		Expect(image.Tag).To(Equal(ptr.To("custom-tag")))
		Expect(image.PullPolicy).To(BeNil())
	})

	It("preserves an explicit image even when it matches the former stock tuple", func() {
		runaiConfig := defaultRunaiConfig()
		runaiConfig.Spec.Internal.Global.ImageRegistry = "runai.example.com/stock"
		runaiConfig.Spec.Internal.Global.Version = "2.26.99"
		runaiConfig.Spec.RunAIScheduler.Image = &runaicommon.Image{
			Registry: ptr.To("runai.example.com/stock"), Name: ptr.To("runai-scheduler"),
			Tag: ptr.To("2.26.99"), PullPolicy: v1.PullAlways,
		}

		image := buildKAIConfig(runaiConfig).Spec.Scheduler.Service.Image
		Expect(image).NotTo(BeNil())
		Expect(image.Repository).To(Equal(ptr.To("runai.example.com/stock")))
		Expect(image.Name).To(Equal(ptr.To("runai-scheduler")))
		Expect(image.Tag).To(Equal(ptr.To("2.26.99")))
		Expect(image.PullPolicy).To(Equal(ptr.To(v1.PullAlways)))
	})

	It("preserves deprecated repository and environment image semantics", func() {
		runaiConfig := defaultRunaiConfig()
		runaiConfig.Spec.RunAIScheduler.Image = &runaicommon.Image{
			Repository: ptr.To("registry.example.com/runai-"),
			Env:        ptr.To("prod"),
			Name:       ptr.To("scheduler-custom"),
			Tag:        ptr.To("kai-compatible"),
		}

		image := buildKAIConfig(runaiConfig).Spec.Scheduler.Service.Image
		Expect(image).NotTo(BeNil())
		Expect(image.Repository).To(Equal(ptr.To("registry.example.com/runai-prod")))
		Expect(image.Name).To(Equal(ptr.To("scheduler-custom")))
		Expect(image.Tag).To(Equal(ptr.To("kai-compatible")))
		Expect(image.PullPolicy).To(BeNil())
	})

	It("does not use the standalone RunAI admission controller as KAI admission configuration", func() {
		runaiConfig := defaultRunaiConfig()
		Expect(runaiConfig.Spec.AdmissionController.Image).NotTo(BeNil())
		runaiConfig.Spec.AdmissionController.Image = &runaicommon.Image{
			Registry: ptr.To("registry.example.com/runai"),
			Name:     ptr.To("runai-admission-custom"),
			Tag:      ptr.To("custom"),
		}

		admission := buildKAIConfig(runaiConfig).Spec.Admission
		Expect(admission.Service).NotTo(BeNil())
		Expect(admission.Service.Image).To(BeNil())
		Expect(admission.Service.Enabled).To(BeNil())
		Expect(admission.Service.Resources).To(BeNil())
		Expect(admission.Service.Affinity).To(BeNil())
		Expect(admission.Replicas).To(BeNil())
	})

	It("leaves component replicas unset so KAI derives them from the global value", func() {
		runaiConfig := defaultRunaiConfig()
		runaiConfig.Spec.Global.ReplicaCount = ptr.To(int32(3))
		runaiConfig.Spec.AdmissionController.Replicas = ptr.To(int32(2))
		config := buildKAIConfig(runaiConfig)
		Expect(config.Spec.Global.ReplicaCount).To(Equal(ptr.To(int32(3))))
		Expect(config.Spec.Scheduler.Replicas).To(BeNil())
		Expect(config.Spec.Binder.Replicas).To(BeNil())
		Expect(config.Spec.PodGrouper.Replicas).To(BeNil())
		Expect(config.Spec.QueueController.Replicas).To(BeNil())
		Expect(config.Spec.PodGroupController.Replicas).To(BeNil())
		Expect(config.Spec.Admission.Replicas).To(BeNil())
	})

	It("points KAI's dedicated namespaces at the RunAI namespaces", func() {
		kaiConfig := buildKAIConfig(defaultRunaiConfig())
		Expect(kaiConfig.Spec.NodeScaleAdjuster.Args).NotTo(BeNil())
		Expect(kaiConfig.Spec.NodeScaleAdjuster.Args.NodeScaleNamespace).To(Equal(ptr.To(runaiv1.ScaleAdjustNamespace)))
		Expect(kaiConfig.Spec.Binder.ResourceReservation).NotTo(BeNil())
		Expect(kaiConfig.Spec.Binder.ResourceReservation.Namespace).To(Equal(ptr.To(runaiv1.ReservationNamespace)))
	})

	It("sets the queue label key to the RunAI queue label", func() {
		kaiConfig := buildKAIConfig(defaultRunaiConfig())
		Expect(kaiConfig.Spec.Global.QueueLabelKey).To(Equal(ptr.To(clustercommonconstants.RunaiQueueLabel)))
	})

	It("sets the RunAI scheduler vocabulary on global, matching the engine adapter", func() {
		runaiConfig := defaultRunaiConfig()
		// The chart delivers the control-plane pull secret through this list.
		runaiConfig.Spec.Internal.Global.AdditionalImagePullSecrets = []string{controlPlaneImagePullSecretName}
		kaiConfig := buildKAIConfig(runaiConfig)
		Expect(kaiConfig.Spec.Global.SchedulerName).To(Equal(ptr.To(clustercommonconstants.RunaiScheduler)))
		Expect(kaiConfig.Spec.Global.NodePoolLabelKey).To(Equal(ptr.To(clustercommonconstants.NodePoolNameLabel)))
		Expect(kaiConfig.Spec.Global.RequireDefaultPodAntiAffinityTerm).To(Equal(ptr.To(false)))
		// pull secrets come from the chart-configured list; deprecated singular field unset
		Expect(kaiConfig.Spec.Global.ImagePullSecret).To(BeNil())
		Expect(kaiConfig.Spec.Global.ImagePullSecrets).To(ContainElement(controlPlaneImagePullSecretName))
	})

	// The chart puts KAI on -fips images for both "on" and "only", so this flag is
	// the only thing separating the two: without it "only" silently degrades to the
	// image default on every operator-managed KAI pod.
	DescribeTable("carries FIPS run-time strictness onto the CR",
		func(fipsMode string, expected bool) {
			runaiConfig := defaultRunaiConfig()
			runaiConfig.Spec.Internal.Global.FipsMode = fipsMode

			Expect(buildKAIConfig(runaiConfig).Spec.Global.FIPSOnly).To(Equal(ptr.To(expected)))
		},
		Entry("only", "only", true),
		Entry("on", "on", false),
		Entry("off", "off", false),
		Entry("unset", "", false),
	)

	It("sets reservation identity and admission/webhook fields like the engine adapter", func() {
		kaiConfig := buildKAIConfig(defaultRunaiConfig())

		rr := kaiConfig.Spec.Binder.ResourceReservation
		Expect(rr.ServiceAccountName).To(Equal(ptr.To(clustercommonconstants.ResourceReservationServiceAccountName)))
		Expect(rr.AppLabel).To(Equal(ptr.To(clustercommonconstants.ResourceReservationAppLabelValue)))
		Expect(rr.RuntimeClassName).To(Equal(ptr.To("")))

		adm := kaiConfig.Spec.Admission
		Expect(adm.GPUSharing).To(Equal(ptr.To(true)))
		Expect(adm.QueueLabelSelector).To(Equal(ptr.To(true)))
		Expect(adm.ValidatingWebhookConfigurationName).To(Equal(ptr.To("validating-engine-admission")))
		Expect(adm.MutatingWebhookConfigurationName).To(Equal(ptr.To("mutating-engine-admission")))

		Expect(kaiConfig.Spec.QueueController.Webhooks).NotTo(BeNil())
		Expect(kaiConfig.Spec.QueueController.Webhooks.WebhookConfigurationNamePrefix).To(Equal(ptr.To("runai-queue-validation-")))
		Expect(kaiConfig.Spec.PodGroupController.Webhooks).NotTo(BeNil())
		Expect(kaiConfig.Spec.PodGroupController.Webhooks.WebhookConfigurationNamePrefix).To(Equal(ptr.To("runai-podgroup-validation-")))
		Expect(kaiConfig.Spec.PodGroupController.Webhooks.EnableValidation).To(Equal(ptr.To(true)))
	})

	It("passes resources through 1:1 to the nested KAI service", func() {
		runaiConfig := defaultRunaiConfig()
		binderResources := &binder.Resources{
			Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse("250m")},
			Limits:   v1.ResourceList{v1.ResourceMemory: resource.MustParse("512Mi")},
		}
		runaiConfig.Spec.Binder.Resources = binderResources

		kaiConfig := buildKAIConfig(runaiConfig)
		Expect(kaiConfig.Spec.Binder.Service).NotTo(BeNil())
		Expect(kaiConfig.Spec.Binder.Service.Resources).NotTo(BeNil())
		Expect((*v1.ResourceRequirements)(kaiConfig.Spec.Binder.Service.Resources)).
			To(Equal((*v1.ResourceRequirements)(binderResources)))
	})

	It("defaults prometheus persistent storage to false", func() {
		kaiConfig := buildKAIConfig(defaultRunaiConfig())
		Expect(kaiConfig.Spec.Prometheus.EnablePersistentStorage).NotTo(BeNil())
		Expect(*kaiConfig.Spec.Prometheus.EnablePersistentStorage).To(BeFalse())
	})

	It("propagates the scheduler GOGC value", func() {
		runaiConfig := defaultRunaiConfig()
		runaiConfig.Spec.RunAIScheduler.GOGC = ptr.To(123)
		kaiConfig := buildKAIConfig(runaiConfig)
		Expect(kaiConfig.Spec.Scheduler.GOGC).To(Equal(ptr.To(123)))
	})
})

var _ = Describe("convertVPA", func() {
	It("returns nil for nil input", func() {
		Expect(convertVPA(nil)).To(BeNil())
	})

	It("converts all fields", func() {
		mode := vpav1.UpdateModeAuto
		vpa := &global.VPASpec{
			Enabled:      ptr.To(true),
			UpdatePolicy: &vpav1.PodUpdatePolicy{UpdateMode: &mode},
		}
		result := convertVPA(vpa)
		Expect(result).NotTo(BeNil())
		Expect(*result.Enabled).To(BeTrue())
		Expect(*result.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateModeAuto))
	})
})

var _ = Describe("KAIConfigStatus", func() {
	It("reports NotReady when the kai-config is missing", func() {
		Expect(kaiv1.AddToScheme(scheme.Scheme)).To(Succeed())
		k8sClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()

		status, err := KAIConfigStatus(context.TODO(), k8sClient, defaultRunaiConfig())
		Expect(err).NotTo(HaveOccurred())
		Expect(status.Ready).To(BeFalse())
	})

	It("reports Ready when the kai-config has reached the Available condition", func() {
		Expect(kaiv1.AddToScheme(scheme.Scheme)).To(Succeed())
		existing := &kaiv1.Config{
			ObjectMeta: metav1.ObjectMeta{Name: kaiConfigCRName, UID: "kai-config-uid"},
			Status: kaiv1.ConfigStatus{
				Conditions: []metav1.Condition{{
					Type:   string(kaiv1.ConditionTypeAvailable),
					Status: metav1.ConditionTrue,
				}},
			},
		}
		k8sClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(existing).Build()

		status, err := KAIConfigStatus(context.TODO(), k8sClient, defaultRunaiConfig())
		Expect(err).NotTo(HaveOccurred())
		Expect(status.Ready).To(BeTrue())
	})
})

func buildKAIConfig(runaiConfig *runaiv1.RunaiConfig) *kaiv1.Config {
	Expect(kaiv1.AddToScheme(scheme.Scheme)).To(Succeed())
	k8sClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	resource, err := KAIConfigForRunaiConfig(context.TODO(), k8sClient, runaiConfig)
	Expect(err).NotTo(HaveOccurred())
	return resource.(*kaiv1.Config)
}
