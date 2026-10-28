/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package controller

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/plugins/inferenceendpoint"
	ep "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromos "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	jobset "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

var _ = Describe("Inference endpoint plugin integration", Serial, Ordered, func() {
	var (
		c          client.Client
		managerCtx context.Context
		stop       context.CancelFunc
		stopped    chan struct{}
	)
	const ns = "inference-endpoint-plugin-test"
	const endpointSpec = `{"workload":{"selector":{"matchLabels":{"app":"model"}},"port":8000},"routing":{"gatewayRef":{"name":"gateway"},"hostnames":["model.example.test"]},"epp":{"profile":"gie-v1.5"}}`
	BeforeAll(func() {
		managerCtx, stop = context.WithCancel(ctx)
		certDir, err := filepath.Abs(webhookCertDir)
		Expect(err).NotTo(HaveOccurred())
		manager, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme.Scheme, Metrics: metricsserver.Options{BindAddress: "0"}, WebhookServer: webhook.NewServer(webhook.Options{CertDir: certDir, Port: webhookPort})})
		Expect(err).NotTo(HaveOccurred())
		c, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).NotTo(HaveOccurred())
		plugin, err := inferenceendpoint.New(managerCtx, framework.PluginDeps{Client: manager.GetClient()})
		Expect(err).NotTo(HaveOccurred())
		fw := framework.NewFramework(framework.NewRegistry(plugin), nil)
		//nolint:staticcheck // Dromos requires the legacy event recorder interface.
		controller := controller.NewWorkloadRunnerController(manager.GetClient(), fw, manager.GetEventRecorderFor("inference-endpoint-test")).WithName("inference-endpoint-test")
		Expect(controller.SetupWithManager(managerCtx, manager)).To(Succeed())
		setupWebhooks(fw, c, managerCtx, manager)
		stopped = make(chan struct{})
		go func() { defer GinkgoRecover(); defer close(stopped); Expect(manager.Start(managerCtx)).To(Succeed()) }()
		Expect(c.Create(managerCtx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})).To(Succeed())
		waitForSetupToComplete(c, managerCtx, webhookPort)
		createKartaWithStatus(managerCtx, c, newKartaWithValidStatus(test.NewValidJobSetKarta()))
	})
	AfterAll(func() {
		deleteAllTestResources(managerCtx, c)
		stop()
		Eventually(stopped, timeout).Should(BeClosed())
	})
	It("creates sibling resources watches endpoint status repairs drift and removes routing intent", func() {
		workload := test.NewJobSetBuilder("model", ns).WithDefault().Build()
		runner := &dromos.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: "model", Namespace: ns}, Spec: dromos.WorkloadRunnerSpec{Template: marshalRawExtension(workload)}}
		Expect(ep.SetInferenceEndpoint(runner, &ep.InferenceEndpointConfig{Spec: json.RawMessage(endpointSpec)})).To(Succeed())
		Expect(c.Create(managerCtx, runner)).To(Succeed())
		key := client.ObjectKeyFromObject(runner)
		endpoint := &unstructured.Unstructured{}
		endpoint.SetGroupVersionKind(ep.InferenceEndpointGVK)
		name, err := ep.InferenceEndpointName(runner)
		Expect(err).NotTo(HaveOccurred())
		endpointKey := client.ObjectKey{Namespace: ns, Name: name}
		Eventually(func() error { return c.Get(managerCtx, endpointKey, endpoint) }, timeout, interval).Should(Succeed())
		Expect(metav1.IsControlledBy(endpoint, runner)).To(BeTrue())
		Eventually(func() error { return c.Get(managerCtx, client.ObjectKeyFromObject(workload), &jobset.JobSet{}) }, timeout, interval).Should(Succeed())
		Eventually(func() dromos.PluginPhase {
			Expect(c.Get(managerCtx, key, runner)).To(Succeed())
			if status := runner.Status.Plugins[inferenceendpoint.PluginName]; status != nil {
				return status.Phase
			}
			return ""
		}, timeout, interval).Should(Equal(dromos.PluginPhasePending))

		endpoint.Object["status"] = map[string]any{
			"observedGeneration": endpoint.GetGeneration(),
			"conditions": []any{map[string]any{
				"type": "Ready", "status": "True", "observedGeneration": endpoint.GetGeneration(),
				"reason": "RouteAccepted", "message": "routing ready", "lastTransitionTime": metav1.Now().Format(time.RFC3339),
			}},
		}

		Expect(c.Status().Update(managerCtx, endpoint)).To(Succeed())
		Eventually(func() dromos.PluginPhase {
			Expect(c.Get(managerCtx, key, runner)).To(Succeed())
			return runner.Status.Plugins[inferenceendpoint.PluginName].Phase
		}, timeout, interval).Should(Equal(dromos.PluginPhaseReady))
		previousUID := endpoint.GetUID()
		Expect(c.Delete(managerCtx, endpoint)).To(Succeed())
		Eventually(func() bool {
			err := c.Get(managerCtx, endpointKey, endpoint)
			return err == nil && endpoint.GetUID() != previousUID
		}, timeout, interval).Should(BeTrue())
		Eventually(func() error {
			if err := c.Get(managerCtx, key, runner); err != nil {
				return err
			}
			delete(runner.Annotations, ep.InferenceEndpointAnnotation)
			return c.Update(managerCtx, runner)
		}, timeout, interval).Should(Succeed())
		Eventually(func() bool { return apierrors.IsNotFound(c.Get(managerCtx, endpointKey, endpoint)) }, timeout, interval).Should(BeTrue())
		// Routing removal must not remove the model workload.
		Expect(c.Get(managerCtx, client.ObjectKeyFromObject(workload), &jobset.JobSet{})).To(Succeed())
	})
	It("reports unknown CRD fields instead of repeatedly reconciling pruned intent", func() {
		workload := test.NewJobSetBuilder("invalidmodel", ns).WithDefault().Build()
		runner := &dromos.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: "invalidmodel", Namespace: ns}, Spec: dromos.WorkloadRunnerSpec{Template: marshalRawExtension(workload)}}
		raw := strings.Replace(endpointSpec, `"profile":"gie-v1.5"`, `"profile":"gie-v1.5","unknownOption":true`, 1)
		Expect(ep.SetInferenceEndpoint(runner, &ep.InferenceEndpointConfig{Spec: json.RawMessage(raw)})).To(Succeed())
		Expect(c.Create(managerCtx, runner)).To(Succeed())
		Eventually(func() dromos.PluginPhase {
			Expect(c.Get(managerCtx, client.ObjectKeyFromObject(runner), runner)).To(Succeed())
			if status := runner.Status.Plugins[inferenceendpoint.PluginName]; status != nil {
				return status.Phase
			}
			return ""
		}, timeout, interval).Should(Equal(dromos.PluginPhaseFailed))
	})

	It("reconciles annotation-only changes after the workload has settled", func() {
		workload := test.NewJobSetBuilder("annotationmodel", ns).WithDefault().Build()
		runner := &dromos.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: "annotationmodel", Namespace: ns}, Spec: dromos.WorkloadRunnerSpec{Template: marshalRawExtension(workload)}}
		Expect(c.Create(managerCtx, runner)).To(Succeed())
		key := client.ObjectKeyFromObject(runner)
		Eventually(func() dromos.PluginPhase {
			Expect(c.Get(managerCtx, key, runner)).To(Succeed())
			if status := runner.Status.Plugins[inferenceendpoint.PluginName]; status != nil {
				return status.Phase
			}
			return ""
		}, timeout, interval).Should(Equal(dromos.PluginPhaseDisabled))
		Consistently(func() string {
			latest := &dromos.WorkloadRunner{}
			Expect(c.Get(managerCtx, key, latest)).To(Succeed())
			return latest.ResourceVersion
		}, 2*time.Second, interval).Should(Equal(runner.ResourceVersion))
		generation := runner.Generation
		Expect(ep.SetInferenceEndpoint(runner, &ep.InferenceEndpointConfig{Spec: json.RawMessage(endpointSpec)})).To(Succeed())
		Expect(c.Update(managerCtx, runner)).To(Succeed())
		Expect(runner.Generation).To(Equal(generation))
		endpoint := &unstructured.Unstructured{}
		endpoint.SetGroupVersionKind(ep.InferenceEndpointGVK)
		name, err := ep.InferenceEndpointName(runner)
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() error { return c.Get(managerCtx, client.ObjectKey{Namespace: ns, Name: name}, endpoint) }, timeout, interval).Should(Succeed())
	})

})
