// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package services_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/services"
)

func TestNetworkingAuthorizationTypeRequired(t *testing.T) {
	for _, tc := range []struct {
		name, payload string
		wantErr       bool
	}{
		{name: "no authorization", payload: `{}`},
		{name: "missing type", payload: `{"authorization":{"authorizedUsers":["alice"]}}`, wantErr: true},
		{name: "null type", payload: `{"authorization":{"authorizationType":null}}`, wantErr: true},
		{name: "explicit type", payload: `{"authorization":{"authorizationType":"authorizedUsers","authorizedUsers":["alice"]}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var service api.NetworkingService
			if err := json.Unmarshal([]byte(tc.payload), &service); err != nil {
				t.Fatal(err)
			}
			err := api.AssertNetworkingServiceRequired(service)
			if (err != nil) != tc.wantErr {
				t.Fatalf("got error %v, want error %t", err, tc.wantErr)
			}
		})
	}
}

func TestNetworkingServiceNameConstraints(t *testing.T) {
	for _, tc := range []struct {
		name    string
		length  int
		wantErr bool
	}{
		{name: "empty", length: 0, wantErr: true},
		{name: "minimum", length: 1},
		{name: "maximum", length: 63},
		{name: "above maximum", length: 64, wantErr: true},
		{name: "previous maximum", length: 253, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			name := strings.Repeat("a", tc.length)
			selector := map[string]string{"app": "workload"}
			for target, service := range map[string]api.NetworkingService{
				"spec": {Spec: &api.ServiceSpec{
					Name: name, Selector: &selector, Ports: []api.ServicePort{{Port: 8080}},
				}},
				"ref": {Ref: &api.ServiceRef{Name: name, Port: 8080}},
			} {
				t.Run(target, func(t *testing.T) {
					err := api.AssertNetworkingServiceConstraints(service)
					if (err != nil) != tc.wantErr {
						t.Fatalf("name length %d: got error %v, want error %t", tc.length, err, tc.wantErr)
					}
				})
			}
		})
	}
}

func TestNetworkingServicePortNameConstraints(t *testing.T) {
	if err := api.AssertServicePortConstraints(api.ServicePort{Port: 8080}); err != nil {
		t.Fatalf("omitted port name: %v", err)
	}
	for _, length := range []int{0, 1, 15, 16, 63, 64} {
		name := strings.Repeat("a", length)
		err := api.AssertServicePortConstraints(api.ServicePort{Name: &name, Port: 8080})
		wantErr := length == 0 || length > 63
		if (err != nil) != wantErr {
			t.Errorf("port name length %d: got error %v, want error %t", length, err, wantErr)
		}
	}
}

func TestNetworkingMetricsServiceNameAtWorkloadLimit(t *testing.T) {
	workloads := &services.WorkloadsV2Impl{}
	name := strings.Repeat("a", services.MaxWorkloadNameLength)
	if err := workloads.ValidateWorkloadNameFormat(name); err != nil {
		t.Fatalf("maximum-length workload name: %v", err)
	}
	if err := workloads.ValidateWorkloadNameFormat(name + "a"); err == nil {
		t.Fatal("expected workload name above the limit to be rejected")
	}
	if err := api.AssertServiceSpecConstraints(api.ServiceSpec{
		Name: name + "-metrics", Ports: []api.ServicePort{{Port: 8000}},
	}); err != nil {
		t.Fatalf("metrics Service name for maximum-length workload: %v", err)
	}
}

func TestNetworkingArrayItemSchemas(t *testing.T) {
	// Check the published schema contract; generated Go decoding does not itself reject every JSON null.
	type schema struct {
		Ref        string `json:"$ref"`
		Type       string
		Nullable   bool
		Items      *schema
		Properties map[string]schema
	}
	var spec struct {
		Components struct{ Schemas map[string]schema }
	}
	data, err := os.ReadFile("../../swagger/_build/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		parent, property, item string
		nullable               bool
	}{
		{parent: "NetworkingConfiguration", item: "NetworkingEndpoint", nullable: true},
		{parent: "NetworkingEndpoint", property: "services", item: "NetworkingService", nullable: true},
		{parent: "ServiceSpec", property: "ports", item: "ServicePort"},
	} {
		t.Run(tc.item, func(t *testing.T) {
			array := spec.Components.Schemas[tc.parent]
			if tc.property != "" {
				array = array.Properties[tc.property]
			}
			if array.Type != "array" || array.Nullable != tc.nullable || array.Items == nil || array.Items.Ref != "#/components/schemas/"+tc.item {
				t.Fatalf("unexpected array schema: %+v", array)
			}
			item := spec.Components.Schemas[tc.item]
			if item.Type != "object" || item.Nullable {
				t.Fatalf("array items must be non-null objects: %+v", item)
			}
		})
	}
}
