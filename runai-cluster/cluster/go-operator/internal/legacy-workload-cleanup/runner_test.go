/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	advicerestartcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/advice-restart-controller"
	anyworkloadcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/anyworkload-controller"
	globalconfig "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	inferencecontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/inference-workload-controller"
	jobcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/job-controller"
	internalv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	workloadcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/workload-controller"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Legacy workload cleanup transaction", func() {
	It("disables all three legacy controllers after every safety check succeeds", func() {
		k8sClient := runnerClient("2.26.0", true)
		runner := testRunner(k8sClient, allowAllReviewer{})

		Expect(runner.Run(context.Background())).To(Succeed())
		config := &runaiv1.RunaiConfig{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "runai", Name: "runai"}, config)).To(Succeed())
		Expect(*config.Spec.WorkloadController.Enabled).To(BeFalse())
		Expect(*config.Spec.InferenceWorkloadController.Enabled).To(BeFalse())
		Expect(*config.Spec.JobController.Enabled).To(BeFalse())
	})

	It("accepts a cleanup request from a pre-3.0 source on a branch build", func() {
		k8sClient := runnerClient("2.26.0", true)
		config := getRunaiConfig(k8sClient)
		config.Spec.Internal.Global.Version = "0.0.0-12345.v300-main"
		Expect(k8sClient.Update(context.Background(), config)).To(Succeed())

		Expect(testRunner(k8sClient, allowAllReviewer{}).Run(context.Background())).To(Succeed())
		expectControllerFlags(k8sClient, false)
	})

	It("keeps every controller enabled when a permission check fails", func() {
		k8sClient := runnerClient("2.26.0", true)
		runner := testRunner(k8sClient, denyAllReviewer{})

		Expect(runner.Run(context.Background())).To(MatchError(ContainSubstring("permission denied")))
		expectControllerFlags(k8sClient, true)
	})

	It("rejects upgrade provenance from a 3.0 source", func() {
		k8sClient := runnerClient("3.0.0", true)
		runner := testRunner(k8sClient, denyAllReviewer{})

		Expect(runner.Run(context.Background())).To(MatchError(ContainSubstring("does not require legacy cleanup")))
		expectControllerFlags(k8sClient, true)
	})

	It("rejects a cleanup request without a source-version annotation", func() {
		k8sClient := runnerClient("2.26.0", true)
		config := getRunaiConfig(k8sClient)
		config.Annotations = nil
		Expect(k8sClient.Update(context.Background(), config)).To(Succeed())

		Expect(testRunner(k8sClient, allowAllReviewer{}).Run(context.Background())).To(
			MatchError(ContainSubstring("cleanup source version annotation is missing")))
		expectControllerFlags(k8sClient, true)
	})

	It("does not repeat the Operator readiness gate inside the cleanup Job", func() {
		k8sClient := runnerClient("2.26.0", false)

		Expect(testRunner(k8sClient, allowAllReviewer{}).Run(context.Background())).To(Succeed())
		expectControllerFlags(k8sClient, false)
	})

	It("does not use Advice Restart Controller state as a cleanup gate", func() {
		k8sClient := runnerClient("2.26.0", true)
		config := &runaiv1.RunaiConfig{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "runai", Name: "runai"}, config)).To(Succeed())
		config.Spec.AdviceRestartController.Enabled = ptr.To(true)
		Expect(k8sClient.Update(context.Background(), config)).To(Succeed())

		Expect(testRunner(k8sClient, allowAllReviewer{}).Run(context.Background())).To(Succeed())
		expectControllerFlags(k8sClient, false)
	})

	It("does not use AnyWorkload enablement as a cleanup gate", func() {
		k8sClient := runnerClient("2.26.0", true)
		config := getRunaiConfig(k8sClient)
		config.Spec.AnyworkloadController = &anyworkloadcontroller.AnyworkloadController{Enabled: ptr.To(false)}
		Expect(k8sClient.Update(context.Background(), config)).To(Succeed())

		Expect(testRunner(k8sClient, allowAllReviewer{}).Run(context.Background())).To(Succeed())
		expectControllerFlags(k8sClient, false)
	})

	It("does not use the WorkloadRunner CRD as a cleanup gate", func() {
		k8sClient := runnerClient("2.26.0", true)
		runner := testRunner(k8sClient, allowAllReviewer{})
		runner.WorkloadDynamic = dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

		Expect(runner.Run(context.Background())).To(Succeed())
		expectControllerFlags(k8sClient, false)
	})

	It("ignores legacy RestartRequests because this Job does not manage them", func() {
		k8sClient := runnerClient("2.26.0", true)
		restartGVR := schema.GroupVersionResource{Group: "run.ai", Version: "v1alpha1", Resource: "restartrequests"}
		request := newObject("run.ai/v1alpha1", "RestartRequest", "project-a", "active", "active-uid")
		dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
			map[schema.GroupVersionResource]string{restartGVR: "RestartRequestList"},
			servedCRD("restartrequests.run.ai", restartGVR.Version, true), request)
		runner := testRunner(k8sClient, allowAllReviewer{})
		runner.WorkloadDynamic = dynamicClient

		Expect(runner.Run(context.Background())).To(Succeed())
		expectControllerFlags(k8sClient, false)
	})

	It("returns success immediately when the controller flags are already off", func() {
		k8sClient := runnerClient("2.26.0", true)
		config := &runaiv1.RunaiConfig{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "runai", Name: "runai"}, config)).To(Succeed())
		setControllerFlags(config, false)
		Expect(k8sClient.Update(context.Background(), config)).To(Succeed())

		Expect(testRunner(k8sClient, allowAllReviewer{}).Run(context.Background())).To(Succeed())
		expectControllerFlags(k8sClient, false)
	})

	It("completes when controllers are disabled without waiting for operand deletion", func() {
		deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      "workload-controller",
		}}
		k8sClient := runnerClient("2.26.0", true, deployment)
		config := getRunaiConfig(k8sClient)
		setControllerFlags(config, false)
		Expect(k8sClient.Update(context.Background(), config)).To(Succeed())

		runner := testRunner(k8sClient, allowAllReviewer{})
		Expect(runner.Run(context.Background())).To(Succeed())
		expectControllerFlags(k8sClient, false)
	})

	It("does not wait for legacy controller ReplicaSets before completing", func() {
		replicaSet := &appsv1.ReplicaSet{ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      "workload-controller-old",
			Labels:    map[string]string{"app": "workload-controller"},
		}}
		k8sClient := runnerClient("2.26.0", true, replicaSet)
		config := getRunaiConfig(k8sClient)
		setControllerFlags(config, false)
		Expect(k8sClient.Update(context.Background(), config)).To(Succeed())

		runner := testRunner(k8sClient, allowAllReviewer{})
		Expect(runner.Run(context.Background())).To(Succeed())
		expectControllerFlags(k8sClient, false)
	})

	It("succeeds when retried after workload preparation already removed the finalizer", func() {
		k8sClient := runnerClient("2.26.0", true,
			validWorkloadOwnershipValidator(), validWorkloadOwnershipMutator())
		config := getRunaiConfig(k8sClient)
		config.Spec.Global = &globalconfig.Global{EnableWorkloadOwnershipProtection: ptr.To(true)}
		Expect(k8sClient.Update(context.Background(), config)).To(Succeed())
		workload := newObject("run.ai/v2alpha1", "TrainingWorkload", "project-a", "training", "training-uid")
		workload.SetFinalizers([]string{"foreign.example/finalizer"})
		listKinds := map[schema.GroupVersionResource]string{
			legacyWorkloadResources[0].GVR:                                     "TrainingWorkloadList",
			{Version: "v1", Resource: "pods"}:                                  "PodList",
			{Version: "v1", Resource: "services"}:                              "ServiceList",
			{Version: "v1", Resource: "persistentvolumeclaims"}:                "PersistentVolumeClaimList",
			{Version: "v1", Resource: "secrets"}:                               "SecretList",
			{Version: "v1", Resource: "configmaps"}:                            "ConfigMapList",
			{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}: "IngressList",
		}
		dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds,
			servedCRD(legacyWorkloadResources[0].CRDName, legacyWorkloadResources[0].GVR.Version, true), workload)
		runner := testRunner(k8sClient, allowAllReviewer{})
		runner.WorkloadDynamic = dynamicClient

		Expect(runner.Run(context.Background())).To(Succeed())
		expectControllerFlags(k8sClient, false)
	})

})

