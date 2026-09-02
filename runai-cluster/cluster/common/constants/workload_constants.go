/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

import (
	"time"
)

const (
	NamespaceRunAi       = "runai"
	RunaiNamespacePrefix = NamespaceRunAi + "-"

	// RunaiOrgDomain is the run:ai DNS-style organization prefix used for
	// `run.ai/*` label keys and for composing the project-controller's
	// finalizer string (`run.ai.finalizers.project`).
	RunaiOrgDomain = "run.ai"

	// RunaiNamespaceVersionLabel is the label key project-controller writes
	// on managed namespaces to track their schema version.
	RunaiNamespaceVersionLabel = "runai/namespace-version"

	GroupKubeflow   = "kubeflow.org"
	VersionKubeflow = "v1alpha2"

	AppsV1       = "apps/v1"
	CoreV1       = "v1"
	NetworkingV1 = "networking.k8s.io/v1"
	GatewayV1    = "gateway.networking.k8s.io/v1"
	RouteV1      = "route.openshift.io/v1"

	IngressBodySizeAnnotationKey = "nginx.ingress.kubernetes.io/proxy-body-size"

	// IngressBodySizeAnnotationDefaultValue is set to  "8g" which proved itself to be a reasonable default for the body size for jupyter notebooks
	IngressBodySizeAnnotationDefaultValue = "8g"

	JobType             = "JobType"
	KindPvc             = "PersistentVolumeClaim"
	KindSecret          = "Secret"
	KindVolume          = "Volume"
	KindService         = "Service"
	KindRoute           = "Route"
	KnativeService      = "KnativeService"
	KindComputeDomain   = "ComputeDomain"
	KindLeaderWorkerSet = "LeaderWorkerSet"

	DistributedMpiJob     = "MPIJob"
	DistributedPyTorchJob = "PyTorchJob"
	DistributedTfJob      = "TFJob"
	DistributedXGBoostJob = "XGBoostJob"

	KindRunaiJob       = "RunaiJob"
	KindWorkloadRunner = "WorkloadRunner"
	KindMpiJob         = "MPIJob"
	KindPyTorchJob     = "PyTorchJob"
	KindTFJob          = "TFJob"
	KindXGBoostJob     = "XGBoostJob"
	KindJAXJob         = "JAXJob"
	KindNIMService     = "NIMService"
	KindDeployment     = "Deployment"
	KindIngress        = "Ingress"
	KindHTTPRoute      = "HTTPRoute"
	KindGRPCRoute      = "GRPCRoute"
	KindConfigMap      = "ConfigMap"

	KindInteractivePolicy            = "InteractivePolicy"
	KindInteractiveWorkload          = "InteractiveWorkload"
	KindTrainingPolicy               = "TrainingPolicy"
	KindTrainingWorkload             = "TrainingWorkload"
	KindInferenceWorkload            = "InferenceWorkload"
	KindInferencePolicy              = "InferencePolicy"
	KindDistributedWorkload          = "DistributedWorkload"
	KindDistributedPolicy            = "DistributedPolicy"
	KindDistributedInferencePolicy   = "DistributedInferencePolicy"
	KindDistributedInferenceWorkload = "DistributedInferenceWorkload"

	WorkloadCategoryLabel           = "run.ai/category"
	WorkloadsConfigMapName          = "workloads-config"
	WorkloadsConfigMapNamespace     = NamespaceRunAi
	WorkloadsConfigMapTypesKey      = "types"
	WorkloadsConfigMapCategoriesKey = "categories"
	WorkloadsTypesConfigMetricLabel = "workloads_config"

	KindTrainingPolicyList    = "TrainingPolicyList"
	KindInteractivePolicyList = "InteractivePolicyList"

	KindInferencePolicyList = "InferencePolicyList"

	AppRunaiJob   = "runaijob"
	AppMpiJob     = "mpijob"
	AppPyTorchJob = "pytorchjob"
	AppTFJob      = "tfjob"
	AppJAXJob     = "jaxjob"
	AppXGBoostJob = "xgboostjob"
	AppConfigMap  = "configmap"

	Spec = "spec"

	AnnotationTemplate = "template"
	AnnotationUserName = "user"
	AnnotationImage    = "image"

	AnnotationGpuFraction                    = "gpu-fraction"
	AnnotationGpuFractionDevices             = "gpu-fraction-num-devices"
	AnnotationGpuMemory                      = "gpu-memory"
	ValidationBypassAnnotationVolumeKey      = "run.ai/patch-volumes"
	ValidationBypassAnnotationVolumeMountKey = "run.ai/patch-volume-mounts"
	AnnotationMigProfilePrefix               = "nvidia.com/mig-"
	AnnotationJobUrl                         = "JobUrl"
	AnnotationS3GuidPrefix                   = "s3-bucket-guid"
	AnnotationRequestedNodeAffinity          = "run.ai/requested-node-affinity"

	RunaiProjectLabel                = "project"
	LabelRelease                     = "release"
	LabelTrainingKubeFlow            = "training.kubeflow.org/job-name"
	LabelApp                         = "app"
	RunaiQueueLabel                  = "runai/queue"
	LabelCreatedBy                   = "createdBy"
	LabelJupyter                     = "jupyter"
	LabelTensorboard                 = "tensorboard"
	LabelWorkloadPrefix              = "run.ai/"
	LabelMasterNodeRole              = "node-role.kubernetes.io/master"
	LabelKnativeServiceUrlVisibility = "networking.knative.dev/visibility"
	LabelWorkloadMetricsService      = "run.ai/workload-metrics-service"
	LabelNIMWorkload                 = "run.ai/nim-workload"
	LabelNIMWorkloadMetricsService   = "run.ai/nim-workload-metrics-service"
	LabelVLLMWorkload                = "run.ai/vllm-workload"
	LabelVLLMWorkloadMetricsService  = "run.ai/vllm-workload-metrics-service"

	LabelWorkloadName      = "workloadName"
	LabelWorkloadKind      = "workloadKind"
	RunaiWorkloadNameLabel = "run.ai/workload-name"

	AnnotationWorkloadApiVersion = "workloadApiVersion"

	NodeTypeKey = "run.ai/type"

	LabelPriorityClassName              = "priorityClassName"
	InteractivePreemptiblePriorityClass = "interactive-preemptible"
	InteractivePriorityClass            = "build"

	LabelPreemptibilityName = "kai.scheduler/preemptibility"

	RunaiScheduler = "runai-scheduler"

	// NodePoolNameLabel is the label key used to associate a resource with a node pool.
	NodePoolNameLabel = "runai/node-pool"

	// GPU resource names.
	NvidiaGpuResourceName = "nvidia.com/gpu"
	AmdGpuResourceName    = "amd.com/gpu"
	GpuResourceName       = "run.ai/gpu"
	GpuMemoryResourceName = "run.ai/gpu.memory"

	// GPU node labels.
	GpuCountLabel    = "nvidia.com/gpu.count"
	NvidiaGPUProduct = "nvidia.com/gpu.product"
	NvidiaGpuMemory  = "nvidia.com/gpu.memory"

	// GPU binding / sharing constants.
	GPUGroup                 = "runai-gpu-group"
	MultiGpuGroupLabelPrefix = GPUGroup + "/"

	// Received resource type annotation values (set by the scheduler/binder on a pod).
	ReceivedTypeMigInstance = "MigInstance"
	ReceivedTypeFraction    = "Fraction"

	// NvidiaVisibleDevices is the env var name holding the visible GPU devices for a container.
	NvidiaVisibleDevices = "NVIDIA_VISIBLE_DEVICES"

	// PodBoundCondition is the pod condition type set once a pod is bound by the binder.
	PodBoundCondition = "PodBound"

	// Resource reservation pod attributes used by the binder's resource-reservation config.
	ResourceReservationServiceAccountName = "runai-reservation"
	ResourceReservationAppLabelValue      = "runai-reservation"

	PullPolicyAlways       = "Always"
	PullPolicyIfNotPresent = "IfNotPresent"
	PullPolicyNever        = "Never"

	ServiceTypeLoadBalancer = "LoadBalancer"
	ServiceTypeNodePort     = "NodePort"
	ServiceTypeClusterIP    = "ClusterIP"
	ServiceTypeServingPort  = "ServingPort"

	ProtocolGRPC = "grpc"
	ProtocolHTTP = "http"

	SeccompProfileTypeRuntimeDefault = "RuntimeDefault"
	SeccompProfileTypeUnconfined     = "Unconfined"

	HomeDirEnvName      = "HOME"
	RunaiDefaultHomeDir = "/home/runai-home"

	MpiNumWorkersEnvName = "RUNAI_MPI_NUM_WORKERS"

	DefaultBackoffLimit = 6
	DefaultCompletions  = 1
	DefaultParallelism  = 1

	DefaultWorkers = 1
	MinWorkers     = 1

	DefaultSlotsPerWorker = 1
	MinSlotsPerWorker     = 1

	DefaultDistributedInferenceWorkers = 1
	MinDistributedInferenceWorkers     = 0

	DefaultDistributedInferenceReplicas = 1
	MinDistributedInferenceReplicas     = 0

	DistributedInferenceStartupPolicyLeaderCreated = "LeaderCreated"
	DistributedInferenceStartupPolicyLeaderReady   = "LeaderReady"
	DefaultDistributedInferenceStartupPolicy       = DistributedInferenceStartupPolicyLeaderCreated

	DistributedInferenceRestartPolicyRecreateGroupOnPodRestart = "RecreateGroupOnPodRestart"
	DistributedInferenceRestartPolicyNone                      = "None"
	DefaultDistributedInferenceRestartPolicy                   = DistributedInferenceRestartPolicyRecreateGroupOnPodRestart

	AddressTypeInternalIp = "InternalIP"

	GpuCountLimitEnvVar  = "RUNAI_GPU_COUNT_LIMIT"
	GpuMemoryLimitEnvVar = "RUNAI_GPU_MEMORY_LIMIT"

	RunaiPublicConfigMapName = "runai-public"
	TolerationsNoExecute     = "NoExecute"
	TolerationsNoSchedule    = "NoSchedule"
	TolerationsPreferNoSched = "PreferNoSchedule"
	TolerationsEqual         = "Equal"
	TolerationsExists        = "Exists"

	// env variables providing uid/gid for jupyter

	JupyterNbUid = "NB_UID"
	JupyterNbGid = "NB_GID"

	RequiredDuringSchedulingIgnoredDuringExecutionValue   = "requiredDuringSchedulingIgnoredDuringExecution"
	RequiredDuringSchedulingIgnoredDuringExecutionUiName  = "Required"
	PreferredDuringSchedulingIgnoredDuringExecutionValue  = "preferredDuringSchedulingIgnoredDuringExecution"
	PreferredDuringSchedulingIgnoredDuringExecutionUiName = "Preferred"

	WorkloadControllerOperandName          = "workload-controller"
	InferenceWorkloadControllerOperandName = "inference-workload-controller"
	ResearcherServiceOperandName           = "researcher-service"
	RuntimeRulesOperandName                = "runtime-rules"

	RateLimiterBaseDelay = time.Second
	RateLimiterMaxDelay  = time.Minute

	ExternalUrlAuthContainerDefaultPort = int32(4180)
	ExternalUrlAuthVolumeName           = "runai-auth"
	ExternalUrlAuthVolumePath           = "/etc/runai/auth/users"

	LabelMarkedUnschedulableByNodePoolController = "runai/unschedulable"

	// Workloads Auth
	WorkloadAuthSecretName           = "workloads-auth"
	WorkloadAuthCookieSecretKey      = "OAUTH2_PROXY_COOKIE_SECRET"
	WorkloadAuthClientSecretKey      = "OAUTH2_PROXY_CLIENT_SECRET"
	WorkloadAuthRedisPasswordKey     = "OAUTH2_PROXY_REDIS_PASSWORD"
	WorkloadControllerNodeNameEnvKey = "WORKLOAD_CONTROLLER_NODE_NAME"

	// knative constants
	KnativeServiceApi        = "serving.knative.dev/v1"
	KnativeServiceGroup      = "serving.knative.dev"
	KnativeServiceVersion    = "v1"
	KnativeServiceKind       = "Service"
	KnativeConfigurationKind = "Configuration"
	KnativeRevisionKind      = "Revision"

	// inference workload serving port auth
	ServingPortAuthorizationTypePublic                  = "public"
	ServingPortAuthorizationTypeAuthenticatedUsers      = "authenticatedUsers"
	ServingPortAuthorizationTypeAuthorizedUsersOrGroups = "authorizedUsersOrGroups"
	ServingPortAuthorizationTypeAuthorizedUsers         = "authorizedUsers"
	ServingPortAuthorizationTypeAuthorizedGroups        = "authorizedGroups"
	ServingPortAuthorizationTypeDefault                 = ServingPortAuthorizationTypePublic

	// inference workload model swap
	InferenceWorkloadModelSwapAnnotationVolumeKey = "run.ai/inference-workload-model-swap"

	// phase labels for not owned resources
	RunaiPhaseAnnotationKey        = "run.ai/phase"
	RunaiPhaseMessageAnnotationKey = "run.ai/phase-message"
	RunaiPhaseVersionAnnotationKey = "run.ai/phase-version"

	// WorkloadRunner lifecycle annotations stamped by the runtime-rules on
	// the WorkloadRunner itself to mark Running↔terminal transitions, plus the
	// auto-deletion duration the control-plane stamps on the wrapped child at
	// submission. The runtime-rules runner converter reads these into the unified
	// JobWorkload shape so the same downstream enforcement code runs unchanged.
	StartTimeAnnotation                       = "run.ai/start-time"
	CompletionTimeAnnotation                  = "run.ai/completion-time"
	AutoDeletionTimeAfterCompletionAnnotation = "run.ai/auto-delete-after-completion-sec"

	// MNNVL
	ComputeDomainStatusReady    = "Ready"
	ComputeDomainStatusNotReady = "NotReady"

	// lws constants
	LwsGroup   = "leaderworkerset.x-k8s.io"
	LwsVersion = "v1"
	LwsKind    = "LeaderWorkerSet"

	ProjectNameLabel    = "run.ai/project"
	ProjectIdLabel      = "run.ai/project-id"
	DepartmentIdLabel   = "run.ai/department-id"
	DepartmentNameLabel = "run.ai/department-name"

	ApplicationIdLabel                  = "run.ai/application-id"
	ApplicationNameLabel                = "run.ai/application-name"
	ControlPlaneApplicationIdAnnotation = "run.ai/application-id"
	ResourceReleaseAnnotationName       = "meta.helm.sh/release-name"

	TopologyPreferredPlacementAnnotation = "kai.scheduler/topology-preferred-placement"
	TopologyRequiredPlacementAnnotation  = "kai.scheduler/topology-required-placement"
	TopologyAnnotation                   = "kai.scheduler/topology"
	KAITopologyAnnotationPrefix          = "kai.scheduler/"

	PodGroupRequestedMigDevicesAnnotation = "runai-podgroup-requested-mig-devices"

	HeimdallWorkflowIdLabel = "heimdall.nvidia.com/workflow-id"
	DisableAJRAnnotation    = "run.ai/disable-ajr-integration"
)
