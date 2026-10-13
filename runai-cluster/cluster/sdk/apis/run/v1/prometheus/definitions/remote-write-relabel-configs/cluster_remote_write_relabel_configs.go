/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package remote_write_relabel_configs

import (
	"fmt"
	"slices"
	"strings"

	runv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"k8s.io/utils/ptr"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

const (
	ClusterMetricsRemoteWriteLabel = "runai"
)

var genericMetrics = []string{
	"kube_pod_info",
	"kube_node_status_capacity",
	"runai_node_nodepool",
	"reporter_push_gateway_.*",
	"kube_node_status_condition",
	"runai_pod_group_phase_with_info",
	"runai_total_preemption_attempts",
	"runai_queue_info",
	"runai_queue_quota_cpu_cores",
	"runai_queue_quota_memory_bytes",
	"runai_pod_group_total_runtime",
	"runai_run_time_seconds_per_workload",
	"runai_pod_group_total_wait_time",
	"runai_wait_time_seconds_per_workload",
	"runai_active_job_memory_requested_bytes",
	"runai_active_job_cpu_requested_cores",
	"runai_cpu_limits_per_active_workload",
	"runai_memory_limits_per_active_workload",
	"runai_gpus_limit_per_workload",
	"runai_gpu_memory_mb_limit_per_workload",
	"runai_allocated_gpus",
	"runai_mig_mode_gpu_count",
	"runai_cluster_name",
	"runai_requested_mig_devices",
	"runai_requested_mig_devices_per_workload",
	"runai_project_info",
	"runai_requested_gpus_per_workload",
	"runai_requested_gpu_memory_mb_per_workload",
	"runai_requested_gpus_per_workload_type_per_queue",
	// nim metrics
	"runai_nim_num_requests_running_per_pod",
	"runai_nim_num_requests_running_per_workload",
	"runai_nim_num_requests_waiting_per_pod",
	"runai_nim_num_requests_waiting_per_workload",
	"runai_nim_num_request_max_per_pod",
	"runai_nim_num_request_max_per_workload",
	"runai_nim_request_success_total_per_pod",
	"runai_nim_request_success_total_per_workload",
	"runai_nim_request_failure_total_per_pod",
	"runai_nim_request_failure_total_per_workload",
	"runai_nim_gpu_cache_usage_perc_per_pod",
	"runai_nim_gpu_cache_usage_perc_per_workload",
	// nim metrics time_to_first_token_seconds
	"runai_nim_time_to_first_token_seconds_bucket_per_pod",
	"runai_nim_time_to_first_token_seconds_sum_per_pod",
	"runai_nim_time_to_first_token_seconds_count_per_pod",
	// nim metrics e2e_request_latency_seconds
	"runai_nim_e2e_request_latency_seconds_bucket_per_pod",
	"runai_nim_e2e_request_latency_seconds_sum_per_pod",
	"runai_nim_e2e_request_latency_seconds_count_per_pod",
	// vLLM metrics are all Run:ai recording rules; never export raw vllm:* series.
	"runai_vllm_.*",
	// knative metrics
	"revision_app_request_count",
	"revision_app_request_latencies_sum",
	"revision_app_request_latencies_count",
	"revision_app_request_latencies_bucket",
	"kn_serving_invocation_duration_seconds_count",
	"kn_serving_invocation_duration_seconds_sum",
	"runai_inference_throughput_per_pod",
	"runai_inference_latency_per_pod",
	"kube_node_labels",
}

var newGenMetrics = []string{
	"runai_idle_allocated_gpu_count_per_workload",
	"runai_idle_allocated_gpu_count_per_node",
	"runai_idle_allocated_gpu_count_per_nodepool",
	"runai_pod_phase_with_info",
	"runai_pod_completion_time",
	"runai_gpu_count_per_node",
	"runai_total_gpus_per_nodepool",
	"runai_gpu_count_per_nodepool",
	"runai_ready_gpu_count_per_nodepool",
	"runai_gpu_count_per_cluster",
	"runai_ready_gpu_count_per_cluster",

	"runai_gpu_nodes_count_per_nodepool",
	"runai_ready_gpu_nodes_count_per_nodepool",
	"runai_gpu_nodes_count_per_cluster",
	"runai_ready_gpu_nodes_count_per_cluster",

	// "runai_pending_requested_gpu_count_per_workload",
	"runai_pending_requested_gpu_count_per_queue",
	// "runai_pending_requested_gpu_count_per_project",
	// "runai_pending_requested_gpu_count_per_department",
	"runai_pending_requested_gpu_count_per_nodepool",
	"runai_pending_requested_gpu_count_per_cluster",

	"runai_pod_elements",
	"runai_requested_gpu_memory_mb_per_pod",
	"runai_cpu_requested_cores_per_pod",
	"runai_cpu_limit_cores_per_pod",
	"runai_memory_requested_bytes_per_pod",
	"runai_memory_limit_bytes_per_pod",
	"runai_allocated_gpu_count_per_pod",
	"runai_allocated_gpu_count_per_pod:hourly",
	"runai_allocated_gpu_count_per_workload",
	"runai_allocated_gpu_count_per_queue_per_workload_type",
	"runai_allocated_gpu_count_per_queue",
	"runai_allocated_gpu_count_per_project",
	"runai_allocated_gpu_count_per_project:hourly",
	"runai_allocated_gpu_count_per_department",
	"runai_allocated_gpu_count_per_department:hourly",
	"runai_allocated_gpu_count_per_gpu",
	"runai_allocated_gpu_count_per_node",
	"runai_allocated_gpu_count_per_nodepool",
	"runai_allocated_gpu_count_per_nodepool:hourly",
	"runai_allocated_gpu_count_per_cluster",
	"runai_allocated_gpu_count_per_cluster:hourly",
	"runai_free_gpu_count_per_node",

	"runai_gpu_allocation_percentage_per_node",
	"runai_gpu_allocation_percentage_per_nodepool",
	"runai_gpu_allocation_percentage_per_cluster",

	"runai_deserved_gpu_count_per_queue",
	"runai_deserved_gpu_count_per_queue:hourly",
	"runai_deserved_gpu_count_per_project",
	"runai_deserved_gpu_count_per_project:hourly",
	"runai_deserved_gpu_count_per_department",
	"runai_deserved_gpu_count_per_department:hourly",

	"runai_gpu_utilization_per_pod_per_gpu",
	"runai_gpu_utilization_per_workload",
	"runai_last_gpu_utilization_time_per_workload",
	"runai_gpu_idle_seconds_per_workload",
	"runai_gpu_idle_hours_per_queue:hourly",
	"runai_gpu_utilization_per_queue_per_workload_type",
	"runai_gpu_utilization_per_queue",
	"runai_gpu_utilization_per_project_per_workload_type",
	"runai_gpu_utilization_per_project",
	"runai_gpu_utilization_per_project:hourly",
	"runai_gpu_utilization_per_department",
	"runai_gpu_utilization_per_department:hourly",
	"runai_gpu_memory_utilization_per_project:hourly",
	"runai_gpu_memory_utilization_per_department:hourly",
	"runai_gpu_memory_utilization_per_queue",
	"runai_gpu_utilization_per_gpu",
	"runai_last_gpu_utilization_time_per_gpu",
	"runai_gpu_idle_seconds_per_gpu",
	"runai_gpu_utilization_per_node",
	"runai_gpu_utilization_per_nodepool",
	"runai_gpu_utilization_per_cluster",

	"runai_allocated_gpu_memory_mb_per_workload",

	"runai_gpu_memory_used_mebibytes_per_pod_per_gpu",
	"runai_gpu_memory_used_mebibytes_per_workload",
	"runai_gpu_memory_used_mebibytes_per_queue_per_workload_type",
	// "runai_gpu_memory_used_mebibytes_per_queue",
	// "runai_gpu_memory_used_mebibytes_per_project",
	// "runai_gpu_memory_used_mebibytes_per_department",
	"runai_gpu_memory_used_mebibytes_per_gpu",
	"runai_gpu_memory_used_mebibytes_per_node",
	"runai_gpu_memory_used_mebibytes_per_nodepool",
	"runai_gpu_memory_used_mebibytes_per_cluster",

	"runai_gpu_memory_total_mebibytes_per_gpu",
	"runai_gpu_memory_total_mebibytes_per_node",
	"runai_gpu_memory_total_mebibytes_per_nodepool",
	"runai_gpu_memory_total_mebibytes_per_cluster",

	"runai_gpu_memory_utilization_per_node",
	"runai_gpu_memory_utilization_per_nodepool",
	"runai_gpu_memory_utilization_per_cluster",

	"runai_pending_requested_cpu_cores_per_workload",
	"runai_pending_requested_cpu_cores_per_queue",
	// "runai_pending_requested_cpu_cores_per_project",
	// "runai_pending_requested_cpu_cores_per_department",
	"runai_pending_requested_cpu_cores_per_nodepool",
	"runai_pending_requested_cpu_cores_per_cluster",

	"runai_allocated_millicpus_per_pod",
	// "runai_allocated_cpu_cores_per_pod",
	"runai_allocated_millicpus_per_pod:hourly",
	"runai_allocated_cpu_cores_per_workload",
	"runai_allocated_cpu_cores_per_queue",
	"runai_allocated_cpu_cores_per_project",
	"runai_allocated_cpu_cores_per_department",
	"runai_allocated_cpu_cores_per_node",
	"runai_allocated_cpu_cores_per_nodepool",
	"runai_allocated_cpu_cores_per_cluster",

	"runai_used_cpu_cores_per_pod",
	"runai_used_cpu_cores_per_pod:hourly",
	"runai_used_cpu_cores_per_workload",
	"runai_used_cpu_cores_per_queue",
	// "runai_used_cpu_cores_per_project",
	// "runai_used_cpu_cores_per_department",
	"runai_used_cpu_cores_per_node",
	"runai_used_cpu_cores_per_nodepool",
	"runai_used_cpu_cores_per_cluster",

	"runai_deserved_cpu_cores_per_queue",
	"runai_deserved_cpu_cores_per_project",
	"runai_deserved_cpu_cores_per_department",

	"runai_cpu_cores_count_per_node",
	"runai_cpu_cores_count_per_nodepool",
	"runai_ready_cpu_cores_count_per_nodepool",
	"runai_cpu_cores_count_per_cluster",
	"runai_ready_cpu_cores_count_per_cluster",

	"runai_cpu_cores_allocation_percentage_per_node",
	"runai_cpu_cores_allocation_percentage_per_nodepool",
	"runai_cpu_cores_allocation_percentage_per_cluster",

	"runai_cpu_utilization_per_pod",
	"runai_cpu_utilization_per_workload",
	"runai_cpu_utilization_per_queue",
	"runai_cpu_utilization_per_project",
	"runai_cpu_utilization_per_department",
	"runai_cpu_utilization_per_node",
	"runai_cpu_utilization_per_nodepool",
	"runai_cpu_utilization_per_cluster",

	"runai_pending_requested_memory_bytes_per_workload",
	"runai_pending_requested_memory_bytes_per_queue",
	// "runai_pending_requested_memory_bytes_per_project",
	// "runai_pending_requested_memory_bytes_per_department",
	"runai_pending_requested_memory_bytes_per_nodepool",
	"runai_pending_requested_memory_bytes_per_cluster",

	"runai_allocated_memory_per_pod",
	// "runai_allocated_memory_bytes_per_pod",
	"runai_allocated_memory_per_pod:hourly",
	// "runai_allocated_memory_bytes_per_workload",
	"runai_allocated_memory_bytes_per_queue",
	// "runai_allocated_memory_bytes_per_project",
	// "runai_allocated_memory_bytes_per_department",
	"runai_allocated_memory_bytes_per_node",
	"runai_allocated_memory_bytes_per_nodepool",
	"runai_allocated_memory_bytes_per_cluster",

	"runai_used_memory_bytes_per_pod",
	"runai_used_memory_bytes_per_pod:hourly",
	"runai_used_memory_bytes_per_workload",
	"runai_used_memory_bytes_per_queue",
	// "runai_used_memory_bytes_per_project",
	// "runai_used_memory_bytes_per_department",
	"runai_used_memory_bytes_per_node",
	"runai_used_memory_bytes_per_nodepool",
	"runai_used_memory_bytes_per_cluster",

	"runai_deserved_memory_bytes_per_queue",
	"runai_deserved_memory_bytes_per_project",
	"runai_deserved_memory_bytes_per_department",

	"runai_total_memory_bytes_per_node",
	"runai_total_memory_bytes_per_nodepool",
	"runai_total_memory_bytes_per_cluster",

	"runai_memory_allocation_percentage_per_node",
	"runai_memory_allocation_percentage_per_nodepool",
	"runai_memory_allocation_percentage_per_cluster",

	// "runai_memory_utilization_per_pod",
	// "runai_memory_utilization_per_workload",
	// "runai_memory_utilization_per_queue",
	// "runai_memory_utilization_per_project",
	// "runai_memory_utilization_per_department",
	"runai_memory_utilization_per_node",
	"runai_memory_utilization_per_nodepool",
	"runai_memory_utilization_per_cluster",

	"runai_workload_info_enriched",
	"runai_workload_count_per_type_per_phase_per_queue",
	// "runai_workload_count_per_type_per_phase_per_project",
	// "runai_workload_count_per_type_per_phase_per_department",
	"runai_workload_count_per_type_per_phase_per_nodepool",
	"runai_workload_count_per_type_per_phase_per_cluster",

	"runai_workload_count_per_category",
	"runai_workload_count_per_project",
	"runai_workload_count_per_department",
	"runai_workload_count_per_nodepool",
	"runai_workload_count_per_cluster",

	"runai_node_memory_mem_available_bytes",
	"runai_node_memory_mem_total_bytes",

	"runai_gpu_oomkill_burst_count_per_gpu",
	"runai_gpu_oomkill_burst_count_per_node",
	"runai_gpu_oomkill_burst_count_per_nodepool",
	"runai_gpu_oomkill_idle_count_per_gpu",
	"runai_gpu_oomkill_idle_count_per_node",
	"runai_gpu_oomkill_idle_count_per_nodepool",
	"runai_gpu_oomkill_priority_count_per_gpu",
	"runai_gpu_oomkill_priority_count_per_node",
	"runai_gpu_oomkill_priority_count_per_nodepool",
	"runai_gpu_oomkill_swap_out_of_ram_count_per_gpu",
	"runai_gpu_oomkill_swap_out_of_ram_count_per_node",
	"runai_gpu_oomkill_swap_out_of_ram_count_per_nodepool",
	"runai_swap_memory_bytes_per_pod_per_gpu",

	"runai_dcgm_nvlink_bandwidth_total_bytes",
	"runai_dcgm_nvlink_bandwidth_total_bytes_per_whole_pod_per_gpu",
	"runai_dcgm_nvlink_bandwidth_total_bytes_per_mig_pod",
	"runai_dcgm_nvlink_bandwidth_total_bytes_per_pod_per_gpu",
	"runai_dcgm_nvlink_bandwidth_total_bytes_per_pod",
	"runai_dcgm_nvlink_bandwidth_total_bytes_per_workload",
	"runai_dcgm_nvlink_bandwidth_total_bytes_per_gpu",
	"runai_dcgm_nvlink_bandwidth_total_bytes_per_node",
}

var advancedGpuMetrics = []string{
	"runai_gpu_gr_engine_active_per_pod_per_gpu",
	"runai_gpu_gr_engine_active_per_workload",
	"runai_gpu_gr_engine_active_per_gpu",
	"runai_gpu_gr_engine_active_per_node",
	"runai_gpu_dram_active_per_pod_per_gpu",
	"runai_gpu_dram_active_per_workload",
	"runai_gpu_dram_active_per_gpu",
	"runai_gpu_dram_active_per_node",
	"runai_gpu_sm_active_per_pod_per_gpu",
	"runai_gpu_sm_active_per_workload",
	"runai_gpu_sm_active_per_gpu",
	"runai_gpu_sm_active_per_node",
	"runai_gpu_sm_occupancy_per_pod_per_gpu",
	"runai_gpu_sm_occupancy_per_workload",
	"runai_gpu_sm_occupancy_per_gpu",
	"runai_gpu_sm_occupancy_per_node",
	"runai_gpu_pipe_tensor_active_per_pod_per_gpu",
	"runai_gpu_pipe_tensor_active_per_workload",
	"runai_gpu_pipe_tensor_active_per_gpu",
	"runai_gpu_pipe_tensor_active_per_node",
	"runai_gpu_pipe_fp64_active_per_pod_per_gpu",
	"runai_gpu_pipe_fp64_active_per_workload",
	"runai_gpu_pipe_fp64_active_per_gpu",
	"runai_gpu_pipe_fp64_active_per_node",
	"runai_gpu_pipe_fp32_active_per_pod_per_gpu",
	"runai_gpu_pipe_fp32_active_per_workload",
	"runai_gpu_pipe_fp32_active_per_gpu",
	"runai_gpu_pipe_fp32_active_per_node",
	"runai_gpu_pipe_fp16_active_per_pod_per_gpu",
	"runai_gpu_pipe_fp16_active_per_workload",
	"runai_gpu_pipe_fp16_active_per_gpu",
	"runai_gpu_pipe_fp16_active_per_node",
	"runai_gpu_nvlink_tx_bytes_per_pod_per_gpu",
	"runai_gpu_nvlink_tx_bytes_per_workload",
	"runai_gpu_nvlink_tx_bytes_per_gpu",
	"runai_gpu_nvlink_tx_bytes_per_node",
	"runai_gpu_nvlink_rx_bytes_per_pod_per_gpu",
	"runai_gpu_nvlink_rx_bytes_per_workload",
	"runai_gpu_nvlink_rx_bytes_per_gpu",
	"runai_gpu_nvlink_rx_bytes_per_node",
	"runai_gpu_pcie_tx_bytes_per_pod_per_gpu",
	"runai_gpu_pcie_tx_bytes_per_workload",
	"runai_gpu_pcie_tx_bytes_per_gpu",
	"runai_gpu_pcie_tx_bytes_per_node",
	"runai_gpu_pcie_rx_bytes_per_pod_per_gpu",
	"runai_gpu_pcie_rx_bytes_per_workload",
	"runai_gpu_pcie_rx_bytes_per_gpu",
	"runai_gpu_pcie_rx_bytes_per_node",
}

var monitoringMetrics = []string{
	"runai_e2e_scheduling_latency_milliseconds",
	"runai_plugin_scheduling_latency_microseconds",
	"runai_action_scheduling_latency_microseconds",
	"runai_task_scheduling_latency_microseconds",
	"runai_total_preemption_attempts",
	"node_memory_MemFree_bytes",
	"machine_cpu_cores",
}

var metricsToDropForFinishedJobs = []string{
	"runai_pod_group_phase_with_info",
	"runai_allocated_gpus",
	"runai_requested_gpu_memory_mb_per_workload",
	"runai_requested_gpus_per_workload",
	"runai_cpu_limits_per_active_workload",
	"runai_memory_limits_per_active_workload",
	"runai_gpus_limit_per_workload",
	"runai_gpu_memory_mb_limit_per_workload",
	"runai_workload_info_enriched",
	"runai_run_time_seconds_per_workload",
	"runai_wait_time_seconds_per_workload",
}

var runaiNamespaceExportedMetrics = []string{
	"up",
	"kube_pod_container_resource_limits",
	"kube_pod_container_resource_requests",
	"container_cpu_usage_seconds_total",
	"container_memory_usage_bytes",
	"container_network_transmit_bytes_total",
	"container_network_receive_bytes_total",
	"kube_pod_container_status_restarts_total",
	"kube_pod_container_status_ready",
}

var allNamespaceExportedMetrics = slices.Concat(
	genericMetrics,
	newGenMetrics,
	advancedGpuMetrics,
	monitoringMetrics,
)

var finishedJobStatuses = []string{
	"Succeeded",
	"Failed",
	"Deleted",
	"Timeout",
}

type ClusterRemoteWriteRelabelConfigs struct {
	internalSpec *runv1_internal.Internal
}

func newDefaultClusterRemoteWriteRelabelConfigs(internalSpec *runv1_internal.Internal) RemoteWriteRelabelConfigs {
	return &ClusterRemoteWriteRelabelConfigs{
		internalSpec: internalSpec,
	}
}

func (d *ClusterRemoteWriteRelabelConfigs) Get() []monitoringv1.RelabelConfig {
	retval := []monitoringv1.RelabelConfig{
		{
			SourceLabels: []monitoringv1.LabelName{"__name__", "namespace"},
			Regex:        d.getRelabelConfigRegex(),
			Action:       "keep",
		},
		{
			TargetLabel: "clusterId",
			Replacement: ptr.To(d.internalSpec.Global.ControlPlane.ClusterUUID),
			Action:      "replace",
		},
		{
			SourceLabels: []monitoringv1.LabelName{"__name__", "detailed_status"},
			Regex:        d.getMetricsToDropToFinishedJobs(),
			Action:       "drop",
		},
	}

	if d.internalSpec.Global.Platform.KubernetesFlavor == runv1_internal.KubernetesFlavorOpenShift {
		// In openshift we use deduplication in the prometheus remote write configuration,
		// for this, we add the runai_replica label to the remote write
		retval = append(retval, monitoringv1.RelabelConfig{
			TargetLabel: "runai_replica",
			Replacement: ptr.To("A"),
			Action:      "replace",
		})
	}
	return retval
}

func (d *ClusterRemoteWriteRelabelConfigs) GetRemoteWriteMatchingLabel() string {
	return ClusterMetricsRemoteWriteLabel
}

func (d *ClusterRemoteWriteRelabelConfigs) getRelabelConfigRegex() string {
	return fmt.Sprintf("(%s);.*|(%s);runai", strings.Join(allNamespaceExportedMetrics, "|"), strings.Join(runaiNamespaceExportedMetrics, "|"))
}

func (d *ClusterRemoteWriteRelabelConfigs) getMetricsToDropToFinishedJobs() string {
	return fmt.Sprintf("(%s);(%s)", strings.Join(metricsToDropForFinishedJobs, "|"), strings.Join(finishedJobStatuses, "|"))
}

func (d *ClusterRemoteWriteRelabelConfigs) getClusterId() string {
	return d.internalSpec.Global.ControlPlane.ClusterUUID
}
