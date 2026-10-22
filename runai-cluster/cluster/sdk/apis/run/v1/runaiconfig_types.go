/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	"reflect"

	accessrule_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/accessrule-controller"
	advice_restart_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/advice-restart-controller"
	anyworkload_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/anyworkload-controller"
	application_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/application-controller"
	binder "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/binder"
	cluster_api "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/cluster-api"
	cluster_redis "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/cluster-redis"
	device_plugin "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/device-plugin"
	engine "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/engine"
	external_workload_integrator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/external-workload-integrator"
	inference_endpoint_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/inference-endpoint-controller"
	kai_operator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/kai-operator"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/knative"
	krm_operator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/krm-operator"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/lws"
	metrics_exporter "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/metrics-exporter"
	node_scale_adjuster "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/node-scale-adjuster"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/prometheus"
	queue_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/queue-controller"
	internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	shared_objects_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/shared-objects-controller"
	workload_exporter "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/workload-exporter"
	"github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/numa_placement_exporter"
	kaiprom "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/prometheus"

	admission_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/admission-controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/agent"
	cluster_sync "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/cluster-sync"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	init_ca "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/init-ca"
	limit_range "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/limit-range"
	mpi_opeator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/mpi-operator"
	mps_server "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/mps-server"
	nodepool_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/nodepool-controller"
	pod_group_assigner "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/pod-group-assigner"
	pod_group_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/pod-group-controller"
	pod_grouper "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/pod-grouper"
	project_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/project-controller"
	runtime_rules "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runtime-rules"

	researcher_service "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/researcher-service"
	runai_container_toolkit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-container-toolkit"
	runai_container_toolkit_reservation "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-container-toolkit-reservation"
	runai_node_exporter "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-node-exporter"
	runai_scheduler "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-scheduler"
	status_updater "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/status-updater"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GroupPrefix = Group
	Finalizer   = GroupPrefix + "/finalize"

	NamespacedNameAnnotation = GroupPrefix + "/primary-resource"
	TypeAnnotation           = GroupPrefix + "/primary-resource-type"

	ReconciliationTriggerConfigmapNameSuffix = "reconciliation-trigger"
	ReconciliationTriggerRequestKey          = "reconcile-request"

	DefaultNamespace         = "runai"
	DefaultName              = DefaultNamespace
	RunaiPublicConfigMapName = "runai-public"
	RunaiConfigKind          = "RunaiConfig"

	ReservationNamespace = "runai-reservation"
	ScaleAdjustNamespace = "runai-scale-adjust"
)

const (
	ConditionTypeDeployed                      ConditionType = "Deployed"
	ConditionTypeAvailable                     ConditionType = "Available"
	ConditionDependenciesFulfilled             ConditionType = "DependenciesFulfilled"
	ConditionTypeLegacyWorkloadCleanupComplete ConditionType = "LegacyWorkloadCleanupComplete"
)

const (
	Deployed    ConditionReason = "deployed"
	NotDeployed ConditionReason = "not_deployed"

	Available    ConditionReason = "available"
	NotAvailable ConditionReason = "not_available"

	Reconciled    ConditionReason = "reconciled"
	NotReconciled ConditionReason = "not_reconciled"

	DependenciesFulfilled    ConditionReason = "dependencies_fulfilled"
	DependenciesNotFulfilled ConditionReason = "dependencies_not_fulfilled"

	LegacyWorkloadCleanupWaitingForReady ConditionReason = "WaitingForReady"
	LegacyWorkloadCleanupCreating        ConditionReason = "Creating"
	LegacyWorkloadCleanupRunning         ConditionReason = "Running"
	LegacyWorkloadCleanupFailed          ConditionReason = "Failed"
	LegacyWorkloadCleanupBlocked         ConditionReason = "Blocked"
	LegacyWorkloadCleanupCompleted       ConditionReason = "Completed"
)

// DeletionPolicy tells the operator if the managed resource should
// be deleted upon deletion
type DeletionPolicy string

const (
	DeletionPolicyAnnotationKey                = GroupPrefix + "/deletion-policy"
	DeletionPolicyRetain        DeletionPolicy = "Retain"
)

