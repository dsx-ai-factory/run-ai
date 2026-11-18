/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { versionUtils } from '.';
import { flowRootState } from './config/config';
import { logger } from './logger/logger';
import { SupportedVersions } from './utils/SupporetedEnvsUtils';

export const metricName = {
    gpuAllocation: 'gpuAllocation',

    gpuComputeUtilization: 'gpuComputeUtilization',

    gpuMemoryAllocation: 'gpuMemoryAllocation',

    gpuMemoryUtilization: 'gpuMemoryUtilization',

    cpuAllocation: 'cpuAllocation',

    overviewTotalCPUs: 'overviewTotalCPUs',

    overviewAllocationCPUs: 'overviewAllocationCPUs',

    cpuComputeUtilization: 'cpuComputeUtilization',

    cpuMemoryAllocation: 'cpuMemoryAllocation',

    cpuMemoryUtilization: 'cpuMemoryUtilization',

    gpuComputeUtilizationPerProject: 'gpuComputeUtilizationPerProject',

    allocatedGPUsPerProject: 'allocatedGPUsPerProject',

    allocatedCPUsPerProject: 'allocatedCPUsPerProject',

    overviewAllocatedGPUsPerProject: 'overviewAllocatedGPUsPerProject',

    overviewUnallocatedGpusPerProject: 'overviewUnallocatedGpusPerProject',

    overviewAllocatedCPUsPerProject: 'overviewAllocatedCPUsPerProject',

    overviewDeservedCPUsPerProject: 'overviewDeservedCPUsPerProject',

    overviewTotalGPUs: 'overviewTotalGPUs',

    basicGpuUtilization: 'basicGpuUtilization',

    overviewAllocatedGpus: 'overviewAllocatedGpus',

    idleAllocatedGpus: 'idleAllocatedGpus',

    basicRunningWorkloads: 'basicRunningWorkloads',

    basicPendingWorkloads: 'basicPendingWorkloads',

    jobsGpuUtilization: 'jobsGpuUtilization',

    jobsUsedGpuMemory: 'jobsUsedGpuMemory',

    jobsCpuMemoryUtilization: 'jobsCpuMemoryUtilization',

    jobsCpuUtilization: 'jobsCpuUtilization',

    JobsUsedCpus: 'JobsUsedCpus',

    jobsGPUTabUtilization: 'jobsGPUTabUtilization',

    nodeUsedGpuMemory: 'nodeUsedGpuMemory',

    nodeUsedCpuMemory: 'nodeUsedCpuMemory',

    nodeUsedCpus: 'nodeUsedCpus',

    nodeTotalCpus: 'nodeTotalCpus',

    nodeUtilization: 'nodeUtilization',

    nodeTotalGpus: 'nodeTotalGpus',

    nodeAllocatedGpus: 'nodeAllocatedGpus',

    nodeReady: 'nodeReady',

    nodePool: 'nodePool',

    nodeGpuIdleTimeOnGpu: 'nodeGpuIdleTimeOnGpu',

    nodeTotalGpuMemoryOnGpu: 'nodeTotalGpuMemoryOnGpu',

    nodeUsedGpuMemoryOnGpu: 'nodeUsedGpuMemoryOnGpu',

    nodeUtilizationOnGpu: 'nodeUtilizationOnGpu',

    runningPodsWithError: 'runningPodsWithError',

    dashboardPendingJobs: 'dashboardPendingJobs',

    dashboardActiveProjectsPendingWorkloads: 'dashboardActiveProjectsPendingWorkloads',

    dashboardQueueingPendingWorkloads: 'dashboardQueueingPendingWorkloads',

    analyticsRequestedGpuOfPendingWorkload: 'analyticsRequestedGpuOfPendingWorkload',

    analyticsTotalRequestedGpuOfPendingWorkload: 'analyticsTotalRequestedGpuOfPendingWorkload',

    analyticsPendingWorkloadsCountPerProject: 'analyticsPendingWorkloadsCountPerProject',

    analyticsTotalPendingWorkloadsCountPerProject: 'analyticsTotalPendingWorkloadsCountPerProject',

    consumptionReportOverQuotaGPU: 'consumptionReportOverQuotaGPU',

    consumptionReportInQuotaGPU: 'consumptionReportInQuotaGPU',

    consumptionReportOverQuotaGPUPerProject: 'consumptionReportOverQuotaGPUPerProject',

    consumptionReportInQuotaGPUPerProject: 'consumptionReportInQuotaGPUPerProject',

    GpuUtilizationPerProject: 'GpuUtilizationPerProject',

    GpuUtilizationPerQueue: 'GpuUtilizationPerQueue',

    TotalGpuNodes: 'TotalGpuNodes'
};

