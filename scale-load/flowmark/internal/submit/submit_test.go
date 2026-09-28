/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	orgunitclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	workloadclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
)

var testProject = ProjectRef{ID: "proj-1", Name: "flowmark-a", Namespace: "runai-flowmark-a", ClusterID: "cluster-1"}

func TestProjectRequest(t *testing.T) {
	pools := []clusterclient.Nodepool{{Id: "np-1", Name: "default"}, {Id: "np-2", Name: "gpu"}}

	request := projectRequest("flowmark-a", "cluster-1", pools)

	if request.Name != "flowmark-a" || request.ClusterId != "cluster-1" {
		t.Fatalf("name/cluster mismatch: %+v", request)
	}
	if request.Description == nil || *request.Description != ownerMarker {
		t.Fatalf("project must carry the flowmark ownership marker, got %v", request.Description)
	}
	if len(request.Resources) != len(pools) {
		t.Fatalf("expected %d resources, got %d", len(pools), len(request.Resources))
	}
	for i, resource := range request.Resources {
		pool := resource.NodePool.Get()
		if pool == nil || pool.Id != pools[i].Id || pool.GetName() != pools[i].Name {
			t.Fatalf("resource %d nodepool mismatch: %+v", i, pool)
		}
		if *resource.Gpu.Deserved != 0 || *resource.Gpu.Limit.Get() != -1 {
			t.Fatalf("resource %d gpu quota mismatch: deserved=%v limit=%v", i, *resource.Gpu.Deserved, *resource.Gpu.Limit.Get())
		}
	}
}

func TestProjectByNameIsExact(t *testing.T) {
	projects := []orgunitclient.Project{
		{Id: "1", Name: "flowmark-ab"},
		{Id: "2", Name: "Flowmark-a"},
		{Id: "3", Name: "flowmark-a"},
	}
	cases := []struct {
		name     string
		projects []orgunitclient.Project
		lookup   string
		wantID   string
	}{
		{name: "exact match wins over prefix and case variants", projects: projects, lookup: "flowmark-a", wantID: "3"},
		{name: "prefix match is not a match", projects: projects[:1], lookup: "flowmark-a"},
		{name: "case variant is not a match", projects: projects[1:2], lookup: "flowmark-a"},
		{name: "empty list", projects: nil, lookup: "flowmark-a"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			project, err := projectByName(tc.projects, tc.lookup)
			if tc.wantID == "" {
				if !errors.Is(err, errProjectNotFound) {
					t.Fatalf("expected errProjectNotFound, got %v", err)
				}
				return
			}
			if err != nil || project.Id != tc.wantID {
				t.Fatalf("got %v (%v), want id %s", project, err, tc.wantID)
			}
		})
	}
}

