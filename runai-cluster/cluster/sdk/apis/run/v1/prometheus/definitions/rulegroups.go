package definitions

import (
	"fmt"
	"slices"
	"strings"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

var podLabels = strings.Join([]string{
	"node", "nodepool", "project", "department", "queue_name",
	"pod_name", "pod_uuid", "pod_namespace",
	"workload_name", "workload_type", "workload_id",
	"gpu", "modelName", "service"}, ", ")

// podLabelsWithoutPodKeys same as podLabels without (no pod_name/pod_namespace)
var podLabelsWithoutPodKeys = strings.Join([]string{
	"node", "nodepool", "project", "department", "queue_name", "pod_uuid",
	"workload_name", "workload_type", "workload_id",
	"gpu", "modelName", "service",
}, ", ")

var podLabelsWithoutGpu = strings.Join([]string{
	"node", "nodepool", "project", "department", "queue_name",
	"pod_name", "pod_namespace",
	"workload_name", "workload_type", "workload_id",
	"modelName"}, ", ")
var podLabelsForAmdMetric = "workload_name, workload_type, workload_id, pod_name, pod_namespace, project, department, queue_name"

var workloadLabels = strings.Join([]string{
	"nodepool", "project", "department", "queue_name",
	"workload_name", "workload_type", "workload_id",
	"user", "is_gpu",
	"phase", "detailed_status", "status", "workload_category"}, ", ")

// should be same for all pods in a workload - phase/status could be different...
var workloadPodLabels = strings.Join([]string{
	"nodepool", "project", "department", "queue_name",
	"workload_name", "workload_type", "workload_id",
	"user"}, ", ")

var workloadLabelsNoWorkloadId = strings.Join([]string{"nodepool", "project", "department", "queue_name", "workload_name",
	"workload_type", "user", "is_gpu", "phase", "detailed_status", "status", "workload_category"}, ",")

var queueLabels = strings.Join([]string{"queue_name", "project", "department", "nodepool"}, ", ")
var projectLabels = strings.Join([]string{"project", "department"}, ", ")
var departmentLabels = strings.Join([]string{"department"}, ", ")
var nodepoolLabels = strings.Join([]string{"nodepool"}, ", ")
var nodeLabels = strings.Join([]string{"node", "nodepool", "modelName", "ready", "schedulable"}, ", ")
var gpuLabels = strings.Join([]string{"UUID", "gpu", "node", "nodepool", "modelName"}, ", ")

var runaiJobCpuUsageLabels = "workload_name, workload_id, pod_namespace, project"
var workloadNameUidLabels = "workload_name, workload_id"
var metricsPerMigPodRuleLabels = "pod_name, pod_namespace, workload_name, workload_type, workload_id, pod_uuid, project, department, queue_name"
var metricsPerWholePodRuleLabels = "workload_name, workload_type, workload_id, pod_uuid, project, department, queue_name"
var allocatedMemoryBytesPerPodMetric = "runai_scheduler_allocated_memory_bytes_per_pod_unique"

var workloadIdLabel = "workload_id"
var podEnrichmentWithWorkloadLabels = "workload_name, workload_type, user"

const migPodFilter = `gpu!="", gpu_instance_id!=""`
const wholePodFilter = `gpu="", gpu_instance_id=""`
const fractionalPodFilter = `gpu!="", gpu_instance_id=""`

type RuleGroupsConfig struct {
	AgentPullPushRateMinForAlert             float32
	ClusterSyncQueueDepthIncreaseMaxForAlert float32
	AdditionalAlertLabels                    map[string]string
}

func DefaultRuleGroups(ruleGroupsConfig *RuleGroupsConfig) []monitoringv1.RuleGroup {
	rgs := []monitoringv1.RuleGroup{}

	// if several metrics are exporting the same metric about the same pod/pod-group
	// (for multi nodepools), reduce it to only 1 metric
	schedulerMetricsReduction := []monitoringv1.Rule{
		{
			Record: "runai_workload_info_unique", // instead of runai_pod_group_info_unique
			Expr:   intstr.FromString(`topk by(workload_id) (1,runai_workload_info)`),
		},
		{
			// fixing the gpu label in runai_pod_gpu_info - gpu label value is actually the UUID so we are taking the UUID from the dcgm metric
			Record: "runai_pod_gpu_info_fixed",
			Expr:   intstr.FromString(`(label_replace(runai_pod_gpu_info{gpu!=""}, "UUID", "$1", "gpu", "(.*)")) * on (UUID) group_left(gpu) (count(runai_dcgm_gpu_utilization) by (UUID, gpu))`),
		},
		{
			Record: "runai_pod_info_unique", // same
			Expr:   intstr.FromString(`topk by (pod_uuid) (1, runai_pod_info)`),
		},
		{
			Record: "runai_scheduler_requested_gpus_per_pod_unique", // instead of runai_requested_gpus but on the pod level
			Expr:   intstr.FromString(`topk by(pod_uuid) (1,runai_scheduler_requested_gpus_per_pod)`),
		},
		{
			Record: "runai_scheduler_requested_gpu_memory_mb_per_pod_unique", // instead of runai_requested_gpus_memory but on the pod level
			Expr:   intstr.FromString(`topk by(pod_uuid) (1,runai_scheduler_requested_gpu_memory_mb_per_pod)`),
		},
		{
			Record: "runai_scheduler_gpus_limit_per_pod_unique", // new metric
			Expr:   intstr.FromString(`topk by(pod_uuid) (1,runai_scheduler_gpus_limit_per_pod)`),
		},
		{
			Record: "runai_scheduler_gpu_memory_mb_limit_per_pod_unique", // new metric
			Expr:   intstr.FromString(`topk by(pod_uuid) (1,runai_scheduler_gpu_memory_mb_limit_per_pod)`),
		},
		{
			Record: "runai_scheduler_allocated_gpus_per_pod_unique", // same
			Expr:   intstr.FromString(`topk by(pod_uuid) (1,runai_scheduler_allocated_gpus_per_pod)`),
		},
		{
			Record: "runai_scheduler_allocated_gpu_memory_mb_per_pod_unique", // new metric
			Expr:   intstr.FromString(`topk by(pod_uuid) (1,runai_scheduler_allocated_gpu_memory_mb_per_pod)`),
		},
		{
			Record: "runai_scheduler_allocated_millicpus_per_pod_unique", // same
			Expr:   intstr.FromString(`topk by(pod_uuid) (1,runai_scheduler_allocated_millicpus_per_pod)`),
		},
		{
			Record: "runai_scheduler_allocated_memory_bytes_per_pod_unique", // instead of runai_scheduler_allocated_memory_per_pod_unique
			Expr:   intstr.FromString(`topk by(pod_uuid) (1,runai_scheduler_allocated_memory_bytes_per_pod)`),
		},
	}

	workloadNewGenRules := []monitoringv1.Rule{
		// =================================== Info ==========================================
		{
			// adding is_gpu label to runai_workload_info_unique
			Record: "runai_workload_info_enriched",
			Expr:   intstr.FromString(fmt.Sprintf(`group(runai_workload_info_unique * on (workload_id) group_left(is_gpu) (label_replace((sum(runai_scheduler_requested_gpus_per_pod_unique) by (workload_id)) > 0, "is_gpu", "true", "pod_name", "(.*)") or on(workload_id) label_replace((sum(runai_scheduler_requested_gpus_per_pod_unique) by (workload_id)) == 0, "is_gpu", "false", "pod_name", "(.*)"))) by (%s)`, workloadLabels)),
		},
		{
			Record: "runai_pod_info_enriched",
			Expr:   intstr.FromString(fmt.Sprintf(`(runai_pod_info_unique * on (node) group_left(modelName) runai_node_info_enriched or on(pod_uuid) runai_pod_info_unique) * on(workload_id) group_left(%s) runai_workload_info_enriched`, podEnrichmentWithWorkloadLabels)),
		},
		{
			Record: "runai_pod_info_enriched_nodepool_unique",
			Expr:   intstr.FromString(fmt.Sprintf(`(runai_pod_info_enriched * on(%s, nodepool) group_left() topk(1, group(runai_pod_info_enriched) by(nodepool, %s)) by(%s))`, workloadIdLabel, workloadIdLabel, workloadIdLabel)),
		},
		{
			Record: "runai_pod_completion_time",
			Expr:   intstr.FromString(`runai_pod_info_enriched_nodepool_unique  * on(pod_uuid) group_left() label_replace(kube_pod_completion_time, "pod_uuid", "$1", "uid", "(.*)")`),
		},
		{
			Record: "runai_pod_info_enriched_nodepool_unique_with_gpu_for_core_pods",
			Expr:   intstr.FromString(fmt.Sprintf(`runai_pod_info_enriched_nodepool_unique * on(pod_uuid) group_right(%s) runai_pod_gpu_info_fixed{gpu!=""}  or on(pod_uuid) runai_pod_info_enriched_nodepool_unique`, podLabelsWithoutGpu)),
		},
		// ================================== Requested GPUs ===================================
		{
			Record: "runai_pending_requested_gpu_count_per_workload",
			Expr:   intstr.FromString(`sum(runai_scheduler_requested_gpus_per_pod_unique) by(workload_id) * on(workload_id) group_right() runai_workload_info_enriched{detailed_status="Pending"}`),
		},
		// ================================== Allocated GPU Memory ===================================
		// new metrics starting from 2.17 version with the correct workload metrics
		{
			Record: "runai_allocated_gpu_memory_mb_per_pod",
			Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_scheduler_allocated_gpu_memory_mb_per_pod_unique) * on (pod_uuid) group_right() runai_pod_info_enriched_nodepool_unique`, podLabels)),
		},
		{
			Record: "runai_requested_gpu_memory_mb_per_pod",
			Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_scheduler_requested_gpu_memory_mb_per_pod_unique) * on (pod_uuid) group_right() runai_pod_info_enriched_nodepool_unique`, podLabels)),
		},
		{
			Record: "runai_allocated_gpu_memory_mb_per_workload",
			Expr:   intstr.FromString(metricsPerWorkloadRule(`runai_allocated_gpu_memory_mb_per_pod`, false)),
		},
	}

	newGenRules := monitoringv1.RuleGroup{
		Name: "new-gen-rules",
		Rules: []monitoringv1.Rule{

			// =================================== Info ==========================================
			{
				Record: "runai_node_nodepool_excluded",
				Expr:   intstr.FromString(fmt.Sprintf(`group((runai_node_nodepool==1 or on(node) kube_node_info) unless on(nodepool) runai_node_nodepool{nodepool="%s"}==1) by(node, nodepool)`, constants.ExcludedNodepoolName)),
			},
			{
				Record: "runai_node_info_enriched",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (kube_node_info * on(node) group_left(nodepool) runai_node_nodepool_excluded) * on(node) group_left(modelName) (group(runai_gpu_memory_total_mebibytes_per_gpu) by(node, modelName) or on(node) kube_node_info) * on(node) group_left(ready) (group(label_replace(kube_node_status_condition{condition="Ready"}, "ready", "$1", "status", "(.*)")==1) by(node, ready)) * on(node) group_left(schedulable) (group(label_replace(kube_node_spec_unschedulable == 0, "schedulable", "true", "schedulable", "(.*)") or on(node) label_replace(kube_node_spec_unschedulable == 1, "schedulable", "false", "schedulable", "(.*)")) by(node,schedulable))`, nodeLabels)),
			},
			{
				Record: "runai_requested_mig_devices_per_workload",
				Expr:   intstr.FromString(`(sum(runai_requested_mig_devices) by (queue_name,workload_name,workload_type,mig_device_type,project)) * on(workload_name,project) group_left(workload_id) runai_workload_info_enriched`),
			},
			// =====================================================================================

			// ==================================== Total GPUs =====================================
			{
				Record: "runai_gpu_count_per_node",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by (%s) ((count by (node) (runai_gpu_memory_total_mebibytes_per_gpu) or on(node) kube_node_info*0) * on(node) group_right() runai_node_info_enriched)`, nodeLabels)),
			},
			// TODO: remove this, use the next one instead
			{
				Record: "runai_total_gpus_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_gpu_count_per_node) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_gpu_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_gpu_count_per_node) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_ready_gpu_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_gpu_count_per_node{ready="true",schedulable="true"}) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_gpu_count_per_cluster",
				Expr:   intstr.FromString(`sum(runai_gpu_count_per_nodepool)`),
			},
			{
				Record: "runai_ready_gpu_count_per_cluster",
				Expr:   intstr.FromString(`sum(runai_ready_gpu_count_per_nodepool)`),
			},
			// =====================================================================================

			// ================================== Total GPU Nodes ==================================
			{
				Record: "runai_gpu_nodes_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`count(runai_gpu_count_per_node>0) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_ready_gpu_nodes_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`count(runai_gpu_count_per_node{ready="true",schedulable="true"}>0) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_gpu_nodes_count_per_cluster",
				Expr:   intstr.FromString(`sum(runai_gpu_nodes_count_per_nodepool)`),
			},
			{
				Record: "runai_ready_gpu_nodes_count_per_cluster",
				Expr:   intstr.FromString(`sum(runai_ready_gpu_nodes_count_per_nodepool)`),
			},
			// =====================================================================================

			// ================================== Requested GPUs ===================================
			{
				Record: "runai_pending_requested_gpu_count_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_pending_requested_gpu_count_per_workload) by (%s)`, queueLabels)),
			},
			// {
			// 	Record: "runai_pending_requested_gpu_count_per_project",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_pending_requested_gpu_count_per_queue) by (%s)`, projectLabels)),
			// },
			// {
			// 	Record: "runai_pending_requested_gpu_count_per_department",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_pending_requested_gpu_count_per_queue) by (%s)`, departmentLabels)),
			// },
			{
				Record: "runai_pending_requested_gpu_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_pending_requested_gpu_count_per_queue) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_pending_requested_gpu_count_per_cluster",
				Expr:   intstr.FromString(`sum(runai_pending_requested_gpu_count_per_nodepool)`),
			},
			// =====================================================================================

			// ================================== Allocated GPUs ===================================
			{
				Record: "runai_allocated_gpu_count_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_scheduler_allocated_gpus_per_pod_unique) * on (pod_uuid) group_right() runai_pod_info_enriched_nodepool_unique`, podLabels)),
			},
			{
				Record: "runai_allocated_gpu_count_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule(`runai_allocated_gpu_count_per_pod`, false)),
			},
			{
				Record: "runai_allocated_gpu_count_per_queue_per_workload_type",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_gpu_count_per_workload>0) by (workload_type, %s)`, queueLabels)),
			},
			{
				Record: "runai_allocated_gpu_count_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_gpu_count_per_pod) by (%s)`, queueLabels)),
			},
			// TODO: remove this, equals to queue
			{
				Record: "runai_allocated_gpu_count_per_project_per_workload_type",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_gpu_count_per_queue_per_workload_type) by (workload_type, %s)`, projectLabels)),
			},
			{
				Record: "runai_allocated_gpu_count_per_project",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_gpu_count_per_queue) by (%s)`, projectLabels)),
			},
			{
				Record: "runai_allocated_gpu_count_per_department",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_gpu_count_per_queue) by (%s)`, departmentLabels)),
			},
			{
				Record: "runai_allocated_gpu_count_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(((sum((runai_allocated_gpu_count_per_pod / on(pod_uuid) group_left() count(runai_pod_gpu_info_fixed{gpu!=""}) by(pod_uuid)) * on(pod_uuid) group_right(node) runai_pod_gpu_info_fixed{gpu!=""}) by(gpu, node) or sum(runai_allocated_gpu_count_per_pod{gpu!=""}) by(gpu, node) or count(runai_gpu_memory_used_mebibytes_per_pod_per_gpu) by(node, gpu)) * on(gpu, node) group_right() group(runai_gpu_memory_total_mebibytes_per_gpu) by(%s)) or on(UUID) runai_gpu_memory_total_mebibytes_per_gpu*0) by(%s)`, gpuLabels, gpuLabels)),
			},
			{
				Record: "runai_allocated_gpu_count_per_node",
				Expr:   intstr.FromString(`((sum by (node) (runai_allocated_gpu_count_per_pod>0)) * on (node) group_right() runai_node_info_enriched) or on(node) runai_node_info_enriched*0`),
			},
			{
				Record: "runai_allocated_gpu_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_gpu_count_per_node) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_allocated_gpu_count_per_cluster",
				Expr:   intstr.FromString(`sum(runai_allocated_gpu_count_per_nodepool)`),
			},
			// =====================================================================================

			// ================================== Free GPUs ===================================
			{
				// free whole gpus per node
				Record: "runai_free_gpu_count_per_node",
				Expr:   intstr.FromString(`count(runai_allocated_gpu_count_per_gpu == 0) by (clusterId, node, nodepool) or on (clusterId, node, nodepool) 0*group(runai_allocated_gpu_count_per_node > 0) by (clusterId, node, nodepool)`),
			},
			// =====================================================================================

			// ================================== GPU Allocation % =================================
			{
				Record: "runai_gpu_allocation_percentage_per_node",
				Expr:   intstr.FromString(`100*(runai_allocated_gpu_count_per_node / runai_gpu_count_per_node>0)`),
			},
			{
				Record: "runai_gpu_allocation_percentage_per_nodepool",
				Expr:   intstr.FromString(`100*(runai_allocated_gpu_count_per_nodepool / runai_gpu_count_per_nodepool>0)`),
			},
			{
				Record: "runai_gpu_allocation_percentage_per_cluster",
				Expr:   intstr.FromString(`100*(runai_allocated_gpu_count_per_cluster / runai_gpu_count_per_cluster>0)`),
			},
			// =====================================================================================

			// =============================== Deserved GPUs ===============================
			{
				Record: "runai_deserved_gpu_count_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_queue_deserved_gpus)`, queueLabels)),
			},
			{
				Record: "runai_deserved_gpu_count_per_project",
				Expr:   intstr.FromString(fmt.Sprintf(`min(runai_deserved_gpu_count_per_queue{project!=""}<0) by(%s) or on(project) sum(runai_deserved_gpu_count_per_queue{project!=""}>=0) by(%s)`, projectLabels, projectLabels)),
			},
			{
				Record: "runai_deserved_gpu_count_per_department",
				Expr:   intstr.FromString(fmt.Sprintf(`min(runai_deserved_gpu_count_per_queue{project=""}<0) by(%s) or on(department) sum(runai_deserved_gpu_count_per_queue{project=""}>=0) by(%s)`, departmentLabels, departmentLabels)),
			},
			// =====================================================================================

			// ================================== GPU Utilization ==================================
			{
				Record: "runai_gpu_oomkill_burst_count_per_gpu",
				Expr:   intstr.FromString(getOomKillerMetric("runai_gpu_oomkill_burst_count", gpuLabels)),
			}, {
				Record: "runai_gpu_oomkill_burst_count_per_node",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_gpu_oomkill_burst_count_per_gpu)`, nodeLabels)),
			},
			{
				Record: "runai_gpu_oomkill_burst_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_gpu_oomkill_burst_count_per_node)`, nodepoolLabels)),
			},
			{
				Record: "runai_gpu_oomkill_idle_count_per_gpu",
				Expr:   intstr.FromString(getOomKillerMetric("runai_gpu_oomkill_idle_count", gpuLabels)),
			},
			{
				Record: "runai_gpu_oomkill_idle_count_per_node",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_gpu_oomkill_idle_count_per_gpu)`, nodeLabels)),
			},
			{
				Record: "runai_gpu_oomkill_idle_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_gpu_oomkill_idle_count_per_node)`, nodepoolLabels)),
			},
			{
				Record: "runai_gpu_oomkill_priority_count_per_gpu",
				Expr:   intstr.FromString(getOomKillerMetric("runai_gpu_oomkill_priority_count", gpuLabels)),
			},
			{
				Record: "runai_gpu_oomkill_priority_count_per_node",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_gpu_oomkill_priority_count_per_gpu)`, nodeLabels)),
			},
			{
				Record: "runai_gpu_oomkill_priority_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_gpu_oomkill_priority_count_per_node)`, nodepoolLabels)),
			},
			{
				Record: "runai_gpu_oomkill_swap_out_of_ram_count_per_gpu",
				Expr:   intstr.FromString(getOomKillerMetric("runai_gpu_oomkill_swap_out_of_ram_count", gpuLabels)),
			},
			{
				Record: "runai_gpu_oomkill_swap_out_of_ram_count_per_node",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_gpu_oomkill_swap_out_of_ram_count_per_gpu)`, nodeLabels)),
			},
			{
				Record: "runai_gpu_oomkill_swap_out_of_ram_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_gpu_oomkill_swap_out_of_ram_count_per_node)`, nodepoolLabels)),
			},
			{
				Record: "runai_swap_memory_bytes_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerFractionalPodRule("runai_pod_gpu_swap_ram_used_bytes")),
			},
			{
				Record: "runai_dcgm_gpu_utilization",
				Expr:   intstr.FromString(dcgmMetricsRule(`DCGM_FI_DEV_GPU_UTIL or ceil(DCGM_FI_PROF_GR_ENGINE_ACTIVE{GPU_I_ID!=""}*100)`)),
			},
			{
				Record: "runai_amd_gpu_utilization",
				Expr:   intstr.FromString(amdMetricsRule(`amd_gpu_use_percent`)),
			},
			{
				Record: "runai_gpu_utilization_per_fractional_pod",
				Expr:   intstr.FromString(gpuMetricsPerFractionalPodRule(`runai_pod_gpu_utilization`)),
			},
			{
				Record: "runai_gpu_utilization_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule(`runai_dcgm_gpu_utilization`)),
			},
			{
				Record: "runai_gpu_utilization_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`%s or on (pod_uuid) %s`, gpuMetricsPerWholePodRule(`runai_dcgm_gpu_utilization`), gpuMetricsPerWholeAmdPodRule(`runai_amd_gpu_utilization`))),
			},
			{
				Record: "runai_gpu_utilization_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule(`runai_gpu_utilization_per_whole_pod_per_gpu`, `runai_gpu_utilization_per_mig_pod`, `runai_gpu_utilization_per_fractional_pod`)),
			},
			{
				Record: "runai_gpu_utilization_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`avg(runai_gpu_utilization_per_pod_per_gpu) by (%s)`, podLabelsWithoutGpu)),
			},
			{
				Record: "runai_last_gpu_utilization_time_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`timestamp(runai_gpu_utilization_per_pod > 2) or on(%s) (runai_last_gpu_utilization_time_per_pod * on(%s) group_left() clamp(runai_allocated_gpu_count_per_pod>0,1,1)) or on(%s) timestamp(runai_gpu_utilization_per_pod)`, podLabelsWithoutGpu, podLabelsWithoutGpu, podLabelsWithoutGpu)),
			},
			{
				Record: "runai_gpu_utilization_per_workload",
				Expr:   intstr.FromString(`(sum(avg(runai_gpu_utilization_per_pod_per_gpu) by(pod_uuid, workload_id) * on(pod_uuid, workload_id) runai_allocated_gpu_count_per_pod) by (workload_id)) / on(workload_id) (runai_allocated_gpu_count_per_workload>0) * on(workload_id) group_right() runai_workload_info_enriched`),
			},
			{
				Record: "runai_last_gpu_utilization_time_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`timestamp(runai_gpu_utilization_per_workload > 2) or on(%s) (runai_last_gpu_utilization_time_per_workload * on(%s) group_left() clamp(runai_allocated_gpu_count_per_workload>0,1,1)) or on(%s) timestamp(runai_gpu_utilization_per_workload)`, workloadIdLabel, workloadIdLabel, workloadIdLabel)),
			},
			{
				Record: "runai_gpu_idle_seconds_per_workload",
				Expr:   intstr.FromString(`time() - runai_last_gpu_utilization_time_per_workload`),
			},
			{
				Record: "runai_gpu_idle_seconds_per_pod",
				Expr:   intstr.FromString(`time() - runai_last_gpu_utilization_time_per_pod`),
			},
			{
				Record: "runai_gpu_utilization_per_queue_per_workload_type",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(avg(runai_gpu_utilization_per_pod_per_gpu) by(pod_uuid, workload_type, %s) * on(pod_uuid, workload_type, %s) runai_allocated_gpu_count_per_pod) by (workload_type, %s) / on(workload_type, %s) (runai_allocated_gpu_count_per_queue_per_workload_type>0)`, queueLabels, queueLabels, queueLabels, queueLabels)),
			},
			{
				Record: "runai_gpu_utilization_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(avg(runai_gpu_utilization_per_pod_per_gpu) by(pod_uuid, %s) * on(pod_uuid, %s) runai_allocated_gpu_count_per_pod) by (%s) / on(%s) (runai_allocated_gpu_count_per_queue>0)`, queueLabels, queueLabels, queueLabels, queueLabels)),
			},
			{
				Record: "runai_gpu_utilization_per_project_per_workload_type",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(avg(runai_gpu_utilization_per_pod_per_gpu) by(pod_uuid, workload_type, %s) * on(pod_uuid, workload_type, %s) runai_allocated_gpu_count_per_pod) by (workload_type, %s) / on(workload_type, %s) (runai_allocated_gpu_count_per_project_per_workload_type>0)`, projectLabels, projectLabels, projectLabels, projectLabels)),
			},
			{
				Record: "runai_gpu_utilization_per_project",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(avg(runai_gpu_utilization_per_pod_per_gpu) by(pod_uuid, %s) * on(pod_uuid, %s) runai_allocated_gpu_count_per_pod) by (%s) / on(%s) (runai_allocated_gpu_count_per_project>0)`, projectLabels, projectLabels, projectLabels, projectLabels)),
			},
			{
				Record: "runai_gpu_utilization_per_department",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(avg(runai_gpu_utilization_per_pod_per_gpu) by(pod_uuid, %s) * on(pod_uuid, %s) runai_allocated_gpu_count_per_pod) by (%s) / on(%s) (runai_allocated_gpu_count_per_department>0)`, departmentLabels, departmentLabels, departmentLabels, departmentLabels)),
			},
			{
				Record: "runai_gpu_memory_utilization_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`clamp_max(100 * sum(runai_gpu_memory_used_mebibytes_per_workload * 1024 * 1024) by (%s) / sum(runai_allocated_gpu_memory_mb_per_workload * 1000 * 1000 > 0) by (%s), 100)`, queueLabels, queueLabels)),
			},
			{
				Record: "runai_gpu_memory_utilization_per_project",
				Expr:   intstr.FromString(fmt.Sprintf(`clamp_max(100 * sum(runai_gpu_memory_used_mebibytes_per_workload * 1024 * 1024) by (%s) / sum(runai_allocated_gpu_memory_mb_per_workload * 1000 * 1000 > 0) by (%s), 100)`, projectLabels, projectLabels)),
			},
			{
				Record: "runai_gpu_memory_utilization_per_department",
				Expr:   intstr.FromString(fmt.Sprintf(`clamp_max(100 * sum(runai_gpu_memory_used_mebibytes_per_workload * 1024 * 1024) by (%s) / sum(runai_allocated_gpu_memory_mb_per_workload * 1000 * 1000 > 0) by (%s), 100)`, departmentLabels, departmentLabels)),
			},
			{
				Record: "runai_gpu_utilization_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`%s or on(UUID) avg by (%s) (runai_dcgm_gpu_utilization) or on(UUID) avg by(%s) (runai_amd_gpu_utilization)`, migGpuAvgMetricRule("runai_gpu_utilization_per_mig_pod"), gpuLabels, gpuLabels)),
			},
			{
				Record: "runai_last_gpu_utilization_time_per_gpu",
				Expr:   intstr.FromString(`timestamp(runai_gpu_utilization_per_gpu > 2) or on(UUID) runai_last_gpu_utilization_time_per_gpu or on(UUID) timestamp(runai_gpu_utilization_per_gpu)`),
			},
			{
				Record: "runai_gpu_idle_seconds_per_gpu",
				Expr:   intstr.FromString(`time() - runai_last_gpu_utilization_time_per_gpu`),
			},
			{
				Record: "runai_gpu_utilization_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule(`runai_gpu_utilization_per_gpu`, true)),
			},
			{
				Record: "runai_gpu_utilization_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`avg by(%s) (runai_gpu_utilization_per_gpu)`, nodepoolLabels)),
			},
			{
				Record: "runai_gpu_utilization_per_cluster",
				Expr:   intstr.FromString(`avg(runai_gpu_utilization_per_gpu)`),
			},
			// =====================================================================================

			// ============================ Used GPU Memory Mebibytes ==============================
			{
				Record: "runai_dcgm_gpu_used_mebibytes",
				Expr:   intstr.FromString(dcgmMetricsRule(`DCGM_FI_DEV_FB_USED`)),
			},
			{
				Record: "runai_amd_gpu_memory_used_mebibytes",
				Expr:   intstr.FromString(amdMetricsRule(`amd_gpu_memory_used/(1024*1024)`)),
			},
			{
				Record: "runai_gpu_memory_used_mebibytes_per_fractional_pod",
				Expr:   intstr.FromString(gpuMetricsPerFractionalPodRule(`runai_pod_gpu_memory_used_bytes/(1024*1024)`)),
			},
			{
				Record: "runai_gpu_memory_used_mebibytes_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`%s or on(pod_uuid) %s`, gpuMetricsPerWholePodRule(`runai_dcgm_gpu_used_mebibytes`), gpuMetricsPerWholeAmdPodRule(`runai_amd_gpu_memory_used_mebibytes`))),
			},
			{
				Record: "runai_gpu_memory_used_mebibytes_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule(`runai_dcgm_gpu_used_mebibytes`)),
			},
			{
				Record: "runai_gpu_memory_used_mebibytes_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule(`runai_gpu_memory_used_mebibytes_per_whole_pod_per_gpu`, `runai_gpu_memory_used_mebibytes_per_mig_pod`, `runai_gpu_memory_used_mebibytes_per_fractional_pod`)),
			},
			{
				Record: "runai_gpu_memory_used_mebibytes_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule(`runai_gpu_memory_used_mebibytes_per_pod_per_gpu`, false)),
			},
			{
				Record: "runai_gpu_memory_used_mebibytes_per_queue_per_workload_type",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_gpu_memory_used_mebibytes_per_workload) by (workload_type, %s)`, queueLabels)),
			},
			// {
			// 	Record: "runai_gpu_memory_used_mebibytes_per_queue",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_gpu_memory_used_mebibytes_per_workload) by (%s)`, queueLabels)),
			// },
			// {
			// 	Record: "runai_gpu_memory_used_mebibytes_per_project",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_gpu_memory_used_mebibytes_per_queue) by (%s)`, projectLabels)),
			// },
			// {
			// 	Record: "runai_gpu_memory_used_mebibytes_per_department",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_gpu_memory_used_mebibytes_per_queue) by (%s)`, departmentLabels)),
			// },
			{
				Record: "runai_gpu_memory_used_mebibytes_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by (%s) (runai_dcgm_gpu_used_mebibytes) or on(UUID) sum by(%s) (runai_amd_gpu_memory_used_mebibytes)`, gpuLabels, gpuLabels)),
			},
			{
				Record: "runai_gpu_memory_used_mebibytes_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule(`runai_gpu_memory_used_mebibytes_per_gpu`, false)),
			},
			{
				Record: "runai_gpu_memory_used_mebibytes_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_gpu_memory_used_mebibytes_per_node)`, nodepoolLabels)),
			},
			{
				Record: "runai_gpu_memory_used_mebibytes_per_cluster",
				Expr:   intstr.FromString(`sum(runai_gpu_memory_used_mebibytes_per_nodepool)`),
			},
			// =====================================================================================

			// ============================ Total GPU Memory Mebibytes =============================
			{
				Record: "runai_dcgm_gpu_total_mebibytes",
				Expr:   intstr.FromString(dcgmMetricsRule(`DCGM_FI_DEV_FB_USED+DCGM_FI_DEV_FB_FREE`)),
			},
			{
				Record: "runai_amd_gpu_memory_total_mebibytes",
				Expr:   intstr.FromString(amdMetricsRule(`amd_gpu_memory_total/(1024*1024)`)),
			},
			{
				Record: "runai_gpu_memory_total_mebibytes_per_mig_mode_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`(group by (%s) (runai_dcgm_gpu_total_mebibytes{gpu_instance_id!="", modelName=~".*H100.*"}))*80000 or (group by (%s) (runai_dcgm_gpu_total_mebibytes{gpu_instance_id!="", modelName=~".*40GB"}))*40000 or (group by (%s) (runai_dcgm_gpu_total_mebibytes{gpu_instance_id!="", modelName=~".*80GB"}))*80000`, gpuLabels, gpuLabels, gpuLabels)),
			},
			{
				Record: "runai_gpu_memory_total_mebibytes_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`runai_gpu_memory_total_mebibytes_per_mig_mode_gpu or on(UUID) sum by (%s) (runai_dcgm_gpu_total_mebibytes) or on(UUID) sum by(%s) (runai_amd_gpu_memory_total_mebibytes)`, gpuLabels, gpuLabels)),
			},
			{
				Record: "runai_gpu_memory_total_mebibytes_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule(`runai_gpu_memory_total_mebibytes_per_gpu`, false)),
			},
			{
				Record: "runai_gpu_memory_total_mebibytes_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_gpu_memory_total_mebibytes_per_node)`, nodepoolLabels)),
			},
			{
				Record: "runai_gpu_memory_total_mebibytes_per_cluster",
				Expr:   intstr.FromString(`sum(runai_gpu_memory_total_mebibytes_per_nodepool)`),
			},
			// =====================================================================================

			// ============================= GPU Memory Utilization ================================
			{
				Record: "runai_gpu_memory_utilization_per_node",
				Expr:   intstr.FromString(`100*(runai_gpu_memory_used_mebibytes_per_node / runai_gpu_memory_total_mebibytes_per_node)`),
			},
			{
				Record: "runai_gpu_memory_utilization_per_nodepool",
				Expr:   intstr.FromString(`100*(runai_gpu_memory_used_mebibytes_per_nodepool / runai_gpu_memory_total_mebibytes_per_nodepool)`),
			},
			{
				Record: "runai_gpu_memory_utilization_per_cluster",
				Expr:   intstr.FromString(`100*(runai_gpu_memory_used_mebibytes_per_cluster / runai_gpu_memory_total_mebibytes_per_cluster)`),
			},
			// =====================================================================================

			// ================================ Requested CPU Cores ================================
			{
				Record: "runai_pending_requested_cpu_cores_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (sum(label_replace(sum(kube_pod_container_resource_requests{resource='cpu'}) by(uid), "pod_uuid" , "$1", "uid", "(.*)") * on(pod_uuid) group_right() runai_pod_info_enriched_nodepool_unique or on(pod_uuid) runai_pod_info_enriched_nodepool_unique*0) by(%s) * on(%s) group_right() runai_workload_info_enriched{detailed_status="Pending"})`, workloadLabels, workloadIdLabel, workloadIdLabel)),
			},
			{
				Record: "runai_pending_requested_cpu_cores_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_pending_requested_cpu_cores_per_workload) by (%s)`, queueLabels)),
			},
			{
				Record: "runai_pending_requested_cpu_cores_per_project",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_pending_requested_cpu_cores_per_queue) by (%s)`, projectLabels)),
			},
			{
				Record: "runai_pending_requested_cpu_cores_per_department",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_pending_requested_cpu_cores_per_queue) by (%s)`, departmentLabels)),
			},
			{
				Record: "runai_pending_requested_cpu_cores_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_pending_requested_cpu_cores_per_queue) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_pending_requested_cpu_cores_per_cluster",
				Expr:   intstr.FromString(`sum(runai_pending_requested_cpu_cores_per_nodepool)`),
			},
			// =====================================================================================

			// ================================ Allocated CPU Cores ================================
			// TODO: remove this, use the next one instead
			{
				Record: "runai_allocated_millicpus_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_scheduler_allocated_millicpus_per_pod_unique) * on (pod_uuid) group_right() runai_pod_info_enriched_nodepool_unique`, podLabels)),
			},
			{
				Record: "runai_allocated_cpu_cores_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) ((runai_scheduler_allocated_millicpus_per_pod_unique/1000) * on (pod_uuid) group_right() runai_pod_info_enriched_nodepool_unique)`, podLabels)),
			},
			{
				Record: "runai_allocated_cpu_cores_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule(`runai_allocated_cpu_cores_per_pod`, false)),
			},
			{
				Record: "runai_allocated_cpu_cores_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_cpu_cores_per_pod) by (%s)`, queueLabels)),
			},
			{
				Record: "runai_allocated_cpu_cores_per_project",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_cpu_cores_per_queue) by (%s)`, projectLabels)),
			},
			{
				Record: "runai_allocated_cpu_cores_per_department",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_cpu_cores_per_queue) by (%s)`, departmentLabels)),
			},
			{
				Record: "runai_allocated_cpu_cores_per_node",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (sum by (node) (kube_pod_container_resource_requests{resource="cpu",unit="core",node!=""} * on(pod, namespace) group_left() (kube_pod_status_phase{phase=~"Pending|Running|Unknown"}==1)) * on(node) group_right() runai_node_info_enriched)`, nodeLabels)),
			},
			{
				Record: "runai_allocated_cpu_cores_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_cpu_cores_per_node) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_allocated_cpu_cores_per_cluster",
				Expr:   intstr.FromString(`sum(runai_allocated_cpu_cores_per_nodepool)`),
			},
			// =====================================================================================

			// ================================== Used CPU Cores ===================================
			{
				Record: "runai_used_cpu_cores_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (label_replace(label_replace((sum(rate(container_cpu_usage_seconds_total{container!=""}[2m])) by (pod, namespace) or max(rate(container_cpu_usage_seconds_total{container=""}[2m])) by (pod, namespace)),"pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)") * on(pod_name, pod_namespace) group_right() runai_pod_info_enriched_nodepool_unique)`, podLabels)),
			},
			{
				Record: "runai_used_cpu_cores_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule(`runai_used_cpu_cores_per_pod`, false)),
			},
			{
				Record: "runai_used_cpu_cores_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_used_cpu_cores_per_workload) by (%s)`, queueLabels)),
			},
			{
				Record: "runai_used_cpu_cores_per_project",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_used_cpu_cores_per_queue) by (%s)`, projectLabels)),
			},
			{
				Record: "runai_used_cpu_cores_per_department",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_used_cpu_cores_per_queue) by (%s)`, departmentLabels)),
			},
			{
				Record: "runai_used_cpu_cores_per_node",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) ((runai_cpu_cores_count_per_node - on(node) sum(rate(node_cpu_seconds_total{job="node-exporter",mode=~"idle|iowait|steal"}[1m])) by(pod, namespace) * on(pod, namespace) group_left(node) kube_pod_info) * on(node) group_right() runai_node_info_enriched)`, nodeLabels)),
			},
			{
				Record: "runai_used_cpu_cores_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_used_cpu_cores_per_node) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_used_cpu_cores_per_cluster",
				Expr:   intstr.FromString(`sum(runai_used_cpu_cores_per_nodepool)`),
			},
			// =====================================================================================

			// ================================= Deserved CPU Cores ================================
			{
				Record: "runai_deserved_cpu_cores_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_queue_quota_cpu_cores) by (%s)`, queueLabels)),
			},
			{
				Record: "runai_deserved_cpu_cores_per_project",
				Expr:   intstr.FromString(fmt.Sprintf(`min(runai_deserved_cpu_cores_per_queue{project!=""}<0) by(%s) or on(project) sum(runai_deserved_cpu_cores_per_queue{project!=""}>=0) by(%s)`, projectLabels, projectLabels)),
			},
			{
				Record: "runai_deserved_cpu_cores_per_department",
				Expr:   intstr.FromString(fmt.Sprintf(`min(runai_deserved_cpu_cores_per_queue{project=""}<0) by(%s) or on(department) sum(runai_deserved_cpu_cores_per_queue{project=""}>=0) by(%s)`, departmentLabels, departmentLabels)),
			},
			// =====================================================================================

			// ================================== Total CPU Cores ==================================
			{
				Record: "runai_cpu_cores_count_per_node",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by (%s) (kube_node_status_capacity{resource='cpu'} * on(node) group_right() runai_node_info_enriched)`, nodeLabels)),
			},
			{
				Record: "runai_cpu_cores_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_cpu_cores_count_per_node) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_ready_cpu_cores_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_cpu_cores_count_per_node{ready="true",schedulable="true"}) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_cpu_cores_count_per_cluster",
				Expr:   intstr.FromString(`sum(runai_cpu_cores_count_per_nodepool)`),
			},
			{
				Record: "runai_ready_cpu_cores_count_per_cluster",
				Expr:   intstr.FromString(`sum(runai_ready_cpu_cores_count_per_nodepool)`),
			},
			// =====================================================================================

			// =============================== CPU Cores Allocation % ==============================
			{
				Record: "runai_cpu_cores_allocation_percentage_per_node",
				Expr:   intstr.FromString(`100*(runai_allocated_cpu_cores_per_node / runai_cpu_cores_count_per_node>0)`),
			},
			{
				Record: "runai_cpu_cores_allocation_percentage_per_nodepool",
				Expr:   intstr.FromString(`100*(runai_allocated_cpu_cores_per_nodepool / runai_cpu_cores_count_per_nodepool>0)`),
			},
			{
				Record: "runai_cpu_cores_allocation_percentage_per_cluster",
				Expr:   intstr.FromString(`100*(runai_allocated_cpu_cores_per_cluster / runai_cpu_cores_count_per_cluster>0)`),
			},
			// =====================================================================================

			// =================================== CPU Utilization =================================
			{
				Record: "runai_cpu_utilization_per_pod",
				Expr:   intstr.FromString(`100*(runai_used_cpu_cores_per_pod / (runai_allocated_cpu_cores_per_pod>0)) or on(pod_uuid) (runai_used_cpu_cores_per_pod*0)+100`),
			},
			{
				Record: "runai_cpu_utilization_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`100*(runai_used_cpu_cores_per_workload / (runai_allocated_cpu_cores_per_workload>0)) or on(%s) (runai_used_cpu_cores_per_workload*0)+100`, workloadIdLabel)),
			},
			{
				Record: "runai_cpu_utilization_per_queue",
				Expr:   intstr.FromString(`100*(runai_used_cpu_cores_per_queue / (runai_allocated_cpu_cores_per_queue>0)) or on(queue_name) (runai_used_cpu_cores_per_queue*0)+100`),
			},
			{
				Record: "runai_cpu_utilization_per_project",
				Expr:   intstr.FromString(`100*(runai_used_cpu_cores_per_project / (runai_allocated_cpu_cores_per_project>0)) or on(project) (runai_used_cpu_cores_per_project*0)+100`),
			},
			{
				Record: "runai_cpu_utilization_per_department",
				Expr:   intstr.FromString(`100*(runai_used_cpu_cores_per_department / (runai_allocated_cpu_cores_per_department>0)) or on(department) (runai_used_cpu_cores_per_department*0)+100`),
			},
			{
				Record: "runai_cpu_utilization_per_node",
				Expr:   intstr.FromString(`100*(runai_used_cpu_cores_per_node / runai_cpu_cores_count_per_node)`),
			},
			{
				Record: "runai_cpu_utilization_per_nodepool",
				Expr:   intstr.FromString(`100*(runai_used_cpu_cores_per_nodepool / runai_cpu_cores_count_per_nodepool)`),
			},
			{
				Record: "runai_cpu_utilization_per_cluster",
				Expr:   intstr.FromString(`100*(runai_used_cpu_cores_per_cluster / runai_cpu_cores_count_per_cluster)`),
			},
			// =====================================================================================

			// =============================== Requested Memory Bytes ==============================
			{
				Record: "runai_pending_requested_memory_bytes_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (sum(label_replace(sum(kube_pod_container_resource_requests{resource='memory'}) by(uid), "pod_uuid" , "$1", "uid", "(.*)") * on(pod_uuid) group_right() runai_pod_info_enriched_nodepool_unique or on(pod_uuid) runai_pod_info_enriched_nodepool_unique*0) by(%s) * on(%s) group_right() runai_workload_info_enriched{detailed_status="Pending"})`, workloadLabels, workloadIdLabel, workloadIdLabel)),
			},
			{
				Record: "runai_pending_requested_memory_bytes_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_pending_requested_memory_bytes_per_workload) by (%s)`, queueLabels)),
			},
			{
				Record: "runai_pending_requested_memory_bytes_per_project",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_pending_requested_memory_bytes_per_queue) by (%s)`, projectLabels)),
			},
			{
				Record: "runai_pending_requested_memory_bytes_per_department",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_pending_requested_memory_bytes_per_queue) by (%s)`, departmentLabels)),
			},
			{
				Record: "runai_pending_requested_memory_bytes_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_pending_requested_memory_bytes_per_queue) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_pending_requested_memory_bytes_per_cluster",
				Expr:   intstr.FromString(`sum(runai_pending_requested_memory_bytes_per_nodepool)`),
			},
			// =====================================================================================

			// =============================== Allocated Memory Bytes ==============================
			// TODO: Remove this, use the next one instead
			{
				Record: "runai_allocated_memory_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (%s) * on (pod_uuid) group_right() runai_pod_info_enriched_nodepool_unique`, podLabels, allocatedMemoryBytesPerPodMetric)),
			},
			// {
			// 	Record: "runai_allocated_memory_bytes_per_pod",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) ((%s) * on (pod_uuid) group_right() runai_pod_info_enriched_nodepool_unique)`, podLabels, allocatedMemoryBytesPerPodMetric)),
			// },
			// {
			// 	Record: "runai_allocated_memory_bytes_per_workload",
			// 	Expr:   intstr.FromString(metricsPerWorkloadRule(`runai_allocated_memory_bytes_per_pod`, false)),
			// },
			{
				Record: "runai_allocated_memory_bytes_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_memory_per_pod) by (%s)`, queueLabels)),
			},
			// {
			// 	Record: "runai_allocated_memory_bytes_per_project",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_memory_bytes_per_queue) by (%s)`, projectLabels)),
			// },
			// {
			// 	Record: "runai_allocated_memory_bytes_per_department",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_memory_bytes_per_queue) by (%s)`, departmentLabels)),
			// },
			{
				Record: "runai_allocated_memory_bytes_per_node",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (sum by (node) (kube_pod_container_resource_requests{resource="memory",unit="byte",node!=""} * on(pod, namespace) group_left() (kube_pod_status_phase{phase=~"Pending|Running|Unknown"}==1)) * on(node) group_right() runai_node_info_enriched)`, nodeLabels)),
			},
			{
				Record: "runai_allocated_memory_bytes_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_allocated_memory_bytes_per_node) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_allocated_memory_bytes_per_cluster",
				Expr:   intstr.FromString(`sum(runai_allocated_memory_bytes_per_nodepool)`),
			},

			// =====================================================================================

			// ================================= Used Memory Bytes =================================
			{
				Record: "runai_used_memory_bytes_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (label_replace(label_replace((sum(container_memory_usage_bytes{ container!=""}) by (pod, namespace) or max(container_memory_usage_bytes{container=""}) by (pod, namespace)),"pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)") * on(pod_name, pod_namespace) group_right() runai_pod_info_enriched_nodepool_unique)`, podLabels)),
			},
			{
				Record: "runai_used_memory_bytes_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule(`runai_used_memory_bytes_per_pod`, false)),
			},
			{
				Record: "runai_used_memory_bytes_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_used_memory_bytes_per_workload) by (%s)`, queueLabels)),
			},
			// {
			// 	Record: "runai_used_memory_bytes_per_project",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_used_memory_bytes_per_queue) by (%s)`, projectLabels)),
			// },
			// {
			// 	Record: "runai_used_memory_bytes_per_department",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_used_memory_bytes_per_queue) by (%s)`, departmentLabels)),
			// },
			{
				Record: "runai_used_memory_bytes_per_node",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) ((node_memory_MemTotal_bytes{job="node-exporter"} - (node_memory_MemAvailable_bytes{job="node-exporter"} or (node_memory_Cached_bytes{job="node-exporter"} + node_memory_Buffers_bytes{job="node-exporter"} + node_memory_MemFree_bytes{job="node-exporter"} + node_memory_Slab_bytes{job="node-exporter"}))) * on(pod, namespace) group_left(node) kube_pod_info * on(node) group_right() runai_node_info_enriched)`, nodeLabels)),
			},
			{
				Record: "runai_used_memory_bytes_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_used_memory_bytes_per_node) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_used_memory_bytes_per_cluster",
				Expr:   intstr.FromString(`sum(runai_used_memory_bytes_per_nodepool)`),
			},
			// =====================================================================================

			// =============================== Deserved Memory Bytes ===============================
			{
				Record: "runai_deserved_memory_bytes_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (runai_queue_quota_memory_bytes)`, queueLabels)),
			},
			{
				Record: "runai_deserved_memory_bytes_per_project",
				Expr:   intstr.FromString(fmt.Sprintf(`min(runai_deserved_memory_bytes_per_queue{project!=""}<0) by(%s) or on(project) sum(runai_deserved_memory_bytes_per_queue{project!=""}>=0) by(%s)`, projectLabels, projectLabels)),
			},
			{
				Record: "runai_deserved_memory_bytes_per_department",
				Expr:   intstr.FromString(fmt.Sprintf(`min(runai_deserved_memory_bytes_per_queue{project=""}<0) by(%s) or on(department) sum(runai_deserved_memory_bytes_per_queue{project=""}>=0) by(%s)`, departmentLabels, departmentLabels)),
			},
			// =====================================================================================

			// ================================ Total Memory Bytes =================================
			{
				Record: "runai_total_memory_bytes_per_node",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by (%s) (kube_node_status_capacity{resource='memory'} * on(node) group_right() runai_node_info_enriched)`, nodeLabels)),
			},
			{
				Record: "runai_total_memory_bytes_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_total_memory_bytes_per_node) by (%s)`, nodepoolLabels)),
			},
			{
				Record: "runai_total_memory_bytes_per_cluster",
				Expr:   intstr.FromString(`sum(runai_total_memory_bytes_per_nodepool)`),
			},
			// =====================================================================================

			// ================================ Memory Allocation % ================================
			{
				Record: "runai_memory_allocation_percentage_per_node",
				Expr:   intstr.FromString(`100*(runai_allocated_memory_bytes_per_node / runai_total_memory_bytes_per_node>0)`),
			},
			{
				Record: "runai_memory_allocation_percentage_per_nodepool",
				Expr:   intstr.FromString(`100*(runai_allocated_memory_bytes_per_nodepool / runai_total_memory_bytes_per_nodepool>0)`),
			},
			{
				Record: "runai_memory_allocation_percentage_per_cluster",
				Expr:   intstr.FromString(`100*(runai_allocated_memory_bytes_per_cluster / runai_total_memory_bytes_per_cluster>0)`),
			},
			// =====================================================================================

			// ================================= Memory Utilization ================================
			// {
			// 	Record: "runai_memory_utilization_per_pod",
			// 	Expr:   intstr.FromString(`100*(runai_used_memory_bytes_per_pod / (runai_allocated_memory_bytes_per_pod>0)) or on(pod_uuid) (runai_used_memory_bytes_per_pod*0)+100`),
			// },
			// {
			// 	Record: "runai_memory_utilization_per_workload",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`100*(runai_used_memory_bytes_per_workload / (runai_allocated_memory_bytes_per_workload>0)) or on(%s) (runai_used_memory_bytes_per_workload*0)+100`, workloadIdLabel)),
			// },
			// {
			// 	Record: "runai_memory_utilization_per_queue",
			// 	Expr:   intstr.FromString(`100*(runai_used_memory_bytes_per_queue / (runai_allocated_memory_bytes_per_queue>0)) or on(queue_name) (runai_used_memory_bytes_per_queue*0)+100`),
			// },
			// {
			// 	Record: "runai_memory_utilization_per_project",
			// 	Expr:   intstr.FromString(`100*(runai_used_memory_bytes_per_project / (runai_allocated_memory_bytes_per_project>0)) or on(project) (runai_used_memory_bytes_per_project*0)+100`),
			// },
			// {
			// 	Record: "runai_memory_utilization_per_department",
			// 	Expr:   intstr.FromString(`100*(runai_used_memory_bytes_per_department / (runai_allocated_memory_bytes_per_department>0)) or on(department) (runai_used_memory_bytes_per_department*0)+100`),
			// },
			{
				Record: "runai_memory_utilization_per_node",
				Expr:   intstr.FromString(`100*(runai_used_memory_bytes_per_node / runai_total_memory_bytes_per_node)`),
			},
			{
				Record: "runai_memory_utilization_per_nodepool",
				Expr:   intstr.FromString(`100*(runai_used_memory_bytes_per_nodepool / runai_total_memory_bytes_per_nodepool)`),
			},
			{
				Record: "runai_memory_utilization_per_cluster",
				Expr:   intstr.FromString(`100*(runai_used_memory_bytes_per_cluster / runai_total_memory_bytes_per_cluster)`),
			},
			// =====================================================================================

			// ================================== Workload Count ===================================
			{
				Record: "runai_workload_count_per_type_per_phase_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`count(runai_workload_info_enriched) by(detailed_status, is_gpu, workload_type, %s)`, queueLabels)),
			},
			{
				Record: "runai_workload_count_per_category",
				Expr:   intstr.FromString("count(runai_workload_info_enriched) by (workload_category)"),
			},
			{
				Record: "runai_workload_count_per_project",
				Expr:   intstr.FromString(fmt.Sprintf("count(runai_workload_info_enriched) by (%s)", projectLabels)),
			},
			{
				Record: "runai_workload_count_per_department",
				Expr:   intstr.FromString(fmt.Sprintf("sum(runai_workload_count_per_project) by (%s)", departmentLabels)),
			},
			{
				Record: "runai_workload_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf("count(runai_workload_info_enriched) by (%s)", nodepoolLabels)),
			},
			{
				Record: "runai_workload_count_per_cluster",
				Expr:   intstr.FromString("sum(runai_workload_count_per_nodepool)"),
			},
			// {
			// 	Record: "runai_workload_count_per_type_per_phase_per_project",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_workload_count_per_type_per_phase_per_queue) by(detailed_status, is_gpu, workload_type, %s)`, projectLabels)),
			// },
			// {
			// 	Record: "runai_workload_count_per_type_per_phase_per_department",
			// 	Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_workload_count_per_type_per_phase_per_queue) by(detailed_status, is_gpu, workload_type, %s)`, departmentLabels)),
			// },
			{
				Record: "runai_workload_count_per_type_per_phase_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_workload_count_per_type_per_phase_per_queue) by(detailed_status, is_gpu, workload_type, %s)`, nodepoolLabels)),
			},
			{
				Record: "runai_workload_count_per_type_per_phase_per_cluster",
				Expr:   intstr.FromString(`sum(runai_workload_count_per_type_per_phase_per_nodepool) by(detailed_status, is_gpu, workload_type)`),
			},
			// =====================================================================================

			// ============================== NVLink Bandwidth Metrics ==============================
			{
				Record: "runai_dcgm_nvlink_bandwidth_total_bytes",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_DEV_NVLINK_BANDWIDTH_TOTAL")),
			},
			{
				Record: "runai_dcgm_nvlink_bandwidth_total_bytes_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_nvlink_bandwidth_total_bytes")),
			},
			{
				Record: "runai_dcgm_nvlink_bandwidth_total_bytes_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_nvlink_bandwidth_total_bytes")),
			},
			{
				Record: "runai_dcgm_nvlink_bandwidth_total_bytes_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_dcgm_nvlink_bandwidth_total_bytes_per_whole_pod_per_gpu", "runai_dcgm_nvlink_bandwidth_total_bytes_per_mig_pod", "")),
			},
			{
				Record: "runai_dcgm_nvlink_bandwidth_total_bytes_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by (%s) (runai_dcgm_nvlink_bandwidth_total_bytes_per_pod_per_gpu)`, podLabelsWithoutGpu)),
			},
			{
				Record: "runai_dcgm_nvlink_bandwidth_total_bytes_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_dcgm_nvlink_bandwidth_total_bytes_per_pod_per_gpu", false)),
			},
			{
				Record: "runai_dcgm_nvlink_bandwidth_total_bytes_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by (%s) (runai_dcgm_nvlink_bandwidth_total_bytes)`, gpuLabels)),
			},
			{
				Record: "runai_dcgm_nvlink_bandwidth_total_bytes_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_dcgm_nvlink_bandwidth_total_bytes_per_gpu", false)),
			},

			// =====================================================================================

			// ==================================== Deprecated =====================================
			{
				Record: "runai_node_cpu_seconds_total",
				Expr:   intstr.FromString(`(node_cpu_seconds_total * on(pod, namespace) group_left(node) kube_pod_info) * on(node) group_left(nodepool) runai_node_nodepool_excluded`),
			},
			{
				Record: "runai_node_memory_mem_available_bytes",
				Expr:   intstr.FromString(`(node_memory_MemAvailable_bytes * on(pod, namespace) group_left(node) kube_pod_info) * on(node) group_left(nodepool) runai_node_nodepool_excluded`),
			},
			{
				Record: "runai_node_memory_mem_total_bytes",
				Expr:   intstr.FromString(`(node_memory_MemTotal_bytes * on(pod, namespace) group_left(node) kube_pod_info) * on(node) group_left(nodepool) runai_node_nodepool_excluded`),
			},
			// =====================================================================================
		},
	}
	newGenRules.Rules = slices.Concat(schedulerMetricsReduction, workloadNewGenRules, newGenRules.Rules)

	dashboardRules := monitoringv1.RuleGroup{
		Name: "runai-dashboard-rules",
		Rules: []monitoringv1.Rule{
			{
				Record: "runai_idle_allocated_gpu_count_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`(runai_gpu_idle_seconds_per_workload > 300)*0 + on(%s) group_left() (runai_allocated_gpu_count_per_workload > 0)`, workloadIdLabel)),
			},
			{
				Record: "runai_idle_allocated_gpu_count_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`(runai_gpu_idle_seconds_per_pod > 300)*0 + on(%s) group_left() (runai_allocated_gpu_count_per_pod > 0)`, podLabelsWithoutGpu)),
			},
			{
				Record: "runai_idle_allocated_gpu_count_per_queue",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_idle_allocated_gpu_count_per_workload) by (%s)`, queueLabels)),
			},
			{
				Record: "runai_idle_allocated_gpu_count_per_node",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_idle_allocated_gpu_count_per_pod) by (%s)`, nodeLabels)),
			},
			{
				Record: "runai_idle_allocated_gpu_count_per_nodepool",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_idle_allocated_gpu_count_per_workload) by (%s)`, nodepoolLabels)),
			},
		},
	}

	workloadRunaiRules := []monitoringv1.Rule{
		// runai_pod_group_phase_with_info usage will be replaced by "runai_workload_info_enriched".
		{
			Record: "runai_pod_phase_with_info",
			Expr:   intstr.FromString(`runai_pod_info_enriched_nodepool_unique`),
		},
		{
			Record: "runai_requested_gpus_per_workload",
			Expr:   intstr.FromString(metricsPerWorkloadRule(`runai_scheduler_requested_gpus_per_pod_unique`, false)),
		},
		{
			Record: "runai_requested_gpu_memory_mb_per_workload",
			Expr:   intstr.FromString(metricsPerWorkloadRule(`runai_scheduler_requested_gpu_memory_mb_per_pod_unique`, false)),
		},
		{
			Record: "runai_gpus_limit_per_workload",
			Expr:   intstr.FromString(metricsPerWorkloadRule(`runai_scheduler_gpus_limit_per_pod_unique`, false)),
		},
		{
			Record: "runai_gpu_memory_mb_limit_per_workload",
			Expr:   intstr.FromString(metricsPerWorkloadRule(`runai_scheduler_gpu_memory_mb_limit_per_pod_unique`, false)),
		},
		{
			Record: "runai_requested_gpus_per_workload_type_per_queue",
			Expr:   intstr.FromString(`sum(max(runai_requested_gpus_per_workload) by (workload_id) * on (workload_id) group_left (queue_name, nodepool, workload_type) (runai_workload_info_enriched{phase="Pending",workload_type!~"Inference|inference|"}==1)) by (nodepool, workload_type, queue_name)`),
		},
	}

	runaiRules := monitoringv1.RuleGroup{
		Name: "runai-rules",
		Rules: []monitoringv1.Rule{
			{
				Record: "runai_active_job_memory_requested_bytes",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(sum(label_replace(label_replace(kube_pod_container_resource_requests{resource="memory",unit="byte"} * on(pod, namespace) group_left() (kube_pod_status_phase{phase=~"Pending|Running|Unknown"}==1), "pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) by (pod_name, pod_namespace) * on(pod_name, pod_namespace) group_left(%s) (runai_pod_info_enriched_nodepool_unique * 0 + 1) ) by (%s)`, workloadNameUidLabels, workloadNameUidLabels)),
			},
			{
				Record: "runai_active_job_cpu_requested_cores",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(sum(label_replace(label_replace(kube_pod_container_resource_requests{resource="cpu",unit="core"} * on(pod, namespace) group_left() (kube_pod_status_phase{phase=~"Pending|Running|Unknown"}==1),"pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) by (pod_name, pod_namespace) * on(pod_name, pod_namespace) group_left(%s) (runai_pod_info_enriched_nodepool_unique * 0 + 1) ) by (%s)`, workloadNameUidLabels, workloadNameUidLabels)),
			},
			{
				Record: "runai_cpu_requested_cores_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (sum(label_replace(label_replace(kube_pod_container_resource_requests{resource="cpu",unit="core"} * on(pod, namespace) group_left() (kube_pod_status_phase{phase=~"Pending|Running|Unknown"}==1),"pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) by (pod_name, pod_namespace) * on(pod_name, pod_namespace) group_right() runai_pod_info_enriched_nodepool_unique)`, podLabels)),
			},
			{
				Record: "runai_cpu_limits_per_active_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(sum(label_replace(label_replace(kube_pod_container_resource_limits{resource="cpu"} * on(pod, namespace) group_left() (kube_pod_status_phase{phase=~"Pending|Running|Unknown"}==1),"pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) by (pod_name, pod_namespace) * on(pod_name, pod_namespace) group_left(%s) (runai_pod_info_enriched_nodepool_unique * 0 + 1) ) by (%s)`, workloadNameUidLabels, workloadNameUidLabels)),
			},
			{
				Record: "runai_cpu_limit_cores_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (sum(label_replace(label_replace(kube_pod_container_resource_limits{resource="cpu"} * on(pod, namespace) group_left() (kube_pod_status_phase{phase=~"Pending|Running|Unknown"}==1),"pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) by (pod_name, pod_namespace) * on(pod_name, pod_namespace) group_right() runai_pod_info_enriched_nodepool_unique)`, podLabels)),
			},
			{
				Record: "runai_memory_limits_per_active_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(sum(label_replace(label_replace(kube_pod_container_resource_limits{resource="memory"} * on(pod, namespace) group_left() (kube_pod_status_phase{phase=~"Pending|Running|Unknown"}==1),"pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) by (pod_name, pod_namespace) * on(pod_name, pod_namespace) group_left(%s) (runai_pod_info_enriched_nodepool_unique * 0 + 1) ) by (%s)`, workloadNameUidLabels, workloadNameUidLabels)),
			},
			{
				Record: "runai_memory_requested_bytes_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (sum(label_replace(label_replace(kube_pod_container_resource_requests{resource="memory",unit="byte"} * on(pod, namespace) group_left() (kube_pod_status_phase{phase=~"Pending|Running|Unknown"}==1),"pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) by (pod_name, pod_namespace) * on(pod_name, pod_namespace) group_right() runai_pod_info_enriched_nodepool_unique)`, podLabels)),
			},
			{
				Record: "runai_memory_limit_bytes_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by(%s) (sum(label_replace(label_replace(kube_pod_container_resource_limits{resource="memory"} * on(pod, namespace) group_left() (kube_pod_status_phase{phase=~"Pending|Running|Unknown"}==1),"pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) by (pod_name, pod_namespace) * on(pod_name, pod_namespace) group_right() runai_pod_info_enriched_nodepool_unique)`, podLabels)),
			},
			{
				Record: "runai_project_info",
				Expr:   intstr.FromString(`sum without(endpoint, instance, job,namespace, pod, queue_name, service, deserved_cpu, deserved_gpu, deserved_memory, nodepool)(runai_queue_info{project!=""})`),
			},
			// NIM metrics
			{
				Record: "runai_nim_num_requests_running_per_pod",
				Expr:   intstr.FromString(`(label_replace(label_replace(sum(num_requests_running) by (pod, namespace), "pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) * on(pod_name, pod_namespace) group_right() runai_pod_info_enriched_nodepool_unique`),
			},
			{
				Record: "runai_nim_num_requests_running_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_nim_num_requests_running_per_pod) by (%s)`, workloadLabels)),
			},
			{
				Record: "runai_nim_num_requests_waiting_per_pod",
				Expr:   intstr.FromString(`(label_replace(label_replace(sum(num_requests_waiting) by (pod, namespace), "pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) * on(pod_name, pod_namespace) group_right() runai_pod_info_enriched_nodepool_unique`),
			},
			{
				Record: "runai_nim_num_requests_waiting_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_nim_num_requests_waiting_per_pod) by (%s)`, workloadLabels)),
			},
			{
				Record: "runai_nim_num_request_max_per_pod",
				Expr:   intstr.FromString(`(label_replace(label_replace(sum(num_request_max) by (pod, namespace), "pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) * on(pod_name, pod_namespace) group_right() runai_pod_info_enriched_nodepool_unique`),
			},
			{
				Record: "runai_nim_num_request_max_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_nim_num_request_max_per_pod) by (%s)`, workloadLabels)),
			},
			{
				Record: "runai_nim_request_success_total_per_pod",
				Expr:   intstr.FromString(`(label_replace(label_replace(sum(request_success_total) by (pod, namespace), "pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) * on(pod_name, pod_namespace) group_right() runai_pod_info_enriched_nodepool_unique`),
			},
			{
				Record: "runai_nim_request_success_total_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_nim_request_success_total_per_pod) by (%s)`, workloadLabels)),
			},
			{
				Record: "runai_nim_request_failure_total_per_pod",
				Expr:   intstr.FromString(`(label_replace(label_replace(sum(request_failure_total) by (pod, namespace), "pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) * on(pod_name, pod_namespace) group_right() runai_pod_info_enriched_nodepool_unique`),
			},
			{
				Record: "runai_nim_request_failure_total_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_nim_request_failure_total_per_pod) by (%s)`, workloadLabels)),
			},
			{
				Record: "runai_nim_gpu_cache_usage_perc_per_pod",
				Expr:   intstr.FromString(`(label_replace(label_replace(avg(gpu_cache_usage_perc) by (pod, namespace), "pod_name" , "$1", "pod", "(.*)"), "pod_namespace" , "$1", "namespace", "(.*)")) * on(pod_name, pod_namespace) group_right() runai_pod_info_enriched_nodepool_unique`),
			},
			{
				Record: "runai_nim_gpu_cache_usage_perc_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`avg(runai_nim_gpu_cache_usage_perc_per_pod) by (%s)`, workloadLabels)),
			},
			{
				Record: "runai_nim_time_to_first_token_seconds_bucket_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`(label_replace(label_replace(sum(time_to_first_token_seconds_bucket) by (pod, namespace, le),"pod_name","$1","pod","(.*)"),"pod_namespace","$1","namespace","(.*)")) * on(pod_name, pod_namespace) group_left(%s) runai_pod_info_enriched_nodepool_unique`, podLabelsWithoutPodKeys)),
			},
			{
				Record: "runai_nim_time_to_first_token_seconds_sum_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`(label_replace(label_replace(sum(time_to_first_token_seconds_sum) by (pod, namespace),"pod_name","$1","pod","(.*)"),"pod_namespace","$1","namespace","(.*)")) * on(pod_name, pod_namespace) group_left(%s) runai_pod_info_enriched_nodepool_unique`, podLabelsWithoutPodKeys)),
			},
			{
				Record: "runai_nim_time_to_first_token_seconds_count_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`(label_replace(label_replace(sum(time_to_first_token_seconds_count) by (pod, namespace),"pod_name","$1","pod","(.*)"),"pod_namespace","$1","namespace","(.*)")) * on(pod_name, pod_namespace) group_left(%s) runai_pod_info_enriched_nodepool_unique`, podLabelsWithoutPodKeys)),
			},
			{
				Record: "runai_nim_e2e_request_latency_seconds_bucket_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`(label_replace(label_replace(sum(e2e_request_latency_seconds_bucket) by (pod, namespace, le),"pod_name","$1","pod","(.*)"),"pod_namespace","$1","namespace","(.*)")) * on(pod_name, pod_namespace) group_left(%s) runai_pod_info_enriched_nodepool_unique`, podLabelsWithoutPodKeys)),
			},
			{
				Record: "runai_nim_e2e_request_latency_seconds_sum_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`(label_replace(label_replace(sum(e2e_request_latency_seconds_sum) by (pod, namespace),"pod_name","$1","pod","(.*)"),"pod_namespace","$1","namespace","(.*)")) * on(pod_name, pod_namespace) group_left(%s) runai_pod_info_enriched_nodepool_unique`, podLabelsWithoutPodKeys)),
			},
			{
				Record: "runai_nim_e2e_request_latency_seconds_count_per_pod",
				Expr:   intstr.FromString(fmt.Sprintf(`(label_replace(label_replace(sum(e2e_request_latency_seconds_count) by (pod, namespace),"pod_name","$1","pod","(.*)"),"pod_namespace","$1","namespace","(.*)")) * on(pod_name, pod_namespace) group_left(%s) runai_pod_info_enriched_nodepool_unique`, podLabelsWithoutPodKeys)),
			},
		},
	}
	runaiRules.Rules = append(workloadRunaiRules, runaiRules.Rules...)

	// This used to be 1s interval, had to change to 30s to match grafana labs requested 2DPM.
	// Need to find a way to this this without deleting the customers jobs runtime data.
	runaiRulesShortInterval := monitoringv1.RuleGroup{
		Name:     "runai-rules-short-interval",
		Interval: ptr.To(monitoringv1.Duration("30s")),
		Rules: []monitoringv1.Rule{
			{
				Record: "runai_workload_run_time_per_30second",
				Expr:   intstr.FromString(`sum_over_time(avg(runai_workload_info_unique{detailed_status="Running"}) by (workload_name, workload_id, workload_type)[30s:1s])`),
			},
			{
				Record: "runai_run_time_seconds_per_workload",
				Expr:   intstr.FromString(`((runai_run_time_seconds_per_workload + runai_workload_run_time_per_30second) or runai_run_time_seconds_per_workload or (runai_workload_run_time_per_30second * 0)) and last_over_time(runai_workload_run_time_per_30second[120s])`),
			},
			{
				Record: "runai_workload_wait_time_per_30second",
				Expr:   intstr.FromString(`sum_over_time(avg(runai_workload_info_unique{detailed_status="Pending"}) by (workload_name, workload_id, workload_type)[30s:1s])`),
			},
			{
				Record: "runai_wait_time_seconds_per_workload",
				Expr:   intstr.FromString(`((runai_wait_time_seconds_per_workload + runai_workload_wait_time_per_30second) or runai_wait_time_seconds_per_workload or (runai_workload_wait_time_per_30second * 0)) and last_over_time(runai_workload_wait_time_per_30second[120s])`),
			},
		},
	}

	runaiRulesOneMinuteInterval := monitoringv1.RuleGroup{
		Name:     "runai-rules-1-minute-interval",
		Interval: ptr.To(monitoringv1.Duration("1m")),
		Rules: []monitoringv1.Rule{
			// Knative Serving - inference metrics:
			// Knative 1.19+ use kn_serving_invocation_duration_seconds_count/sum with labels from target_info, and label names such as k8s_namespace_name, k8s_pod_name, kn_revision_name
			// older versions use revision_app_request_count/latencies_count/sum with labels such as namespace_name, pod_name, revision_name.
			// The following records take care for backward compatibility.
			// The new metrics are pushed by Knative to the Prometheus OTLP endpoint every 1 minute, so we're using
			// a 1m interval as well, otherwise (with the 30s default sampling interval) the metric values are not stable.
			{
				// Join every status counter with one target; e.g. pod-1's 200 and 500 series both receive pod labels.
				Record: "runai_kn_serving_invocation_duration_seconds_count",
				Expr:   intstr.FromString(`label_replace(label_replace(label_replace(kn_serving_invocation_duration_seconds_count * on(instance) group_left(k8s_namespace_name, k8s_pod_name, kn_revision_name) target_info, "namespace_name", "$1", "k8s_namespace_name", "(.*)"), "pod_name", "$1", "k8s_pod_name", "(.*)"), "revision_name", "$1", "kn_revision_name", "(.*)")`),
			},
			{
				// Join every status duration with one target; e.g. pod-1's 200 and 500 series both receive pod labels.
				Record: "runai_kn_serving_invocation_duration_seconds_sum",
				Expr:   intstr.FromString(`label_replace(label_replace(label_replace(kn_serving_invocation_duration_seconds_sum * on(instance) group_left(k8s_namespace_name, k8s_pod_name, kn_revision_name) target_info, "namespace_name", "$1", "k8s_namespace_name", "(.*)"), "pod_name", "$1", "k8s_pod_name", "(.*)"), "revision_name", "$1", "kn_revision_name", "(.*)")`),
			},
			{
				// Sum status rates into one pod throughput; e.g. 1 rps (200) + 0.5 rps (500) = 1.5 rps.
				// Legacy revision_app_request_count remains the fallback.
				Record: "runai_inference_throughput_per_pod",
				Expr:   intstr.FromString(`(label_replace(sum without(http_response_status_code) (rate(runai_kn_serving_invocation_duration_seconds_count[2m])), "pod_namespace", "$1", "namespace_name", "(.*)") or on(namespace_name, pod_name) label_replace(rate(revision_app_request_count[2m]), "pod_namespace", "$1", "namespace_name", "(.*)")) * on(pod_namespace, pod_name) group_left(department, namespace, node, nodepool, pod_uuid, workload_id, workload_name, workload_type) runai_pod_info_enriched_nodepool_unique`),
			},
			{
				// Weight latency across statuses; e.g. (30s + 60s) / (60 + 30 requests) = 1s.
				// Convert new seconds to legacy milliseconds; revision_app_request_latencies_* remains the fallback.
				Record: "runai_inference_latency_per_pod",
				Expr:   intstr.FromString(`(label_replace(sum without(http_response_status_code) (rate(runai_kn_serving_invocation_duration_seconds_sum[2m])) / (sum without(http_response_status_code) (rate(runai_kn_serving_invocation_duration_seconds_count[2m])) > 0) * 1000, "pod_namespace", "$1", "namespace_name", "(.*)") or on(namespace_name, pod_name) label_replace(rate(revision_app_request_latencies_sum[2m]) / (rate(revision_app_request_latencies_count[2m]) > 0), "pod_namespace", "$1", "namespace_name", "(.*)")) * on(pod_namespace, pod_name) group_left(department, namespace, node, nodepool, pod_uuid, workload_id, workload_name, workload_type) runai_pod_info_enriched_nodepool_unique`),
			},
		},
	}

	runaiRulesHourlyInterval := monitoringv1.RuleGroup{
		Name:     "runai-rules-hourly-interval",
		Interval: ptr.To(monitoringv1.Duration("1h")),
		Rules: []monitoringv1.Rule{
			{
				Record: "runai_allocated_gpu_count_per_pod:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_allocated_gpu_count_per_pod[1h:1s])`),
			},
			{
				Record: "runai_allocated_millicpus_per_pod:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_allocated_millicpus_per_pod[1h:1s])`),
			},
			{
				Record: "runai_allocated_memory_per_pod:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_allocated_memory_per_pod[1h:1s])`),
			},
			{
				Record: "runai_used_cpu_cores_per_pod:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_used_cpu_cores_per_pod[1h:1s])`),
			},
			{
				Record: "runai_used_memory_bytes_per_pod:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_used_memory_bytes_per_pod[1h:1s])`),
			},
			{
				Record: "runai_gpu_idle_hours_per_queue:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_idle_allocated_gpu_count_per_queue[1h:1s])`),
			},
			{
				Record: "runai_allocated_gpu_count_per_project:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_allocated_gpu_count_per_project[1h:1s])`),
			},
			{
				Record: "runai_allocated_gpu_count_per_department:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_allocated_gpu_count_per_department[1h:1s])`),
			},
			{
				Record: "runai_gpu_utilization_per_project:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_gpu_utilization_per_project[1h:1s])`),
			},
			{
				Record: "runai_gpu_utilization_per_department:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_gpu_utilization_per_department[1h:1s])`),
			},
			{
				Record: "runai_gpu_memory_utilization_per_project:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_gpu_memory_utilization_per_project[1h:1s])`),
			},
			{
				Record: "runai_gpu_memory_utilization_per_department:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_gpu_memory_utilization_per_department[1h:1s])`),
			},
			{
				Record: "runai_allocated_gpu_count_per_cluster:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_allocated_gpu_count_per_cluster[1h:1s])`),
			},
			{
				Record: "runai_allocated_gpu_count_per_nodepool:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_allocated_gpu_count_per_nodepool[1h:1s])`),
			},
			{
				Record: "runai_deserved_gpu_count_per_project:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_deserved_gpu_count_per_project[1h:1s])`),
			},
			{
				Record: "runai_deserved_gpu_count_per_queue:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_deserved_gpu_count_per_queue[1h:1s])`),
			},
			{
				Record: "runai_deserved_gpu_count_per_department:hourly",
				Expr:   intstr.FromString(`sum_over_time(runai_deserved_gpu_count_per_department[1h:1s])`),
			},
		},
	}

	runaiAlerts := monitoringv1.RuleGroup{
		Name: "runai-alerts",
		Rules: []monitoringv1.Rule{
			{
				Alert:  "RunaiAgentPullRateLow",
				Expr:   intstr.FromString(fmt.Sprintf(`(rate(successful_backend_sync_count{sync_type="Pull"}[5m]) < %.2f or absent(successful_backend_sync_count{sync_type="Pull"}))`, ruleGroupsConfig.AgentPullPushRateMinForAlert)),
				For:    ptr.To(monitoringv1.Duration("10m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary": "Runai Agent fails pulling configuration from the backend (instance {{ $labels.instance }})",
				},
			},
			{
				Alert:  "RunaiAgentClusterInfoPushRateLow",
				Expr:   intstr.FromString(fmt.Sprintf(`(rate(successful_backend_sync_count{sync_type="ClusterInfoPush"}[5m]) < %.2f or absent(successful_backend_sync_count{sync_type="ClusterInfoPush"}))`, ruleGroupsConfig.AgentPullPushRateMinForAlert)),
				For:    ptr.To(monitoringv1.Duration("10m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary": "Runai cluster-sync fails pushing cluster info the backend (instance {{ $labels.instance }})",
				},
			},
			{
				Alert:  "RunaiClusterSyncHandlingRateLow",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(increase(workqueue_depth{container="cluster-sync"}[1m]) > %.2f) by (instance, container, name)`, ruleGroupsConfig.ClusterSyncQueueDepthIncreaseMaxForAlert)),
				For:    ptr.To(monitoringv1.Duration("10m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "warning"}),
				Annotations: map[string]string{
					"summary": "Runai cluster-sync has a low object handling rate (for objects of type {{ $labels.name }})",
				},
			},
			{
				Alert:  "RunaiDeploymentInsufficientReplicas",
				Expr:   intstr.FromString(`sum(min_over_time(kube_deployment_status_replicas_available{namespace=~"runai|runai-backend"}[30s])) < sum(min_over_time(kube_deployment_spec_replicas{namespace=~"runai|runai-backend"}[30s]))`),
				For:    ptr.To(monitoringv1.Duration("5m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary": "Runai Deployment has not matched the expected number of replicas.",
				},
			},
			{
				Alert:  "RunaiProjectControllerReconcileFailure",
				Expr:   intstr.FromString(`increase(controller_runtime_reconcile_errors_total{controller="project"}[10m]) > 0`),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary": "Runai project controller service had a runtime error while reconciling",
				},
			},
			{
				Alert:  "RunaiStatefulSetInsufficientReplicas",
				Expr:   intstr.FromString(`sum(min_over_time(kube_statefulset_status_replicas_ready{namespace=~"runai|runai-backend"}[30s])) < sum(min_over_time(kube_statefulset_status_replicas{namespace=~"runai|runai-backend"}[30s]))`),
				For:    ptr.To(monitoringv1.Duration("5m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary": "Runai StatefulSet has not matched the expected number of replicas.",
				},
			},
			{
				Alert:  "RunaiCriticalProblem",
				Expr:   intstr.FromString(`sum(ALERTS{alertname=~"Runai.*", severity="critical"} unless ALERTS{alertname="RunaiCriticalProblem"})> 0`),
				For:    ptr.To(monitoringv1.Duration("5m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary": "Runai platform has a critical problem",
				},
			},
			{
				Alert:  "RunaiDaemonSetRolloutStuck",
				Expr:   intstr.FromString("((kube_daemonset_status_current_number_scheduled{namespace=~\"runai|runai-backend\"}!=kube_daemonset_status_desired_number_scheduled{namespace=~\"runai|runai-backend\"}) or (kube_daemonset_status_number_misscheduled{namespace=~\"runai|runai-backend\"}!=0) or (kube_daemonset_status_updated_number_scheduled{namespace=~\"runai|runai-backend\"}!=kube_daemonset_status_desired_number_scheduled{namespace=~\"runai|runai-backend\"}) or (kube_daemonset_status_number_available{namespace=~\"runai|runai-backend\"}!=kube_daemonset_status_desired_number_scheduled{namespace=~\"runai|runai-backend\"}))"),
				For:    ptr.To(monitoringv1.Duration("5m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary":     "Runai DaemonSet rollout is stuck.",
					"description": "Runai DaemonSet {{ $labels.namespace }}/{{ $labels.daemonset }} has not finished or progressed for at least 5 minutes.",
				},
			},
			{
				Alert:  "RunaiDeploymentUnavailableReplicas",
				Expr:   intstr.FromString("kube_deployment_status_replicas_unavailable{namespace=~\"runai|runai-backend\"} > 0"),
				For:    ptr.To(monitoringv1.Duration("5m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "warning"}),
				Annotations: map[string]string{
					"summary":     "Runai Deployment has unavailable replicas.",
					"description": "Runai Deployment {{ $labels.namespace }}/{{ $labels.deployment }} has unavailable replicas.",
				},
			},
			{
				Alert:  "RunaiDeploymentNoAvailableReplicas",
				Expr:   intstr.FromString("(kube_deployment_status_replicas_available{namespace=~\"runai|runai-backend\"}==0) and (kube_deployment_spec_replicas{namespace=~\"runai|runai-backend\"}>0)"),
				For:    ptr.To(monitoringv1.Duration("5m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary":     "Runai Deployment has no available replicas.",
					"description": "Runai Deployment {{ $labels.namespace }}/{{ $labels.deployment }} has no available replicas.",
				},
			},
			{
				Alert:  "RunaiDaemonSetUnavailableOnNodes",
				Expr:   intstr.FromString("kube_daemonset_status_number_unavailable{namespace=~\"runai|runai-backend\"} > 0"),
				For:    ptr.To(monitoringv1.Duration("5m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary":     "Runai Daemonset is unavailable on nodes.",
					"description": "Runai Daemonset {{ $labels.namespace }}/{{ $labels.daemonset }} is unavailable on certain nodes.",
				},
			},
			{
				Alert:  "RunaiStatefulSetNoAvailableReplicas",
				Expr:   intstr.FromString("(kube_statefulset_status_replicas_available{namespace=~\"runai|runai-backend\"}==0) and (kube_statefulset_replicas{namespace=~\"runai|runai-backend\"}>0)"),
				For:    ptr.To(monitoringv1.Duration("5m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary":     "Runai StatefulSet has no available replicas.",
					"description": "Runai StatefulSet {{ $labels.namespace }}/{{ $labels.statefulset }} has no available replicas.",
				},
			},
			{
				Alert:  "RunaiContainerRestarting",
				Expr:   intstr.FromString("delta(kube_pod_container_status_restarts_total{namespace=~\"runai|runai-backend\"}[10m]) >= 2"),
				For:    ptr.To(monitoringv1.Duration("10s")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "warning"}),
				Annotations: map[string]string{
					"summary":     "Runai container is restarting.",
					"description": "Runai container {{ $labels.namespace }}/{{ $labels.pod }}/{{ $labels.container }} has restarted more than two times in the last 10 min.",
				},
			},
			{
				Alert:  "RunaiContainerMemoryUsageWarning",
				Expr:   intstr.FromString("((container_memory_working_set_bytes{namespace=~\"runai|runai-backend\"}>0) / on(container, pod, namespace) (kube_pod_container_resource_limits{namespace=~\"runai|runai-backend\", unit='byte'}>0)) > 0.8"),
				For:    ptr.To(monitoringv1.Duration("2m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "warning"}),
				Annotations: map[string]string{
					"summary":     "Runai container is using more than 80% of its memory limit.",
					"description": "Runai container {{ $labels.namespace }}/{{ $labels.pod }}/{{ $labels.container }} is using more than 80% of its memory limit.",
				},
			},
			{
				Alert:  "RunaiContainerMemoryUsageCritical",
				Expr:   intstr.FromString("((container_memory_working_set_bytes{namespace=~\"runai|runai-backend\"}>0) / on(container, pod, namespace) (kube_pod_container_resource_limits{namespace=~\"runai|runai-backend\", unit=\"byte\"}>0)) > 0.9"),
				For:    ptr.To(monitoringv1.Duration("30s")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary":     "Runai container is using more than 90% of its memory limit.",
					"description": "Runai container {{ $labels.namespace }}/{{ $labels.pod }}/{{ $labels.container }} is using more than 90% of its memory limit.",
				},
			},
			{
				Alert:  "RunaiCpuUsageWarning",
				Expr:   intstr.FromString("((rate(container_cpu_usage_seconds_total{namespace=~\"runai|runai-backend\"}[2m])>0) / on(container, pod, namespace) (kube_pod_container_resource_limits{namespace=~\"runai|runai-backend\", unit=\"core\"}>0)) > 0.8"),
				For:    ptr.To(monitoringv1.Duration("2m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "warning"}),
				Annotations: map[string]string{
					"summary":     "Runai container is using more than 80% of its CPU limit.",
					"description": "Runai container {{ $labels.namespace }}/{{ $labels.pod }}/{{ $labels.container }} is using more than 80% of its CPU limit.",
				},
			},
			{
				Alert:  "RunaiNodeMemoryUsageCritical",
				Expr:   intstr.FromString("runai_used_memory_bytes_per_node / runai_total_memory_bytes_per_node * 100 > 90"),
				For:    ptr.To(monitoringv1.Duration("30s")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary":     "Node memory usage is critically high",
					"description": "Node {{ $labels.node }} is using more than 90% of its memory. Current memory usage: {{ printf \"%.2f\" $value }}%",
				},
			},
			{
				Alert:  "RunaiNodeMemoryUsageWarning",
				Expr:   intstr.FromString("runai_used_memory_bytes_per_node / runai_total_memory_bytes_per_node * 100 > 80"),
				For:    ptr.To(monitoringv1.Duration("2m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "warning"}),
				Annotations: map[string]string{
					"summary":     "Node memory usage is high",
					"description": "Node {{ $labels.node }} is using more than 80% of its memory. Current memory usage: {{ printf \"%.2f\" $value }}%",
				},
			},
			{
				Alert:  "RunaiNodeUnschedulableOrNotReady",
				Expr:   intstr.FromString("runai_node_info_enriched{ready=~\"false|unknown\"} or runai_node_info_enriched{schedulable=\"false\"}"),
				For:    ptr.To(monitoringv1.Duration("2m")),
				Labels: mergeLabels(ruleGroupsConfig.AdditionalAlertLabels, map[string]string{"severity": "critical"}),
				Annotations: map[string]string{
					"summary":     "Node {{ $labels.node }} is either unschedulable or has unknown status",
					"description": "The node {{ $labels.node }} is in one of the following states:\n 1) `Ready=Unknown`: The control plane cannot communicate with the node.\n 2) `Ready=False`: The node is not healthy.\n 3)`Unschedulable=True`: The node is marked as unschedulable.",
				},
			},
		},
	}

	// vLLM v0.24.0 exposes colon-prefixed metrics. Raw counters are converted
	// to rates before pod aggregation so resets in one replica cannot affect
	// another replica.
	vllmRules := monitoringv1.RuleGroup{
		Name: "runai-vllm-rules",
		Rules: []monitoringv1.Rule{
			monitoringv1.Rule{
				Record: "runai_vllm_num_requests_running_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:num_requests_running) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_num_requests_running_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_num_requests_running_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_num_requests_waiting_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:num_requests_waiting) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_num_requests_waiting_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_num_requests_waiting_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_success_per_second_per_pod",
				Expr: intstr.FromString(vllmEnrichedPodExpr(
					`sum(rate(vllm:request_success_total[2m])) by (pod, namespace, finished_reason)`,
				)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_success_per_second_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_request_success_per_second_per_pod) by (%s, finished_reason)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_success_rate_per_pod",
				Expr:   intstr.FromString(`100 * (sum without(finished_reason) (runai_vllm_request_success_per_second_per_pod{finished_reason=~"stop|length"}) or sum without(finished_reason) (runai_vllm_request_success_per_second_per_pod) * 0) / sum without(finished_reason) (runai_vllm_request_success_per_second_per_pod)`),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_success_rate_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`100 * (sum(runai_vllm_request_success_per_second_per_workload{finished_reason=~"stop|length"}) by (%s) or sum(runai_vllm_request_success_per_second_per_workload) by (%s) * 0) / sum(runai_vllm_request_success_per_second_per_workload) by (%s)`, workloadLabels, workloadLabels, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_failure_rate_per_pod",
				Expr:   intstr.FromString(`100 * (sum without(finished_reason) (runai_vllm_request_success_per_second_per_pod{finished_reason=~"error|abort"}) or sum without(finished_reason) (runai_vllm_request_success_per_second_per_pod) * 0) / sum without(finished_reason) (runai_vllm_request_success_per_second_per_pod)`),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_failure_rate_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`100 * (sum(runai_vllm_request_success_per_second_per_workload{finished_reason=~"error|abort"}) by (%s) or sum(runai_vllm_request_success_per_second_per_workload) by (%s) * 0) / sum(runai_vllm_request_success_per_second_per_workload) by (%s)`, workloadLabels, workloadLabels, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_prompt_tokens_per_second_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(rate(vllm:prompt_tokens_total[2m])) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_prompt_tokens_per_second_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_prompt_tokens_per_second_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_generation_tokens_per_second_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(rate(vllm:generation_tokens_total[2m])) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_generation_tokens_per_second_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_generation_tokens_per_second_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_kv_cache_usage_perc_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`avg(vllm:kv_cache_usage_perc) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_kv_cache_usage_perc_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`avg(runai_vllm_kv_cache_usage_perc_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_kv_cache_usage_perc_max_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`max(runai_vllm_kv_cache_usage_perc_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_time_to_first_token_seconds_bucket_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:time_to_first_token_seconds_bucket) by (pod, namespace, le)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_time_to_first_token_seconds_bucket_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_time_to_first_token_seconds_bucket_per_pod) by (%s, le)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_time_to_first_token_seconds_sum_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:time_to_first_token_seconds_sum) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_time_to_first_token_seconds_sum_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_time_to_first_token_seconds_sum_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_time_to_first_token_seconds_count_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:time_to_first_token_seconds_count) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_time_to_first_token_seconds_count_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_time_to_first_token_seconds_count_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_e2e_request_latency_seconds_bucket_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:e2e_request_latency_seconds_bucket) by (pod, namespace, le)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_e2e_request_latency_seconds_bucket_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_e2e_request_latency_seconds_bucket_per_pod) by (%s, le)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_e2e_request_latency_seconds_sum_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:e2e_request_latency_seconds_sum) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_e2e_request_latency_seconds_sum_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_e2e_request_latency_seconds_sum_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_e2e_request_latency_seconds_count_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:e2e_request_latency_seconds_count) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_e2e_request_latency_seconds_count_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_e2e_request_latency_seconds_count_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_queue_time_seconds_bucket_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:request_queue_time_seconds_bucket) by (pod, namespace, le)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_queue_time_seconds_bucket_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_request_queue_time_seconds_bucket_per_pod) by (%s, le)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_queue_time_seconds_sum_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:request_queue_time_seconds_sum) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_queue_time_seconds_sum_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_request_queue_time_seconds_sum_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_queue_time_seconds_count_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:request_queue_time_seconds_count) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_queue_time_seconds_count_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_request_queue_time_seconds_count_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_prefill_time_seconds_bucket_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:request_prefill_time_seconds_bucket) by (pod, namespace, le)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_prefill_time_seconds_bucket_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_request_prefill_time_seconds_bucket_per_pod) by (%s, le)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_prefill_time_seconds_sum_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:request_prefill_time_seconds_sum) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_prefill_time_seconds_sum_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_request_prefill_time_seconds_sum_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_prefill_time_seconds_count_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:request_prefill_time_seconds_count) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_prefill_time_seconds_count_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_request_prefill_time_seconds_count_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_decode_time_seconds_bucket_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:request_decode_time_seconds_bucket) by (pod, namespace, le)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_decode_time_seconds_bucket_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_request_decode_time_seconds_bucket_per_pod) by (%s, le)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_decode_time_seconds_sum_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:request_decode_time_seconds_sum) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_decode_time_seconds_sum_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_request_decode_time_seconds_sum_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_decode_time_seconds_count_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:request_decode_time_seconds_count) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_request_decode_time_seconds_count_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_request_decode_time_seconds_count_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_inter_token_latency_seconds_bucket_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:inter_token_latency_seconds_bucket) by (pod, namespace, le)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_inter_token_latency_seconds_bucket_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_inter_token_latency_seconds_bucket_per_pod) by (%s, le)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_inter_token_latency_seconds_sum_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:inter_token_latency_seconds_sum) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_inter_token_latency_seconds_sum_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_inter_token_latency_seconds_sum_per_pod) by (%s)`, workloadLabels)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_inter_token_latency_seconds_count_per_pod",
				Expr:   intstr.FromString(vllmEnrichedPodExpr(`sum(vllm:inter_token_latency_seconds_count) by (pod, namespace)`)),
			},
			monitoringv1.Rule{
				Record: "runai_vllm_inter_token_latency_seconds_count_per_workload",
				Expr:   intstr.FromString(fmt.Sprintf(`sum(runai_vllm_inter_token_latency_seconds_count_per_pod) by (%s)`, workloadLabels)),
			},
		},
	}

	rgs = append(rgs, newGenRules, dashboardRules, runaiRules, vllmRules, runaiRulesShortInterval, runaiRulesOneMinuteInterval, runaiRulesHourlyInterval, runaiAlerts)

	return rgs
}

func vllmEnrichedPodExpr(expr string) string {
	return fmt.Sprintf(
		`label_replace(label_replace(%s, "pod_name", "$1", "pod", "(.*)"), "pod_namespace", "$1", "namespace", "(.*)") * on(pod_name, pod_namespace) group_left(%s) runai_pod_info_enriched_nodepool_unique`,
		expr,
		podLabelsWithoutPodKeys,
	)
}

func getOomKillerMetric(metricName string, labels string) string {
	return fmt.Sprintf(`sum by(%s) (label_replace(%s, "pod_ip", "$1", "instance", "(.*):(.*)") * on (pod_ip) group_left(node) topk(1, kube_pod_info{pod_ip!=""}) by (pod_ip) * on (node) group_left(nodepool, modelName) runai_node_info_enriched)`, labels, metricName)
}

func AdvancedRuleGroup() monitoringv1.RuleGroup {
	return monitoringv1.RuleGroup{
		Name: "advanced-gpu-rules",
		Rules: []monitoringv1.Rule{
			{
				Record: "runai_dcgm_gr_engine_active",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_PROF_GR_ENGINE_ACTIVE*100")),
			},
			{
				Record: "runai_gpu_gr_engine_active_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_gr_engine_active")),
			},
			{
				Record: "runai_gpu_gr_engine_active_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_gr_engine_active")),
			},
			{
				Record: "runai_gpu_gr_engine_active_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_gpu_gr_engine_active_per_whole_pod_per_gpu", "runai_gpu_gr_engine_active_per_mig_pod", "")),
			},
			{
				Record: "runai_gpu_gr_engine_active_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_gpu_gr_engine_active_per_pod_per_gpu", true)),
			},
			{
				Record: "runai_gpu_gr_engine_active_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`%s or on(UUID) avg by (%s) (runai_dcgm_gr_engine_active)`, migGpuAvgMetricRule("runai_gpu_gr_engine_active_per_mig_pod"), gpuLabels)),
			},
			{
				Record: "runai_gpu_gr_engine_active_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_gpu_gr_engine_active_per_gpu", true)),
			},

			{
				Record: "runai_dcgm_dram_active",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_PROF_DRAM_ACTIVE*100")),
			},
			{
				Record: "runai_gpu_dram_active_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_dram_active")),
			},
			{
				Record: "runai_gpu_dram_active_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_dram_active")),
			},
			{
				Record: "runai_gpu_dram_active_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_gpu_dram_active_per_whole_pod_per_gpu", "runai_gpu_dram_active_per_mig_pod", "")),
			},
			{
				Record: "runai_gpu_dram_active_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_gpu_dram_active_per_pod_per_gpu", true)),
			},
			{
				Record: "runai_gpu_dram_active_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`%s or on(UUID) avg by (%s) (runai_dcgm_dram_active)`, migGpuAvgMetricRule("runai_gpu_dram_active_per_mig_pod"), gpuLabels)),
			},
			{
				Record: "runai_gpu_dram_active_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_gpu_dram_active_per_gpu", true)),
			},

			{
				Record: "runai_dcgm_sm_active",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_PROF_SM_ACTIVE*100")),
			},
			{
				Record: "runai_gpu_sm_active_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_sm_active")),
			},
			{
				Record: "runai_gpu_sm_active_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_sm_active")),
			},
			{
				Record: "runai_gpu_sm_active_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_gpu_sm_active_per_whole_pod_per_gpu", "runai_gpu_sm_active_per_mig_pod", "")),
			},
			{
				Record: "runai_gpu_sm_active_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_gpu_sm_active_per_pod_per_gpu", true)),
			},
			{
				Record: "runai_gpu_sm_active_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`%s or on(UUID) avg by (%s) (runai_dcgm_sm_active)`, migGpuAvgMetricRule("runai_gpu_sm_active_per_mig_pod"), gpuLabels)),
			},
			{
				Record: "runai_gpu_sm_active_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_gpu_sm_active_per_gpu", true)),
			},

			{
				Record: "runai_dcgm_sm_occupancy",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_PROF_SM_OCCUPANCY*100")),
			},
			{
				Record: "runai_gpu_sm_occupancy_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_sm_occupancy")),
			},
			{
				Record: "runai_gpu_sm_occupancy_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_sm_occupancy")),
			},
			{
				Record: "runai_gpu_sm_occupancy_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_gpu_sm_occupancy_per_whole_pod_per_gpu", "runai_gpu_sm_occupancy_per_mig_pod", "")),
			},
			{
				Record: "runai_gpu_sm_occupancy_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_gpu_sm_occupancy_per_pod_per_gpu", true)),
			},
			{
				Record: "runai_gpu_sm_occupancy_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`%s or on(UUID) avg by (%s) (runai_dcgm_sm_occupancy)`, migGpuAvgMetricRule("runai_gpu_sm_occupancy_per_mig_pod"), gpuLabels)),
			},
			{
				Record: "runai_gpu_sm_occupancy_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_gpu_sm_occupancy_per_gpu", true)),
			},

			{
				Record: "runai_dcgm_pipe_tensor_active",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_PROF_PIPE_TENSOR_ACTIVE*100")),
			},
			{
				Record: "runai_gpu_pipe_tensor_active_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_pipe_tensor_active")),
			},
			{
				Record: "runai_gpu_pipe_tensor_active_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_pipe_tensor_active")),
			},
			{
				Record: "runai_gpu_pipe_tensor_active_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_gpu_pipe_tensor_active_per_whole_pod_per_gpu", "runai_gpu_pipe_tensor_active_per_mig_pod", "")),
			},
			{
				Record: "runai_gpu_pipe_tensor_active_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_gpu_pipe_tensor_active_per_pod_per_gpu", true)),
			},
			{
				Record: "runai_gpu_pipe_tensor_active_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`%s or on(UUID) avg by (%s) (runai_dcgm_pipe_tensor_active)`, migGpuAvgMetricRule("runai_gpu_pipe_tensor_active_per_mig_pod"), gpuLabels)),
			},
			{
				Record: "runai_gpu_pipe_tensor_active_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_gpu_pipe_tensor_active_per_gpu", true)),
			},

			{
				Record: "runai_dcgm_pipe_fp64_active",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_PROF_PIPE_FP64_ACTIVE*100")),
			},
			{
				Record: "runai_gpu_pipe_fp64_active_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_pipe_fp64_active")),
			},
			{
				Record: "runai_gpu_pipe_fp64_active_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_pipe_fp64_active")),
			},
			{
				Record: "runai_gpu_pipe_fp64_active_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_gpu_pipe_fp64_active_per_whole_pod_per_gpu", "runai_gpu_pipe_fp64_active_per_mig_pod", "")),
			},
			{
				Record: "runai_gpu_pipe_fp64_active_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_gpu_pipe_fp64_active_per_pod_per_gpu", true)),
			},
			{
				Record: "runai_gpu_pipe_fp64_active_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`%s or on(UUID) avg by (%s) (runai_dcgm_pipe_fp64_active)`, migGpuAvgMetricRule("runai_gpu_pipe_fp64_active_per_mig_pod"), gpuLabels)),
			},
			{
				Record: "runai_gpu_pipe_fp64_active_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_gpu_pipe_fp64_active_per_gpu", true)),
			},

			{
				Record: "runai_dcgm_pipe_fp32_active",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_PROF_PIPE_FP32_ACTIVE*100")),
			},
			{
				Record: "runai_gpu_pipe_fp32_active_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_pipe_fp32_active")),
			},
			{
				Record: "runai_gpu_pipe_fp32_active_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_pipe_fp32_active")),
			},
			{
				Record: "runai_gpu_pipe_fp32_active_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_gpu_pipe_fp32_active_per_whole_pod_per_gpu", "runai_gpu_pipe_fp32_active_per_mig_pod", "")),
			},
			{
				Record: "runai_gpu_pipe_fp32_active_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_gpu_pipe_fp32_active_per_pod_per_gpu", true)),
			},
			{
				Record: "runai_gpu_pipe_fp32_active_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`%s or on(UUID) avg by (%s) (runai_dcgm_pipe_fp32_active)`, migGpuAvgMetricRule("runai_gpu_pipe_fp32_active_per_mig_pod"), gpuLabels)),
			},
			{
				Record: "runai_gpu_pipe_fp32_active_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_gpu_pipe_fp32_active_per_gpu", true)),
			},

			{
				Record: "runai_dcgm_pipe_fp16_active",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_PROF_PIPE_FP16_ACTIVE*100")),
			},
			{
				Record: "runai_gpu_pipe_fp16_active_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_pipe_fp16_active")),
			},
			{
				Record: "runai_gpu_pipe_fp16_active_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_pipe_fp16_active")),
			},
			{
				Record: "runai_gpu_pipe_fp16_active_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_gpu_pipe_fp16_active_per_whole_pod_per_gpu", "runai_gpu_pipe_fp16_active_per_mig_pod", "")),
			},
			{
				Record: "runai_gpu_pipe_fp16_active_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_gpu_pipe_fp16_active_per_pod_per_gpu", true)),
			},
			{
				Record: "runai_gpu_pipe_fp16_active_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`%s or on(UUID) avg by (%s) (runai_dcgm_pipe_fp16_active)`, migGpuAvgMetricRule("runai_gpu_pipe_fp16_active_per_mig_pod"), gpuLabels)),
			},
			{
				Record: "runai_gpu_pipe_fp16_active_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_gpu_pipe_fp16_active_per_gpu", true)),
			},

			{
				Record: "runai_dcgm_nvlink_tx_bytes",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_PROF_NVLINK_TX_BYTES")),
			},
			{
				Record: "runai_gpu_nvlink_tx_bytes_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_nvlink_tx_bytes")),
			},
			{
				Record: "runai_gpu_nvlink_tx_bytes_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_nvlink_tx_bytes")),
			},
			{
				Record: "runai_gpu_nvlink_tx_bytes_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_gpu_nvlink_tx_bytes_per_whole_pod_per_gpu", "runai_gpu_nvlink_tx_bytes_per_mig_pod", "")),
			},
			{
				Record: "runai_gpu_nvlink_tx_bytes_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_gpu_nvlink_tx_bytes_per_pod_per_gpu", false)),
			},
			{
				Record: "runai_gpu_nvlink_tx_bytes_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by (%s) (runai_dcgm_nvlink_tx_bytes)`, gpuLabels)),
			},
			{
				Record: "runai_gpu_nvlink_tx_bytes_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_gpu_nvlink_tx_bytes_per_gpu", false)),
			},

			{
				Record: "runai_dcgm_nvlink_rx_bytes",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_PROF_NVLINK_RX_BYTES")),
			},
			{
				Record: "runai_gpu_nvlink_rx_bytes_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_nvlink_rx_bytes")),
			},
			{
				Record: "runai_gpu_nvlink_rx_bytes_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_nvlink_rx_bytes")),
			},
			{
				Record: "runai_gpu_nvlink_rx_bytes_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_gpu_nvlink_rx_bytes_per_whole_pod_per_gpu", "runai_gpu_nvlink_rx_bytes_per_mig_pod", "")),
			},
			{
				Record: "runai_gpu_nvlink_rx_bytes_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_gpu_nvlink_rx_bytes_per_pod_per_gpu", false)),
			},
			{
				Record: "runai_gpu_nvlink_rx_bytes_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by (%s) (runai_dcgm_nvlink_rx_bytes)`, gpuLabels)),
			},
			{
				Record: "runai_gpu_nvlink_rx_bytes_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_gpu_nvlink_rx_bytes_per_gpu", false)),
			},

			{
				Record: "runai_dcgm_pcie_tx_bytes",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_PROF_PCIE_TX_BYTES")),
			},
			{
				Record: "runai_gpu_pcie_tx_bytes_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_pcie_tx_bytes")),
			},
			{
				Record: "runai_gpu_pcie_tx_bytes_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_pcie_tx_bytes")),
			},
			{
				Record: "runai_gpu_pcie_tx_bytes_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_gpu_pcie_tx_bytes_per_whole_pod_per_gpu", "runai_gpu_pcie_tx_bytes_per_mig_pod", "")),
			},
			{
				Record: "runai_gpu_pcie_tx_bytes_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_gpu_pcie_tx_bytes_per_pod_per_gpu", false)),
			},
			{
				Record: "runai_gpu_pcie_tx_bytes_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by (%s) (runai_dcgm_pcie_tx_bytes)`, gpuLabels)),
			},
			{
				Record: "runai_gpu_pcie_tx_bytes_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_gpu_pcie_tx_bytes_per_gpu", false)),
			},

			{
				Record: "runai_dcgm_pcie_rx_bytes",
				Expr:   intstr.FromString(dcgmMetricsRule("DCGM_FI_PROF_PCIE_RX_BYTES")),
			},
			{
				Record: "runai_gpu_pcie_rx_bytes_per_whole_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerWholePodRule("runai_dcgm_pcie_rx_bytes")),
			},
			{
				Record: "runai_gpu_pcie_rx_bytes_per_mig_pod",
				Expr:   intstr.FromString(gpuMetricsPerMigPodRule("runai_dcgm_pcie_rx_bytes")),
			},
			{
				Record: "runai_gpu_pcie_rx_bytes_per_pod_per_gpu",
				Expr:   intstr.FromString(gpuMetricsPerPodRule("runai_gpu_pcie_rx_bytes_per_whole_pod_per_gpu", "runai_gpu_pcie_rx_bytes_per_mig_pod", "")),
			},
			{
				Record: "runai_gpu_pcie_rx_bytes_per_workload",
				Expr:   intstr.FromString(metricsPerWorkloadRule("runai_gpu_pcie_rx_bytes_per_pod_per_gpu", false)),
			},
			{
				Record: "runai_gpu_pcie_rx_bytes_per_gpu",
				Expr:   intstr.FromString(fmt.Sprintf(`sum by (%s) (runai_dcgm_pcie_rx_bytes)`, gpuLabels)),
			},
			{
				Record: "runai_gpu_pcie_rx_bytes_per_node",
				Expr:   intstr.FromString(gpuMetricsPerNodeRule("runai_gpu_pcie_rx_bytes_per_gpu", false)),
			},
		},
	}
}