// UpdatePolicy tells the operator if the managed resource should
// be updated upon update
type UpdatePolicy string

const (
	UpdatePolicyAnnotationKey              = GroupPrefix + "/update-policy"
	UpdatePolicySkip          UpdatePolicy = "Skip"
)

// ServiceGroup represents which scaling group a service belongs to
type ServiceGroup string

const (
	SchedulingServicesGroup ServiceGroup = "scheduling"
	SyncServicesGroup       ServiceGroup = "sync"
	WorkloadServicesGroup   ServiceGroup = "workload"
)

// ServiceGroupMapping defines which services belong to which group.
// This is the single source of truth used by both Resources and VPA resolution.
// The key is the field name in RunaiConfigSpec.
var ServiceGroupMapping = map[string]ServiceGroup{
	// Scheduling Services
	"RunAIScheduler":        SchedulingServicesGroup,
	"StatusUpdater":         SchedulingServicesGroup,
	"MetricsExporter":       SchedulingServicesGroup,
	"PodGrouper":            SchedulingServicesGroup,
	"PodGroupAssigner":      SchedulingServicesGroup,
	"PodGroupController":    SchedulingServicesGroup,
	"NumaPlacementExporter": SchedulingServicesGroup,
	"QueueController":       SchedulingServicesGroup,
	"NodePoolController":    SchedulingServicesGroup,
	"Binder":                SchedulingServicesGroup,

	// Sync Services
	"Agent":       SyncServicesGroup,
	"ClusterSync": SyncServicesGroup,

	// Workload Services
	"RuntimeRules":                WorkloadServicesGroup,
	"AnyworkloadController":       WorkloadServicesGroup,
	"ExternalWorkloadIntegrator":  WorkloadServicesGroup,
	"InferenceEndpointController": WorkloadServicesGroup,
	"ClusterRedis":                WorkloadServicesGroup,
	"ClusterApi":                  WorkloadServicesGroup,
	"ResearcherService":           WorkloadServicesGroup,
	"SharedObjectsController":     WorkloadServicesGroup,
	"WorkloadExporter":            WorkloadServicesGroup,
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RunaiConfigSpec defines the desired state of RunAIConfig
type RunaiConfigSpec struct {
	// Note to Dev: if you are changing anything in this struct, please make sure to update
	// the script in "cluster/new-service-template/hooks/post_gen_project.py" (add_app_to_runaiconfig_spec func)

	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// LimitRange TODO
	// +kubebuilder:validation:Optional
	LimitRange *limit_range.LimitRange `json:"limitRange,omitempty"`

	// Global specifies global configuration fields
	// +kubebuilder:validation:Optional
	Global *global.Global `json:"global,omitempty"`

	// RunAIScheduler specifies configuration for the runai-scheduler
	// +kubebuilder:validation:Optional
	RunAIScheduler *runai_scheduler.RunAIScheduler `json:"runai-scheduler,omitempty"`

	// Agent specifies configuration for the RunAI agent
	// +kubebuilder:validation:Optional
	Agent *agent.Agent `json:"agent,omitempty"`

	// AdmissionController specifies configuration for the admission-controller
	// +kubebuilder:validation:Optional
	AdmissionController *admission_controller.AdmissionController `json:"admission-controller,omitempty"`

	// InitCA specifies configuration for the init-ca job
	// +kubebuilder:validation:Optional
	InitCA *init_ca.InitCA `json:"init-ca,omitempty"`

	// MPSServer specifies configuration for the mps-server
	// +kubebuilder:validation:Optional
	MPSServer *mps_server.MPSServer `json:"mps-server,omitempty"`

	// NodeScaleAdjuster specifies configuration for the node-scale-adjuster
	// +kubebuilder:validation:Optional
	NodeScaleAdjuster *node_scale_adjuster.NodeScaleAdjuster `json:"node-scale-adjuster,omitempty"`

	// PodGrouper specifies configuration for the pod-grouper
	// +kubebuilder:validation:Optional
	PodGrouper *pod_grouper.PodGrouper `json:"pod-grouper,omitempty"`

	// PodGroupAssigner specifies configuration for the pod-group-assigner
	// +kubebuilder:validation:Optional
	PodGroupAssigner *pod_group_assigner.PodGroupAssigner `json:"pod-group-assigner,omitempty"`

	// PodGroupController specifies configuration for the pod-group-controller
	// +kubebuilder:validation:Optional
	PodGroupController *pod_group_controller.PodGroupController `json:"pod-group-controller,omitempty"`

	// NumaPlacementExporter specifies configuration for the numa-placement-exporter
	// +kubebuilder:validation:Optional
	NumaPlacementExporter *numa_placement_exporter.NumaPlacementExporter `json:"numa-placement-exporter,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	RunAIContainerToolkit *runai_container_toolkit.RunAIContainerToolkit `json:"runai-container-toolkit,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	RunAIContainerToolkitReservation *runai_container_toolkit_reservation.RunAIContainerToolkitReservation `json:"runai-container-toolkit-reservation,omitempty"`

	// ProjectController specifies configuration for the project-controller
	// +kubebuilder:validation:Optional
	ProjectController *project_controller.ProjectController `json:"project-controller,omitempty"`

	// QueueController specifies configuration for the queue-controller
	// +kubebuilder:validation:Optional
	QueueController *queue_controller.QueueController `json:"queue-controller,omitempty"`

	// Binder specifies configuration for the binder
	// +kubebuilder:validation:Optional
	Binder *binder.Binder `json:"binder,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	DevicePlugin *device_plugin.DevicePlugin `json:"device-plugin,omitempty"`

	// NodePoolController specifies configuration for the nodepool-controller
	// +kubebuilder:validation:Optional
	NodePoolController *nodepool_controller.NodePoolController `json:"nodepool-controller,omitempty"`

	// AccessRuleController specifies configuration for the accessrule-controller
	// +kubebuilder:validation:Optional
	AccessRuleController *accessrule_controller.AccessRuleController `json:"accessrule-controller,omitempty"`

	// ResearcherService specifies configuration for the researcher-service
	// +kubebuilder:validation:Optional
	ResearcherService *researcher_service.ResearcherService `json:"researcher-service,omitempty"`

	// ClusterSync specifies configuration for the cluster-sync component
	// +kubebuilder:validation:Optional
	ClusterSync *cluster_sync.ClusterSync `json:"cluster-sync,omitempty"`

	// MpiOperator specifies configuration for the mpi-operator component
	// +kubebuilder:validation:Optional
	MpiOperator *mpi_opeator.MpiOperator `json:"mpi-operator,omitempty"`

	// RuntimeRules specifies configuration for the runtime-rules component
	// +kubebuilder:validation:Optional
	RuntimeRules *runtime_rules.RuntimeRulesConfiguration `json:"runtime-rules,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	Engine *engine.Engine `json:"engine,omitempty"`

	// Deprecated: retained as inert data for rollback to 2.26.
	// +kubebuilder:validation:Optional
	NodeExporter *runai_node_exporter.NodeExporter `json:"runai-node-exporter,omitempty"`

	// KRMOperator specifies configuration for the KRM operator operand
	// +kubebuilder:validation:Optional
	KRMOperator *krm_operator.KRMOperator `json:"krm-operator,omitempty"`

	// KAIOperator specifies configuration for the KAI operator operand
	// +kubebuilder:validation:Optional
	KAIOperator *kai_operator.KAIOperator `json:"kai-operator,omitempty"`

	// Prometheus specifies the Prometheus configurations set by the operator
	// +kubebuilder:validation:Optional
	Prometheus *prometheus.Prometheus `json:"prometheus,omitempty"`

	// HistoricalUsagePrometheus specifies configuration for the historical-usage-prometheus component
	// +kubebuilder:validation:Optional
	HistoricalUsagePrometheus *kaiprom.Prometheus `json:"historical-usage-prometheus,omitempty"`

	// MetricsExporter specifies configuration for the metrics-exporter
	// +kubebuilder:validation:Optional
	MetricsExporter *metrics_exporter.MetricsExporter `json:"metrics-exporter,omitempty"`

	// StatusUpdater specifies configuration for the status-updater
	// +kubebuilder:validation:Optional
	StatusUpdater *status_updater.StatusUpdater `json:"status-updater,omitempty"`

	// ExternalWorkloadIntegrator specifies configuration for the external-workload-integrator
	// +kubebuilder:validation:Optional
	ExternalWorkloadIntegrator *external_workload_integrator.ExternalWorkloadIntegrator `json:"external-workload-integrator,omitempty"`

	// AdviceRestartController specifies configuration for the advice-restart-controller
	// +kubebuilder:validation:Optional
	AdviceRestartController *advice_restart_controller.AdviceRestartController `json:"advice-restart-controller,omitempty"`

	// InferenceEndpointController configures inference routing on the workload cluster.
	// +kubebuilder:validation:Optional
	InferenceEndpointController *inference_endpoint_controller.InferenceEndpointController `json:"inference-endpoint-controller,omitempty"`

	// AnyworkloadController specifies configuration for the anyworkload-controller
	// +kubebuilder:validation:Optional
	AnyworkloadController *anyworkload_controller.AnyworkloadController `json:"anyworkload-controller,omitempty"`

	// ApplicationController specifies configuration for the application-controller
	// +kubebuilder:validation:Optional
	ApplicationController *application_controller.ApplicationController `json:"application-controller,omitempty"`

	// SharedObjectsController specifies configuration for the shared-objects-controller
	// +kubebuilder:validation:Optional
	SharedObjectsController *shared_objects_controller.SharedObjectsController `json:"shared-objects-controller,omitempty"`

	// ClusterApi specifies configuration for the cluster-api
	// +kubebuilder:validation:Optional
	ClusterApi *cluster_api.ClusterApi `json:"cluster-api,omitempty"`

	// Knative specifies configuration for knative serving
	// +kubebuilder:validation:Optional
	Knative *knative.Knative `json:"knative,omitempty"`

	// LWS specifies configuration for the LWS (LeaderWorkerSet) operator component
	// +kubebuilder:validation:Optional
	LWS *lws.LWS `json:"lws,omitempty"`

	// WorkloadExporter specifies configuration for the workload-exporter
	// +kubebuilder:validation:Optional
	WorkloadExporter *workload_exporter.WorkloadExporter `json:"workload-exporter,omitempty"`

	// ClusterRedis specifies configuration for the cluster redis
	// +kubebuilder:validation:Optional
	ClusterRedis *cluster_redis.ClusterRedis `json:"cluster-redis,omitempty"`

	// Internal specifies internal configuration for the Run:AI deployment. Do not manually set this field.
	// +kubebuilder:validation:Optional
	Internal *internal.Internal `json:"__internal,omitempty" jsonschema:"-"`
}

// SetDefaultsWhereNeeded sets default for unset fields
// Note to Dev: if you are changing anything in this func, please make sure to update
// the script in "cluster/new-service-template/hooks/post_gen_project.py" (add_app_to_runaiconfig_set_defaults_where_needed func)
func (runaiConfig *RunaiConfig) SetDefaultsWhereNeeded() {
	runaiConfig.TypeMeta.APIVersion = GroupVersion.Group + "/" + GroupVersion.Version
	runaiConfig.TypeMeta.Kind = RunaiConfigKind

	spec := &runaiConfig.Spec
	if spec.Global == nil {
		spec.Global = &global.Global{}
	}
	if spec.LimitRange == nil {
		spec.LimitRange = &limit_range.LimitRange{}
	}
	if spec.RunAIScheduler == nil {
		spec.RunAIScheduler = &runai_scheduler.RunAIScheduler{}
	}
	if spec.Agent == nil {
		spec.Agent = &agent.Agent{}
	}
	if spec.AdmissionController == nil {
		spec.AdmissionController = &admission_controller.AdmissionController{}
	}
	if spec.InitCA == nil {
		spec.InitCA = &init_ca.InitCA{}
	}
	if spec.MPSServer == nil {
		spec.MPSServer = &mps_server.MPSServer{}
	}
	if spec.NodeScaleAdjuster == nil {
		spec.NodeScaleAdjuster = &node_scale_adjuster.NodeScaleAdjuster{}
	}
	if spec.PodGrouper == nil {
		spec.PodGrouper = &pod_grouper.PodGrouper{}
	}
	if spec.PodGroupAssigner == nil {
		spec.PodGroupAssigner = &pod_group_assigner.PodGroupAssigner{}
	}
	if spec.PodGroupController == nil {
		spec.PodGroupController = &pod_group_controller.PodGroupController{}
	}
	if spec.ProjectController == nil {
		spec.ProjectController = &project_controller.ProjectController{}
	}
	if spec.QueueController == nil {
		spec.QueueController = &queue_controller.QueueController{}
	}
	if spec.Binder == nil {
		spec.Binder = &binder.Binder{}
	}
	if spec.NodePoolController == nil {
		spec.NodePoolController = &nodepool_controller.NodePoolController{}
	}
	if spec.AccessRuleController == nil {
		spec.AccessRuleController = &accessrule_controller.AccessRuleController{}
	}
	if spec.ResearcherService == nil {
		spec.ResearcherService = &researcher_service.ResearcherService{}
	}
	if spec.ClusterSync == nil {
		spec.ClusterSync = &cluster_sync.ClusterSync{}
	}
	if spec.MpiOperator == nil {
		spec.MpiOperator = &mpi_opeator.MpiOperator{}
	}
	if spec.RuntimeRules == nil {
		spec.RuntimeRules = &runtime_rules.RuntimeRulesConfiguration{}
	}
	if spec.KRMOperator == nil {
		spec.KRMOperator = &krm_operator.KRMOperator{}
	}
	if spec.KAIOperator == nil {
		spec.KAIOperator = &kai_operator.KAIOperator{}
	}
	if spec.Prometheus == nil {
		spec.Prometheus = &prometheus.Prometheus{}
	}
	if spec.MetricsExporter == nil {
		spec.MetricsExporter = &metrics_exporter.MetricsExporter{}
	}
	if spec.StatusUpdater == nil {
		spec.StatusUpdater = &status_updater.StatusUpdater{}
	}
	if spec.ExternalWorkloadIntegrator == nil {
		spec.ExternalWorkloadIntegrator = &external_workload_integrator.ExternalWorkloadIntegrator{}
	}
	if spec.InferenceEndpointController == nil {
		spec.InferenceEndpointController = &inference_endpoint_controller.InferenceEndpointController{}
	}
	if spec.AdviceRestartController == nil {
		spec.AdviceRestartController = &advice_restart_controller.AdviceRestartController{}
	}
	if spec.AnyworkloadController == nil {
		spec.AnyworkloadController = &anyworkload_controller.AnyworkloadController{}
	}
	if spec.ApplicationController == nil {
		spec.ApplicationController = &application_controller.ApplicationController{}
	}
	if spec.SharedObjectsController == nil {
		spec.SharedObjectsController = &shared_objects_controller.SharedObjectsController{}
	}
	if spec.ClusterApi == nil {
		spec.ClusterApi = &cluster_api.ClusterApi{}
	}
	if spec.Knative == nil {
		spec.Knative = &knative.Knative{}
	}
	if spec.LWS == nil {
		spec.LWS = &lws.LWS{}
	}
	if spec.WorkloadExporter == nil {
		spec.WorkloadExporter = &workload_exporter.WorkloadExporter{}
	}
	if spec.ClusterRedis == nil {
		spec.ClusterRedis = &cluster_redis.ClusterRedis{}
	}
	if spec.Internal == nil {
		spec.Internal = &internal.Internal{}
	}

	spec.Internal.SetDefaultsWhereNeeded()
	spec.Global.SetDefaultsWhereNeeded(spec.Internal)
	setCommonResourceRequirementsWhereNeeded(spec)
	setCommonVPAWhereNeeded(spec)
	spec.LimitRange.SetDefaultsWhereNeeded()
	spec.RunAIScheduler.SetDefaultsWhereNeeded()
	spec.Agent.SetDefaultsWhereNeeded(spec.Internal, spec.Global.ReplicaCount)
	spec.AdmissionController.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.InitCA.SetDefaultsWhereNeeded(spec.Internal, spec.Global.ReplicaCount)
	spec.MPSServer.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry)
	spec.PodGroupAssigner.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.ProjectController.SetDefaultsWhereNeeded(spec.Internal, spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.NodePoolController.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.AccessRuleController.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.ResearcherService.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.ClusterSync.SetDefaultsWhereNeeded(spec.Internal, spec.Global.ReplicaCount)
	spec.MpiOperator.SetDefaultsWhereNeeded()
	spec.RuntimeRules.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.KRMOperator.SetDefaultsWhereNeeded()
	spec.KAIOperator.SetDefaultsWhereNeeded()
	spec.QueueController.SetDefaultsWhereNeeded()
	spec.WorkloadExporter.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.Prometheus.SetDefaultsWhereNeeded(spec.Internal, spec.Agent.SyncIntervalMillis, spec.Global.ReplicaCount)
	spec.MetricsExporter.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.StatusUpdater.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.ExternalWorkloadIntegrator.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.InferenceEndpointController.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.AdviceRestartController.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry)
	spec.AnyworkloadController.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry)
	spec.ApplicationController.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.SharedObjectsController.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.ClusterApi.SetDefaultsWhereNeeded(spec.Internal.Global.ImageRegistry, spec.Global.ReplicaCount)
	spec.Knative.SetDefaultsWhereNeeded()
	spec.LWS.SetDefaultsWhereNeeded()
	spec.ClusterRedis.SetDefaultsWhereNeeded(spec.Internal, spec.Global)
}