func TestRequireManaged(t *testing.T) {
	const flowmark = "flowmark-ci@run.ai"
	cases := []struct {
		name        string
		description string
		createdBy   string
		phase       orgunitclient.ProjectPhase
		identity    string
		wantErr     string
	}{
		{name: "flowmark project ready", description: ownerMarker, createdBy: flowmark, phase: orgunitclient.PROJECTPHASE_READY, identity: flowmark},
		{name: "flowmark project still creating", description: ownerMarker, createdBy: flowmark, phase: orgunitclient.PROJECTPHASE_CREATING, identity: flowmark},
		{name: "marker forged by another user", description: ownerMarker, createdBy: "alice@run.ai", phase: orgunitclient.PROJECTPHASE_READY, identity: flowmark, wantErr: "not created by flowmark"},
		{name: "same user but no marker", description: "team quota", createdBy: flowmark, phase: orgunitclient.PROJECTPHASE_READY, identity: flowmark, wantErr: "not created by flowmark"},
		{name: "unknown creator", description: ownerMarker, phase: orgunitclient.PROJECTPHASE_READY, identity: flowmark, wantErr: "not created by flowmark"},
		{name: "unknown identity refuses everything", description: ownerMarker, createdBy: flowmark, phase: orgunitclient.PROJECTPHASE_READY, wantErr: "identity is unknown"},
		{name: "deleting project refused", description: ownerMarker, createdBy: flowmark, phase: orgunitclient.PROJECTPHASE_DELETING, identity: flowmark, wantErr: "being deleted"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			phase := tc.phase
			project := &orgunitclient.Project{Name: "p", Description: tc.description, Status: orgunitclient.ProjectStatus{Phase: &phase}}
			if tc.createdBy != "" {
				project.CreatedBy = orgunitclient.PtrString(tc.createdBy)
			}
			err := requireManaged(project, tc.identity)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestProjectFilterIsExact(t *testing.T) {
	filter := projectFilter("cluster-1", "flowmark-a")

	want := []string{"clusterId==cluster-1", "name==flowmark-a"}
	if strings.Join(filter, "|") != strings.Join(want, "|") {
		t.Fatalf("got %v, want %v", filter, want)
	}
}

func TestTrainingRequest(t *testing.T) {
	spec := WorkloadSpec{Name: "train-1", Project: testProject, Image: "ubuntu", Command: "sleep infinity", GPUs: 2, Labels: map[string]string{"b": "2", "a": "1"}}

	request := trainingRequest(spec)

	if request.Name != "train-1" || request.ProjectId != "proj-1" || request.ClusterId != "cluster-1" {
		t.Fatalf("identity mismatch: %+v", request)
	}
	if *request.Spec.Command.Get() != "sleep infinity" || *request.Spec.Image.Get() != "ubuntu" {
		t.Fatalf("command/image mismatch: %+v", request.Spec)
	}
	if *request.Spec.ImagePullPolicy.Get() != workloadclient.IMAGEPULLPOLICY_IF_NOT_PRESENT {
		t.Fatalf("pull policy: %v", request.Spec.ImagePullPolicy.Get())
	}
	assertCompute(t, request.Spec.Compute, 2)
	assertLabels(t, request.Spec.Labels, "a=1", "b=2")
}

func TestWorkspaceRequest(t *testing.T) {
	spec := WorkloadSpec{Name: "ws-1", Project: testProject, Image: "jupyter", GPUs: 1}

	request := workspaceRequest(spec)

	if request.Name != "ws-1" || request.ProjectId != "proj-1" || request.ClusterId != "cluster-1" {
		t.Fatalf("identity mismatch: %+v", request)
	}
	if request.Spec.Command.IsSet() {
		t.Fatal("empty command must stay unset")
	}
	if *request.Spec.Image.Get() != "jupyter" {
		t.Fatalf("image: %v", request.Spec.Image.Get())
	}
	assertCompute(t, request.Spec.Compute, 1)
	if len(request.Spec.Labels) != 0 {
		t.Fatalf("expected no labels, got %v", request.Spec.Labels)
	}
}

func assertCompute(t *testing.T, compute workloadclient.NullableWorkspaceSpecSpecAllOfCompute, gpus int32) {
	t.Helper()
	value := compute.Get()
	if value == nil {
		t.Fatal("compute not set")
	}
	if *value.GpuDevicesRequest.Get() != gpus || *value.CpuCoreRequest.Get() != cpuCoreRequest || *value.CpuMemoryRequest.Get() != cpuMemoryRequest {
		t.Fatalf("compute mismatch: gpu=%v cpu=%v mem=%v", *value.GpuDevicesRequest.Get(), *value.CpuCoreRequest.Get(), *value.CpuMemoryRequest.Get())
	}
}

func assertLabels(t *testing.T, got []workloadclient.Label, want ...string) {
	t.Helper()
	pairs := make([]string, 0, len(got))
	for _, label := range got {
		pairs = append(pairs, *label.Name.Get()+"="+*label.Value.Get())
	}
	if strings.Join(pairs, ",") != strings.Join(want, ",") {
		t.Fatalf("labels %v, want %v", pairs, want)
	}
}

func TestDeploymentObject(t *testing.T) {
	userLabels := map[string]string{"flowmark": "scenario-1"}
	spec := DeploymentSpec{Name: "dep-1", Namespace: "runai-flowmark-a", Image: "nginx", Replicas: 3, Labels: userLabels}

	deployment := deploymentObject(spec)

	if deployment.Name != "dep-1" || deployment.Namespace != "runai-flowmark-a" || *deployment.Spec.Replicas != 3 {
		t.Fatalf("metadata mismatch: %+v", deployment.ObjectMeta)
	}
	if deployment.Spec.Template.Spec.Containers[0].Image != "nginx" {
		t.Fatalf("image: %v", deployment.Spec.Template.Spec.Containers)
	}
	podLabels := deployment.Spec.Template.Labels
	if podLabels["app"] != "dep-1" || deployment.Spec.Selector.MatchLabels["app"] != "dep-1" {
		t.Fatalf("selector must match pod labels: %v / %v", deployment.Spec.Selector.MatchLabels, podLabels)
	}
	if podLabels["flowmark"] != "scenario-1" {
		t.Fatalf("user labels must reach the pods: %v", podLabels)
	}
	if _, mutated := userLabels["app"]; mutated {
		t.Fatal("caller's label map must not be mutated")
	}
}

func TestClusterID(t *testing.T) {
	cases := []struct {
		name     string
		clusters string
		pick     string
		want     string
		wantErr  string
	}{
		{name: "single cluster chosen by default", clusters: `[{"uuid":"u1","tenantId":1,"name":"kind","createdAt":"2026-01-01T00:00:00Z"}]`, want: "u1"},
		{name: "several clusters need a name", clusters: `[{"uuid":"u1","tenantId":1,"name":"a","createdAt":"2026-01-01T00:00:00Z"},{"uuid":"u2","tenantId":1,"name":"b","createdAt":"2026-01-01T00:00:00Z"}]`, wantErr: "2 clusters"},
		{name: "named cluster", clusters: `[{"uuid":"u1","tenantId":1,"name":"a","createdAt":"2026-01-01T00:00:00Z"},{"uuid":"u2","tenantId":1,"name":"b","createdAt":"2026-01-01T00:00:00Z"}]`, pick: "b", want: "u2"},
		{name: "unknown name", clusters: `[{"uuid":"u1","tenantId":1,"name":"a","createdAt":"2026-01-01T00:00:00Z"}]`, pick: "zzz", wantErr: "not found"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/clusters" {
					t.Errorf("unexpected request %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.clusters))
			}))
			t.Cleanup(server.Close)
			clients := &client.Clients{ClusterService: clusterclient.NewAPIClient(&clusterclient.Configuration{
				Servers: clusterclient.ServerConfigurations{{URL: server.URL}},
			})}

			got, err := ClusterID(context.Background(), clients, tc.pick)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %q (%v), want %q", got, err, tc.want)
			}
		})
	}
}
