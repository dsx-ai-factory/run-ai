// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package services_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
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

func TestNetworkingArrayItemSchemas(t *testing.T) {
	// Check the published schema contract; generated Go decoding does not itself reject every JSON null.
	type schema struct {
		Type       string
		Nullable   bool
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
	for _, name := range []string{"NetworkingEndpoint", "NetworkingService", "ServicePort"} {
		t.Run(name, func(t *testing.T) {
			item := spec.Components.Schemas[name]
			if item.Type != "object" || item.Nullable {
				t.Fatalf("array items must be non-null objects: %+v", item)
			}
		})
	}
	if !spec.Components.Schemas["NetworkingConfiguration"].Nullable || !spec.Components.Schemas["NetworkingEndpoint"].Properties["services"].Nullable {
		t.Fatal("optional networking arrays must remain nullable")
	}
}