// getGroupResources returns the group-level Resources for a given ServiceGroup
func getGroupResources(spec *RunaiConfigSpec, group ServiceGroup) *global.Resources {
	switch group {
	case SchedulingServicesGroup:
		if spec.Global.SchedulingServices != nil {
			return spec.Global.SchedulingServices.Resources
		}
	case SyncServicesGroup:
		if spec.Global.SyncServices != nil {
			return spec.Global.SyncServices.Resources
		}
	case WorkloadServicesGroup:
		if spec.Global.WorkloadServices != nil {
			return spec.Global.WorkloadServices.Resources
		}
	}
	return nil
}

func setCommonResourceRequirementsWhereNeeded(spec *RunaiConfigSpec) {
	specValue := reflect.ValueOf(spec).Elem()

	for serviceName, group := range ServiceGroupMapping {
		serviceField := specValue.FieldByName(serviceName)
		if !serviceField.IsValid() || serviceField.IsNil() {
			continue
		}

		groupResources := getGroupResources(spec, group)
		if groupResources == nil {
			continue
		}

		setServiceResources(serviceField, groupResources)
	}
}

func setServiceResources(serviceField reflect.Value, resources *global.Resources) {
	resourcesField := serviceField.Elem().FieldByName("Resources")
	if !resourcesField.IsValid() {
		return
	}

	if resourcesField.IsNil() {
		resourcesField.Set(reflect.New(resourcesField.Type().Elem()))
	}

	// Special handling for researcher-service which has nested Resources.ResearcherService structure
	researcherServiceField := resourcesField.Elem().FieldByName("ResearcherService")
	if researcherServiceField.IsValid() {
		researcherServiceField.FieldByName("Requests").Set(reflect.ValueOf(resources.Requests))
		researcherServiceField.FieldByName("Limits").Set(reflect.ValueOf(resources.Limits))
	} else {
		resourcesField.Elem().FieldByName("Requests").Set(reflect.ValueOf(resources.Requests))
		resourcesField.Elem().FieldByName("Limits").Set(reflect.ValueOf(resources.Limits))
	}
}

