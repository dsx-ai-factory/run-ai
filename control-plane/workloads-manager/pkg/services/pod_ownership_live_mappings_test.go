/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"os"
	"testing"

	workloadsAPI "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	karta "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestPodOwnershipStampLiveMappings(t *testing.T) {
	for _, tc := range []struct {
		name, fixture, manifest string
		labels                  [][]string
	}{
		{"pytorch", "krt-v1-kubeflow-org-pytorchjob", `{"apiVersion":"kubeflow.org/v1","kind":"PyTorchJob","metadata":{},"spec":{"pytorchReplicaSpecs":{"Master":{"template":{"spec":{"containers":[]}}},"Worker":{"template":{"spec":{"containers":[]}}}}}}`, [][]string{{"spec", "pytorchReplicaSpecs", "Master", "template", "metadata", "labels"}, {"spec", "pytorchReplicaSpecs", "Worker", "template", "metadata", "labels"}}},
		{"mpi", "krt-v2beta1-kubeflow-org-mpijob", `{"apiVersion":"kubeflow.org/v2beta1","kind":"MPIJob","metadata":{},"spec":{"mpiReplicaSpecs":{"Launcher":{"template":{"spec":{"containers":[]}}},"Worker":{"template":{"spec":{"containers":[]}}}}}}`, [][]string{{"spec", "mpiReplicaSpecs", "Launcher", "template", "metadata", "labels"}, {"spec", "mpiReplicaSpecs", "Worker", "template", "metadata", "labels"}}},
		{"lws", "krt-v1-leaderworkerset-x-k8s-io-leaderworkerset", `{"apiVersion":"leaderworkerset.x-k8s.io/v1","kind":"LeaderWorkerSet","metadata":{},"spec":{"leaderWorkerTemplate":{"leaderTemplate":{"spec":{"containers":[]}},"workerTemplate":{"spec":{"containers":[]}}}}}`, [][]string{{"spec", "leaderWorkerTemplate", "leaderTemplate", "metadata", "labels"}, {"spec", "leaderWorkerTemplate", "workerTemplate", "metadata", "labels"}}},
		{"kserve-model", "krt-v1beta1-serving-kserve-io-inferenceservice", `{"apiVersion":"serving.kserve.io/v1beta1","kind":"InferenceService","metadata":{},"spec":{"predictor":{"model":{"modelFormat":{"name":"sklearn"},"storageUri":"s3://example/model"}}}}`, [][]string{{"spec", "predictor", "labels"}}},
		{"kserve-container", "krt-v1beta1-serving-kserve-io-inferenceservice", `{"apiVersion":"serving.kserve.io/v1beta1","kind":"InferenceService","metadata":{},"spec":{"predictor":{"containers":[{"name":"predictor","image":"example"}]}}}`, [][]string{{"spec", "predictor", "labels"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wt := ownershipLiveType(t, tc.fixture)
			result, stamped, err := ownershipTestStamp(t, ownershipTestManifest(t, tc.manifest), wt)
			if err != nil || !stamped {
				t.Fatalf("stamp: %v, %v", stamped, err)
			}
			for _, path := range tc.labels {
				labels, found, err := unstructured.NestedStringMap(result, path...)
				if err != nil || !found || labels["run.ai/top-owner-uid"] != "cp-uuid" || labels["run.ai/workload-id"] != "cp-uuid" {
					t.Fatalf("missing canonical labels at %v: %v (%v)", path, labels, err)
				}
			}
			if tc.name == "kserve-model" || tc.name == "kserve-container" {
				if _, found, _ := unstructured.NestedFieldNoCopy(result, "spec", "transformer"); found {
					t.Fatal("invented optional transformer")
				}
			}
		})
	}
	t.Run("argo-multiple-templates", func(t *testing.T) {
		manifest := ownershipTestManifest(t, `{"apiVersion":"argoproj.io/v1alpha1","kind":"Workflow","metadata":{},"spec":{"templates":[{"name":"one","container":{"image":"example"}},{"name":"two","container":{"image":"example"}}]}}`)
		result, stamped, err := ownershipTestStamp(t, manifest, ownershipLiveType(t, "krt-v1alpha1-argoproj-io-workflow"))
		if err != nil || !stamped {
			t.Fatalf("stamp: %v, %v", stamped, err)
		}
		templates, _, _ := unstructured.NestedSlice(result, "spec", "templates")
		if len(templates) != 2 {
			t.Fatalf("changed templates: %v", templates)
		}
		for _, tpl := range templates {
			m := tpl.(map[string]interface{})
			labels, _, _ := unstructured.NestedMap(m, "metadata", "labels")
			ownershipTestLabels(t, labels)
		}
	})
	t.Run("ray-multiple-worker-groups", func(t *testing.T) {
		manifest := ownershipTestManifest(t, `{"apiVersion":"ray.io/v1","kind":"RayCluster","metadata":{},"spec":{"headGroupSpec":{"template":{"spec":{"containers":[]}}},"workerGroupSpecs":[{"groupName":"a","template":{"spec":{"containers":[]}}},{"groupName":"b","template":{"spec":{"containers":[]}}}]}}`)
		result, stamped, err := ownershipTestStamp(t, manifest, ownershipLiveType(t, "krt-v1-ray-io-raycluster"))
		if err != nil || !stamped {
			t.Fatalf("stamp: %v, %v", stamped, err)
		}
		labels, _, _ := unstructured.NestedMap(result, "spec", "headGroupSpec", "template", "metadata", "labels")
		ownershipTestLabels(t, labels)
		workers, _, _ := unstructured.NestedSlice(result, "spec", "workerGroupSpecs")
		if len(workers) != 2 {
			t.Fatal("changed worker groups")
		}
		for _, w := range workers {
			labels, _, _ := unstructured.NestedMap(w.(map[string]interface{}), "template", "metadata", "labels")
			ownershipTestLabels(t, labels)
		}
	})
}

func ownershipLiveType(t *testing.T, name string) *workloadsAPI.WorkloadTypeConfig {
	t.Helper()
	data, err := os.ReadFile("testdata/ownership/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var k karta.Karta
	if err := json.Unmarshal(data, &k); err != nil {
		t.Fatal(err)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatal(err)
	}
	version := k.Spec.StructureDefinition.RootComponent.Kind.Version
	return &workloadsAPI.WorkloadTypeConfig{Kartas: []workloadsAPI.WorkloadTypeKartasWithVersionInner{{Version: &version, Karta: obj}}}
}
