// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nodepool

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGPUTypesUsesMinimalInventory(t *testing.T) {
	pages := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/node-pools/minimal" {
			t.Errorf("unexpected endpoint requiring additional permissions: %s", r.URL.Path)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.URL.Query().Get("filterBy") != "clusterId==selected-cluster" {
			t.Errorf("missing cluster filter: %s", r.URL.RawQuery)
		}
		pages++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("offset") == "0" {
			fmt.Fprint(w, `{"nodepools":[`)
			for i := 0; i < 500; i++ {
				if i > 0 {
					fmt.Fprint(w, ",")
				}
				fmt.Fprintf(w, `{"id":"pool-%d","name":"pool","clusterId":"selected-cluster","gpuTypes":[{"type":"NVIDIA-A100-SXM4-40GB","totalDevices":1}]}`, i)
			}
			fmt.Fprint(w, `]}`)
		} else {
			fmt.Fprint(w, `{"nodepools":[{"id":"last","name":"pool","clusterId":"selected-cluster","gpuTypes":[{"type":"Tesla-V100","totalDevices":1},{"type":"","totalDevices":0}]}]}`)
		}
	}))
	defer server.Close()
	labels, err := GetGPUTypes(context.Background(), server.URL, "test-token", "selected-cluster")
	if err != nil || !reflect.DeepEqual(labels, []string{"NVIDIA-A100-SXM4-40GB", "Tesla-V100"}) || pages != 2 {
		t.Fatalf("labels=%v pages=%d err=%v", labels, pages, err)
	}
}

func TestGPUTypesPreservesInventoryFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusForbidden) }))
	defer server.Close()
	if _, err := GetGPUTypes(context.Background(), server.URL, "test-token", "cluster"); err == nil {
		t.Fatal("expected inventory error")
	}
}