func amdMetricsRule(amdMetric string) string {
	return fmt.Sprintf(`label_join(label_replace(label_replace(label_replace(label_replace(%s, "node", "$1", "node_id", "(.*)"), "modelName", "$1", "productname", "(.*)"), "gpu", "$1", "gpu_id", "(.*)"), "pod_uuid", "$1", "pod", "(.*)"), "UUID", "-", "node", "gpu") * on(node) group_left(nodepool) runai_node_nodepool_excluded`, amdMetric)
}

func dcgmMetricsRule(dcgmMetric string) string {
	return fmt.Sprintf(`label_replace(label_replace(label_replace(label_replace(%s, "pod_ip", "$1", "instance", "(.*):(.*)"), "pod_name", "$1", "exported_pod", "(.+)"), "pod_namespace", "$1", "exported_namespace", "(.+)"), "gpu_instance_id",
	"$1", "GPU_I_ID", "(.+)") * on (pod_ip) group_left(node) kube_pod_info{created_by_name=~".*dcgm-exporter", pod_ip!=""} * on(node) group_left(nodepool) runai_node_nodepool_excluded`, dcgmMetric)
}

func gpuMetricsPerWholePodRule(runaiDcgmMetric string) string {
	return fmt.Sprintf(`runai_pod_info_enriched_nodepool_unique_with_gpu_for_core_pods{%s} * on(pod_name, pod_namespace) group_right(%s) %s`,
		wholePodFilter, metricsPerWholePodRuleLabels, runaiDcgmMetric)
}

