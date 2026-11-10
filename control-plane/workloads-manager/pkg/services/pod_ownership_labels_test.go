/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	workloadsAPI "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
)

func ownershipTestType(t *testing.T, root *kartav1alpha1.SpecDefinition, children ...kartav1alpha1.ComponentDefinition) *workloadsAPI.WorkloadTypeConfig {
	t.Helper()
	karta := kartav1alpha1.Karta{Spec: kartav1alpha1.KartaSpec{StructureDefinition: kartav1alpha1.StructureDefinition{
		RootComponent: kartav1alpha1.ComponentDefinition{Name: "root", SpecDefinition: root}, ChildComponents: children,
	}}}
	data, err := json.Marshal(karta)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]interface{}
	if err := json.Unmarshal(data, &object); err != nil {
		t.Fatal(err)
	}
	version := "v1"
	return &workloadsAPI.WorkloadTypeConfig{Kartas: []workloadsAPI.WorkloadTypeKartasWithVersionInner{
		{Version: &version, Karta: object},
	}}
}

func ownershipTestManifest(t *testing.T, body string) map[string]interface{} {
	t.Helper()
	var manifest map[string]interface{}
	if err := json.Unmarshal([]byte(body), &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest
}

func ownershipTestStamp(t *testing.T, manifest map[string]interface{}, wt *workloadsAPI.WorkloadTypeConfig) (map[string]interface{}, bool, error) {
	t.Helper()
	info, err := NewK8sResourceUtil().GetResourceInfo(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return stampPodOwnershipLabels(context.Background(), manifest, "cp-uuid", wt, info)
}

func ownershipTestLabels(t *testing.T, labels interface{}) {
	t.Helper()
	m, ok := labels.(map[string]interface{})
	if !ok {
		t.Fatalf("missing label map: %#v", labels)
	}
	if m["run.ai/top-owner-uid"] != "cp-uuid" || m["run.ai/workload-id"] != "cp-uuid" {
		t.Fatalf("wrong ownership labels: %#v", m)
	}
	if _, ok := m["run.ai/cluster-top-owner-uid"]; ok {
		t.Fatal("stale cluster UID must be left for EWI to calculate")
	}
}

func TestPodOwnershipStampDeclaredTemplatesPreserveUnknownData(t *testing.T) {
	path := ".spec.template"
	manifest := ownershipTestManifest(t, `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"test"},"spec":{"template":{"metadata":{"labels":{"app":"keep","run.ai/top-owner-uid":"stale","run.ai/workload-id":"stale","run.ai/cluster-top-owner-uid":"old-job"}},"spec":{"containers":[],"operator-extension":{"keep":true}}}},"status":{"shadow":{"spec":{"containers":[]}}}}`)
	before, _ := json.Marshal(manifest)
	result, stamped, err := ownershipTestStamp(t, manifest, ownershipTestType(t, &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: &path}))
	if err != nil || !stamped {
		t.Fatalf("stamp failed: %v / %v", stamped, err)
	}
	template := result["spec"].(map[string]interface{})["template"].(map[string]interface{})
	labels := template["metadata"].(map[string]interface{})["labels"]
	ownershipTestLabels(t, labels)
	if labels.(map[string]interface{})["app"] != "keep" {
		t.Fatal("removed user label")
	}
	if !reflect.DeepEqual(template["spec"], manifest["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"]) {
		t.Fatal("changed unknown Pod fields")
	}
	if !reflect.DeepEqual(result["status"], manifest["status"]) {
		t.Fatal("stamped unrelated status data")
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) {
		t.Fatal("mutated original input")
	}
}

func TestPodOwnershipStampArrayAndOptionalComponents(t *testing.T) {
	path, missing := ".spec.workers[].template", ".spec.master.template"
	manifest := ownershipTestManifest(t, `{"apiVersion":"example/v1","kind":"Training","metadata":{},"spec":{"workers":[{"template":{"spec":{"containers":[]}}},{"template":{"metadata":null,"spec":{"containers":[]}}}],"master":null}}`)
	wt := ownershipTestType(t, nil,
		kartav1alpha1.ComponentDefinition{Name: "workers", SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: &path}},
		kartav1alpha1.ComponentDefinition{Name: "master", SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: &missing}})
	result, stamped, err := ownershipTestStamp(t, manifest, wt)
	if err != nil || !stamped {
		t.Fatalf("stamp failed: %v / %v", stamped, err)
	}
	spec := result["spec"].(map[string]interface{})
	for _, worker := range spec["workers"].([]interface{}) {
		template := worker.(map[string]interface{})["template"].(map[string]interface{})
		ownershipTestLabels(t, template["metadata"].(map[string]interface{})["labels"])
	}
	if spec["master"] != nil {
		t.Fatal("manufactured absent optional Master")
	}
}

func TestPodOwnershipStampFragmentedAndSeparateMetadata(t *testing.T) {
	for _, fragmented := range []bool{false, true} {
		t.Run(map[bool]string{false: "separate metadata", true: "fragmented labels"}[fragmented], func(t *testing.T) {
			manifest := ownershipTestManifest(t, `{"apiVersion":"example/v1","kind":"Training","metadata":{},"spec":{"podSpec":{"containers":[]},"container":{"image":"keep"},"podLabels":{"user":"keep"}}}`)
			pod, meta, labels, container := ".spec.podSpec", ".spec.podMetadata", ".spec.podLabels", ".spec.container"
			def := &kartav1alpha1.SpecDefinition{PodSpecPath: &pod, MetadataPath: &meta}
			if fragmented {
				def = &kartav1alpha1.SpecDefinition{FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{LabelsPath: &labels, ContainerPath: &container}}
			}
			result, stamped, err := ownershipTestStamp(t, manifest, ownershipTestType(t, def))
			if err != nil || !stamped {
				t.Fatalf("stamp failed: %v / %v", stamped, err)
			}
			spec := result["spec"].(map[string]interface{})
			if fragmented {
				ownershipTestLabels(t, spec["podLabels"])
				if spec["podLabels"].(map[string]interface{})["user"] != "keep" {
					t.Fatal("lost existing labels")
				}
			} else {
				ownershipTestLabels(t, spec["podMetadata"].(map[string]interface{})["labels"])
			}
		})
	}
}

