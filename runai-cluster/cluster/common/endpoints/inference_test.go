/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package endpoints

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const inferenceSpec = `{
"workload":{"selector":{"matchLabels":{"app":"model"}},"port":8000},
"routing":{"gatewayRef":{"name":"gateway","namespace":"infra"},"hostnames":["model.example.test"]},
"epp":{"profile":"gie-v1.5","scheduling":{"scorers":[{"type":"queue","weight":30}]}},
"rewrites":[{"model":"alias","targets":[{"model":"model","weight":100}]}]}`

func TestInferenceEndpointAbsent(t *testing.T) {
	cfg, err := GetInferenceEndpoint(&metav1.ObjectMeta{})
	if cfg != nil || err != nil {
		t.Fatalf("absent intent: %v, %v", cfg, err)
	}
}

func TestInferenceEndpointRejectsInvalidIntent(t *testing.T) {
	for _, raw := range []string{
		"", "null", "{",
		`{"schemaVersion":"v2","spec":` + inferenceSpec + `}`,
		`{"schemaVersion":"v1"}`,
		`{"schemaVersion":"v1","spec":null}`,
		`{"schemaVersion":"v1","spec":[]}`,
		`{"schemaVersion":"v1","spec":{}}`,
		`{"schemaVersion":"v1","spec":{"workload":{},"routing":{},"epp":null}}`, `{"schemaVersion":"v1","metadata":{"namespace":"other"},"spec":` + inferenceSpec + `}`} {
		t.Run(raw, func(t *testing.T) {
			_, err := GetInferenceEndpoint(&metav1.ObjectMeta{Annotations: map[string]string{InferenceEndpointAnnotation: raw}})
			if err == nil {
				t.Fatal("expected invalid intent to be rejected")
			}
		})
	}
}

func TestInferenceEndpointRoundTrip(t *testing.T) {
	runner := &metav1.ObjectMeta{Name: strings.Repeat("a", 200), Namespace: "project", UID: "0146a461-a4d2-42c3-a030-f020c3039297", Annotations: map[string]string{"keep": "value"}}
	cfg := &InferenceEndpointConfig{Spec: json.RawMessage(inferenceSpec)}
	original := string(cfg.Spec)
	if err := SetInferenceEndpoint(runner, cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.SchemaVersion != "" || string(cfg.Spec) != original {
		t.Fatal("setter mutated source config")
	}
	if runner.Annotations["keep"] != "value" {
		t.Fatal("lost unrelated annotation")
	}
	decoded, err := GetInferenceEndpoint(runner)
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := BuildInferenceEndpoint(runner, decoded)
	if err != nil {
		t.Fatal(err)
	}
	if endpoint.GetName() != "inference-0146a461-a4d2-42c3-a030-f020c3039297" || endpoint.GetNamespace() != "project" {
		t.Fatalf("wrong identity: %v", endpoint.Object)
	}
	if endpoint.GetAPIVersion() != "inference.platform.run.ai/v1alpha1" || endpoint.GetKind() != "InferenceEndpoint" {
		t.Fatal("wrong GVK")
	}
	port, found, err := unstructured.NestedInt64(endpoint.Object, "spec", "workload", "port")
	if err != nil || !found || port != 8000 {
		t.Fatalf("numeric fields did not survive: %d %v", port, err)
	}
	var want, got any
	if err := json.Unmarshal([]byte(inferenceSpec), &want); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(endpoint.Object["spec"])
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatal("spec fields did not round-trip")
	}
	endpoint.Object["spec"].(map[string]any)["epp"] = nil
	again, err := BuildInferenceEndpoint(runner, decoded)
	if err != nil || again.Object["spec"].(map[string]any)["epp"] == nil {
		t.Fatal("builder aliased configuration")
	}
}

func TestInferenceEndpointRequiresSavedNamespacedRunner(t *testing.T) {
	cfg := &InferenceEndpointConfig{SchemaVersion: "v1", Spec: json.RawMessage(inferenceSpec)}
	for _, runner := range []*metav1.ObjectMeta{{Namespace: "project"}, {UID: "uid"}, {Namespace: "project", UID: "INVALID/UID"}} {
		if _, err := BuildInferenceEndpoint(runner, cfg); err == nil {
			t.Fatal("accepted invalid identity")
		}
	}
}

func TestInferenceEndpointSetterRejectsInvalidWithoutMutation(t *testing.T) {
	obj := &metav1.ObjectMeta{Annotations: map[string]string{"keep": "value"}}
	for _, cfg := range []*InferenceEndpointConfig{nil, {Spec: json.RawMessage(`{}`)}, {SchemaVersion: "v2", Spec: json.RawMessage(inferenceSpec)}} {
		if err := SetInferenceEndpoint(obj, cfg); err == nil {
			t.Fatal("accepted invalid config")
		}
		if len(obj.Annotations) != 1 {
			t.Fatal("mutated annotations on error")
		}
	}
}

func TestInferenceEndpointRejectsNullValuesThatCRDWouldPrune(t *testing.T) {
	raw := strings.Replace(inferenceSpec, `"profile":"gie-v1.5"`, `"image":null,"profile":"gie-v1.5"`, 1)
	obj := &metav1.ObjectMeta{}
	if err := SetInferenceEndpoint(obj, &InferenceEndpointConfig{Spec: json.RawMessage(raw)}); err == nil {
		t.Fatal("accepted a null field that would be pruned and prevent readiness")
	}
}