func gpuMetricsPerWholeAmdPodRule(runaiAmdMetric string) string {
	return fmt.Sprintf(`runai_pod_info_enriched_nodepool_unique_with_gpu_for_core_pods{%s} * on(pod_uuid) group_right(%s) %s`, wholePodFilter, podLabelsForAmdMetric, runaiAmdMetric)
}

func gpuMetricsPerFractionalPodRule(runaiFractionalMetric string) string {
	return fmt.Sprintf(`runai_pod_info_enriched_nodepool_unique_with_gpu_for_core_pods{%s} * on (pod_uuid, gpu) group_right(%s) %s or on(pod_uuid) runai_pod_info_enriched_nodepool_unique_with_gpu_for_core_pods{%s}*0`, fractionalPodFilter, podLabelsWithoutGpu, runaiFractionalMetric, fractionalPodFilter)
}

func gpuMetricsPerMigPodRule(runaiDcgmMetric string) string {
	return fmt.Sprintf(`runai_pod_info_enriched_nodepool_unique_with_gpu_for_core_pods{%s} * on(node, gpu, gpu_instance_id) group_right(%s) %s`, migPodFilter, metricsPerMigPodRuleLabels, runaiDcgmMetric)
}

func migGpuAvgMetricRule(migGpuMetric string) string {
	return fmt.Sprintf(`sum by (%s) (%s * on(pod_uuid) group_left() runai_allocated_gpu_count_per_pod)`, gpuLabels, migGpuMetric)
}

