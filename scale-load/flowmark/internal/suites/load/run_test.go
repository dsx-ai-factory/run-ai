/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package load

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	workloadclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/submit"
)

const testNamespace = "runai-flowmark-load"

type fakeControlPlane struct {
	kube        k8sclient.WithWatch
	failDeletes bool
	deletes     atomic.Int32

	mu    sync.Mutex
	names map[string]string
}

func (cp *fakeControlPlane) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resource, id, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/api/v1/workloads/"), "/")
	switch {
	case r.Method == http.MethodPost && resource == "workspaces" && id == "":
		var request struct {
			Name string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		runner := &unstructured.Unstructured{}
		runner.SetGroupVersionKind(measure.WorkloadRunnerKind)
		runner.SetNamespace(testNamespace)
		runner.SetName(request.Name)
		_ = cp.kube.Create(r.Context(), runner)
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: testNamespace, Name: request.Name + "-0-0"}}
		_ = cp.kube.Create(r.Context(), pod)
		cp.mu.Lock()
		if cp.names == nil {
			cp.names = map[string]string{}
		}
		cp.names["wl-"+request.Name] = request.Name
		cp.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"workloadId":"wl-%s","name":"%s"}`, request.Name, request.Name)
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/workloads":
		cp.mu.Lock()
		response := workloadclient.GetWorkloads200Response{Workloads: []workloadclient.Workload{}}
		for id, name := range cp.names {
			response.Workloads = append(response.Workloads, workloadclient.Workload{
				Id: id, Name: name, Type: submit.WorkspaceType, ProjectId: "p1", ClusterId: "c1", Namespace: testNamespace,
				Source: workloadclient.SOURCE_CONTROL_PLANE, Phase: workloadclient.PHASE_RUNNING,
				Conditions: []workloadclient.Condition{}, Category: "workspace",
			})
		}
		cp.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	case r.Method == http.MethodDelete && resource == "workspaces":
		cp.deletes.Add(1)
		if cp.failDeletes {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		cp.mu.Lock()
		delete(cp.names, id)
		cp.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func fakeSuite(t *testing.T, controlPlane *fakeControlPlane) Suite {
	t.Helper()
	kube := fake.NewClientBuilder().Build()
	controlPlane.kube = kube
	server := httptest.NewServer(controlPlane)
	t.Cleanup(server.Close)
	clients := &client.Clients{Kube: kube, Workloads: workloadclient.NewAPIClient(&workloadclient.Configuration{
		Servers: workloadclient.ServerConfigurations{{URL: server.URL}},
	})}
	return Suite{openSession: func(context.Context, core.Target, client.Limits, string, string, time.Duration) (*submit.Session, error) {
		return &submit.Session{Clients: clients, ClusterID: "c1", RunID: "1a2b3c4d", Project: submit.ProjectRef{ID: "p1", Namespace: testNamespace, ClusterID: "c1"}}, nil
	}}
}

func quickScenario() core.Scenario {
	return core.Scenario{Suite: suiteName, Knobs: map[string]any{"users": 2, "rate": 5, "duration": "1s", "timeout": "5s", "setupTimeout": "5s"}}
}

func TestSuiteRunMeasuresEverySpanWithAFakeSession(t *testing.T) {
	controlPlane := &fakeControlPlane{}
	suite := fakeSuite(t, controlPlane)

	section, err := suite.Run(context.Background(), core.Target{}, quickScenario())

	if err != nil || section.Failed {
		t.Fatalf("a clean paced minute must pass: %v / %+v", err, section)
	}
	for _, key := range []string{"api_accepted_p90_ms", "cluster_synced_p90_ms", "pod_created_p90_ms", "ttj_p90_ms", "failed_count"} {
		if _, present := section.Metrics[key]; !present {
			t.Fatalf("%s missing from %v", key, section.Metrics)
		}
	}
	if section.Metrics["failed_count"] != 0 {
		t.Fatalf("every workspace must be accepted and followed to its pod: %v", section.Metrics)
	}
	if got := controlPlane.deletes.Load(); got < 4 || got > 6 {
		t.Fatalf("about 5 workspaces are submitted in one second at 5/s and each is deleted once, got %d deletes", got)
	}
}

func TestSuiteRunFailsTheSectionWhenCleanupDeletesFail(t *testing.T) {
	suite := fakeSuite(t, &fakeControlPlane{failDeletes: true})

	section, err := suite.Run(context.Background(), core.Target{}, quickScenario())

	if err != nil || !section.Failed || !strings.Contains(section.Error, "deleting workspace") {
		t.Fatalf("delete failures must fail the section and name the workload: %v / %+v", err, section)
	}
	if _, present := section.Metrics["ttj_p90_ms"]; !present {
		t.Fatalf("the measurements are still reported next to the cleanup failure: %v", section.Metrics)
	}
}
