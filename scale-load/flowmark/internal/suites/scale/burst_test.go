/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scale

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	workloadclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/submit"
)

const testNamespace = "runai-flowmark-burst"

func TestParseSettings(t *testing.T) {
	cases := []struct {
		name    string
		knobs   map[string]any
		want    settings
		wantErr string
	}{
		{name: "defaults", knobs: nil, want: settings{rounds: []int{100, 250, 500}, workloadType: "Deployment", pods: 1, image: defaultImage, concurrency: 16, limits: client.DefaultLimits, timeout: 15 * time.Minute, setupTimeout: 5 * time.Minute, settle: 90 * time.Second}},
		{name: "yaml knobs", knobs: map[string]any{"workloads": []any{10, 20}, "podsPerWorkload": 0, "image": "mirror.local/pause:3.9", "submitConcurrency": 20, "kubeRequestsPerSecond": 100, "controlPlaneRequestsPerSecond": 10, "timeout": "30s", "setupTimeout": "1m", "settle": "5s", "cluster": "kind", "slo": map[string]any{"p90_ms": 5000}},
			want: settings{rounds: []int{10, 20}, workloadType: "Deployment", pods: 0, image: "mirror.local/pause:3.9", concurrency: 20, limits: client.Limits{KubeRequestsPerSecond: 100, ControlPlaneRequestsPerSecond: 10}, timeout: 30 * time.Second, setupTimeout: time.Minute, settle: 5 * time.Second, cluster: "kind", slo: map[string]int{"p90_ms": 5000}}},
		{name: "training", knobs: map[string]any{"workloadType": "Training"}, want: settings{rounds: []int{100, 250, 500}, workloadType: "Training", pods: 1, image: defaultImage, concurrency: 16, limits: client.DefaultLimits, timeout: 15 * time.Minute, setupTimeout: 5 * time.Minute, settle: 90 * time.Second}},
		{name: "workspace", knobs: map[string]any{"workloadType": "Workspace"}, want: settings{rounds: []int{100, 250, 500}, workloadType: "Workspace", pods: 1, image: defaultImage, concurrency: 16, limits: client.DefaultLimits, timeout: 15 * time.Minute, setupTimeout: 5 * time.Minute, settle: 90 * time.Second}},
		{name: "too much concurrency", knobs: map[string]any{"submitConcurrency": 500}, wantErr: "submitConcurrency must be between 1 and 64"},
		{name: "zero request rate", knobs: map[string]any{"kubeRequestsPerSecond": 0}, wantErr: "kubeRequestsPerSecond must be between 1 and 1000"},
		{name: "empty image", knobs: map[string]any{"image": ""}, wantErr: "image must not be empty"},
		{name: "training with pods", knobs: map[string]any{"workloadType": "Training", "podsPerWorkload": 2}, wantErr: "podsPerWorkload must be 1 for Training"},
		{name: "workspace with pods", knobs: map[string]any{"workloadType": "Workspace", "podsPerWorkload": 2}, wantErr: "podsPerWorkload must be 1 for Workspace"},
		{name: "unsupported type", knobs: map[string]any{"workloadType": "StatefulSet"}, wantErr: "is not supported, use Deployment, Training or Workspace"},
		{name: "misspelled knob", knobs: map[string]any{"workloadTipe": "Training"}, wantErr: `unknown knob "workloadTipe"; known knobs: workloads, workloadType`},
		{name: "empty rounds", knobs: map[string]any{"workloads": []any{}}, wantErr: "at least one round"},
		{name: "zero round", knobs: map[string]any{"workloads": []any{100, 0}}, wantErr: "at least 1"},
		{name: "repeated size", knobs: map[string]any{"workloads": []any{100, 100}}, wantErr: "must not repeat a round size, got 100 twice"},
		{name: "too many pods", knobs: map[string]any{"podsPerWorkload": 50}, wantErr: "between 0 and 10"},
		{name: "wrong list type", knobs: map[string]any{"workloads": "many"}, wantErr: "must be a list"},
		{name: "slo not a number", knobs: map[string]any{"slo": map[string]any{"p90_ms": "5s"}}, wantErr: `entry "p90_ms" must be a whole number`},
		{name: "slo nested", knobs: map[string]any{"slo": map[string]any{"tiers": map[string]any{"gold": 1}}}, wantErr: `entry "tiers" must be a whole number`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSettings(core.Scenario{Knobs: tc.knobs})
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got.rounds, tc.want.rounds) || got.workloadType != tc.want.workloadType || got.pods != tc.want.pods ||
				got.image != tc.want.image || got.concurrency != tc.want.concurrency || got.limits != tc.want.limits ||
				got.timeout != tc.want.timeout || got.setupTimeout != tc.want.setupTimeout || got.settle != tc.want.settle || got.cluster != tc.want.cluster ||
				!maps.Equal(got.slo, tc.want.slo) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func fakeKube() k8sclient.WithWatch {
	return fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Create: func(ctx context.Context, c k8sclient.WithWatch, obj k8sclient.Object, opts ...k8sclient.CreateOption) error {
			if obj.GetUID() == "" {
				obj.SetUID(types.UID("uid-" + obj.GetName()))
			}
			return c.Create(ctx, obj, opts...)
		},
	}).Build()
}