func gpuMetricsPerPodRule(wholePodsGpuMetric string, migPodsGpuMetrics string, fractionalPodsGpuMetric string) string {
	if fractionalPodsGpuMetric != "" {
		return fmt.Sprintf(`sum by(%s) (%s or on (pod_uuid) %s or on (pod_uuid) %s)`, podLabels, fractionalPodsGpuMetric, migPodsGpuMetrics, wholePodsGpuMetric)
	}
	return fmt.Sprintf(`sum by(%s) (%s or on (pod_uuid) %s or on (pod_uuid) (runai_pod_info_enriched_nodepool_unique_with_gpu_for_core_pods{%s}*0))`, podLabels, migPodsGpuMetrics, wholePodsGpuMetric, fractionalPodFilter)
}

func metricsPerWorkloadRule(podGpuMetric string, isAvg bool) string {
	if isAvg {
		return fmt.Sprintf(`avg(%s) by (%s) * on(%s) group_left(%s) runai_workload_info_enriched`, podGpuMetric, workloadPodLabels, workloadIdLabel, workloadLabelsNoWorkloadId)
	}
	return fmt.Sprintf(`sum(%s) by (%s) * on(%s) group_left(%s) runai_workload_info_enriched`, podGpuMetric, workloadPodLabels, workloadIdLabel, workloadLabelsNoWorkloadId)
}

func gpuMetricsPerNodeRule(gpuMeticPerGpu string, isAvg bool) string {
	if isAvg {
		return fmt.Sprintf(`avg by(%s) (%s)`, nodeLabels, gpuMeticPerGpu)
	}
	return fmt.Sprintf(`sum by(%s) (%s)`, nodeLabels, gpuMeticPerGpu)
}

func mergeLabels(labels, overrideLabels map[string]string) map[string]string {
	mergedMap := map[string]string{}

	if labels != nil {
		for k, v := range labels {
			mergedMap[k] = v
		}
	}
	if overrideLabels != nil {
		for k, v := range overrideLabels {
			mergedMap[k] = v
		}
	}
	return mergedMap
}
