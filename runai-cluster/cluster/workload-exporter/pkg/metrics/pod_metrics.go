package metrics

import (
	"context"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"

	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/metrics/data"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	podInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "pod_info",
			Help:      "Pod info",
		}, []string{
			PodNameLabel, PodUUIDLabel, NamespaceLabel,
			queueLabel, nodepoolLabel, projectLabel, departmentLabel,
			WorkloadIdLabel,
			nodeLabel,
			phaseLabel, detailedStatusLabel},
	)

	podRequestedGpus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "scheduler_requested_gpus_per_pod",
			Help:      "Requested gpus per pod",
		}, []string{PodNameLabel, PodUUIDLabel, NamespaceLabel, WorkloadIdLabel},
	)

	podRequestedGPUMemory = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "scheduler_requested_gpu_memory_mb_per_pod",
			Help:      "Requested gpu memory per pod, in MB",
		}, []string{PodNameLabel, PodUUIDLabel, NamespaceLabel, WorkloadIdLabel},
	)

	podLimitGpus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "scheduler_gpus_limit_per_pod",
			Help:      "Gpus limit per pod",
		}, []string{PodNameLabel, PodUUIDLabel, NamespaceLabel, WorkloadIdLabel},
	)

	podLimitGPUMemory = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "scheduler_gpu_memory_mb_limit_per_pod",
			Help:      "Gpu memory limit per pod, in MB",
		}, []string{PodNameLabel, PodUUIDLabel, NamespaceLabel, WorkloadIdLabel},
	)

	podAllocatedGPUs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "scheduler_allocated_gpus_per_pod",
			Help:      "Allocated gpus per pod",
		}, []string{PodNameLabel, PodUUIDLabel, NamespaceLabel, WorkloadIdLabel},
	)

	podAllocatedGPUMemory = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "scheduler_allocated_gpu_memory_mb_per_pod",
			Help:      "Allocated gpu memory per pod, in MB",
		}, []string{PodNameLabel, PodUUIDLabel, NamespaceLabel, WorkloadIdLabel},
	)

	podAllocatedCPUs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "scheduler_allocated_millicpus_per_pod",
			Help:      "Allocated milliCPUs per pod",
		}, []string{PodNameLabel, PodUUIDLabel, NamespaceLabel, WorkloadIdLabel},
	)

	podAllocatedCpuMemory = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "scheduler_allocated_memory_bytes_per_pod",
			Help:      "Allocated memory per pod, in bytes",
		}, []string{PodNameLabel, PodUUIDLabel, NamespaceLabel, WorkloadIdLabel},
	)

	podElements = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "pod_elements",
			Help:      "Pod element hierarchy mapping",
		}, []string{PodNameLabel, NamespaceLabel, PodUUIDLabel, WorkloadIdLabel, elementIDLabel, elementLevelLabel},
	)
)