type allowAllReviewer struct{}

func (allowAllReviewer) Review(context.Context, requiredPermission) (bool, string, error) {
	return true, "", nil
}

type denyAllReviewer struct{}

func (denyAllReviewer) Review(context.Context, requiredPermission) (bool, string, error) {
	return false, "test denial", nil
}

func testRunner(k8sClient client.Client, reviewer permissionReviewer) *Runner {
	dynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	return &Runner{
		ManagementClient:   k8sClient,
		WorkloadClient:     k8sClient,
		WorkloadDynamic:    dynamicClient,
		ManagementReviewer: reviewer,
		WorkloadReviewer:   reviewer,
	}
}

func runnerClient(sourceVersion string, ready bool, objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	Expect(admissionv1.AddToScheme(scheme)).To(Succeed())
	Expect(appsv1.AddToScheme(scheme)).To(Succeed())
	Expect(runaiv1.AddToScheme(scheme)).To(Succeed())

	config := readyRunaiConfig(1)
	config.Namespace = runaiv1.DefaultNamespace
	config.Name = runaiv1.DefaultName
	config.Spec.WorkloadController = &workloadcontroller.WorkloadController{Enabled: ptr.To(true)}
	config.Spec.InferenceWorkloadController = &inferencecontroller.InferenceWorkloadController{Enabled: ptr.To(true)}
	config.Spec.JobController = &jobcontroller.JobController{Enabled: ptr.To(true)}
	config.Spec.Internal = &internalv1.Internal{Global: &internalv1.Global{Version: "3.0.0"}}
	config.Spec.AdviceRestartController = &advicerestartcontroller.AdviceRestartController{Enabled: ptr.To(false)}
	config.Spec.AnyworkloadController = &anyworkloadcontroller.AnyworkloadController{Enabled: ptr.To(true)}
	if !ready {
		config.Status.ObservedGeneration = 0
	}
	config.Annotations = map[string]string{SourceVersionAnnotation: sourceVersion}
	allObjects := []client.Object{config}
	allObjects = append(allObjects, objects...)
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(allObjects...).Build()
}

