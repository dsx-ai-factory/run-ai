/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"context"
	"testing"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestServiceMonitorReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ServiceMonitor Reconciler Suite")
}

var _ = Describe("ServiceMonitor Reconciler", func() {
	var (
		reconciler *ServiceMonitorReconciler
		k8sClient  client.Client
		project    kaiorg.Project
		namespace  string
		ctx        context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		namespace = "runai-test-project"
		project = kaiorg.Project{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-project",
				UID:  types.UID("test-project-uid"),
			},
			Status: kaiorg.ProjectStatus{
				Namespace: namespace,
			},
		}

		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}

		scheme := scheme_init.Scheme()
		Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
		k8sClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&kaiorg.Project{}).
			WithObjects(project.DeepCopy(), ns).
			Build()
		reconciler = NewServiceMonitorReconciler(k8sClient)
	})

	Context("when all ServiceMonitors need to be created", func() {
		BeforeEach(func() {
			Expect(reconciler.Reconcile(ctx, &project)).To(Succeed())
		})

		It("creates the Knative Serving ServiceMonitor with correct configuration", func() {
			sm := newSMUnstructured()
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: KnativeServingServiceMonitorName, Namespace: namespace}, sm)).To(Succeed())

			spec := sm.Object["spec"].(map[string]interface{})
			endpoints := spec["endpoints"].([]interface{})
			Expect(endpoints).To(HaveLen(1))
			Expect(endpoints[0].(map[string]interface{})["port"]).To(Equal(PortHTTPUsermetric))

			selector := spec["selector"].(map[string]interface{})
			matchLabels := selector["matchLabels"].(map[string]interface{})
			Expect(matchLabels).To(HaveLen(1))
			Expect(matchLabels[LabelKnativeServiceType]).To(Equal(ValuePrivate))
		})

		It("creates the NIM inference ServiceMonitor with correct configuration", func() {
			sm := newSMUnstructured()
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: NIMServiceMonitorInferenceName, Namespace: namespace}, sm)).To(Succeed())

			spec := sm.Object["spec"].(map[string]interface{})
			endpoints := spec["endpoints"].([]interface{})
			Expect(endpoints).To(HaveLen(4))

			ep0 := endpoints[0].(map[string]interface{})
			ep1 := endpoints[1].(map[string]interface{})
			ep2 := endpoints[2].(map[string]interface{})
			ep3 := endpoints[3].(map[string]interface{})
			Expect(ep0["port"]).To(Equal(PortMetrics))
			Expect(ep0["path"]).To(Equal(PathV1Metrics))
			Expect(ep1["port"]).To(Equal(PortMetrics))
			Expect(ep1["path"]).To(Equal(PathMetrics))
			Expect(ep2["port"]).To(Equal(PortServingPort))
			Expect(ep2["path"]).To(Equal(PathV1Metrics))
			Expect(ep3["port"]).To(Equal(PortServingPort))
			Expect(ep3["path"]).To(Equal(PathMetrics))

			selector := spec["selector"].(map[string]interface{})
			matchLabels := selector["matchLabels"].(map[string]interface{})
			Expect(matchLabels).To(HaveLen(3))
			Expect(matchLabels[LabelNIMWorkload]).To(Equal(constants.BoolTrueAsString))
			Expect(matchLabels[LabelWorkloadKind]).To(Equal(ValueInferenceWorkload))
			Expect(matchLabels[constants.LabelNIMWorkloadMetricsService]).To(Equal(constants.BoolTrueAsString))
		})

		It("creates the NIM external workloads ServiceMonitor with correct configuration", func() {
			sm := newSMUnstructured()
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: NIMServiceMonitor, Namespace: namespace}, sm)).To(Succeed())

			spec := sm.Object["spec"].(map[string]interface{})
			endpoints := spec["endpoints"].([]interface{})
			Expect(endpoints).To(HaveLen(4))

			ep0 := endpoints[0].(map[string]interface{})
			ep1 := endpoints[1].(map[string]interface{})
			ep2 := endpoints[2].(map[string]interface{})
			ep3 := endpoints[3].(map[string]interface{})
			Expect(ep0["port"]).To(Equal(PortMetrics))
			Expect(ep0["path"]).To(Equal(PathV1Metrics))
			Expect(ep1["port"]).To(Equal(PortMetrics))
			Expect(ep1["path"]).To(Equal(PathMetrics))
			Expect(ep2["port"]).To(Equal(PortServingPort))
			Expect(ep2["path"]).To(Equal(PathV1Metrics))
			Expect(ep3["port"]).To(Equal(PortServingPort))
			Expect(ep3["path"]).To(Equal(PathMetrics))

			selector := spec["selector"].(map[string]interface{})
			matchLabels := selector["matchLabels"].(map[string]interface{})
			Expect(matchLabels).To(HaveLen(1))
			Expect(matchLabels[LabelNIMWorkload]).To(Equal(constants.BoolTrueAsString))

			matchExpressions := selector["matchExpressions"].([]interface{})
			Expect(matchExpressions).To(HaveLen(1))
			expr := matchExpressions[0].(map[string]interface{})
			Expect(expr["key"]).To(Equal(LabelWorkloadKind))
			Expect(expr["operator"]).To(Equal("NotIn"))
			Expect(expr["values"]).To(ConsistOf(ValueInferenceWorkload))
		})

		It("creates the vLLM ServiceMonitor with the exact selector and endpoint", func() {
			sm := newSMUnstructured()
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: VLLMServiceMonitorName, Namespace: namespace}, sm)).To(Succeed())
			Expect(sm.GetName()).To(Equal("runai-vllm"))

			spec := sm.Object["spec"].(map[string]interface{})
			Expect(spec["endpoints"]).To(Equal([]interface{}{
				map[string]interface{}{"port": PortMetrics, "path": PathMetrics},
			}))
			Expect(spec["selector"]).To(Equal(map[string]interface{}{
				"matchLabels": map[string]interface{}{
					LabelWorkloadKind:                         ValueInferenceWorkload,
					constants.LabelVLLMWorkload:               constants.BoolTrueAsString,
					constants.LabelVLLMWorkloadMetricsService: constants.BoolTrueAsString,
				},
			}))
		})

	})

	Context("when ServiceMonitors already exist", func() {
		var existingUIDs map[string]string
		var existingSpecs map[string]interface{}

		BeforeEach(func() {
			existingUIDs = make(map[string]string)
			existingSpecs = make(map[string]interface{})

			for _, cfg := range []ServiceMonitorConfig{
				knativeServingServiceMonitorConfig,
				nimServiceMonitorConfigInference,
				nimServiceMonitorConfig,
				vllmServiceMonitorConfig,
			} {
				sm := reconciler.buildServiceMonitor(namespace, project.Name, project.UID, cfg)
				sm.SetUID(types.UID(cfg.Name + "-uid"))
				Expect(k8sClient.Create(ctx, sm)).To(Succeed())
				existingUIDs[cfg.Name] = string(sm.GetUID())
				persisted := newSMUnstructured()
				Expect(k8sClient.Get(ctx, client.ObjectKey{Name: cfg.Name, Namespace: namespace}, persisted)).To(Succeed())
				existingSpecs[cfg.Name] = persisted.Object["spec"]
			}

			Expect(reconciler.Reconcile(ctx, &project)).To(Succeed())
		})

		It("does not recreate existing ServiceMonitors", func() {
			for name, uid := range existingUIDs {
				sm := newSMUnstructured()
				Expect(k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, sm)).To(Succeed())
				Expect(string(sm.GetUID())).To(Equal(uid))
				actualSpec := sm.Object["spec"].(map[string]interface{})
				expectedSpec := existingSpecs[name].(map[string]interface{})
				Expect(actualSpec["endpoints"]).To(Equal(expectedSpec["endpoints"]))
				Expect(actualSpec["selector"]).To(Equal(expectedSpec["selector"]))
			}
		})

	})

	Context("when the vLLM ServiceMonitor spec is stale", func() {
		It("repairs it in place without changing existing NIM or Knative monitors", func() {
			unchangedSpecs := make(map[string]interface{})
			var uid types.UID
			for _, cfg := range []ServiceMonitorConfig{
				knativeServingServiceMonitorConfig,
				nimServiceMonitorConfigInference,
				nimServiceMonitorConfig,
				vllmServiceMonitorConfig,
			} {
				sm := reconciler.buildServiceMonitor(namespace, project.Name, project.UID, cfg)
				sm.SetUID(types.UID(cfg.Name + "-uid"))
				if cfg.Name == VLLMServiceMonitorName {
					uid = sm.GetUID()
					sm.Object["spec"] = map[string]interface{}{
						"endpoints": []interface{}{map[string]interface{}{"port": "stale", "path": "/stale"}},
						"selector":  map[string]interface{}{"matchLabels": map[string]interface{}{"stale": "true"}},
					}
				}
				Expect(k8sClient.Create(ctx, sm)).To(Succeed())
				if cfg.Name != VLLMServiceMonitorName {
					persisted := newSMUnstructured()
					Expect(k8sClient.Get(ctx, client.ObjectKey{Name: cfg.Name, Namespace: namespace}, persisted)).To(Succeed())
					unchangedSpecs[cfg.Name] = persisted.Object["spec"]
				}
			}

			Expect(reconciler.Reconcile(ctx, &project)).To(Succeed())
			vllm := newSMUnstructured()
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: VLLMServiceMonitorName, Namespace: namespace}, vllm)).To(Succeed())
			Expect(vllm.GetUID()).To(Equal(uid))
			actualSpec := vllm.Object["spec"].(map[string]interface{})
			expectedSpec := reconciler.buildServiceMonitor(namespace, project.Name, project.UID, vllmServiceMonitorConfig).Object["spec"].(map[string]interface{})
			Expect(actualSpec["endpoints"]).To(Equal(expectedSpec["endpoints"]))
			Expect(actualSpec["selector"]).To(Equal(expectedSpec["selector"]))

			for name, spec := range unchangedSpecs {
				sm := newSMUnstructured()
				Expect(k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, sm)).To(Succeed())
				actualSpec := sm.Object["spec"].(map[string]interface{})
				expectedSpec := spec.(map[string]interface{})
				Expect(actualSpec["endpoints"]).To(Equal(expectedSpec["endpoints"]))
				Expect(actualSpec["selector"]).To(Equal(expectedSpec["selector"]))
			}
		})
	})

	Context("when the project namespace is not yet set", func() {
		It("skips reconciliation without error", func() {
			projectWithoutNS := project.DeepCopy()
			projectWithoutNS.Status.Namespace = ""

			Expect(reconciler.Reconcile(ctx, projectWithoutNS)).To(Succeed())

			for _, name := range []string{KnativeServingServiceMonitorName, NIMServiceMonitorInferenceName, NIMServiceMonitor, VLLMServiceMonitorName} {
				sm := newSMUnstructured()
				err := k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, sm)
				Expect(err).To(HaveOccurred())
			}
		})
	})

})

func newSMUnstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": ServiceMonitorAPIVersion,
			"kind":       ServiceMonitorKind,
		},
	}
}

func buildTestServiceMonitor(name, ns, projectName string, projectUID types.UID, cfg ServiceMonitorConfig) *unstructured.Unstructured {
	endpoints := make([]interface{}, len(cfg.Endpoints))
	for i, ep := range cfg.Endpoints {
		epMap := map[string]interface{}{"port": ep.Port}
		if ep.Path != "" {
			epMap["path"] = ep.Path
		}
		if len(ep.MetricRelabelings) > 0 {
			relabelings := make([]interface{}, len(ep.MetricRelabelings))
			for j, r := range ep.MetricRelabelings {
				rm := make(map[string]interface{})
				if r.Regex != "" {
					rm["regex"] = r.Regex
				}
				if r.Action != "" {
					rm["action"] = r.Action
				}
				relabelings[j] = rm
			}
			epMap["metricRelabelings"] = relabelings
		}
		endpoints[i] = epMap
	}

	matchLabels := make(map[string]interface{})
	for k, v := range cfg.LabelSelector {
		matchLabels[k] = v
	}

	app := cfg.App
	if app == "" {
		app = cfg.Name
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": ServiceMonitorAPIVersion,
			"kind":       ServiceMonitorKind,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": ns,
				"labels":    map[string]interface{}{appLabel: app},
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion": kaiorg.GroupVersion.Identifier(),
						"kind":       "Project",
						"name":       projectName,
						"uid":        string(projectUID),
					},
				},
			},
			"spec": map[string]interface{}{
				"endpoints": endpoints,
				"selector":  map[string]interface{}{"matchLabels": matchLabels},
			},
		},
	}
}
