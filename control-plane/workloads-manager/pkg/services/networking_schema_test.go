// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package services_test

import (
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
)

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