func SetPodMetrics(ctx context.Context, pod *corev1.Pod, client client.Client) error {
	name := pod.Name
	namespace := pod.Namespace
	uuid := string(pod.UID)
	nodeName := pod.Spec.NodeName

	podDetailedPhase := data.GetPodDetailedPhase(pod)

	orgData, err := utils.GetOrgData(ctx, namespace, client)
	if err != nil {
		return err
	}
	projectName := orgData.ProjectName
	departmentName := orgData.DepartmentName

	nodePoolName, err := data.GetNodePoolOfPod(ctx, pod, podDetailedPhase, client)
	if err != nil {
		return err
	}
	queueName, err := utils.GetQueueName(ctx, nodePoolName, projectName, client)
	if err != nil {
		return err
	}

	workloadId, err := data.GetWorkloadIdForPod(pod)
	if err != nil {
		return err
	}

	phaseInfo, err := data.GetPodPhaseMetricLabelsInfo(pod, podDetailedPhase)
	if err != nil {
		return err
	}

	resourcesRequested, resourcesLimit := data.GetPodResources(pod)
	requestedGPUs := resourcesRequested.Gpus
	requestedGPUsMemory := resourcesRequested.GpuMemoryMegaBytes
	limitGPUs := resourcesLimit.Gpus
	limitGPUsMemory := resourcesLimit.GpuMemoryMegaBytes

	allocatedResources := data.GetPodAllocatedResources(pod, podDetailedPhase)
	allocatedGPUs := allocatedResources.Gpus
	allocatedGpuMemory := allocatedResources.GpuMemoryMegaBytes
	allocatedCPUs := allocatedResources.CpuMilliCores
	allocatedCpuMemory := allocatedResources.CpuMemoryBytes

	ResetPodMetrics(name, namespace)

	podInfo.WithLabelValues(
		name, uuid, namespace,
		queueName, nodePoolName, projectName, departmentName,
		workloadId,
		nodeName,
		string(phaseInfo.Phase), phaseInfo.DetailedStatus).Set(1)
	podRequestedGpus.WithLabelValues(name, uuid, namespace, workloadId).Set(requestedGPUs)
	podRequestedGPUMemory.WithLabelValues(name, uuid, namespace, workloadId).Set(requestedGPUsMemory)
	podLimitGpus.WithLabelValues(name, uuid, namespace, workloadId).Set(limitGPUs)
	podLimitGPUMemory.WithLabelValues(name, uuid, namespace, workloadId).Set(limitGPUsMemory)
	podAllocatedGPUs.WithLabelValues(name, uuid, namespace, workloadId).Set(allocatedGPUs)
	podAllocatedGPUMemory.WithLabelValues(name, uuid, namespace, workloadId).Set(allocatedGpuMemory)
	podAllocatedCPUs.WithLabelValues(name, uuid, namespace, workloadId).Set(allocatedCPUs)
	podAllocatedCpuMemory.WithLabelValues(name, uuid, namespace, workloadId).Set(allocatedCpuMemory)

	for level := 1; level <= commonconstants.MaxStoredElementLevels; level++ {
		elementID := pod.Labels[fmt.Sprintf("%s%d-id", commonconstants.ElementLabelPrefix, level)]
		if elementID == "" {
			break
		}
		podElements.WithLabelValues(name, namespace, uuid, workloadId, elementID, strconv.Itoa(level)).Set(1)
	}

	log.Info().Msgf("Updated Pod %s/%s metrics successfully", namespace, name)
	return nil
}

func ResetPodMetrics(podName, podNamespace string) {
	podIdentifier := prometheus.Labels{NamespaceLabel: podNamespace, PodNameLabel: podName}
	podRequestedGpus.DeletePartialMatch(podIdentifier)
	podRequestedGPUMemory.DeletePartialMatch(podIdentifier)
	podLimitGpus.DeletePartialMatch(podIdentifier)
	podLimitGPUMemory.DeletePartialMatch(podIdentifier)
	podAllocatedGPUs.DeletePartialMatch(podIdentifier)
	podAllocatedGPUMemory.DeletePartialMatch(podIdentifier)
	podAllocatedCPUs.DeletePartialMatch(podIdentifier)
	podAllocatedCpuMemory.DeletePartialMatch(podIdentifier)
	podInfo.DeletePartialMatch(podIdentifier)
	podElements.DeletePartialMatch(podIdentifier)
}

func GetPodInfoMetric() *prometheus.GaugeVec {
	return podInfo
}

func GetPodElementsMetric() *prometheus.GaugeVec {
	return podElements
}

func GetPodResourcesMetrics() []*prometheus.GaugeVec {
	return []*prometheus.GaugeVec{
		podRequestedGpus, podRequestedGPUMemory,
		podLimitGpus, podLimitGPUMemory,
		podAllocatedGPUs, podAllocatedGPUMemory,
		podAllocatedCPUs, podAllocatedCpuMemory,
	}
}
