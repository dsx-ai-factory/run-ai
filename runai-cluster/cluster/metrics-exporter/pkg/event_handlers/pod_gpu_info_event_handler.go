package event_handlers

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/metrics_exporter/pkg/exporters/pod_exporter"
	"github.com/run-ai/runai/runai-cluster/cluster/metrics_exporter/pkg/metrics"
)

type PodGpuInfoEventHandler struct {
	k8sClient client.Client
}

func NewPodGpuInfoEventHandler(k8sClient client.Client) *PodGpuInfoEventHandler {
	return &PodGpuInfoEventHandler{k8sClient: k8sClient}
}

func (pe *PodGpuInfoEventHandler) ResetMetrics(identifier ObjectIdentifier) {
	metrics.ResetPodGpuInfoMetrics(identifier.Name, identifier.Namespace)
}

func (pe *PodGpuInfoEventHandler) SetMetrics(ctx context.Context, eventObject interface{}) error {
	pod, ok := eventObject.(*v1.Pod)
	if !ok {
		return fmt.Errorf("PodGpuInfoEventHandler: eventObject is not a pod")
	}

	podExporter := pod_exporter.NewPodExporter(ctx, pe.k8sClient, pod)
	return metrics.SetPodGpuInfoMetrics(podExporter)
}
