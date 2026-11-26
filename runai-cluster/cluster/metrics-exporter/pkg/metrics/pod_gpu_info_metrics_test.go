/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"context"
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/exporters/pod_exporter"
)

const (
	gpuInfoTestPodName      = "dra-gpu-info-pod"
	gpuInfoTestPodNamespace = "dra-gpu-info-ns"
	gpuInfoMetricName       = "runai_pod_gpu_info"
)

// gpuInfoStubResolver returns a fixed GPU count. The first failures calls return an error.
type gpuInfoStubResolver struct {
	gpus     int64
	failures int
}

func (s *gpuInfoStubResolver) ResolvePodRequestedGpuCount(_ context.Context, _ *v1.Pod) (int64, error) {
	if s.failures > 0 {
		s.failures--
		return 0, errors.New("claim is not readable yet")
	}
	return s.gpus, nil
}

func newGpuInfoDRATestPod() *v1.Pod {
	claimName := "gpu-claim"
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        gpuInfoTestPodName,
			Namespace:   gpuInfoTestPodNamespace,
			UID:         types.UID("dra-gpu-info-pod-uid"),
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		},
		Spec: v1.PodSpec{
			NodeName:       "gpu-node",
			Containers:     []v1.Container{{Name: "main"}},
			ResourceClaims: []v1.PodResourceClaim{{Name: "gpu", ResourceClaimName: &claimName}},
		},
		Status: v1.PodStatus{Phase: v1.PodRunning},
	}
}

// gpuInfoSeriesOfTestPod returns the labels and value of every runai_pod_gpu_info series of the test pod.
func gpuInfoSeriesOfTestPod(t *testing.T) []map[string]string {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	var series []map[string]string
	for _, family := range families {
		if family.GetName() != gpuInfoMetricName {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := map[string]string{}
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["pod_name"] != gpuInfoTestPodName || labels["pod_namespace"] != gpuInfoTestPodNamespace {
				continue
			}
			labels["value"] = "1"
			if metric.GetGauge().GetValue() != 1 {
				labels["value"] = "not-1"
			}
			series = append(series, labels)
		}
	}
	return series
}

func TestSetPodGpuInfoMetricsForDRAPod(t *testing.T) {
	t.Cleanup(func() { ResetPodGpuInfoMetrics(gpuInfoTestPodName, gpuInfoTestPodNamespace) })

	resolver := &gpuInfoStubResolver{gpus: 2}
	podExporter := pod_exporter.NewPodExporter(context.TODO(), nil, newGpuInfoDRATestPod(),
		pod_exporter.WithDRAGpuResolver(resolver))

	require.NoError(t, SetPodGpuInfoMetrics(podExporter))

	series := gpuInfoSeriesOfTestPod(t)
	require.Len(t, series, 1, "DRA pod: runai_pod_gpu_info is created once")
	require.Equal(t, "2", series[0]["gpu_num_devices"], "DRA pod with a two-GPU claim: gpu_num_devices is 2")
	require.Equal(t, "1", series[0]["value"])
}

func TestSetPodGpuInfoMetricsForDRAPodLookupFailureKeepsNoWrongSeries(t *testing.T) {
	t.Cleanup(func() { ResetPodGpuInfoMetrics(gpuInfoTestPodName, gpuInfoTestPodNamespace) })

	resolver := &gpuInfoStubResolver{gpus: 2, failures: 1}
	podExporter := pod_exporter.NewPodExporter(context.TODO(), nil, newGpuInfoDRATestPod(),
		pod_exporter.WithDRAGpuResolver(resolver))

	// Lookup failure: an error is returned, so the reconcile is retried, and no series with a wrong count exists.
	require.Error(t, SetPodGpuInfoMetrics(podExporter))
	require.Empty(t, gpuInfoSeriesOfTestPod(t))

	// Recovery: the retry exports the real device count.
	require.NoError(t, SetPodGpuInfoMetrics(podExporter))
	series := gpuInfoSeriesOfTestPod(t)
	require.Len(t, series, 1)
	require.Equal(t, "2", series[0]["gpu_num_devices"])
}