// metricQueriesV1 is the queries for the old version 2.9 and below
export const metricQueriesV1 = {
    gpuAllocation:
        '100*(sum(max(runai_allocated_gpus) by (pod_group_uuid)) or vector(0))/clamp_min((count(group(runai_node_gpu_utilization) by(node,gpu)) or vector(0)) + (sum(runai_mig_mode_gpu_count) or vector(0)), 1)',

    gpuComputeUtilization: 'avg(runai_node_gpu_utilization) or vector(0)',

    gpuMemoryAllocation:
        '100*sum((sum(max(runai_allocated_gpus!=0) by (pod_group_uuid) * on (pod_group_uuid) group_left(node) label_replace(runai_pod_group_phase_with_info{phase=~"Running",detailed_status=~"^$|Running"}==1,"node","$1","node_name", "(.+)")) by (node)) / on(node) group_left() (count(runai_node_gpu_last_not_idle_time) by (node)) * on (node) group_left() sum(runai_node_gpu_total_memory) by (node)) / sum(runai_node_gpu_total_memory>0) or vector(0)',

    gpuMemoryUtilization:
        '100 * sum(runai_node_gpu_used_memory) / sum(runai_node_gpu_total_memory>0) or vector(0)',

    cpuAllocation:
        '100 *sum(runai_node_cpu_requested) / sum(kube_node_status_capacity{resource="cpu"}>0) or vector(0)',

    cpuComputeUtilization: '100 * avg(runai_cpu_utilization) or vector(0)',

    cpuMemoryAllocation:
        '100 * sum(runai_node_requested_memory_bytes) / sum(kube_node_status_capacity{resource="memory"}>0) or vector(0)',

    cpuMemoryUtilization: '100 * avg(runai_memory_utilization) or vector(0)',

    gpuComputeUtilizationPerProject: 'avg(runai_node_gpu_utilization) or vector(0)',

    allocatedGPUsPerProject: 'sum(max(runai_allocated_gpus) by(pod_group_uuid))',

    overviewAllocatedGPUsPerProject:
        'sum(max(runai_allocated_gpus > 0) by (pod_group_uuid, project)) by (project)',

    overviewUnallocatedGpusPerProject:
        '((count(group(runai_node_gpu_utilization) by(node,gpu)) or vector(0)) + (sum(runai_mig_mode_gpu_count) or vector(0))) - (sum(max(runai_allocated_gpus) by (pod_group_uuid)) or vector(0))',

    overviewTotalGPUs:
        '(count(runai_node_gpu_utilization{}) or vector(0)) + (sum(sum(runai_mig_mode_gpu_count{}) by(node)) without (node) or vector(0))',

    basicGpuUtilization: 'runai_node_gpu_utilization',

    overviewAllocatedGpus: 'runai_allocated_gpus',

    basicRunningWorkloads: 'runai_pod_group_phase_with_info{phase=~"Running"}==1',

    basicPendingWorkloads: 'runai_pod_group_phase_with_info{phase=~"Pending"}==1',

    jobsGpuUtilization: '(runai_utilization_full_gpu_jobs) or (runai_utilization_shared_gpu_jobs)',

    jobsUsedGpuMemory: 'runai_pod_group_used_gpu_memory * 1024 * 1024',

    jobsCpuMemoryUtilization: 'runai_job_memory_used_bytes',

    jobsCpuUtilization: 'runai_job_cpu_usage',

    jobsGPUTabUtilization: `sum by (pod_group_uuid) (runai_pod_group_gpu_utilization)
    / on (pod_group_uuid)
    (count(runai_pod_group_gpu_utilization) by (pod_group_uuid)
    or
    sum(runai_utilization_shared_gpu_jobs) by (pod_group_uuid))`,

    nodeUsedGpuMemory: '(sum(runai_node_gpu_used_memory * 1024 * 1024) by (node))',

    nodeUsedCpuMemory: 'runai_node_memory_used_bytes',

    nodeUsedCpus: 'avg(runai_node_cpu_utilization) by(node) * 100',

    nodeUtilization:
        '((sum(runai_node_gpu_utilization) by (node)) / on (node) (count(runai_node_gpu_utilization) by (node)))',

    nodeTotalGpus: '(count(runai_gpus_is_running_with_pod2) by (node))',

    nodeAllocatedGpus:
        '((sum(runai_gpus_is_running_with_pod2) by (node))) + (sum(runai_used_shared_gpu_per_node) by (node))',

    nodeReady: 'sum(kube_node_status_condition{condition="Ready",status="true"}) by (node)',

    nodePool: 'runai_node_nodepool==1',

    nodeGpuIdleTimeOnGpu: '(sum(time()-runai_node_gpu_last_not_idle_time) by (node, gpu))',

    nodeTotalGpuMemoryOnGpu: '(sum(runai_node_gpu_total_memory_excluding_mig * 1024 * 1024) by (node, gpu))',

    nodeUsedGpuMemoryOnGpu: '(sum(runai_node_gpu_used_memory_excluding_mig * 1024 * 1024) by (node, gpu))',

    nodeUtilizationOnGpu: '(sum(runai_node_gpu_utilization) by (node, gpu))',

    runningPodsWithError:
        'count(runai_pod_group_phase_with_info{detailed_status=~".*(Err|Error|Fail|BackOff|ContainerCannotRun).*"}==1) or vector(0)',

    dashboardPendingJobs:
        'sum (runai_pod_group_phase_with_info{phase="Pending", workload_type!~"Inference|inference|"}==1) or vector(0)',

    dashboardActiveProjectsPendingWorkloads:
        'count(runai_pod_group_phase_with_info{phase=~"Pending", workload_type!~"Inference|inference|"}==1) by (queue_name) or 0*(sum(runai_queue_deserved_gpus * on (queue_name) group_left() max(runai_pod_group_phase_with_info{phase=~"Running|Pending"}==1) by (queue_name)) by (queue_name))',
    dashboardQueueingPendingWorkloads:
        'sum(runai_pod_group_phase_with_info{phase="Pending", workload_type!~"Inference|inference|"}==1) by (pod_group_uuid, workload_type)',

    analyticsRequestedGpuOfPendingWorkload:
        '(sum(max(runai_requested_gpus) by (pod_group_uuid) * on (pod_group_uuid) group_left (queue_name) (runai_pod_group_phase_with_info{phase="Pending",workload_type!~"Inference|inference|"}==1)) by (queue_name))',

    analyticsTotalRequestedGpuOfPendingWorkload:
        '(sum(max(runai_requested_gpus) by (pod_group_uuid) * on (pod_group_uuid) group_left (queue_name) (runai_pod_group_phase_with_info{phase="Pending",workload_type!~"Inference|inference|"}==1))) or vector(0)',

    analyticsPendingWorkloadsCountPerProject:
        'count(runai_pod_group_phase_with_info{phase="Pending",workload_type!~"Inference|inference|"}==1) by (queue_name)',

    analyticsTotalPendingWorkloadsCountPerProject:
        'count(runai_pod_group_phase_with_info{phase="Pending",workload_type!~"Inference|inference|"}==1) or vector(0)'
};

