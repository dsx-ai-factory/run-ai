/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"path/filepath"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/mutator/gvk"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/test/plugins"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	webhookControllerName = "webhook-integration-controller"
)

var _ = Describe("Webhook Integration Tests", Serial, Ordered, func() {
	var (
		k8sManager    ctrl.Manager
		k8sClient     client.Client
		managerCtx    context.Context
		managerCancel context.CancelFunc
	)

	BeforeAll(func() {
		managerCtx, managerCancel = context.WithCancel(ctx)

		// Use static certs from testdata
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

		registry := framework.NewRegistry(
			plugins.NewWebhookValidateNamespaceExistsPlugin(),
			plugins.NewWebhookMutateAddLabelPlugin(),
		)
		fw := framework.NewFramework(registry, nil)
		webhookCtrl := controller.NewWorkloadRunnerController(k8sClient, fw, k8sManager.GetEventRecorderFor(webhookControllerName)).WithName(webhookControllerName)
		err = webhookCtrl.SetupWithManager(managerCtx, k8sManager)
		Expect(err).NotTo(HaveOccurred())

		setupWebhooks(fw, k8sClient, managerCtx, k8sManager)

		go func() {
			defer GinkgoRecover()
			err = k8sManager.Start(managerCtx)
			Expect(err).NotTo(HaveOccurred())
		}()

		waitForSetupToComplete(k8sClient, managerCtx, webhookPort)

	})

	AfterAll(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		managerCancel()
		waitForWebhookToBeClosed(webhookUrlValidate)

	})

	Describe("Webhook Validation", func() {
		It("rejects WorkloadRunner with template missing name", func() {
			template := newPodRawExtension("", "test-namespace")
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-no-name-",
					Namespace:    "default",
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: template,
				},
			}

			err := k8sClient.Create(managerCtx, runner)

			Expect(err).To(HaveOccurred())
			Expect(apierrors.IsInvalid(err)).To(BeTrue())
			expectInvalidCause(err, field.ErrorTypeRequired, "spec.template.metadata.name")
		})

		It("rejects WorkloadRunner with template missing namespace", func() {
			template := newPodRawExtension("test-pod", "")
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-no-ns-",
					Namespace:    "default",
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: template,
				},
			}

			err := k8sClient.Create(managerCtx, runner)

			Expect(err).To(HaveOccurred())
			Expect(apierrors.IsInvalid(err)).To(BeTrue())
			expectInvalidCause(err, field.ErrorTypeRequired, "spec.template.metadata.namespace")
		})

		It("rejects WorkloadRunner with template missing kind", func() {
			template := runtime.RawExtension{
				Raw: []byte(`{"apiVersion":"v1","metadata":{"name":"test","namespace":"test-ns"}}`),
			}
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-no-kind-",
					Namespace:    "default",
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: template,
				},
			}

			err := k8sClient.Create(managerCtx, runner)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("spec.template.kind: Required value"))
		})

		It("rejects WorkloadRunner with template missing apiVersion", func() {
			template := runtime.RawExtension{
				Raw: []byte(`{"kind":"Pod","metadata":{"name":"test","namespace":"test-ns"}}`),
			}
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-no-apiversion-",
					Namespace:    "default",
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: template,
				},
			}

			err := k8sClient.Create(managerCtx, runner)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("spec.template.apiVersion: Required value"))
		})

		It("rejects WorkloadRunner when template name is changed on update", func() {
			runner := createWorkloadRunner(managerCtx, k8sClient, "runner-name-change", "default", newTestPod("test-pod", "test-namespace"), nil)
			defer func() {
				_ = k8sClient.Delete(managerCtx, runner)
			}()

			updatedRunner := runner.DeepCopy()
			updatedRunner.Spec.Template = newPodRawExtension("new-pod-name", "test-namespace")

			err := patchExpectInvalid(k8sClient, managerCtx, updatedRunner)
			expectInvalidCause(err, field.ErrorTypeForbidden, "spec.template.metadata.name")
		})

		It("rejects WorkloadRunner when template namespace is changed on update", func() {
			runner := createWorkloadRunner(managerCtx, k8sClient, "runner-namespace-change", "default", newTestPod("test-pod", "test-namespace"), nil)
			defer func() {
				_ = k8sClient.Delete(managerCtx, runner)
			}()

			updatedRunner := runner.DeepCopy()
			updatedRunner.Spec.Template = newPodRawExtension("test-pod", "different-namespace")

			err := patchExpectInvalid(k8sClient, context.Background(), updatedRunner)
			expectInvalidCause(err, field.ErrorTypeForbidden, "spec.template.metadata.namespace")
		})

		It("rejects WorkloadRunner when GVK is changed on update", func() {
			runner := createWorkloadRunner(managerCtx, k8sClient, "runner-gvk-change", "default", newTestPod("test-pod", "test-namespace"), nil)
			defer func() {
				_ = k8sClient.Delete(managerCtx, runner)
			}()

			updatedRunner := runner.DeepCopy()
			updatedRunner.Spec.Template = newJobRawExtension("test-job", "test-namespace")

			err := patchExpectInvalid(k8sClient, context.Background(), updatedRunner)
			expectInvalidCause(err, field.ErrorTypeForbidden, "spec.template")
		})

		It("accepts WorkloadRunner update when template name, namespace and GVK unchanged", func() {
			runner := createWorkloadRunner(managerCtx, k8sClient, "runner-valid-update", "default", newTestPod("test-pod", "test-namespace"), nil)
			defer func() {
				_ = k8sClient.Delete(managerCtx, runner)
			}()
			Expect(runner.Labels).ToNot(HaveKeyWithValue("updated", "true"))

			updatedRunner := runner.DeepCopy()
			if updatedRunner.Labels == nil {
				updatedRunner.Labels = make(map[string]string)
			}
			updatedRunner.Labels["updated"] = "true"

			Eventually(func() string {
				return updateGetErrorString(k8sClient, context.Background(), updatedRunner)
			}, timeout, interval).Should(Equal(""))

			check := &dromosv1alpha1.WorkloadRunner{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(runner), check)
				if err != nil {
					return false
				}
				_, found := check.Labels["updated"]
				return found
			}, timeout, interval).Should(BeTrue())
			Expect(check.Labels["updated"]).To(Equal("true"))
		})
	})

	Describe("Webhook Mutation", func() {
		It("adds example label when labels are nil", func() {
			runner := createWorkloadRunner(managerCtx, k8sClient, "runner-with-labels-nil", "default", newTestPod("test-pod", "test-namespace"), nil)
			defer func() {
				_ = k8sClient.Delete(managerCtx, runner)
			}()

			created := &dromosv1alpha1.WorkloadRunner{}
			Eventually(func() error {
				return k8sClient.Get(managerCtx, client.ObjectKeyFromObject(runner), created)
			}, timeout, interval).Should(Succeed())
			Expect(created.Labels).NotTo(BeNil())
			Expect(created.Labels["example"]).To(Equal("true"))
		})

		It("check GVK mutation create and enforce labels", func() {
			runner := createWorkloadRunner(managerCtx, k8sClient, "runner-gvk-labels", "default", newTestPod("test-pod", "test-namespace"), nil)

			defer func() {
				_ = k8sClient.Delete(managerCtx, runner)
			}()

			Expect(runner.Labels[gvk.LabelKeyGroup]).To(Equal(""))
			Expect(runner.Labels[gvk.LabelKeyKind]).To(Equal("Pod"))
			Expect(runner.Labels[gvk.LabelKeyVersion]).To(Equal("v1"))

			updatedRunner := runner.DeepCopy()
			updatedRunner.Labels[gvk.LabelKeyGroup] = "ShouldBeEmpty"
			updatedRunner.Labels[gvk.LabelKeyKind] = "NotPod"
			updatedRunner.Labels[gvk.LabelKeyVersion] = "v200"

			Eventually(func() string {
				return updateGetErrorString(k8sClient, context.Background(), updatedRunner)
			}, timeout, interval).Should(Equal(""))

			check := &dromosv1alpha1.WorkloadRunner{}
			Eventually(func() error {
				return k8sClient.Get(context.Background(), client.ObjectKeyFromObject(runner), check)
			}, timeout, interval).Should(Succeed())

			Expect(check.Labels[gvk.LabelKeyGroup]).To(Equal(""))
			Expect(check.Labels[gvk.LabelKeyKind]).To(Equal("Pod"))
			Expect(check.Labels[gvk.LabelKeyVersion]).To(Equal("v1"))
		})
	})
})

