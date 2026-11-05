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

const inferenceBlock = `{
"epp":{"profile":"gie-v1.5","scheduling":{"scorers":[{"type":"queue","weight":30}]}},
"objectives":[{"model":"model","priority":5}],
"rewrites":[{"model":"alias","targets":[{"model":"model","weight":100}]}]}`

func inferenceAnnotation(block string) string {
	return `{"schemaVersion":"v1","endpoints":[{"services":[{"create":{"metadata":{"name":"model"},"spec":{"selector":{"app":"model"},"ports":[{"port":8000}]}}}],` +
		`"externalAccess":{"httpRoute":{"spec":{"hostnames":["model.example.test"]}},"inference":` + block + `}}]}`
}

func testTarget() InferenceEndpointTarget {
	return InferenceEndpointTarget{
		Selector:           map[string]string{"app": "model"},
		Port:               8000,
		GatewayName:        "gateway",
		GatewayNamespace:   "infra",
		Hostnames:          []string{"model.example.test"},
		DefaultBackendName: "model",
		DefaultBackendPort: 8000,
	}
}

func jsonEqual(t *testing.T, want string, got any) {
	t.Helper()
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var wantValue, gotValue any
	if err := json.Unmarshal([]byte(want), &wantValue); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &gotValue); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wantValue, gotValue) {
		t.Fatalf("got %s, want %s", encoded, want)
	}
}

func TestInferenceBlockRoundTripsThroughTheEndpointsAnnotation(t *testing.T) {
	obj := &metav1.ObjectMeta{Annotations: map[string]string{Annotation: inferenceAnnotation(inferenceBlock)}}
	cfg, err := Get(obj)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.InferenceEndpointIndex(); got != 0 {
		t.Fatalf("inference endpoint index = %d", got)
	}
	if !HasInference(obj) {
		t.Fatal("HasInference = false")
	}
	if err := Set(obj, cfg); err != nil {
		t.Fatal(err)
	}
	again, err := Get(obj)
	if err != nil {
		t.Fatal(err)
	}
	jsonEqual(t, inferenceBlock, again.Endpoints[0].ExternalAccess.Inference)
}

func TestEndpointsWithoutInferenceReportNone(t *testing.T) {
	for _, obj := range []*metav1.ObjectMeta{
		{},
		{Annotations: map[string]string{Annotation: `{"schemaVersion":"v1","endpoints":[{"services":[{"ref":{"name":"a","port":1}}]}]}`}},
		{Annotations: map[string]string{Annotation: "{"}},
	} {
		if HasInference(obj) {
			t.Fatalf("HasInference = true for %v", obj.Annotations)
		}
	}
	var cfg *Config
	if cfg.InferenceEndpointIndex() != -1 {
		t.Fatal("nil config reported an inference endpoint")
	}
}

func TestInferenceBlockRejectsUnknownKeys(t *testing.T) {
	obj := &metav1.ObjectMeta{Annotations: map[string]string{
		Annotation: inferenceAnnotation(`{"epp":{"profile":"gie-v1.5"},"workload":{"port":1}}`),
	}}
	if _, err := Get(obj); err == nil || !strings.Contains(err.Error(), "workload") {
		t.Fatalf("expected unknown key to be rejected, got %v", err)
	}
}

func TestInferenceBlockValidation(t *testing.T) {
	for name, block := range map[string]string{
		"missing epp":       `{}`,
		"null epp":          `{"epp":null}`,
		"empty epp":         `{"epp":{}}`,
		"array epp":         `{"epp":[]}`,
		"object objectives": `{"epp":{"profile":"gie-v1.5"},"objectives":{}}`,
		"empty rewrites":    `{"epp":{"profile":"gie-v1.5"},"rewrites":[]}`,
		"nested null":       `{"epp":{"profile":"gie-v1.5","image":null}}`,
	} {
		t.Run(name, func(t *testing.T) {
			var cfg InferenceConfig
			if err := json.Unmarshal([]byte(block), &cfg); err != nil {
				t.Fatal(err)
			}
			if err := cfg.Validate(); err == nil {
				t.Fatal("expected invalid inference block to be rejected")
			}
		})
	}
	var missing *InferenceConfig
	if missing.Validate() == nil {
		t.Fatal("nil inference block accepted")
	}
}

