/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/testutil"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const dynamoNameField = "name"

// Paths match the Dynamo Kartas in control-plane/workloads/pkg/database/go_migrations/ri/source.
func dynamoKarta(beta bool) *kartav1alpha1.Karta {
	version := "v1alpha1"
	prefix := ".spec.services | .[]"
	podPrefix := prefix + " | .extraPodSpec"
	containerPath := podPrefix + ".mainContainer"
	instancePath := ".spec.services | to_entries[] | .key"
	replicasPath := ".spec.services[] | (.replicas // 1) * (.multinode.nodeCount // 1)"
	resourcesPath := ptr.To(prefix + " | .resources")
	labelsPath := prefix + " | .labels"
	if beta {
		version = "v1beta1"
		podPrefix = ".spec.components[] | .podTemplate.spec"
		containerPath = podPrefix + ".containers[] | select(.name == \"main\")"
		instancePath = ".spec.components[] | .name"
		replicasPath = ".spec.components[] | (.replicas // 1) * (.multinode.nodeCount // 1)"
		resourcesPath = nil
		labelsPath = ".spec.components[] | .podTemplate.metadata.labels"
	}
	return &kartav1alpha1.Karta{Spec: kartav1alpha1.KartaSpec{StructureDefinition: kartav1alpha1.StructureDefinition{
		RootComponent: kartav1alpha1.ComponentDefinition{Name: "dynamographdeployment", Kind: &kartav1alpha1.GroupVersionKind{Group: "nvidia.com", Version: version, Kind: "DynamoGraphDeployment"}},
		ChildComponents: []kartav1alpha1.ComponentDefinition{{Name: "service", OwnerRef: ptr.To("dynamographdeployment"), InstanceIdPath: ptr.To(instancePath),
			ScaleDefinition: &kartav1alpha1.ScaleDefinition{ReplicasPath: ptr.To(replicasPath)},
			SpecDefinition: &kartav1alpha1.SpecDefinition{FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
				LabelsPath: ptr.To(labelsPath), ResourcesPath: resourcesPath, ContainerPath: ptr.To(containerPath), ResourceClaimsPath: ptr.To(podPrefix + ".resourceClaims"),
				NodeAffinityPath: ptr.To(podPrefix + ".affinity.nodeAffinity"), PodAffinityPath: ptr.To(podPrefix + ".affinity.podAffinity"),
			}},
		}},
	}}}
}

func dynamoResource(t *testing.T, beta bool, worker string) *unstructured.Unstructured {
	t.Helper()
	field, version, frontend := "services", "v1alpha1", `{"replicas":1,"resources":{"limits":{"gpu":"0"}},"extraPodSpec":{"mainContainer":{"image":"frontend","resources":{"limits":{"cpu":"1"}}}}}`
	if beta {
		field, version, frontend = "components", "v1beta1", `{"replicas":1,"podTemplate":{"spec":{"containers":[{"name":"main","image":"frontend","resources":{"limits":{"cpu":"1"}}}]}}}`
	}
	obj := &unstructured.Unstructured{}
	data := fmt.Sprintf(`{
		"apiVersion": "nvidia.com/%s",
		"kind": "DynamoGraphDeployment",
		"metadata": {"name": "probe", "namespace": "runai-test"},
		"spec": {"%s": {"frontend": %s, "w": %s}}
	}`, version, field, frontend, worker)
	NewWithT(t).Expect(json.Unmarshal([]byte(data), &obj.Object)).To(Succeed())
	if beta {
		components := obj.Object["spec"].(map[string]interface{})["components"].(map[string]interface{})
		frontend := components["frontend"].(map[string]interface{})
		frontend[dynamoNameField] = "frontend"
		worker := components["w"].(map[string]interface{})
		worker[dynamoNameField] = "w"
		obj.Object["spec"].(map[string]interface{})["components"] = []interface{}{frontend, worker}
	}
	return obj
}

