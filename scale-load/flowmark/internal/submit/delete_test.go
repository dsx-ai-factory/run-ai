/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	workloadclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
)

func workloadsClient(t *testing.T, handler http.HandlerFunc) *client.Clients {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &client.Clients{Workloads: workloadclient.NewAPIClient(&workloadclient.Configuration{
		Servers: workloadclient.ServerConfigurations{{URL: server.URL}},
	})}
}

func TestDeleteWorkloadsIgnoresGoneWorkloadsAndReportsTheRest(t *testing.T) {
	var deletes atomic.Int32
	clients := workloadsClient(t, func(w http.ResponseWriter, r *http.Request) {
		deletes.Add(1)
		switch {
		case r.Method != http.MethodDelete || !strings.HasPrefix(r.URL.Path, "/api/v1/workloads/workspaces/"):
			w.WriteHeader(http.StatusMethodNotAllowed)
		case strings.HasSuffix(r.URL.Path, "/gone"):
			w.WriteHeader(http.StatusNotFound)
		case strings.HasSuffix(r.URL.Path, "/broken"):
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	undeleted, err := DeleteWorkloads(context.Background(), clients, WorkspaceType, []string{"a", "gone", "b", "broken", "c"}, 3)

	if deletes.Load() != 5 {
		t.Fatalf("every id must be sent exactly once, got %d deletes", deletes.Load())
	}
	if !slices.Equal(undeleted, []string{"broken"}) {
		t.Fatalf("only the server error is undeleted, a 404 means the workload is already gone: %v", undeleted)
	}
	if err == nil || !strings.Contains(err.Error(), "deleting workspace broken") {
		t.Fatalf("the failure must name the workload: %v", err)
	}
}

func TestDeleteWorkloadsStopsWhenTheContextEndsAndCapsTheErrorList(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var deletes atomic.Int32
	clients := workloadsClient(t, func(w http.ResponseWriter, r *http.Request) {
		if deletes.Add(1) == 3 {
			cancel()
		}
		w.WriteHeader(http.StatusInternalServerError)
	})
	ids := make([]string, 50)
	for i := range ids {
		ids[i] = "w" + string(rune('a'+i%26)) + string(rune('a'+i/26))
	}

	undeleted, err := DeleteWorkloads(ctx, clients, TrainingType, ids, 1)

	if deletes.Load() >= 50 {
		t.Fatalf("deletes must stop once the context is cancelled, got %d", deletes.Load())
	}
	if len(undeleted) != 50 {
		t.Fatalf("every id is either failed or never sent, got %d undeleted", len(undeleted))
	}
	if message := err.Error(); !strings.Contains(message, "deletes left: context canceled") || strings.Count(message, "\n") > reportedDeleteErrors+1 {
		t.Fatalf("the error must mention the stop instead of listing 50 lines:\n%s", message)
	}
}

func TestSummarizeCapsTheErrorList(t *testing.T) {
	var errs []error
	for i := range 9 {
		errs = append(errs, fmt.Errorf("deleting workspace w%d: boom", i))
	}
	message := summarize(errs).Error()
	if strings.Count(message, "\n") != reportedDeleteErrors || !strings.HasSuffix(message, "and 4 more") {
		t.Fatalf("9 errors must be shown as %d lines plus a count:\n%s", reportedDeleteErrors, message)
	}
	if summarize(nil) != nil {
		t.Fatal("no errors is nil")
	}
}

func TestDeleteWorkloadsRejectsTypesWithoutAControlPlaneAPI(t *testing.T) {
	undeleted, err := DeleteWorkloads(context.Background(), &client.Clients{}, "StatefulSet", []string{"x"}, 1)
	if err == nil || len(undeleted) != 1 {
		t.Fatalf("a StatefulSet is not deleted through the control plane: %v / %v", undeleted, err)
	}
}