func getGroupVPA(spec *RunaiConfigSpec, group ServiceGroup) *global.VPASpec {
	switch group {
	case SchedulingServicesGroup:
		if spec.Global.SchedulingServices != nil {
			return spec.Global.SchedulingServices.VPA
		}
	case SyncServicesGroup:
		if spec.Global.SyncServices != nil {
			return spec.Global.SyncServices.VPA
		}
	case WorkloadServicesGroup:
		if spec.Global.WorkloadServices != nil {
			return spec.Global.WorkloadServices.VPA
		}
	}
	return nil
}

func setCommonVPAWhereNeeded(spec *RunaiConfigSpec) {
	specValue := reflect.ValueOf(spec).Elem()

	for serviceName, group := range ServiceGroupMapping {
		serviceField := specValue.FieldByName(serviceName)
		if !serviceField.IsValid() || serviceField.IsNil() {
			continue
		}

		vpaField := serviceField.Elem().FieldByName("VPA")
		if !vpaField.IsValid() {
			continue
		}

		if !vpaField.IsNil() {
			continue
		}

		groupVPA := getGroupVPA(spec, group)
		if groupVPA != nil {
			vpaField.Set(reflect.ValueOf(groupVPA))
			continue
		}

		if spec.Global != nil && spec.Global.VPA != nil {
			vpaField.Set(reflect.ValueOf(spec.Global.VPA))
		}
	}
}