func TestPodOwnershipStampDoesNotInventUnsupportedTargets(t *testing.T) {
	path, containers := ".spec.absent", ".spec.containers"
	for name, wt := range map[string]*workloadsAPI.WorkloadTypeConfig{
		"no RI":                   {},
		"absent template":         ownershipTestType(t, &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: &path}),
		"fragment without labels": ownershipTestType(t, &kartav1alpha1.SpecDefinition{FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{ContainersPath: &containers}}),
		"different child GVK":     ownershipTestType(t, nil, kartav1alpha1.ComponentDefinition{Name: "different", Kind: &kartav1alpha1.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"}, SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: &path}}),
	} {
		t.Run(name, func(t *testing.T) {
			manifest := ownershipTestManifest(t, `{"apiVersion":"example/v1","kind":"Training","metadata":{},"spec":{"containers":[]}}`)
			result, stamped, err := ownershipTestStamp(t, manifest, wt)
			if name == "absent template" || name == "different child GVK" {
				if err != nil || !stamped || !reflect.DeepEqual(result["spec"], manifest["spec"]) {
					t.Fatalf("optional template changed: %v / %v / %#v", stamped, err, result)
				}
				return
			}
			if err != nil || stamped || !reflect.DeepEqual(result, manifest) {
				t.Fatalf("unexpected mutation: %v / %v / %#v", stamped, err, result)
			}
		})
	}
}

func TestPodOwnershipStampInvalidPathDoesNotMutateInput(t *testing.T) {
	path := ".spec.["
	manifest := ownershipTestManifest(t, `{"apiVersion":"example/v1","kind":"Training","metadata":{}}`)
	before, _ := json.Marshal(manifest)
	_, stamped, err := ownershipTestStamp(t, manifest, ownershipTestType(t, &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: &path}))
	if err == nil || stamped {
		t.Fatal("invalid RI path accepted")
	}
	after, _ := json.Marshal(manifest)
	if string(before) != string(after) {
		t.Fatal("partial mutation on failure")
	}
}

func TestPodOwnershipStampMissingVersion(t *testing.T) {
	path := ".spec.template"
	wt := ownershipTestType(t, &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: &path})
	*wt.Kartas[0].Version = "v2"
	manifest := ownershipTestManifest(t, `{"apiVersion":"example/v1","kind":"Training","metadata":{}}`)
	result, stamped, err := ownershipTestStamp(t, manifest, wt)
	if err != nil || stamped || !reflect.DeepEqual(result, manifest) {
		t.Fatalf("missing-version RI changed submission: %v / %v", stamped, err)
	}
}

func TestPodOwnershipStampAbsentArrays(t *testing.T) {
	path := `.spec.workers[].template`
	for _, spec := range []string{`{}`, `{"workers":null}`, `{"workers":[]}`} {
		manifest := ownershipTestManifest(t, `{"apiVersion":"example/v1","kind":"Training","metadata":{},"spec":`+spec+`}`)
		result, stamped, err := ownershipTestStamp(t, manifest, ownershipTestType(t, &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: &path}))
		if err != nil || !stamped || !reflect.DeepEqual(result["spec"], manifest["spec"]) {
			t.Fatalf("optional array changed: %s / %v / %v", spec, stamped, err)
		}
	}
}

func TestPodOwnershipStampNestedArraysAndQuotedKeys(t *testing.T) {
	path := `.spec["workers[]"][].replicas[].template`
	manifest := ownershipTestManifest(t, `{"apiVersion":"example/v1","kind":"Training","metadata":{},"spec":{"workers[]":[{"replicas":null},{"replicas":[{"template":{"spec":{"containers":[]}}}]}]}}`)
	result, stamped, err := ownershipTestStamp(t, manifest, ownershipTestType(t, &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: &path}))
	if err != nil || !stamped {
		t.Fatalf("nested path failed: %v / %v", stamped, err)
	}
	workers := result["spec"].(map[string]interface{})["workers[]"].([]interface{})
	if workers[0].(map[string]interface{})["replicas"] != nil {
		t.Fatal("manufactured optional replicas")
	}
	template := workers[1].(map[string]interface{})["replicas"].([]interface{})[0].(map[string]interface{})["template"].(map[string]interface{})
	ownershipTestLabels(t, template["metadata"].(map[string]interface{})["labels"])
}

func TestPodOwnershipStampInvalidArrayType(t *testing.T) {
	path := `.spec.workers[].template`
	manifest := ownershipTestManifest(t, `{"apiVersion":"example/v1","kind":"Training","metadata":{},"spec":{"workers":"bad"}}`)
	_, stamped, err := ownershipTestStamp(t, manifest, ownershipTestType(t, &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: &path}))
	if err == nil || stamped {
		t.Fatal("invalid array type silently skipped")
	}
}

func TestPodOwnershipStampMalformedResourceInterface(t *testing.T) {
	path := ".spec.template"
	wt := ownershipTestType(t, &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: &path})
	wt.Kartas[0].Karta["spec"] = "invalid"
	manifest := ownershipTestManifest(t, `{"apiVersion":"example/v1","kind":"Training","metadata":{}}`)
	_, stamped, err := ownershipTestStamp(t, manifest, wt)
	if err == nil || stamped {
		t.Fatal("corrupt RI silently skipped")
	}
}