func dynamoComponent(t *testing.T, obj *unstructured.Unstructured, beta bool, name string) map[string]interface{} {
	t.Helper()
	if beta {
		components, _, err := unstructured.NestedSlice(obj.Object, "spec", "components")
		NewWithT(t).Expect(err).NotTo(HaveOccurred())
		for _, component := range components {
			if component.(map[string]interface{})[dynamoNameField] == name {
				return component.(map[string]interface{})
			}
		}
		t.Fatalf("component %s not found", name)
	}
	component, _, err := unstructured.NestedMap(obj.Object, "spec", "services", name)
	NewWithT(t).Expect(err).NotTo(HaveOccurred())
	return component
}

func TestDynamoComputeDomain(t *testing.T) {
	for _, elastic := range []bool{false, true} {
		name := "static"
		if elastic {
			name = "elastic"
		}
		t.Run(name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			obj := dynamoResource(t, false, `{"replicas":1,"multinode":{"nodeCount":2},"resources":{"limits":{"gpu":"1"}},"extraPodSpec":{"mainContainer":{"image":"worker"}}}`)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			cl := testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(scheme.Scheme).Build())
			state := reconcile.New()
			g.Expect(newTestPlugin(elastic).HandleAdditionalResources(ctx, runner, obj, dynamoKarta(false), cl, state)).To(Succeed())
			g.Expect(state.Issues()).To(BeEmpty())
			domains := &nvapi.ComputeDomainList{}
			g.Expect(cl.List(ctx, domains)).To(Succeed())
			g.Expect(domains.Items).To(HaveLen(1))
			want := 2
			if elastic {
				want = 0
			}
			g.Expect(domains.Items[0].Spec.NumNodes).To(Equal(want))
		})
	}
}

