/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/plugins/knativescaling"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/test"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const knativeMinimum = "autoscaling.knative.dev/min-scale"

var _ = Describe("Knative minimum through WorkloadRunner reconciliation", Serial, Ordered, func() {
	var (
		apiClient     client.Client
		managerCancel context.CancelFunc
		managerDone   chan error
		faultClient   *knativeApplyConflictClient
		crds          []*apiextensionsv1.CustomResourceDefinition
	)

	BeforeAll(func() {
		var err error
		apiClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).NotTo(HaveOccurred())
		// These structural fixtures exercise API-server apply and resourceVersion
		// conflicts. Knative admission, revision creation and scaling need live tests.
		crds, err = envtest.InstallCRDs(cfg, envtest.CRDInstallOptions{CRDs: []*apiextensionsv1.CustomResourceDefinition{
			knativeFixtureCRD("Service", "services"),
			knativeFixtureCRD("Configuration", "configurations"),
			knativeFixtureCRD("Revision", "revisions"),
		}})
		Expect(err).NotTo(HaveOccurred())

		managerCtx, cancelManager := context.WithCancel(ctx)
		managerCancel = cancelManager
		manager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:        scheme.Scheme,
			Metrics:       metricsserver.Options{BindAddress: "0"},
			WebhookServer: webhook.NewServer(webhook.Options{CertDir: webhookCertDir, Port: webhookPort}),
		})
		Expect(err).NotTo(HaveOccurred())
		plugin, err := knativescaling.New(managerCtx, framework.PluginDeps{})
		Expect(err).NotTo(HaveOccurred())
		fw := framework.NewFramework(framework.NewRegistry(plugin), nil)
		faultClient = &knativeApplyConflictClient{Client: manager.GetClient()}
		workloadController := controller.NewWorkloadRunnerController(faultClient, fw,
			&record.FakeRecorder{}).WithName("knative-scaling-test")
		Expect(workloadController.SetupWithManager(managerCtx, manager)).To(Succeed())
		setupWebhooks(fw, apiClient, managerCtx, manager)
		managerDone = make(chan error, 1)
		go func() { managerDone <- manager.Start(managerCtx) }()
	})

	AfterAll(func() {
		managerCancel()
		Eventually(managerDone, timeout).Should(Receive(Succeed()))
		waitForWebhookToBeClosed(webhookUrlValidate)
		Expect(envtest.UninstallCRDs(cfg, envtest.CRDInstallOptions{CRDs: crds})).To(Succeed())
	})

	It("keeps the serving template on a minimum decrease, retries a partial apply, and carries the requested minimum to a later image change", func() {
		const namespace, name = "knative-scaling-test", "inference"
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
		Expect(apiClient.Create(ctx, ns)).To(Succeed())
		DeferCleanup(func() { Expect(apiClient.Delete(ctx, ns)).To(Succeed()) })

		karta := test.NewValidJobSetKarta()
		karta.Name = "knative-scaling-test"
		karta.Labels = map[string]string{
			"run.ai/karta-group": "serving.knative.dev", "run.ai/karta-version": "v1", "run.ai/karta-kind": "Service",
		}
		karta.Spec.StructureDefinition.RootComponent.Kind = &kartav1alpha1.GroupVersionKind{
			Group: "serving.knative.dev", Version: "v1", Kind: "Service",
		}
		karta.Spec.StructureDefinition.ChildComponents = nil
		karta = createKartaWithStatus(ctx, apiClient, newKartaWithValidStatus(karta))
		DeferCleanup(func() { deleteAndPollUntilDeleted(ctx, apiClient, karta) })

		desired := knativeFixtureObject("Service", namespace, name)
		desired.SetLabels(map[string]string{
			constants.ConvertedFromNativeLabel:                 "true",
			constants.ControlPlaneWorkloadKindLabelForWorkload: constants.V2InferenceWorkloadKind,
		})
		desired.Object["spec"] = map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{"annotations": map[string]interface{}{knativeMinimum: "1"}},
				"spec": map[string]interface{}{"containers": []interface{}{
					map[string]interface{}{"name": "inference", "image": "example.invalid/inference:one"},
				}},
			},
		}
		runner := createWorkloadRunner(ctx, apiClient, name, namespace, desired, nil)
		DeferCleanup(func() { deleteAndPollUntilDeleted(ctx, apiClient, runner) })
		service := knativeFixtureObject("Service", namespace, name)
		Eventually(func() error { return apiClient.Get(ctx, client.ObjectKeyFromObject(service), service) }, timeout, interval).Should(Succeed())
		Expect(metav1.IsControlledBy(service, runner)).To(BeTrue())

		configuration := knativeFixtureObject("Configuration", namespace, name)
		Expect(ctrl.SetControllerReference(service, configuration, scheme.Scheme)).To(Succeed())
		Expect(apiClient.Create(ctx, configuration)).To(Succeed())
		revision := knativeFixtureObject("Revision", namespace, name+"-00001")
		revision.SetAnnotations(map[string]string{knativeMinimum: "1"})
		Expect(ctrl.SetControllerReference(configuration, revision, scheme.Scheme)).To(Succeed())
		Expect(apiClient.Create(ctx, revision)).To(Succeed())
		service.Object["status"] = map[string]interface{}{
			"observedGeneration": service.GetGeneration(), "latestCreatedRevisionName": revision.GetName(), "latestReadyRevisionName": revision.GetName(),
			"conditions": []interface{}{map[string]interface{}{"type": "Ready", "status": "True"}},
			"traffic":    []interface{}{map[string]interface{}{"revisionName": revision.GetName(), "percent": int64(100)}},
		}
		Expect(apiClient.Status().Update(ctx, service)).To(Succeed())
		originalSpec := service.DeepCopy().Object["spec"]
		originalGeneration := service.GetGeneration()

		By("lowering the requested minimum and retrying after the Revision patch succeeds but the Service apply conflicts")
		Expect(unstructured.SetNestedField(desired.Object, "0", "spec", "template", "metadata", "annotations", knativeMinimum)).To(Succeed())
		faultClient.failNextApply.Store(true)
		Eventually(func() error {
			if err := apiClient.Get(ctx, client.ObjectKeyFromObject(runner), runner); err != nil {
				return err
			}
			runner.Spec.Template = marshalRawExtension(desired)
			return apiClient.Update(ctx, runner)
		}, timeout, interval).Should(Succeed())
		Eventually(func(g Gomega) {
			g.Expect(apiClient.Get(ctx, client.ObjectKeyFromObject(revision), revision)).To(Succeed())
			g.Expect(revision.GetAnnotations()[knativeMinimum]).To(Equal("0"))
			g.Expect(faultClient.failedApplies.Load()).To(Equal(int32(1)))
			g.Expect(faultClient.successfulRetries.Load()).To(BeNumerically(">=", 1))
			g.Expect(apiClient.Get(ctx, client.ObjectKeyFromObject(service), service)).To(Succeed())
			g.Expect(service.Object["spec"]).To(Equal(originalSpec))
			g.Expect(service.GetGeneration()).To(Equal(originalGeneration))
		}, timeout, interval).Should(Succeed())
		Expect(apiClient.Get(ctx, client.ObjectKeyFromObject(runner), runner)).To(Succeed())
		storedDesired := &unstructured.Unstructured{}
		Expect(storedDesired.UnmarshalJSON(runner.Spec.Template.Raw)).To(Succeed())
		minimum, found, err := unstructured.NestedString(storedDesired.Object, "spec", "template", "metadata", "annotations", knativeMinimum)
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(minimum).To(Equal("0"))

		By("using that requested minimum when a subsequent image update needs a rollout")
		Expect(unstructured.SetNestedSlice(desired.Object, []interface{}{
			map[string]interface{}{"name": "inference", "image": "example.invalid/inference:two"},
		}, "spec", "template", "spec", "containers")).To(Succeed())
		Eventually(func() error {
			if err := apiClient.Get(ctx, client.ObjectKeyFromObject(runner), runner); err != nil {
				return err
			}
			runner.Spec.Template = marshalRawExtension(desired)
			return apiClient.Update(ctx, runner)
		}, timeout, interval).Should(Succeed())
		Eventually(func(g Gomega) {
			g.Expect(apiClient.Get(ctx, client.ObjectKeyFromObject(service), service)).To(Succeed())
			g.Expect(service.Object["spec"]).To(Equal(desired.Object["spec"]))
			g.Expect(service.GetGeneration()).To(BeNumerically(">", originalGeneration))
		}, timeout, interval).Should(Succeed())
	})
})