// metricQueriesV2 is the new version of the queries, it is used for the new dashboard 2.9.0 and above
export const metricQueriesV2 = {
    gpuAllocation: '(sum(runai_dashboard:analytics:gpu_allocation)) ',

    gpuComputeUtilization: '(sum(runai_dashboard:analytics:gpu_compute_utilization))',

    gpuMemoryAllocation: '(sum(runai_dashboard:analytics:gpu_memory_allocation))',

    gpuMemoryUtilization: '(sum(runai_dashboard:analytics:gpu_memory_utilization))',

    cpuAllocation:
        '100 *sum(runai_node_cpu_requested) / sum(kube_node_status_capacity{resource="cpu"}>0) or vector(0)',

    overviewTotalCPUs:
        'sum(runai_cpu_cores_count_per_node{modelName!="", node!="runai-cluster-control-plane"}) or vector(0)',

    TotalGpuNodes:
        'count(runai_cpu_cores_count_per_node{modelName!="", node!="runai-cluster-control-plane"}) or vector(0)',

    overviewAllocationCPUs:
        'sum(runai_allocated_cpu_cores_per_node{modelName!="", node!="runai-cluster-control-plane"}) ' +
        'or vector(0)',

    cpuComputeUtilization: '100 * avg(runai_cpu_utilization) or vector(0)',

    cpuMemoryAllocation:
        '100 * sum(runai_node_requested_memory_bytes) / sum(kube_node_status_capacity{resource="memory"}>0) or vector(0)',

    cpuMemoryUtilization: '100 * avg(runai_memory_utilization) or vector(0)',

    gpuComputeUtilizationPerProject:
        'sum(runai_dashboard:analytics:gpu_compute_utilization_per_project:project) by(project)',

    allocatedGPUsPerProject: '(sum(runai_dashboard:analytics:allocated_gpus_per_project:total_allocated))',

    overviewAllocatedGPUsPerProject: 'runai_dashboard:overview:allocated_gpus_per_project:project',

    overviewUnallocatedGpusPerProject: 'runai_dashboard:overview:allocated_gpus_per_project:unallocated',

    overviewAllocatedCPUsPerProject: 'runai_dashboard:overview:allocated_cpus_per_project:project',

    overviewDeservedCPUsPerProject: 'runai_dashboard:overview:deserved_cpu_cores_per_project:project',

    nodeTotalCpus: 'sum(runai_cpu_cores_count_per_node)',

    overviewTotalGPUs: 'sum(runai_total_gpus_per_nodepool)',

    basicGpuUtilization: 'runai_node_gpu_utilization',

    overviewAllocatedGpus: 'runai_dashboard:overview:allocated_gpus',

    basicRunningWorkloads: 'runai_pod_group_phase_with_info{phase=~"Running"}==1',

    basicPendingWorkloads: 'runai_pod_group_phase_with_info{phase=~"Pending"}==1',

    jobsGpuUtilization: 'runai_gpu_utilization_per_workload',

    jobsUsedGpuMemory: 'runai_gpu_memory_used_mebibytes_per_workload * 1024 * 1024',

    jobsCpuMemoryUtilization: 'runai_job_memory_used_bytes',

    jobsCpuUtilization: 'runai_job_cpu_usage',

    jobsGPUTabUtilization: '(avg(runai_gpu_utilization_per_pod_per_gpu) by(pod_group_uuid))',

    nodeUsedGpuMemory: '(sum(runai_gpu_memory_used_mebibytes_per_node * 1024 * 1024) by (node))',

    nodeUsedCpuMemory: 'runai_node_memory_used_bytes',

    nodeUsedCpus: 'avg(runai_node_cpu_utilization) by(node) * 100',

    JobsUsedCpus: 'sum(runai_cpu_utilization_per_workload) by(workload_name)',

    nodeUtilization:
        '((sum(runai_node_gpu_utilization) by (node)) / on (node) (count(runai_node_gpu_utilization) by (node)))',

    nodeTotalGpus: '(sum(runai_gpu_count_per_node) by (node))',

    nodeAllocatedGpus: 'sum(runai_allocated_gpu_count_per_node) by (node)',

    nodeReady: 'sum(kube_node_status_condition{condition="Ready",status="true"}) by (node)',

    nodePool: 'runai_node_nodepool==1',

    nodeGpuIdleTimeOnGpu: '(sum(time()-runai_last_gpu_utilization_time_per_gpu) by (node, gpu))',

    nodeTotalGpuMemoryOnGpu: '(sum(runai_gpu_memory_total_mebibytes_per_node * 1024 * 1024) by (node, gpu))',

    nodeUsedGpuMemoryOnGpu: '(sum(runai_gpu_memory_used_mebibytes_per_node * 1024 * 1024) by (node, gpu))',

    nodeUtilizationOnGpu: '(sum(runai_gpu_utilization_per_node) by (node, gpu))',

    runningPodsWithError:
        'count(runai_pod_group_phase_with_info{detailed_status=~".*(Err|Error|Fail|BackOff|ContainerCannotRun).*"}==1) or vector(0)',

    dashboardPendingJobs:
        'sum(runai_pod_group_phase_with_info{phase="Pending", workload_type!~"Inference|inference|"}==1) by (pod_group_uuid, workload_name)',

    dashboardActiveProjectsPendingWorkloads:
        'count(runai_pod_group_phase_with_info{phase=~"Pending", workload_type!~"Inference|inference|"}==1) by (project) or 0*(sum(runai_queue_deserved_gpus * on (project) group_left() max(runai_pod_group_phase_with_info{phase=~"Running|Pending"}==1) by (project)) by (project))',

    dashboardQueueingPendingWorkloads:
        'sum(runai_pod_group_phase_with_info{phase="Pending", workload_type!~"Inference|inference|"}==1) by (pod_group_uuid, workload_type)',

    analyticsRequestedGpuOfPendingWorkload:
        '(sum(max(runai_requested_gpus) by (pod_group_uuid) * on (pod_group_uuid) group_left (queue_name) (runai_pod_group_phase_with_info{phase="Pending",workload_type!~"Inference|inference|"}==1)) by (queue_name))',

    analyticsTotalRequestedGpuOfPendingWorkload:
        '(sum(max(runai_requested_gpus) by (pod_group_uuid) * on (pod_group_uuid) group_left (queue_name) (runai_pod_group_phase_with_info{phase="Pending",workload_type!~"Inference|inference|"}==1))) or vector(0)',

    analyticsPendingWorkloadsCountPerProject:
        'count(runai_pod_group_phase_with_info{phase="Pending",workload_type!~"Inference|inference|"}==1) by (queue_name)',

    analyticsTotalPendingWorkloadsCountPerProject:
        'count(runai_pod_group_phase_with_info{phase="Pending",workload_type!~"Inference|inference|"}==1) or vector(0)',

    consumptionReportOverQuotaGPU:
        'sum(((sum (runai_allocated_gpu_count_per_pod) by (project) - sum(runai_project_guaranteed_gpus) by (project)) > 0))',

    consumptionReportInQuotaGPU:
        'sum(sum(runai_allocated_gpu_count_per_pod) by (project) - clamp_min(sum (runai_allocated_gpu_count_per_pod) by (project) - sum(runai_project_guaranteed_gpus) by (project), 0))',

    consumptionReportOverQuotaGPUPerProject:
        'sum(((sum (runai_allocated_gpu_count_per_pod) by (project) - sum(runai_project_guaranteed_gpus) by (project)) > 0)) by (project)',

    consumptionReportInQuotaGPUPerProject:
        'sum(sum(runai_allocated_gpu_count_per_pod) by (project) - clamp_min(sum (runai_allocated_gpu_count_per_pod) by (project) - sum(runai_project_guaranteed_gpus) by (project), 0)) by (project)',

    GpuUtilizationPerProject: 'runai_gpu_utilization_per_project',

    GpuUtilizationPerQueue: 'runai_gpu_utilization_per_queue'
};