func TestDynamoClaimRoundTrip(t *testing.T) {
	cases := []struct {
		name, worker          string
		beta, containerClaims bool
	}{
		{"alpha service GPU", `{
  "replicas": 1,
  "multinode": {
    "nodeCount": 2
  },
  "resources": {
    "limits": {
      "gpu": "1"
    },
    "claims": [
      {
        "name": "existing-service"
      }
    ]
  },
  "extraPodSpec": {
    "mainContainer": {
      "image": "worker"
    }
  }
}`, false, false},
		{"alpha main claims override", `{
  "replicas": 1,
  "multinode": {
    "nodeCount": 2
  },
  "resources": {
    "limits": {
      "gpu": "1"
    },
    "claims": [
      {
        "name": "existing-service"
      }
    ]
  },
  "extraPodSpec": {
    "mainContainer": {
      "image": "worker",
      "resources": {
        "claims": [
          {
            "name": "existing-main"
          }
        ]
      }
    }
  }
}`, false, true},
		{"alpha without main container", `{"replicas":1,"multinode":{"nodeCount":2},"resources":{"limits":{"gpu":"1"},"claims":[{"name":"existing-service"}]}}`, false, false},
		{"alpha service and main GPU inherit service claims", `{
  "replicas": 1,
  "multinode": {
    "nodeCount": 2
  },
  "resources": {
    "limits": {
      "gpu": "1"
    },
    "claims": [
      {
        "name": "existing-service"
      }
    ]
  },
  "extraPodSpec": {
    "mainContainer": {
      "image": "worker",
      "resources": {
        "limits": {
          "nvidia.com/gpu": "1"
        }
      }
    }
  }
}`, false, true},
		{"alpha main GPU", `{
  "replicas": 1,
  "multinode": {
    "nodeCount": 2
  },
  "extraPodSpec": {
    "mainContainer": {
      "image": "worker",
      "resources": {
        "limits": {
          "nvidia.com/gpu": "1"
        },
        "claims": [
          {
            "name": "existing-main"
          }
        ]
      }
    }
  }
}`, false, true},
		{"beta main GPU", `{
  "replicas": 1,
  "multinode": {
    "nodeCount": 2
  },
  "podTemplate": {
    "spec": {
      "containers": [
        {
          "name": "main",
          "image": "worker",
          "resources": {
            "limits": {
              "nvidia.com/gpu": "1"
            },
            "claims": [
              {
                "name": "existing-main"
              }
            ]
          }
        }
      ]
    }
  }
}`, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			obj := dynamoResource(t, tc.beta, tc.worker)
			podPath := []string{"extraPodSpec"}
			if tc.beta {
				podPath = []string{"podTemplate", "spec"}
			}
			originalFrontend := dynamoComponent(t, obj, tc.beta, "frontend")
			originalWorker := dynamoComponent(t, obj, tc.beta, "w")
			originalObject := obj.DeepCopy().Object
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			for range 2 {
				state := reconcile.New()
				g.Expect(plugin.MutateResource(ctx, runner, obj, dynamoKarta(tc.beta), state)).To(Succeed())
				g.Expect(state.Issues()).To(BeEmpty())
				assertNoIntroducedNulls(t, originalObject, obj.Object, "")
				count, err := plugin.calculateNumNodes(ctx, obj, dynamoKarta(tc.beta))
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(count).To(Equal(2))
				worker := dynamoComponent(t, obj, tc.beta, "w")
				labelPath := []string{"labels"}
				if tc.beta {
					labelPath = []string{"podTemplate", "metadata", "labels"}
				}
				workerLabels, _, err := unstructured.NestedStringMap(worker, labelPath...)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(workerLabels).To(HaveKeyWithValue(WorkloadNameLabel, runner.Name))
				pod, _, err := unstructured.NestedMap(worker, podPath...)
				g.Expect(err).NotTo(HaveOccurred())
				claims, _, err := unstructured.NestedSlice(pod, "resourceClaims")
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(claims).To(Equal([]interface{}{
					map[string]interface{}{
						dynamoNameField:             getResourceClaimName(),
						"resourceClaimTemplateName": composeResourceClaimTemplateName(testWorkloadName, workloadRunnerInitials),
					},
				}))
				g.Expect(pod["affinity"]).NotTo(BeNil())
				if !tc.beta {
					if originalResources, ok := originalWorker["resources"].(map[string]interface{}); ok {
						resources, _, err := unstructured.NestedMap(worker, "resources")
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(resources["limits"]).To(Equal(originalResources["limits"]))
						g.Expect(resources["claims"]).To(Equal([]interface{}{map[string]interface{}{dynamoNameField: "existing-service"}, map[string]interface{}{dynamoNameField: getResourceClaimName()}}))
					} else {
						g.Expect(worker["resources"]).To(BeNil())
					}
					_, hadMain, _ := unstructured.NestedMap(originalWorker, "extraPodSpec", "mainContainer")
					_, hasMain, _ := unstructured.NestedMap(worker, "extraPodSpec", "mainContainer")
					g.Expect(hasMain).To(Equal(hadMain))
				}
				if tc.containerClaims {
					var container map[string]interface{}
					if tc.beta {
						container = pod["containers"].([]interface{})[0].(map[string]interface{})
					} else {
						container = pod["mainContainer"].(map[string]interface{})
					}
					var parsed corev1.Container
					g.Expect(runtime.DefaultUnstructuredConverter.FromUnstructured(container, &parsed)).To(Succeed())
					inheritedClaim := "existing-main"
					if tc.name == "alpha service and main GPU inherit service claims" {
						inheritedClaim = "existing-service"
					}
					g.Expect(parsed.Resources.Claims).To(Equal([]corev1.ResourceClaim{{Name: inheritedClaim}, {Name: getResourceClaimName()}}))
					g.Expect(parsed.Image).To(Equal("worker"))
					g.Expect(parsed.Resources.Limits).NotTo(HaveKey(corev1.ResourceName("gpu")))
				}
				frontend := dynamoComponent(t, obj, tc.beta, "frontend")
				frontendLabels, _, err := unstructured.NestedStringMap(frontend, labelPath...)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(frontendLabels).NotTo(HaveKey(WorkloadNameLabel))
				frontendPod, _, err := unstructured.NestedMap(frontend, podPath...)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(frontendPod["resourceClaims"]).To(BeNil())
				affinity, _, err := unstructured.NestedMap(frontendPod, "affinity")
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(affinity["nodeAffinity"]).To(BeNil())
				g.Expect(affinity["podAffinity"]).To(BeNil())
				originalPod, _, err := unstructured.NestedMap(originalFrontend, podPath...)
				g.Expect(err).NotTo(HaveOccurred())
				containerKey := "mainContainer"
				if tc.beta {
					containerKey = "containers"
				}
				// Compare typed containers to ignore Karta's existing JSON defaulting.
				var before, after corev1.Container
				beforeContainer, afterContainer := originalPod[containerKey], frontendPod[containerKey]
				if tc.beta {
					beforeContainer = beforeContainer.([]interface{})[0]
					afterContainer = afterContainer.([]interface{})[0]
				}
				g.Expect(runtime.DefaultUnstructuredConverter.FromUnstructured(beforeContainer.(map[string]interface{}), &before)).To(Succeed())
				g.Expect(runtime.DefaultUnstructuredConverter.FromUnstructured(afterContainer.(map[string]interface{}), &after)).To(Succeed())
				g.Expect(after).To(Equal(before))
			}
		})
	}
}

