/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package prometheus_monitor

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/sampler_actions/utils/kubernetes"
	time_utils "github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/sampler_actions/utils/time"
	controller_runtime_utils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/controller-runtime"

	rand "math/rand/v2"

	"github.com/phayes/freeport"
	"github.com/dsx-ai-factory/run-ai/scale-tests/sdk/reports"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
	lm "github.com/samber/lo/mutable"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

const (
	TestResourcesNamespace = "runai-test"
)

type portForwarder interface {
	PortForward(namespace, podName string, localPort, remotePort int, stopCh <-chan struct{}, readyCh chan struct{}, steams genericclioptions.IOStreams) error
}

// PrometheusMonitor sets up a Promtheus instance for monitoring Run:AI's Prometheus instance.
// Requires Prometheus Operator to be installed in the cluster.
type PrometheusMonitor struct {
	k8sClient     client.Client
	portForwarder portForwarder
	kubeResources []client.Object
	clusterConfig *scale.ClusterConfig
}

func NewPrometheusMonitor(k8sClient client.Client, config *rest.Config, clusterConfig *scale.ClusterConfig, metricLabels map[string]string) *PrometheusMonitor {
	return &PrometheusMonitor{
		k8sClient:     k8sClient,
		portForwarder: kubernetes.NewPortForwarder(config),
		kubeResources: generatePrometheusMonitorResources(clusterConfig, metricLabels),
		clusterConfig: clusterConfig,
	}
}

func (pm *PrometheusMonitor) Setup(ctx context.Context) error {
	for _, res := range pm.kubeResources {
		err := controller_runtime_utils.CreateOrUpdate(ctx, pm.k8sClient, res)
		if err != nil {
			return fmt.Errorf("failed to create or update resource: %w", err)
		}
	}

	return pm.waitForPrometheusPod(ctx)
}

func (pm *PrometheusMonitor) Sample(ctx context.Context, since time.Time) (map[scale.MetricType]reports.PerformanceMetric, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	port, err := pm.portForwardPrometheusMonitor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to port forward Prometheus Monitor pod: %w", err)
	}

	client, err := api.NewClient(api.Config{
		Address: fmt.Sprintf("http://localhost:%d", port),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	v1api := v1.NewAPI(client)
	performanceMetrics := map[scale.MetricType]reports.PerformanceMetric{}
	for _, extractor := range prometheusPerformanceMetricExtractors {
		pm, err := extractor(ctx, v1api, since)
		if err != nil {
			return nil, fmt.Errorf("failed to extract performance metric: %w", err)
		}

		performanceMetrics[pm.typ] = pm.val
	}

	return performanceMetrics, nil
}

func (pm *PrometheusMonitor) Teardown(ctx context.Context) error {
	// If remote write is enabled, wait until the samples are flushed to the remote write
	if pm.clusterConfig.PrometheusRemoteWriteEnabled {
		err := pm.waitUntilSamplesFlushed(ctx)
		if err != nil {
			return fmt.Errorf("failed to wait until samples are flushed: %w", err)
		}
	}

	reversedKubeResources := append([]client.Object{}, pm.kubeResources...)
	lm.Reverse(reversedKubeResources)

	for _, res := range reversedKubeResources {
		err := pm.k8sClient.Delete(ctx, res)
		if err != nil {
			return fmt.Errorf("failed to delete resource %s/%s: %w", res.GetObjectKind().GroupVersionKind().Kind, res.GetName(), err)
		}
	}

	return nil
}

// Get current prometheus_remote_storage_highest_timestamp_in_seconds and wait until
// it is greater than or equal to prometheus_remote_storage_queue_highest_sent_timestamp_seconds
func (pm *PrometheusMonitor) waitUntilSamplesFlushed(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 6*time.Minute)
	defer cancel()

	port, err := pm.portForwardPrometheusMonitor(ctx)
	if err != nil {
		return fmt.Errorf("failed to port forward Prometheus Monitor pod: %w", err)
	}

	client, err := api.NewClient(api.Config{
		Address: fmt.Sprintf("http://localhost:%d", port),
	})
	if err != nil {
		return fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	v1api := v1.NewAPI(client)

	highestTimestamp, err := extractUniqueInstantVectorValue(ctx, v1api, "prometheus_remote_storage_highest_timestamp_in_seconds")
	if err != nil {
		return fmt.Errorf("failed to extract prometheus_remote_storage_highest_timestamp_in_seconds: %w", err)
	}
	if highestTimestamp == nil {
		return fmt.Errorf("prometheus_remote_storage_highest_timestamp_in_seconds is nil")
	}

	// Wait until samples are flushed
	return time_utils.WaitFor(ctx, func() bool {
		queueTimestamp, err := extractUniqueInstantVectorValue(ctx, v1api, "prometheus_remote_storage_queue_highest_sent_timestamp_seconds")
		if err != nil {
			return false
		}
		if queueTimestamp == nil {
			return false
		}

		return float64(*queueTimestamp) >= float64(*highestTimestamp)
	}, 5*time.Minute, 5*time.Second)
}

func (pm *PrometheusMonitor) portForwardPrometheusMonitor(ctx context.Context) (int, error) {
	freePort, err := freeport.GetFreePort()
	if err != nil {
		return 0, fmt.Errorf("failed to find free port: %w", err)
	}

	readyCh := make(chan struct{})
	errCh := make(chan error)
	defer close(errCh)
	go func() {
		err := pm.portForwarder.PortForward(TestResourcesNamespace, "prometheus-runai-prometheus-monitor-0", freePort, 9090, ctx.Done(), readyCh, genericclioptions.IOStreams{
			ErrOut: os.Stderr,
		})
		if err != nil {
			errCh <- fmt.Errorf("failed to port forward Prometheus pod: %w", err)
		}
	}()

	select {
	case <-readyCh:
	case <-time.After(5 * time.Minute):
		return 0, fmt.Errorf("timed out waiting for port forward to be ready")
	case err := <-errCh:
		return 0, err
	}

	return freePort, nil
}

func (pm *PrometheusMonitor) waitForPrometheusPod(ctx context.Context) error {
	return time_utils.WaitFor(ctx, func() bool {
		pod := &corev1.Pod{}
		err := pm.k8sClient.Get(ctx, client.ObjectKey{Name: "prometheus-runai-prometheus-monitor-0", Namespace: TestResourcesNamespace}, pod)
		if err != nil {
			return false
		}

		if len(pod.Spec.Containers) > len(pod.Status.ContainerStatuses) {
			return false
		}

		for _, container := range pod.Status.ContainerStatuses {
			if !container.Ready {
				return false
			}
		}

		return true
	}, 5*time.Minute, 5*time.Second)
}

func generateRandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}
