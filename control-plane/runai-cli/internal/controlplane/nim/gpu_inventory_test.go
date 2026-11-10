// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nim

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/spf13/cobra"
)

type inventorySubmission struct {
	workload.WithUpdateAction
	t      *testing.T
	called bool
}

func (c *inventorySubmission) Submit(ctx context.Context, _, _, clusterID string, gs *spec.GeneralSpec) (string, error) {
	c.called = true
	s, err := gs.GenerateNimServiceSpec(ctx, clusterID)
	if err != nil {
		return "", err
	}
	if s.GetGpuType() != "Tesla-V100" || s.Compute.Get() == nil || s.Compute.Get().GetGpuDevicesRequest() != 1 {
		c.t.Fatalf("unexpected GPU spec: %+v", s)
	}
	return "submitted", nil
}
func TestSubmitGPUTypeWithoutNodeRead(t *testing.T) {
	for _, label := range []string{"Tesla-V100", "tesla-v100"} {
		t.Run(label, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/node-pools/minimal" {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"nodepools":[{"id":"pool","name":"pool","clusterId":"cluster","gpuTypes":[{"type":"Tesla-V100","totalDevices":1}]}]}`)
			}))
			defer server.Close()
			previous := wsClient
			client := &inventorySubmission{t: t}
			wsClient = client
			t.Cleanup(func() { wsClient = previous })
			gs := spec.NewNimSpec()
			cmd := &cobra.Command{}
			project := "project"
			gs.SetFlags(cmd, &project)
			if err := cmd.ParseFlags([]string{"--gpu-type", label, "-g", "1"}); err != nil {
				t.Fatal(err)
			}
			id, err := Submit(transaction.NewTxn(), server.URL, "test-token", "nim", "project", "cluster", gs)
			if err != nil || id != "submitted" || !client.called {
				t.Fatalf("id=%s submitted=%v err=%v", id, client.called, err)
			}
		})
	}
}