func waitForSetupToComplete(k8sClient client.Client, managerCtx context.Context, _ int) {
	// Wait for manager cache to be ready
	Eventually(func() error {
		return k8sClient.List(managerCtx, &dromosv1alpha1.WorkloadRunnerList{})
	}).Should(Succeed())

	// Wait for webhook server to be ready by checking if it's listening
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	Eventually(func() error {
		resp, err := httpClient.Get(webhookUrlValidate)
		if err != nil {
			return err
		}

		defer func() {
			_ = resp.Body.Close()
		}()

		return nil
	}, timeout, interval).Should(Succeed())
}

func updateGetErrorString(k8sClient client.Client, ctx context.Context, runner client.Object) string {
	current := &dromosv1alpha1.WorkloadRunner{}
	err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(runner), current)
	if err != nil {
		return err.Error()
	}

	err = k8sClient.Patch(ctx, runner, client.MergeFrom(current))
	if err != nil {
		return err.Error()
	}
	return ""
}

func patchExpectInvalid(k8sClient client.Client, ctx context.Context, runner client.Object) error {
	GinkgoHelper()
	var patchErr error
	Eventually(func() bool {
		current := &dromosv1alpha1.WorkloadRunner{}
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), current); err != nil {
			return false
		}
		patchErr = k8sClient.Patch(ctx, runner, client.MergeFrom(current))
		return apierrors.IsInvalid(patchErr)
	}, timeout, interval).Should(BeTrue())
	return patchErr
}

func expectInvalidCause(err error, causeType field.ErrorType, fieldPath string) {
	GinkgoHelper()
	statusErr := &apierrors.StatusError{}
	Expect(errors.As(err, &statusErr)).To(BeTrue())
	Expect(statusErr.ErrStatus.Reason).To(Equal(metav1.StatusReasonInvalid))
	Expect(statusErr.ErrStatus.Code).To(Equal(int32(http.StatusUnprocessableEntity)))
	Expect(statusErr.ErrStatus.Details).NotTo(BeNil())
	Expect(statusErr.ErrStatus.Details.Causes).To(ContainElement(SatisfyAll(
		HaveField("Type", metav1.CauseType(causeType)),
		HaveField("Field", fieldPath),
	)))
}