// A nil lookup cannot distinguish absence from an explicit null rejected by a CRD.
func assertNoIntroducedNulls(t *testing.T, original, updated interface{}, path string) {
	t.Helper()
	switch value := updated.(type) {
	case map[string]interface{}:
		before, _ := original.(map[string]interface{})
		for key, child := range value {
			old, existed := before[key]
			if child == nil && !existed {
				t.Errorf("introduced null at %s/%s", path, key)
			}
			assertNoIntroducedNulls(t, old, child, path+"/"+key)
		}
	case []interface{}:
		before, _ := original.([]interface{})
		for index, child := range value {
			var old interface{}
			if index < len(before) {
				old = before[index]
			}
			assertNoIntroducedNulls(t, old, child, fmt.Sprintf("%s/%d", path, index))
		}
	}
}

func TestDynamoCPUAfterWorker(t *testing.T) {
	for _, beta := range []bool{false, true} {
		t.Run(fmt.Sprintf("beta=%t", beta), func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			worker := `{"replicas":1,"multinode":{"nodeCount":2},"resources":{"limits":{"gpu":"1"}}}`
			if beta {
				worker = `{"replicas":1,"multinode":{"nodeCount":2},"podTemplate":{"spec":{"containers":[{"name":"main","resources":{"limits":{"nvidia.com/gpu":"1"}}}]}}}`
			}
			obj := dynamoResource(t, beta, worker)
			spec := obj.Object["spec"].(map[string]interface{})
			if beta {
				components := spec["components"].([]interface{})
				components[0], components[1] = components[1], components[0]
			} else {
				services := spec["services"].(map[string]interface{})
				services["zfrontend"] = services["frontend"]
				delete(services, "frontend")
			}
			original := obj.DeepCopy().Object
			plugin := newTestPlugin(false)
			runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
			for range 2 {
				state := reconcile.New()
				g.Expect(plugin.MutateResource(ctx, runner, obj, dynamoKarta(beta), state)).To(Succeed())
				g.Expect(state.Issues()).To(BeEmpty())
				assertNoIntroducedNulls(t, original, obj.Object, "")
				count, err := plugin.calculateNumNodes(ctx, obj, dynamoKarta(beta))
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(count).To(Equal(2))
			}
		})
	}
}

