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
	globalconfig "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Legacy workload cleanup transaction", func() {
	It("deletes legacy deployments without source annotations or controller flags", func() {
		deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      "workload-controller",
		}}
		k8sClient := runnerClient(deployment)

		Expect(testRunner(k8sClient, allowAllReviewer{}).Run(context.Background())).To(Succeed())
		err := k8sClient.Get(context.Background(), client.ObjectKey{
			Namespace: runaiv1.DefaultNamespace,
			Name:      "workload-controller",
		}, &appsv1.Deployment{})
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("checks every permission before deleting a deployment", func() {
		deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      "workload-controller",
		}}
		k8sClient := runnerClient(deployment)

		Expect(testRunner(k8sClient, denyAllReviewer{}).Run(context.Background())).To(MatchError(ContainSubstring("permission denied")))
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{
			Namespace: runaiv1.DefaultNamespace, Name: "workload-controller",
		}, &appsv1.Deployment{})).To(Succeed())
	})

	It("does not repeat the Operator readiness gate inside the cleanup Job", func() {
		k8sClient := runnerClient()

		Expect(testRunner(k8sClient, allowAllReviewer{}).Run(context.Background())).To(Succeed())
	})

	It("succeeds when retried after workload preparation already removed the finalizer", func() {
		k8sClient := runnerClient(
			validWorkloadOwnershipValidator(), validWorkloadOwnershipMutator())
		config := &runaiv1.RunaiConfig{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "runai", Name: "runai"}, config)).To(Succeed())
		config.Spec.Global = &globalconfig.Global{EnableWorkloadOwnershipProtection: ptr.To(true)}
		Expect(k8sClient.Update(context.Background(), config)).To(Succeed())
		workload := newObject("run.ai/v2alpha1", "TrainingWorkload", "project-a", "training", "training-uid")
		workload.SetFinalizers([]string{"foreign.example/finalizer"})
		listKinds := map[schema.GroupVersionResource]string{
			legacyWorkloadResources[0].GVR:                                                "TrainingWorkloadList",
			{Version: "v1", Resource: "pods"}:                                             "PodList",
			{Version: "v1", Resource: "services"}:                                         "ServiceList",
			{Version: "v1", Resource: "persistentvolumeclaims"}:                           "PersistentVolumeClaimList",
			{Version: "v1", Resource: "secrets"}:                                          "SecretList",
			{Version: "v1", Resource: "configmaps"}:                                       "ConfigMapList",
			{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}:            "IngressList",
			{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}: "RoleBindingList",
		}
		dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds,
			servedCRD(legacyWorkloadResources[0].CRDName, legacyWorkloadResources[0].GVR.Version, true), workload)
		runner := testRunner(k8sClient, allowAllReviewer{})
		runner.WorkloadDynamic = dynamicClient

		Expect(runner.Run(context.Background())).To(Succeed())
		Expect(runner.Run(context.Background())).To(Succeed())
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
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}: "RoleBindingList",
		})
	return &Runner{
		ManagementClient:   k8sClient,
		WorkloadClient:     k8sClient,
		WorkloadDynamic:    dynamicClient,
		ManagementReviewer: reviewer,
		WorkloadReviewer:   reviewer,
	}
}

func runnerClient(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	Expect(admissionv1.AddToScheme(scheme)).To(Succeed())
	Expect(appsv1.AddToScheme(scheme)).To(Succeed())
	Expect(rbacv1.AddToScheme(scheme)).To(Succeed())
	Expect(runaiv1.AddToScheme(scheme)).To(Succeed())

	config := &runaiv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{
		Namespace: runaiv1.DefaultNamespace, Name: runaiv1.DefaultName,
	}}
	allObjects := []client.Object{config}
	allObjects = append(allObjects, objects...)
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(allObjects...).Build()
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