func expectControllerFlags(k8sClient client.Client, enabled bool) {
	config := &runaiv1.RunaiConfig{}
	Expect(k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "runai", Name: "runai"}, config)).To(Succeed())
	Expect(*config.Spec.WorkloadController.Enabled).To(Equal(enabled))
	Expect(*config.Spec.InferenceWorkloadController.Enabled).To(Equal(enabled))
	Expect(*config.Spec.JobController.Enabled).To(Equal(enabled))
}

func validWorkloadOwnershipValidator() *admissionv1.ValidatingWebhookConfiguration {
	return &admissionv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: workloadOwnershipValidationWebhook},
		Webhooks: []admissionv1.ValidatingWebhook{{
			ClientConfig: admissionv1.WebhookClientConfig{Service: &admissionv1.ServiceReference{
				Namespace: "runai", Name: "anyworkload-controller",
			}},
			Rules: []admissionv1.RuleWithOperations{{
				Operations: []admissionv1.OperationType{admissionv1.Update, admissionv1.Delete},
				Rule:       admissionv1.Rule{APIGroups: []string{"run.ai"}, APIVersions: []string{"v2alpha1"}, Resources: []string{"trainingworkloads"}},
			}},
		}},
	}
}

func validWorkloadOwnershipMutator() *admissionv1.MutatingWebhookConfiguration {
	return &admissionv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: workloadOwnershipMutationWebhook},
		Webhooks: []admissionv1.MutatingWebhook{{
			ClientConfig: admissionv1.WebhookClientConfig{Service: &admissionv1.ServiceReference{
				Namespace: "runai", Name: "anyworkload-controller",
			}},
			Rules: []admissionv1.RuleWithOperations{{
				Operations: []admissionv1.OperationType{admissionv1.Update},
				Rule:       admissionv1.Rule{APIGroups: []string{"run.ai"}, APIVersions: []string{"v2alpha1"}, Resources: []string{"trainingworkloads"}},
			}},
		}},
	}
}
