// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
)

func TestValidateNetworkingEndpoints(t *testing.T) {
	for _, tc := range []struct {
		name, input, wantErr string
	}{
		{name: "null networking", input: `null`},
		{name: "empty networking", input: `[]`},
		{name: "null endpoint", input: `[null]`, wantErr: "networking[0] must contain a service or externalAccess"},
		{name: "empty endpoint", input: `[{}]`, wantErr: "networking[0] must contain a service or externalAccess"},
		{name: "empty services", input: `[{"services":[]}]`, wantErr: "networking[0] must contain a service or externalAccess"},
		{name: "empty external access", input: `[{"externalAccess":{}}]`, wantErr: "networking[0] must contain a service or externalAccess"},
		{name: "null service", input: `[{"services":[null]}]`, wantErr: "networking[0].services[0] must contain spec or ref"},
		{name: "empty service", input: `[{"services":[{}]}]`, wantErr: "networking[0].services[0] must contain spec or ref"},
		{name: "service with both spec and ref", input: `[{"services":[{"spec":{"name":"one","selector":{"app":"one"},"ports":[{"port":80}]},"ref":{"name":"one","port":80}}]}]`, wantErr: "networking[0].services[0] must not contain both spec and ref"},
		{name: "multiple services", input: `[{"services":[{"ref":{"name":"one","port":80}},{"ref":{"name":"two","port":80}}]}]`, wantErr: "networking[0].services must contain at most one service"},
		{name: "later invalid endpoint", input: `[{"externalAccess":{"ingress":{}}},null]`, wantErr: "networking[1] must contain a service or externalAccess"},
		{name: "created service", input: `[{"services":[{"spec":{"name":"one","ports":[{"port":80}]}}]}]`},
		{name: "referenced service", input: `[{"services":[{"ref":{"name":"one","port":80}}]}]`},
		{name: "ingress only", input: `[{"services":[],"externalAccess":{"ingress":{}}}]`},
		{name: "route only", input: `[{"externalAccess":{"route":{}}}]`},
		{name: "HTTPRoute only", input: `[{"externalAccess":{"httpRoute":{}}}]`},
		{name: "NIM reference with authorization", input: `[{"services":[{"ref":{"name":"nim","port":8000},"authorization":{"authorizationType":"authenticatedUsers"}}],"externalAccess":{"ingress":{},"httpRoute":{}}}]`},
		{name: "multiple endpoints", input: `[{"services":[{"ref":{"name":"one","port":80}}]},{"services":[{"ref":{"name":"two","port":80}}]}]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var endpoints []api.NetworkingEndpoint
			if err := json.Unmarshal([]byte(tc.input), &endpoints); err != nil {
				t.Fatal(err)
			}
			err := validateNetworkingEndpoints(endpoints)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %+v", err)
				}
				return
			}
			if err == nil || err.Code != http.StatusBadRequest || !strings.Contains(err.Message, tc.wantErr) {
				t.Fatalf("got %+v, want HTTP 400 containing %q", err, tc.wantErr)
			}
		})
	}
}