// A failed final apply leaves the Revision updated. The next normal reconcile
// must recover without copying the original minimum back to that Revision.
type knativeApplyConflictClient struct {
	client.Client
	failNextApply     atomic.Bool
	failedApplies     atomic.Int32
	successfulRetries atomic.Int32
}

func (c *knativeApplyConflictClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	options := (&client.PatchOptions{}).ApplyOptions(opts)
	isServiceApply := obj.GetObjectKind().GroupVersionKind() == (schema.GroupVersionKind{Group: "serving.knative.dev", Version: "v1", Kind: "Service"}) && len(options.DryRun) == 0
	if isServiceApply && obj.GetResourceVersion() != "" && c.failNextApply.CompareAndSwap(true, false) {
		c.failedApplies.Add(1)
		return apierrors.NewConflict(schema.GroupResource{Group: "serving.knative.dev", Resource: "services"}, obj.GetName(), errors.New("injected apply conflict"))
	}
	err := c.Client.Patch(ctx, obj, patch, opts...)
	if err == nil && isServiceApply && c.failedApplies.Load() > 0 {
		c.successfulRetries.Add(1)
	}
	return err
}

func knativeFixtureObject(kind, namespace, name string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion("serving.knative.dev/v1")
	obj.SetKind(kind)
	obj.SetNamespace(namespace)
	obj.SetName(name)
	return obj
}

func knativeFixtureCRD(kind, plural string) *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: plural + ".serving.knative.dev", Labels: map[string]string{"app.kubernetes.io/version": "1.21.0"}},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "serving.knative.dev", Scope: apiextensionsv1.NamespaceScoped,
			Names: apiextensionsv1.CustomResourceDefinitionNames{Plural: plural, Kind: kind, ListKind: kind + "List"},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{
				Name: "v1", Served: true, Storage: true,
				Subresources: &apiextensionsv1.CustomResourceSubresources{Status: &apiextensionsv1.CustomResourceSubresourceStatus{}},
				Schema: &apiextensionsv1.CustomResourceValidation{OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
					Type: "object", Properties: map[string]apiextensionsv1.JSONSchemaProps{
						"spec":   {Type: "object", XPreserveUnknownFields: ptr.To(true)},
						"status": {Type: "object", XPreserveUnknownFields: ptr.To(true)},
					},
				}},
			}},
		},
	}
}