// metricQueriesV3 is the new version of the queries, it is used for the new dashboard 2.12.0 and above
export const metricQueriesV3 = Object.assign({}, metricQueriesV2);
metricQueriesV3.overviewAllocatedGpus = 'sum(runai_allocated_gpu_count_per_nodepool)';
// it's 2 and not 300 here for test purposes
metricQueriesV3['idleAllocatedGpus'] =
    '(sum((runai_gpu_idle_seconds_per_workload > 2)*0 + on(job_uuid, nodepool) runai_allocated_gpu_count_per_workload))';
metricQueriesV3.overviewAllocatedGPUsPerProject =
    'sum(runai_allocated_gpu_count_per_project_nodepool>0) by (project)';
metricQueriesV3.overviewUnallocatedGpusPerProject =
    'sum(runai_gpu_count_per_node) - sum(runai_allocated_gpu_count_per_project_nodepool) or vector(0)';
metricQueriesV3.overviewAllocatedCPUsPerProject =
    'sum(runai_allocated_cpu_cores_per_queue{project!=""}) by (project)';
metricQueriesV3.overviewDeservedCPUsPerProject = 'sum(runai_deserved_cpu_cores_per_queue>0) by (project)';
metricQueriesV3.nodeTotalCpus = 'sum(runai_cpu_cores_count_per_node) or vector(0)';