func fakeIntegrator(t *testing.T, ctx context.Context, kube k8sclient.WithWatch) {
	t.Helper()
	events, err := kube.Watch(ctx, &appsv1.DeploymentList{}, k8sclient.InNamespace(testNamespace))
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		defer events.Stop()
		for event := range events.ResultChan() {
			deployment, ok := event.Object.(*appsv1.Deployment)
			if !ok {
				continue
			}
			external := &unstructured.Unstructured{}
			external.SetGroupVersionKind(measure.ExternalWorkloadKind)
			external.SetNamespace(testNamespace)
			external.SetName("ew-deployment-" + deployment.Name)
			_ = unstructured.SetNestedField(external.Object, string(deployment.UID), "spec", "originalWorkloadUID")
			switch event.Type {
			case watch.Added:
				_ = kube.Create(ctx, external)
			case watch.Deleted:
				_ = kube.Delete(ctx, external)
			}
		}
	}()
}

func quick() settings {
	return settings{pods: 1, image: defaultImage, concurrency: 4, timeout: 5 * time.Second, setupTimeout: 5 * time.Second, settle: 10 * time.Millisecond}
}

func TestDeploymentRoundMeasuresEveryWorkloadAndQuiesces(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	kube := fakeKube()
	fakeIntegrator(t, ctx, kube)

	result := runRound(ctx, &deploymentFlow{kube: kube, concurrency: 4}, kube, testNamespace, "1a2b3c4d", 2, 7, quick())

	if result.err != nil {
		t.Fatalf("round must succeed: %v", result.err)
	}
	summary := measure.Stats(result.timings)
	if summary.SampleSize != 7 || summary.NeverArrivedCount != 0 || summary.FailedCount != 0 {
		t.Fatalf("all 7 deployments must be measured: %+v", summary)
	}
	for _, timing := range result.timings {
		if !strings.HasPrefix(timing.Name, "flowmark-burst-1a2b3c4d-r2-") {
			t.Fatalf("timings must be reported under the workload name, got %q", timing.Name)
		}
	}
	var left appsv1.DeploymentList
	if err := kube.List(ctx, &left, k8sclient.InNamespace(testNamespace)); err != nil || len(left.Items) != 0 {
		t.Fatalf("quiesce must delete the round's deployments, %d left (%v)", len(left.Items), err)
	}
	if result.submitMakespan <= 0 || result.quiesce <= 0 || result.unmatched != 0 {
		t.Fatalf("makespans must be measured and nothing may be unmatched: %+v", result)
	}
}