func TestBuildInferenceEndpointDerivesSpec(t *testing.T) {
	runner := &metav1.ObjectMeta{Name: strings.Repeat("a", 200), Namespace: "project", UID: "0146a461-a4d2-42c3-a030-f020c3039297"}
	var cfg InferenceConfig
	if err := json.Unmarshal([]byte(inferenceBlock), &cfg); err != nil {
		t.Fatal(err)
	}
	endpoint, err := BuildInferenceEndpoint(runner, &cfg, testTarget())
	if err != nil {
		t.Fatal(err)
	}
	if endpoint.GetName() != "inference-0146a461-a4d2-42c3-a030-f020c3039297" || endpoint.GetNamespace() != "project" {
		t.Fatalf("wrong identity: %v", endpoint.Object)
	}
	if endpoint.GetAPIVersion() != "inference.platform.run.ai/v1alpha1" || endpoint.GetKind() != "InferenceEndpoint" {
		t.Fatal("wrong GVK")
	}
	jsonEqual(t, `{
"workload":{"selector":{"matchLabels":{"app":"model"}},"port":8000},
"routing":{"gatewayRef":{"name":"gateway","namespace":"infra"},"hostnames":["model.example.test"],"defaultBackendRef":{"name":"model","port":8000}},
"epp":{"profile":"gie-v1.5","scheduling":{"scorers":[{"type":"queue","weight":30}]}},
"objectives":[{"model":"model","priority":5}],
"rewrites":[{"model":"alias","targets":[{"model":"model","weight":100}]}]}`, endpoint.Object["spec"])
	port, found, err := unstructured.NestedInt64(endpoint.Object, "spec", "workload", "port")
	if err != nil || !found || port != 8000 {
		t.Fatalf("numeric fields are not int64: %d %v", port, err)
	}
	endpoint.Object["spec"].(map[string]any)["epp"] = nil
	again, err := BuildInferenceEndpoint(runner, &cfg, testTarget())
	if err != nil || again.Object["spec"].(map[string]any)["epp"] == nil {
		t.Fatal("builder aliased configuration")
	}
}

func TestBuildInferenceEndpointOmitsUnsetOptionalFields(t *testing.T) {
	runner := &metav1.ObjectMeta{Namespace: "project", UID: "uid"}
	target := testTarget()
	target.GatewayNamespace = ""
	endpoint, err := BuildInferenceEndpoint(runner, &InferenceConfig{Epp: json.RawMessage(`{"profile":"gie-v1.5"}`)}, target)
	if err != nil {
		t.Fatal(err)
	}
	if _, found, _ := unstructured.NestedString(endpoint.Object, "spec", "routing", "gatewayRef", "namespace"); found {
		t.Fatal("empty Gateway namespace was written")
	}
	spec := endpoint.Object["spec"].(map[string]any)
	for _, key := range []string{"objectives", "rewrites"} {
		if _, found := spec[key]; found {
			t.Fatalf("absent %s were written", key)
		}
	}
}

func TestBuildInferenceEndpointRejectsIncompleteInput(t *testing.T) {
	cfg := &InferenceConfig{Epp: json.RawMessage(`{"profile":"gie-v1.5"}`)}
	for _, runner := range []*metav1.ObjectMeta{{Namespace: "project"}, {UID: "uid"}, {Namespace: "project", UID: "INVALID/UID"}} {
		if _, err := BuildInferenceEndpoint(runner, cfg, testTarget()); err == nil {
			t.Fatal("accepted invalid identity")
		}
	}
	runner := &metav1.ObjectMeta{Namespace: "project", UID: "uid"}
	if _, err := BuildInferenceEndpoint(runner, nil, testTarget()); err == nil {
		t.Fatal("accepted missing inference block")
	}
	for name, mutate := range map[string]func(*InferenceEndpointTarget){
		"selector":        func(t *InferenceEndpointTarget) { t.Selector = nil },
		"port":            func(t *InferenceEndpointTarget) { t.Port = 0 },
		"gateway":         func(t *InferenceEndpointTarget) { t.GatewayName = "" },
		"hostnames":       func(t *InferenceEndpointTarget) { t.Hostnames = nil },
		"default backend": func(t *InferenceEndpointTarget) { t.DefaultBackendName = "" },
		"backend port":    func(t *InferenceEndpointTarget) { t.DefaultBackendPort = 70000 },
	} {
		target := testTarget()
		mutate(&target)
		if _, err := BuildInferenceEndpoint(runner, cfg, target); err == nil {
			t.Fatalf("accepted target without a valid %s", name)
		}
	}
}