// metricQueriesV4 is the new version of the queries, it is used for the new dashboard 2.17.0 and above
export const metricQueriesV4 = Object.assign({}, metricQueriesV3);
metricQueriesV4.analyticsRequestedGpuOfPendingWorkload =
    '(sum(max(runai_requested_gpus_per_workload) by (workload_id) * on (workload_id) group_left (queue_name) (runai_workload_info_enriched{phase="Pending",workload_type!~"Inference|inference|"}==1)) by (queue_name))';
metricQueriesV4.analyticsTotalRequestedGpuOfPendingWorkload =
    '(sum(max(runai_requested_gpus_per_workload) by (workload_id) * on (workload_id) group_left (queue_name) (runai_workload_info_enriched{phase="Pending",workload_type!~"Inference|inference|"}==1))) or vector(0)';
metricQueriesV4.basicRunningWorkloads = 'runai_workload_info_enriched{phase=~"Running"}==1';
metricQueriesV4.basicPendingWorkloads = 'runai_workload_info_enriched{phase=~"Pending"}==1';
metricQueriesV4.runningPodsWithError =
    'count(runai_workload_info_enriched{detailed_status=~".*(Err|Error|Fail|BackOff|ContainerCannotRun).*"}==1) or vector(0)';
metricQueriesV4.dashboardPendingJobs =
    'sum(runai_workload_info_enriched{phase="Pending", workload_type!~"Inference|inference|"}==1) by (workload_id, workload_name)';