func TestDeploymentRoundLabelsObjectsWithRoundAndSize(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	kube := fakeKube()
	var seen atomic.Pointer[map[string]string]
	events, err := kube.Watch(ctx, &appsv1.DeploymentList{}, k8sclient.InNamespace(testNamespace))
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for event := range events.ResultChan() {
			if deployment, ok := event.Object.(*appsv1.Deployment); ok && event.Type == watch.Added {
				labels := deployment.Spec.Template.Labels
				seen.Store(&labels)
			}
		}
	}()
	knobs := quick()
	knobs.timeout = 50 * time.Millisecond

	runRound(ctx, &deploymentFlow{kube: kube, concurrency: 4}, kube, testNamespace, "1a2b3c4d", 3, 2, knobs)

	labels := seen.Load()
	if labels == nil || (*labels)[roundLabel] != "3" || (*labels)[roundSizeLabel] != "2" || (*labels)[runLabel] != "1a2b3c4d" {
		t.Fatalf("pod template must carry run, round index and round size labels, got %v", labels)
	}
}

func TestDeploymentRoundWithoutIntegratorReportsNeverArrivedAndStillCleansUp(t *testing.T) {
	ctx := context.Background()
	kube := fakeKube()
	knobs := quick()
	knobs.timeout = 100 * time.Millisecond

	result := runRound(ctx, &deploymentFlow{kube: kube, concurrency: 4}, kube, testNamespace, "1a2b3c4d", 1, 3, knobs)

	summary := measure.Stats(result.timings)
	if summary.NeverArrivedCount != 3 || summary.SampleSize != 0 {
		t.Fatalf("without an integrator every deployment must be marked never arrived: %+v", summary)
	}
	if result.err != nil {
		t.Fatalf("quiesce must still succeed when nothing arrived: %v", result.err)
	}
	var left appsv1.DeploymentList
	if err := kube.List(ctx, &left, k8sclient.InNamespace(testNamespace)); err != nil || len(left.Items) != 0 {
		t.Fatalf("deployments must be deleted even after a timeout, %d left", len(left.Items))
	}
}

type fakeControlPlane struct {
	kube               k8sclient.WithWatch
	deletes            atomic.Int32
	failAfterStore     atomic.Int32
	keepListingDeleted atomic.Bool

	mu        sync.Mutex
	workloads map[string]storedWorkload
}

type storedWorkload struct {
	name         string
	workloadType string
}

var controlPlaneResources = map[string]string{"trainings": trainingType, "workspaces": workspaceType}