func TestRemoveIntroducedNullFields(t *testing.T) {
	for _, tc := range []struct{ name, original, updated, want string }{
		{"new empty parents", `{}`, `{"affinity":{"nodeAffinity":null,"podAffinity":null},"resourceClaims":null}`, `{}`},
		{"existing empty parent", `{"affinity":{}}`, `{"affinity":{"nodeAffinity":null,"podAffinity":null}}`, `{"affinity":{}}`},
		{"explicit null parent", `{"affinity":null}`, `{"affinity":{"nodeAffinity":null,"podAffinity":null}}`, `{"affinity":null}`},
		{"explicit null field", `{"resourceClaims":null}`, `{"resourceClaims":null,"resources":null}`, `{"resourceClaims":null}`},
		{
			"preserve new nonnull values",
			`{}`,
			`{"affinity":{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[]}},"podAffinity":null},"resources":{}}`,
			`{"affinity":{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[]}}},"resources":{}}`,
		},
		{
			"preserve array ordering",
			`{"components":[{"name":"worker"},{"name":"frontend"}]}`,
			`{"components":[{"name":"worker","resourceClaims":[{"name":"compute-domain"}]},{"name":"frontend","resourceClaims":null,"affinity":{"nodeAffinity":null}}]}`,
			`{"components":[{"name":"worker","resourceClaims":[{"name":"compute-domain"}]},{"name":"frontend"}]}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			var original, updated, want interface{}
			g.Expect(json.Unmarshal([]byte(tc.original), &original)).To(Succeed())
			g.Expect(json.Unmarshal([]byte(tc.updated), &updated)).To(Succeed())
			g.Expect(json.Unmarshal([]byte(tc.want), &want)).To(Succeed())
			removeIntroducedNullFields(original, updated)
			g.Expect(updated).To(Equal(want))
			removeIntroducedNullFields(original, updated)
			g.Expect(updated).To(Equal(want))
		})
	}
}

func TestDynamoWithoutGPUsSkipsComputeDomain(t *testing.T) {
	for _, beta := range []bool{false, true} {
		for _, elastic := range []bool{false, true} {
			t.Run(fmt.Sprintf("beta=%t/elastic=%t", beta, elastic), func(t *testing.T) {
				g := NewWithT(t)
				ctx := context.Background()
				worker := `{"replicas":1,"resources":{"limits":{"gpu":"0"}},"extraPodSpec":{"mainContainer":{"image":"worker"}}}`
				if beta {
					worker = `{"replicas":1,"podTemplate":{"spec":{"containers":[{"name":"main","image":"worker","resources":{"limits":{"cpu":"1"}}}]}}}`
				}
				obj := dynamoResource(t, beta, worker)
				runner := newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
				cl := testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(scheme.Scheme).Build())
				plugin := newTestPlugin(elastic)
				state := reconcile.New()
				g.Expect(plugin.HandleAdditionalResources(ctx, runner, obj, dynamoKarta(beta), cl, state)).To(Succeed())
				g.Expect(state.Issues()).To(BeEmpty())
				domains := &nvapi.ComputeDomainList{}
				g.Expect(cl.List(ctx, domains)).To(Succeed())
				g.Expect(domains.Items).To(BeEmpty())

				g.Expect(plugin.MutateResource(ctx, runner, obj, dynamoKarta(beta), state)).To(Succeed())
				g.Expect(state.Issues()).To(BeEmpty())
				for _, name := range []string{"frontend", "w"} {
					component := dynamoComponent(t, obj, beta, name)
					path := []string{"extraPodSpec", "resourceClaims"}
					if beta {
						path = []string{"podTemplate", "spec", "resourceClaims"}
					}
					claims, _, err := unstructured.NestedSlice(component, path...)
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(claims).To(BeEmpty())
				}
			})
		}
	}
}

func TestDynamoWithoutGPUsAndMNNVLDisabled(t *testing.T) {
	for _, mode := range []dromosv1alpha1.MnnvlMode{"", dromosv1alpha1.MnnvlModeNone} {
		t.Run(string(mode), func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			obj := dynamoResource(t, false, `{"replicas":1,"extraPodSpec":{"mainContainer":{"image":"worker"}}}`)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			if mode != "" {
				runner = newTestRunner(testWorkloadName, testNamespace, newMnnvlConfig(mode, testGPULabel))
			}
			cl := testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(scheme.Scheme).Build())
			state := reconcile.New()
			plugin := newTestPlugin(false)
			g.Expect(plugin.HandleAdditionalResources(ctx, runner, obj, dynamoKarta(false), cl, state)).To(Succeed())
			g.Expect(state.Issues()).To(BeEmpty())
		})
	}
}
