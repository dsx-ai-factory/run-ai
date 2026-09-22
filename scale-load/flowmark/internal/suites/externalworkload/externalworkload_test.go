/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkload

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
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

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
)

const testNamespace = "runai-flowmark-ew"

func TestParseSettings(t *testing.T) {
	cases := []struct {
		name    string
		knobs   map[string]any
		want    settings
		wantErr string
	}{
		{name: "defaults", knobs: nil, want: settings{rounds: []int{100, 250, 500}, workloadType: "Deployment", pods: 1, timeout: 15 * time.Minute, setupTimeout: 5 * time.Minute, settle: 90 * time.Second}},
		{name: "yaml knobs", knobs: map[string]any{"workloads": []any{10, 20}, "podsPerWorkload": 0, "timeout": "30s", "setupTimeout": "1m", "settle": "5s", "cluster": "kind", "slo": map[string]any{"create_to_ew_p90_ms": 5000}},
			want: settings{rounds: []int{10, 20}, workloadType: "Deployment", pods: 0, timeout: 30 * time.Second, setupTimeout: time.Minute, settle: 5 * time.Second, cluster: "kind", slo: map[string]int{"create_to_ew_p90_ms": 5000}}},
		{name: "repeated size", knobs: map[string]any{"workloads": []any{100, 100}}, wantErr: "must not repeat a round size, got 100 twice"},
		{name: "slo not a number", knobs: map[string]any{"slo": map[string]any{"create_to_ew_p90_ms": "5s"}}, wantErr: `entry "create_to_ew_p90_ms" must be a whole number`},
		{name: "slo nested", knobs: map[string]any{"slo": map[string]any{"tiers": map[string]any{"gold": 1}}}, wantErr: `entry "tiers" must be a whole number`},
		{name: "unsupported type", knobs: map[string]any{"workloadType": "StatefulSet"}, wantErr: "not supported yet"},
		{name: "empty rounds", knobs: map[string]any{"workloads": []any{}}, wantErr: "at least one round"},
		{name: "zero round", knobs: map[string]any{"workloads": []any{100, 0}}, wantErr: "at least 1"},
		{name: "too many pods", knobs: map[string]any{"podsPerWorkload": 50}, wantErr: "between 0 and 10"},
		{name: "wrong list type", knobs: map[string]any{"workloads": "many"}, wantErr: "must be a list"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSettings(core.Config{Knobs: tc.knobs})
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

func TestRunRoundMeasuresEveryDeploymentAndQuiesces(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	kube := fakeKube()
	fakeIntegrator(t, ctx, kube)

	result := runRound(ctx, kube, testNamespace, "1a2b3c4d", 2, 7, settings{pods: 1, timeout: 5 * time.Second, setupTimeout: 5 * time.Second, settle: 10 * time.Millisecond})

	if result.err != nil {
		t.Fatalf("round must succeed: %v", result.err)
	}
	summary := measure.Stats(result.timings)
	if summary.SampleSize != 7 || summary.PendingCount != 0 || summary.FailedCount != 0 {
		t.Fatalf("all 7 deployments must be measured: %+v", summary)
	}
	for _, timing := range result.timings {
		if !strings.HasPrefix(timing.Name, "flowmark-ew-1a2b3c4d-r2-") {
			t.Fatalf("timings must be reported under the deployment name, got %q", timing.Name)
		}
	}
	var left appsv1.DeploymentList
	if err := kube.List(ctx, &left, k8sclient.InNamespace(testNamespace)); err != nil || len(left.Items) != 0 {
		t.Fatalf("quiesce must delete the round's deployments, %d left (%v)", len(left.Items), err)
	}
	if result.submitMakespan <= 0 || result.quiesce <= 0 {
		t.Fatalf("makespans must be measured: %+v", result)
	}
}

func TestRunRoundWithoutIntegratorReportsPendingAndStillCleansUp(t *testing.T) {
	ctx := context.Background()
	kube := fakeKube()

	result := runRound(ctx, kube, testNamespace, "1a2b3c4d", 1, 3, settings{pods: 1, timeout: 100 * time.Millisecond, setupTimeout: 5 * time.Second, settle: 10 * time.Millisecond})

	summary := measure.Stats(result.timings)
	if summary.PendingCount != 3 || summary.SampleSize != 0 {
		t.Fatalf("without an integrator every deployment must be pending: %+v", summary)
	}
	if result.err != nil {
		t.Fatalf("quiesce must still succeed when nothing arrived: %v", result.err)
	}
	var left appsv1.DeploymentList
	if err := kube.List(ctx, &left, k8sclient.InNamespace(testNamespace)); err != nil || len(left.Items) != 0 {
		t.Fatalf("deployments must be deleted even after a timeout, %d left", len(left.Items))
	}
}

func labelledPod(name string, labels map[string]string) *corev1.Pod {
	pod := &corev1.Pod{}
	pod.Name, pod.Namespace, pod.Labels = name, testNamespace, labels
	pod.Spec.Containers = []corev1.Container{{Name: "main", Image: pauseImage}}
	return pod
}

func TestQuiesceWaitsForExternalWorkloadsAndPodsOfTheRound(t *testing.T) {
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
		{name: "external workload remains", objects: []k8sclient.Object{stale}, wantErr: "0 deployments, 1 external workloads and 0 pods still present"},
		{name: "pod of the round remains", objects: []k8sclient.Object{labelledPod("p1", round)}, wantErr: "0 deployments, 0 external workloads and 1 pods still present"},
		{name: "pod of another round does not count", objects: []k8sclient.Object{labelledPod("p2", map[string]string{runLabel: "x", roundLabel: "2"})}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kube := fakeKube()
			for _, obj := range tc.objects {
				if err := kube.Create(ctx, obj); err != nil {
					t.Fatal(err)
				}
			}
			err := quiesce(ctx, kube, testNamespace, round, 50*time.Millisecond)
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

	if err := quiesce(ctx, kube, testNamespace, round, 5*time.Second); err != nil {
		t.Fatalf("quiesce must succeed once the pod is gone: %v", err)
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
	knobs := settings{rounds: []int{2, 3}, workloadType: "Deployment", pods: 1, slo: map[string]int{"create_to_ew_p90_ms": 5000}}
	rounds := []round{
		{index: 1, size: 2, submitMakespan: time.Second, quiesce: 2 * time.Second, timings: []measure.Timing{timing("a", measure.Arrived, time.Second, nil), timing("b", measure.Arrived, 3*time.Second, nil)}},
		{index: 2, size: 3, submitMakespan: time.Second, quiesce: 2 * time.Second, timings: []measure.Timing{timing("c", measure.Arrived, 4*time.Second, nil), timing("d", measure.Arrived, 5*time.Second, nil), timing("e", measure.Arrived, 6*time.Second, nil)}},
	}

	section := buildSection(rounds, knobs)

	if section.Failed || section.Error != "" {
		t.Fatalf("clean rounds must not fail: %+v", section)
	}
	if section.Metrics["create_to_ew_slo_ms"] != 5000 || !maps.Equal(section.Metrics["slo"].(map[string]int), knobs.slo) {
		t.Fatalf("SLO must be copied from the yaml: %v / %v", section.Metrics["create_to_ew_slo_ms"], section.Metrics["slo"])
	}
	documented := []string{"create_to_ew_p50_ms", "create_to_ew_p90_ms", "create_to_ew_p99_ms", "create_to_ew_max_ms", "throughput_per_sec", "ew_makespan_ms", "submit_makespan_ms", "quiesce_ms", "sample_size", "pending_count", "failed_count", "success_rate"}
	for _, size := range knobs.rounds {
		for _, key := range documented {
			if _, present := section.Metrics[fmt.Sprintf("%s@%d", key, size)]; !present {
				t.Fatalf("documented metric %s@%d is missing", key, size)
			}
		}
	}
	if section.Metrics["create_to_ew_p50_ms@2"] != int64(1000) || section.Metrics["create_to_ew_p90_ms@3"] != int64(6000) || section.Metrics["pending_count@3"] != 0 {
		t.Fatalf("per-round metrics wrong: %v", section.Metrics)
	}
	if section.Metrics["quiesce_ms@2"] != int64(2000) || section.Metrics["submit_makespan_ms@3"] != int64(1000) {
		t.Fatalf("round timings must be reported: %v", section.Metrics)
	}
}

func TestBuildSectionPendingRoundFailsWithoutZeros(t *testing.T) {
	knobs := settings{rounds: []int{2}, workloadType: "Deployment", pods: 1}
	rounds := []round{{index: 1, size: 2, timings: []measure.Timing{
		timing("a", measure.Pending, 0, nil),
		timing("b", measure.Pending, 0, errors.New("watch closed by the server")),
	}}}

	section := buildSection(rounds, knobs)

	if !section.Failed || !strings.Contains(section.Error, "round 1 (2 deployments): 2 never got an ExternalWorkload") || !strings.Contains(section.Error, "watch closed") {
		t.Fatalf("pending round must fail with a clear message: %+v", section)
	}
	if _, present := section.Metrics["create_to_ew_p50_ms@2"]; present {
		t.Fatal("no latency may be reported when nothing arrived")
	}
	if section.Metrics["pending_count@2"] != 2 {
		t.Fatalf("pending_count must be reported: %v", section.Metrics)
	}
}

func TestBuildSectionStoppedRunAndFailedCreates(t *testing.T) {
	knobs := settings{rounds: []int{1, 2, 3}, workloadType: "Deployment", pods: 1}
	rounds := []round{
		{index: 1, size: 1, timings: []measure.Timing{timing("a", measure.Failed, 0, errors.New("admission webhook denied"))}},
		{index: 2, size: 2, err: errors.New("namespace did not quiesce within 5m"), timings: []measure.Timing{timing("b", measure.Arrived, time.Second, nil), timing("c", measure.Arrived, time.Second, nil)}},
	}

	section := buildSection(rounds, knobs)

	for _, want := range []string{"round 1 (1 deployments): 1 creates failed (admission webhook denied)", "round 2 (2 deployments): namespace did not quiesce", "run stopped after 2 of 3 rounds"} {
		if !strings.Contains(section.Error, want) {
			t.Fatalf("error must contain %q, got %q", want, section.Error)
		}
	}
	if section.Metrics["failed_count@1"] != 1 || section.Metrics["sample_size@2"] != 2 {
		t.Fatalf("counts wrong: %v", section.Metrics)
	}
}

func TestQuiesceDeletesEveryLabelledDeploymentAndNothingElse(t *testing.T) {
	ctx := context.Background()
	kube := fakeKube()
	round := map[string]string{runLabel: "x", roundLabel: "100"}
	other := map[string]string{runLabel: "x", roundLabel: "250"}
	for name, labels := range map[string]map[string]string{"r100-1": round, "r100-2": round, "r250-1": other} {
		deployment := &appsv1.Deployment{}
		deployment.Name, deployment.Namespace, deployment.Labels = name, testNamespace, labels
		if err := kube.Create(ctx, deployment); err != nil {
			t.Fatal(err)
		}
	}

	if err := quiesce(ctx, kube, testNamespace, round, 5*time.Second); err != nil {
		t.Fatal(err)
	}

	var left appsv1.DeploymentList
	if err := kube.List(ctx, &left, k8sclient.InNamespace(testNamespace)); err != nil {
		t.Fatal(err)
	}
	if len(left.Items) != 1 || left.Items[0].Name != "r250-1" {
		t.Fatalf("only the other round's deployment may survive, got %v", left.Items)
	}
}