func (cp *fakeControlPlane) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resource, id, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/api/v1/workloads/"), "/")
	workloadType, typed := controlPlaneResources[resource]
	switch {
	case r.Method == http.MethodPost && typed && id == "":
		var request struct {
			Name string `json:"name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		runner := &unstructured.Unstructured{}
		runner.SetGroupVersionKind(measure.WorkloadRunnerKind)
		runner.SetNamespace(testNamespace)
		runner.SetName(request.Name)
		_ = cp.kube.Create(r.Context(), runner)
		cp.mu.Lock()
		if cp.workloads == nil {
			cp.workloads = map[string]storedWorkload{}
		}
		cp.workloads["wl-"+request.Name] = storedWorkload{name: request.Name, workloadType: workloadType}
		cp.mu.Unlock()
		if cp.failAfterStore.Add(-1) >= 0 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"workloadId":"wl-%s","name":"%s"}`, request.Name, request.Name)
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/workloads":
		cp.mu.Lock()
		response := workloadclient.GetWorkloads200Response{Workloads: []workloadclient.Workload{}}
		for id, stored := range cp.workloads {
			response.Workloads = append(response.Workloads, workloadclient.Workload{
				Id: id, Name: stored.name, Type: stored.workloadType, ProjectId: "p1", ClusterId: "c1", Namespace: testNamespace,
				Source: workloadclient.SOURCE_CONTROL_PLANE, Phase: workloadclient.PHASE_RUNNING,
				Conditions: []workloadclient.Condition{}, Category: strings.ToLower(stored.workloadType),
			})
		}
		cp.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	case r.Method == http.MethodDelete && typed && strings.HasPrefix(id, "wl-"):
		runner := &unstructured.Unstructured{}
		runner.SetGroupVersionKind(measure.WorkloadRunnerKind)
		runner.SetNamespace(testNamespace)
		runner.SetName(strings.TrimPrefix(id, "wl-"))
		_ = cp.kube.Delete(r.Context(), runner)
		if !cp.keepListingDeleted.Load() {
			cp.mu.Lock()
			delete(cp.workloads, id)
			cp.mu.Unlock()
		}
		cp.deletes.Add(1)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func TestTrainingTeardownSendsOneDeletePerWorkloadEvenWhileTheControlPlaneStillListsThem(t *testing.T) {
	ctx := context.Background()
	kube := fakeKube()
	controlPlane := &fakeControlPlane{kube: kube}
	controlPlane.keepListingDeleted.Store(true)
	server := httptest.NewServer(controlPlane)
	t.Cleanup(server.Close)
	clients := &client.Clients{Kube: kube, Workloads: workloadclient.NewAPIClient(&workloadclient.Configuration{
		Servers: workloadclient.ServerConfigurations{{URL: server.URL}},
	})}
	flow := &controlPlaneFlow{workloadType: trainingType, clients: clients, project: submit.ProjectRef{ID: "p1", Name: "proj", Namespace: testNamespace, ClusterID: "c1"}, concurrency: 4, created: map[string]string{}}

	result := runRound(ctx, flow, kube, testNamespace, "1a2b3c4d", 1, 5, quick())

	if result.err != nil {
		t.Fatalf("the round must quiesce once the runners are gone: %v", result.err)
	}
	if _, err := flow.teardown(ctx, testNamespace, nil); err != nil {
		t.Fatalf("a later teardown pass must not fail: %v", err)
	}
	if got := controlPlane.deletes.Load(); got != 5 {
		t.Fatalf("each training must be deleted exactly once although the control plane keeps listing it, got %d deletes", got)
	}
}

func TestTrainingTeardownDeletesWorkloadsWhoseCreateFailedAfterTheServerStoredThem(t *testing.T) {
	ctx := context.Background()
	kube := fakeKube()
	controlPlane := &fakeControlPlane{kube: kube}
	controlPlane.failAfterStore.Store(1)
	server := httptest.NewServer(controlPlane)
	t.Cleanup(server.Close)
	clients := &client.Clients{Kube: kube, Workloads: workloadclient.NewAPIClient(&workloadclient.Configuration{
		Servers: workloadclient.ServerConfigurations{{URL: server.URL}},
	})}
	flow := &controlPlaneFlow{workloadType: trainingType, clients: clients, project: submit.ProjectRef{ID: "p1", Name: "proj", Namespace: testNamespace, ClusterID: "c1"}, concurrency: 4, created: map[string]string{}}

	result := runRound(ctx, flow, kube, testNamespace, "1a2b3c4d", 1, 5, quick())

	if result.err != nil {
		t.Fatalf("the round must still quiesce: %v", result.err)
	}
	summary := measure.Stats(result.timings)
	if summary.SampleSize != 4 || summary.FailedCount != 1 {
		t.Fatalf("one create failed client-side, four arrived: %+v", summary)
	}
	if result.unmatched != 0 {
		t.Fatalf("the stored workload's arrival is not a correlation problem, got %d unmatched", result.unmatched)
	}
	if got := controlPlane.deletes.Load(); got != 5 {
		t.Fatalf("all 5 stored trainings must be deleted, including the one whose create failed, got %d", got)
	}
	runners, err := countAllInNamespace(ctx, kube, measure.WorkloadRunnerKind, testNamespace)
	if err != nil || runners != 0 {
		t.Fatalf("no WorkloadRunner may remain, got %d (%v)", runners, err)
	}
}

func TestControlPlaneRoundMeasuresEveryWorkloadAndQuiesces(t *testing.T) {
	for _, workloadType := range []string{trainingType, workspaceType} {
		t.Run(workloadType, func(t *testing.T) {
			ctx := context.Background()
			kube := fakeKube()
			controlPlane := &fakeControlPlane{kube: kube}
			server := httptest.NewServer(controlPlane)
			t.Cleanup(server.Close)
			clients := &client.Clients{Kube: kube, Workloads: workloadclient.NewAPIClient(&workloadclient.Configuration{
				Servers: workloadclient.ServerConfigurations{{URL: server.URL}},
			})}
			flow := &controlPlaneFlow{workloadType: workloadType, clients: clients, project: submit.ProjectRef{ID: "p1", Name: "proj", Namespace: testNamespace, ClusterID: "c1"}, concurrency: 4, created: map[string]string{}}

			result := runRound(ctx, flow, kube, testNamespace, "1a2b3c4d", 1, 5, quick())

			if result.err != nil {
				t.Fatalf("round must succeed: %v", result.err)
			}
			if want := strings.ToLower(workloadType) + "_to_workload_runner"; flow.name() != want {
				t.Fatalf("flow name must follow the type, got %q", flow.name())
			}
			summary := measure.Stats(result.timings)
			if summary.SampleSize != 5 || summary.NeverArrivedCount != 0 {
				t.Fatalf("all 5 workloads must be measured: %+v", summary)
			}
			for _, timing := range result.timings {
				if !strings.HasPrefix(timing.Name, "flowmark-burst-1a2b3c4d-r1-") {
					t.Fatalf("timings must be reported under the workload name, got %q", timing.Name)
				}
			}
			controlPlane.mu.Lock()
			for id, stored := range controlPlane.workloads {
				if stored.workloadType != workloadType {
					t.Fatalf("%s was created through the %s API", id, stored.workloadType)
				}
			}
			controlPlane.mu.Unlock()
			if got := controlPlane.deletes.Load(); got != 5 {
				t.Fatalf("every workload must be deleted through the control plane once, got %d deletes", got)
			}
			for id := range flow.created {
				if !flow.deleteRequested[id] {
					t.Fatalf("%s was never deleted", id)
				}
			}
			runners, err := countAllInNamespace(ctx, kube, measure.WorkloadRunnerKind, testNamespace)
			if err != nil || runners != 0 {
				t.Fatalf("no WorkloadRunner may remain, got %d (%v)", runners, err)
			}
		})
	}
}

func labelledPod(name string, labels map[string]string) *corev1.Pod {
	pod := &corev1.Pod{}
	pod.Name, pod.Namespace, pod.Labels = name, testNamespace, labels
	pod.Spec.Containers = []corev1.Container{{Name: "main", Image: defaultImage}}
	return pod
}

func TestQuiesceWaitsForWorkloadObjectsAndEveryPodInTheNamespace(t *testing.T) {
	ctx := context.Background()
	round := map[string]string{runLabel: "x", roundLabel: "1"}
	stale := &unstructured.Unstructured{}
	stale.SetGroupVersionKind(measure.ExternalWorkloadKind)
	stale.SetNamespace(testNamespace)
	stale.SetName("ew-deployment-stale")

	cases := []struct {
		name    string
		objects []k8sclient.Object
		wantErr string
	}{
		{name: "external workload remains", objects: []k8sclient.Object{stale}, wantErr: "1 workload objects and 0 pods still present"},
		{name: "pod of the round remains", objects: []k8sclient.Object{labelledPod("p1", round)}, wantErr: "0 workload objects and 1 pods still present"},
		{name: "any pod in the namespace counts", objects: []k8sclient.Object{labelledPod("p2", map[string]string{"app": "other"})}, wantErr: "0 workload objects and 1 pods still present"},
		{name: "empty namespace", objects: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kube := fakeKube()
			for _, obj := range tc.objects {
				if err := kube.Create(ctx, obj); err != nil {
					t.Fatal(err)
				}
			}
			err := quiesce(ctx, &deploymentFlow{kube: kube, concurrency: 4}, kube, testNamespace, round, 50*time.Millisecond)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("quiesce must succeed: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestQuiesceReturnsOnceTheLastPodIsGone(t *testing.T) {
	ctx := context.Background()
	kube := fakeKube()
	round := map[string]string{runLabel: "x", roundLabel: "1"}
	pod := labelledPod("p1", round)
	if err := kube.Create(ctx, pod); err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = kube.Delete(ctx, pod)
	}()

	if err := quiesce(ctx, &deploymentFlow{kube: kube, concurrency: 4}, kube, testNamespace, round, 5*time.Second); err != nil {
		t.Fatalf("quiesce must succeed once the pod is gone: %v", err)
	}
}

func TestQuiesceDeletesEveryLabelledDeploymentAndNothingElse(t *testing.T) {
	ctx := context.Background()
	kube := fakeKube()
	round := map[string]string{runLabel: "x", roundLabel: "1"}
	other := map[string]string{runLabel: "x", roundLabel: "2"}
	for name, labels := range map[string]map[string]string{"r1-1": round, "r1-2": round, "r2-1": other} {
		deployment := &appsv1.Deployment{}
		deployment.Name, deployment.Namespace, deployment.Labels = name, testNamespace, labels
		if err := kube.Create(ctx, deployment); err != nil {
			t.Fatal(err)
		}
	}

	if err := quiesce(ctx, &deploymentFlow{kube: kube, concurrency: 4}, kube, testNamespace, round, 5*time.Second); err != nil {
		t.Fatal(err)
	}

	var left appsv1.DeploymentList
	if err := kube.List(ctx, &left, k8sclient.InNamespace(testNamespace)); err != nil {
		t.Fatal(err)
	}
	if len(left.Items) != 1 || left.Items[0].Name != "r2-1" {
		t.Fatalf("only the other round's deployment may survive, got %v", left.Items)
	}
}

func timing(name string, outcome measure.Outcome, d time.Duration, err error) measure.Timing {
	created := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	item := measure.Timing{Name: name, Outcome: outcome, Created: created, Err: err}
	if outcome == measure.Arrived {
		item.Arrived = created.Add(d)
		item.Duration = d
	}
	return item
}

func TestBuildSectionReportsEveryRoundAndTheSLO(t *testing.T) {
	knobs := settings{rounds: []int{2, 3}, workloadType: "Deployment", pods: 1, settle: time.Second, slo: map[string]int{"p90_ms": 5000}}
	rounds := []round{
		{index: 1, size: 2, submitMakespan: time.Second, quiesce: 2 * time.Second, timings: []measure.Timing{timing("a", measure.Arrived, time.Second, nil), timing("b", measure.Arrived, 3*time.Second, nil)}},
		{index: 2, size: 3, submitMakespan: time.Second, quiesce: 2 * time.Second, timings: []measure.Timing{timing("c", measure.Arrived, 4*time.Second, nil), timing("d", measure.Arrived, 5*time.Second, nil), timing("e", measure.Arrived, 6*time.Second, nil)}},
	}

	section := buildSection("deployment_to_external_workload", rounds, knobs)

	if section.Failed || section.Error != "" {
		t.Fatalf("clean rounds must not fail: %+v", section)
	}
	if section.Metrics["flow"] != "deployment_to_external_workload" || section.Metrics["settle_ms"] != int64(1000) {
		t.Fatalf("flow and settle must be reported: %v", section.Metrics)
	}
	for _, key := range []string{"image", "submit_concurrency", "kube_requests_per_sec", "control_plane_requests_per_sec"} {
		if _, present := section.Metrics[key]; !present {
			t.Fatalf("submission settings must be reported, %s missing", key)
		}
	}
	if section.Metrics["slo_p90_ms"] != 5000 || !maps.Equal(section.Metrics["slo"].(map[string]int), knobs.slo) {
		t.Fatalf("SLO must be copied from the yaml: %v / %v", section.Metrics["slo_p90_ms"], section.Metrics["slo"])
	}
	documented := []string{"p50_ms", "p90_ms", "p99_ms", "mean_ms", "max_ms", "throughput_per_sec", "makespan_ms", "submit_makespan_ms", "quiesce_ms", "sample_size", "never_arrived_count", "failed_count", "unmatched_arrivals", "success_rate"}
	for _, size := range knobs.rounds {
		for _, key := range documented {
			if _, present := section.Metrics[fmt.Sprintf("%s@%d", key, size)]; !present {
				t.Fatalf("documented metric %s@%d is missing", key, size)
			}
		}
	}
	if section.Metrics["p50_ms@2"] != int64(1000) || section.Metrics["p90_ms@3"] != int64(6000) || section.Metrics["never_arrived_count@3"] != 0 {
		t.Fatalf("per-round metrics wrong: %v", section.Metrics)
	}
}

func TestBuildSectionNeverArrivedRoundFailsWithoutZeros(t *testing.T) {
	knobs := settings{rounds: []int{2}, workloadType: "Deployment", pods: 1}
	rounds := []round{{index: 1, size: 2, unmatched: 2, timings: []measure.Timing{
		timing("a", measure.NeverArrived, 0, nil),
		timing("b", measure.NeverArrived, 0, errors.New("watch closed by the server")),
	}}}

	section := buildSection("deployment_to_external_workload", rounds, knobs)

	if !section.Failed || !strings.Contains(section.Error, "round 1 (2 workloads): 2 never reached the cluster") || !strings.Contains(section.Error, "watch closed") {
		t.Fatalf("a round where nothing arrived must fail with a clear message: %+v", section)
	}
	if !strings.Contains(section.Error, "2 objects arrived that matched no submitted workload") || section.Metrics["unmatched_arrivals@2"] != 2 {
		t.Fatalf("unmatched arrivals must be reported next to the never arrived count: %+v", section)
	}
	if _, present := section.Metrics["p50_ms@2"]; present {
		t.Fatal("no latency may be reported when nothing arrived")
	}
	if section.Metrics["never_arrived_count@2"] != 2 {
		t.Fatalf("never_arrived_count must be reported: %v", section.Metrics)
	}
}

func TestBuildSectionStoppedRunAndFailedCreates(t *testing.T) {
	knobs := settings{rounds: []int{1, 2, 3}, workloadType: "Training", pods: 1}
	rounds := []round{
		{index: 1, size: 1, timings: []measure.Timing{timing("a", measure.Failed, 0, errors.New("HTTP 429 from the control plane"))}},
		{index: 2, size: 2, err: errors.New("namespace did not quiesce within 5m"), timings: []measure.Timing{timing("b", measure.Arrived, time.Second, nil), timing("c", measure.Arrived, time.Second, nil)}},
	}

	section := buildSection("training_to_workload_runner", rounds, knobs)

	for _, want := range []string{"round 1 (1 workloads): 1 creates failed (HTTP 429 from the control plane)", "round 2 (2 workloads): namespace did not quiesce", "run stopped after 2 of 3 rounds"} {
		if !strings.Contains(section.Error, want) {
			t.Fatalf("error must contain %q, got %q", want, section.Error)
		}
	}
	if section.Metrics["failed_count@1"] != 1 || section.Metrics["sample_size@2"] != 2 {
		t.Fatalf("counts wrong: %v", section.Metrics)
	}
}

func TestDescribeRound(t *testing.T) {
	cases := []struct {
		name   string
		r      round
		want   []string
		absent []string
	}{
		{
			name: "every workload arrived",
			r:    round{size: 2, timings: []measure.Timing{timing("a", measure.Arrived, time.Second, nil), timing("b", measure.Arrived, 2*time.Second, nil)}},
			want: []string{"2/2 arrived", "p90 2s", "workloads/s"},
		},
		{
			name:   "nothing arrived shows no latency",
			r:      round{size: 2, timings: []measure.Timing{timing("a", measure.NeverArrived, 0, nil), timing("b", measure.Failed, 0, errors.New("quota"))}},
			want:   []string{"0/2 arrived", "1 never arrived", "1 failed"},
			absent: []string{"p90", "workloads/s"},
		},
		{
			name:   "one arrival is not a rate",
			r:      round{size: 1, timings: []measure.Timing{timing("a", measure.Arrived, 8*time.Millisecond, nil)}},
			want:   []string{"1/1 arrived", "p90 8ms"},
			absent: []string{"workloads/s"},
		},
		{
			name:   "round error is appended",
			r:      round{size: 3, unmatched: 1, err: errors.New("watch: connection refused")},
			want:   []string{"0/3 arrived", "1 unmatched arrivals", ": watch: connection refused"},
			absent: []string{"p90"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := describeRound(tc.r)
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("describeRound() = %q, want it to contain %q", got, want)
				}
			}
			for _, absent := range tc.absent {
				if strings.Contains(got, absent) {
					t.Errorf("describeRound() = %q, must not contain %q", got, absent)
				}
			}
		})
	}
}