metricQueriesV4.dashboardActiveProjectsPendingWorkloads =
    'count(runai_workload_info_enriched{phase=~"Pending", workload_type!~"Inference|inference|"}==1) by (project) or 0*(sum(runai_queue_deserved_gpus * on (project) group_left() max(runai_workload_info_enriched{phase=~"Running|Pending"}==1) by (project)) by (project))';
metricQueriesV4.dashboardQueueingPendingWorkloads =
    'sum(runai_workload_info_enriched{phase="Pending", workload_type!~"Inference|inference|"}==1) by (workload_id, workload_type)';
metricQueriesV4.analyticsPendingWorkloadsCountPerProject =
    'count(runai_workload_info_enriched{phase="Pending",workload_type!~"Inference|inference|"}==1) by (queue_name)';
metricQueriesV4.analyticsTotalPendingWorkloadsCountPerProject =
    'count(runai_workload_info_enriched{phase="Pending",workload_type!~"Inference|inference|"}==1) or vector(0)';
metricQueriesV4.jobsGPUTabUtilization = '(avg(runai_gpu_utilization_per_pod_per_gpu) by(workload_id))';
metricQueriesV4['idleAllocatedGpus'] =
    '(sum((runai_gpu_idle_seconds_per_workload > 2)*0 + on(workload_id, nodepool) runai_allocated_gpu_count_per_workload))';

export function getAnalyticsQueries(metric: string): string {
    if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_17)) {
        logger.debug(
            `getAnalyticsQueries cluster version is: ${flowRootState.cluster.version}, return metricQueriesV4 (new dashboard)`
        );
        return metricQueriesV4[metric];
    }

    if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_13)) {
        logger.debug(
            `getAnalyticsQueries cluster version is: ${flowRootState.cluster.version}, return metricQueriesV3 (mid dashboard)`
        );
        return metricQueriesV3[metric];
    }

    if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_9)) {
        logger.debug(
            `getAnalyticsQueries cluster version is: ${flowRootState.cluster.version}, return metricQueriesV2 (mid dashboard)`
        );
        return metricQueriesV2[metric];
    }

    logger.debug(
        `getAnalyticsQueries cluster version is: ${flowRootState.cluster.version}, return metricQueriesV1 (old dashboard)`
    );
    return metricQueriesV1[metric];
}
