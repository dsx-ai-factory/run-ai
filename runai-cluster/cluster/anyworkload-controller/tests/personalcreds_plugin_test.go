/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/anyworkload-controller/plugins/personalcreds"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	pcreds "github.com/run-ai/runai/runai-cluster/cluster/common/personalcreds"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

// known-good fixture: personalcredsKey decrypts personalcredsEncryptedValue to "private-value-1"
const personalcredsKey = "905c18f0839affcddca117dcac633f79" // #nosec G101 -- test fixture, not a real credential

var personalcredsEncryptedValue = []byte{
	168, 221, 244, 73, 135, 225, 25, 104, 145, 32, 118, 20, 196, 29, 72, 165,
	224, 40, 97, 180, 83, 157, 205, 140, 158, 186, 60, 73, 173, 193, 146, 219,
	201, 199, 65, 36, 241, 251, 52, 131, 38, 155, 0,
}

var _ = Describe("Personalcreds Plugin Integration Tests", Serial, Ordered, func() {
	var (
		k8sManager     ctrl.Manager
		k8sClient      client.Client
		managerCtx     context.Context
		managerCancel  context.CancelFunc
		workloadNS     = "runai-personalcreds-test"
		controllerName = "personalcreds-plugin-test-controller"
	)

	BeforeAll(func() {
		managerCtx, managerCancel = context.WithCancel(ctx)

		Expect(os.Setenv(commonconstants.RunaiAdsEncryptionKeyEnvVar, personalcredsKey)).To(Succeed())

		absCertDir, err := filepath.Abs(webhookCertDir)
		Expect(err).NotTo(HaveOccurred())

		k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
			Metrics: metricsserver.Options{
				BindAddress: "0",
			},
			WebhookServer: webhook.NewServer(webhook.Options{
				CertDir: absCertDir,
				Port:    webhookPort,
			}),
		})
		Expect(err).NotTo(HaveOccurred())

		k8sClient = k8sManager.GetClient()

		plugin := personalcreds.NewPlugin(k8sClient, commonconstants.NamespaceRunAi)
		registry := framework.NewRegistry(plugin)
		fw := framework.NewFramework(registry, nil)
		c := controller.NewWorkloadRunnerController(k8sClient, fw, k8sManager.GetEventRecorderFor("workloadrunner-controller-test")).WithName(controllerName)
		Expect(c.SetupWithManager(managerCtx, k8sManager)).To(Succeed())

		setupWebhooks(fw, k8sClient, managerCtx, k8sManager)

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(managerCtx)).To(Succeed())
		}()

		Expect(k8sClient.Create(managerCtx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: commonconstants.NamespaceRunAi},
		})).To(Succeed())
		Expect(k8sClient.Create(managerCtx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: workloadNS},
		})).To(Succeed())

		waitForSetupToComplete(k8sClient, managerCtx, webhookPort)
	})

	AfterAll(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		deleteAllForType(managerCtx, k8sClient, &corev1.SecretList{})
		managerCancel()
		waitForWebhookToBeClosed(webhookUrlValidate)
		Expect(os.Unsetenv(commonconstants.RunaiAdsEncryptionKeyEnvVar)).To(Succeed())
	})

	AfterEach(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		deleteAllForType(managerCtx, k8sClient, &corev1.SecretList{})
	})

	It("rejects a runner whose source secret is missing in the runai namespace", func() {
		jobset := test.NewJobSetBuilder("wl-missing", workloadNS).WithDefault().Build()
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runner-missing-source",
				Namespace: workloadNS,
				Annotations: map[string]string{
					pcreds.Annotation: "- sourceSecret: not-there\n",
				},
			},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{Template: marshalRawExtension(jobset)},
		}
		err := k8sClient.Create(managerCtx, runner)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("not-there"))
	})

	It("materializes a destination secret with decrypted data and ownerRef → runner", func() {
		source := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-cred",
				Namespace: commonconstants.NamespaceRunAi,
			},
			Data: map[string][]byte{"token": personalcredsEncryptedValue},
		}
		Expect(k8sClient.Create(managerCtx, source)).To(Succeed())

		jobset := test.NewJobSetBuilder("wl-ok", workloadNS).WithDefault().Build()
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runner-ok",
				Namespace: workloadNS,
				Annotations: map[string]string{
					pcreds.Annotation: "- sourceSecret: my-cred\n",
				},
			},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{Template: marshalRawExtension(jobset)},
		}
		Expect(k8sClient.Create(managerCtx, runner)).To(Succeed())

		dest := &corev1.Secret{}
		Eventually(func() error {
			return k8sClient.Get(managerCtx, types.NamespacedName{
				Name:      "pc-runner-ok-my-cred",
				Namespace: workloadNS,
			}, dest)
		}, timeout, interval).Should(Succeed())

		Expect(dest.Data).To(HaveKeyWithValue("token", []byte("private-value-1")))
		Expect(dest.OwnerReferences).To(HaveLen(1))
		Expect(dest.OwnerReferences[0].Kind).To(Equal("WorkloadRunner"))
		Expect(dest.OwnerReferences[0].Name).To(Equal(runner.Name))
		Expect(metav1.IsControlledBy(dest, runner)).To(BeTrue())

		// JobSet still produced as usual
		js := &jobsetv1alpha2.JobSet{}
		Eventually(func() error {
			return k8sClient.Get(managerCtx, types.NamespacedName{
				Name:      "wl-ok",
				Namespace: workloadNS,
			}, js)
		}, timeout, interval).Should(Succeed())
	})
})