// OperandStatus specifies the status of an operand.
// And operand manages resources, some of which are not immediately available
// and need their status monitored, such as Deployments, Daemonsets, Ingressess etc.
// Some resources are immediately available and should not be monitored once accepted
// by the cluster, such as Services, ConfigMaps, Secrets etc.
// If all resources of an operand are ready, a Status with Ready set to 'true' and
// a nil 'Reasons' slice should be returned, otherwise, a Status with 'Reasons'
// slice containig the reason why the Operand is not ready (Deployment pods are not ready
// for example.)
type OperandStatus struct {
	// Ready specifies if the operand is ready or not
	Ready bool `json:"ready"`

	// Reasons specifies the reasons why the operand is not ready
	Reasons []string `json:"reasons,omitempty"`
}

type TimedOperandStatus struct {
	OperandStatus `json:",inline"`

	// LastTransitionTime specifies the last time the operand readiness changed
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// RunaiConfigStatus defines the observed state of RunAIConfig
type RunaiConfigStatus struct {
	RunaiCustomTypeStatus `json:",inline"`

	// OperandStatuses specifies the status of the managed operands
	OperandStatuses map[string]TimedOperandStatus `json:"operands,omitempty" yaml:"operands,omitempty"`

	// RemoteConfig actual configuration available to cluster
	RemoteConfig *internal.ClusterFullConfig `json:"remoteConfig,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Available",type="string",JSONPath=".status.conditions[?(@.type=='Available')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// RunaiConfig is the Schema for the runaiconfigs API
type RunaiConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunaiConfigSpec   `json:"spec,omitempty"`
	Status RunaiConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RunaiConfigList contains a list of RunAIConfig
type RunaiConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RunaiConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RunaiConfig{}, &RunaiConfigList{})
}

func (runaiConfig *RunaiConfig) GetRunaiCustomStatus() *RunaiCustomTypeStatus {
	return &runaiConfig.Status.RunaiCustomTypeStatus
}

func (runaiConfig *RunaiConfig) IsRunaiConfigAvailableAndReconciled() bool {
	status := runaiConfig.GetRunaiCustomStatus()
	if len(status.Conditions) == 0 {
		return false
	}
	reconcileLoopEndedSuccessfully := true
	for _, condition := range status.Conditions {
		if condition.Type == string(ConditionTypeReconciled) {
			reconcileLoopEndedSuccessfully = reconcileLoopEndedSuccessfully && condition.Status == metav1.ConditionTrue
		}
		if condition.Type == string(ConditionTypeAvailable) {
			reconcileLoopEndedSuccessfully = reconcileLoopEndedSuccessfully && condition.Status == metav1.ConditionTrue
		}
	}
	return reconcileLoopEndedSuccessfully
}

func (runaiConfig *RunaiConfig) ExtractReconciledConditionMessage() string {
	status := runaiConfig.GetRunaiCustomStatus()
	reconcileCondition := status.GetCondition(ConditionTypeReconciled)
	if reconcileCondition == nil {
		return ConditionTypeReconciledLoopHaveNotRun
	}
	return reconcileCondition.Message
}

func (runaiConfig *RunaiConfig) ShouldSkipDependencyChecks() bool {
	return runaiConfig.Spec.Internal.Global.NMC.IsEnabled() || runaiConfig.Spec.Internal.Development.SkipDependencyChecks
}
