/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package consts

// Inference consts
const (
	InferenceShortMsg             = "inference management"
	DistributedInferenceShortMsg  = "Runs multiple coordinated inference processes across multiple nodes. Required for models too large to run on a single node."
	FollowingInference            = "To track the workload's status, run 'runai inference describe %s'"
	FollowingDistributedInference = "To track the workload's status, run 'runai inference distributed describe %s'"
	NimInferenceShortMsg          = "[Experimental] Runs NVIDIA NIM (NVIDIA Inference Microservices) workloads. Optimized for deploying foundation models."
	FollowingNimInference         = "To track the workload's status, run 'runai inference nim describe %s'"
)

// Submit Inference const's
const (
	// serving-port
	AttributeNameServingPort = "serving-port"

	// serving-port for Inference
	AttributeUsageServingPort = "Defines various attributes for the serving port. " +
		"Usage formats: " +
		"(1) Simplified format: --serving-port=CONTAINER_PORT " +
		"(2) Full format: --serving-port=container=CONTAINER_PORT," +
		"[authorization-type=public|authenticatedUsers|authorizedUsersOrGroups],[authorized-users=USER1:USER2:APP1...],[authorized-groups=GROUP1:GROUP2...],[cluster-local-access-only],[protocol=http|grpc]"

	// auto-scaling
	AttributeNameMinReplicas  = "min-replicas"
	AttributeUsageMinReplicas = "The minimum number of replicas for autoscaling. Defaults to 1. Use 0 to allow scale-to-zero"

	AttributeNameMaxReplicas  = "max-replicas"
	AttributeUsageMaxReplicas = "The maximum number of replicas for autoscaling. Defaults to minReplicas, or to 1 if minReplicas is set to 0"

	AttributeNameInitialReplicas  = "initial-replicas"
	AttributeUsageInitialReplicas = "The number of replicas to run when initializing the workload for the first time. Defaults to minReplicas, or to 1 if minReplicas is set to 0"

	AttributeNameActivationReplicas  = "activation-replicas"
	AttributeUsageActivationReplicas = "The number of replicas to run when scaling-up from zero. Defaults to minReplicas, or to 1 if minReplicas is set to 0"

	AttributeNameConcurrencyHardLimit  = "concurrency-hard-limit"
	AttributeUsageConcurrencyHardLimit = "The maximum number of requests allowed to flow to a single replica at any time. 0 means no limit"

	AttributeNameScaleToZeroRetentionSeconds  = "scale-to-zero-retention-seconds"
	AttributeUsageScaleToZeroRetentionSeconds = "The minimum amount of time (in seconds) that the last replica will remain active after a scale-to-zero decision. Defaults to 0. Available only if minReplicas is set to 0"

	AttributeNameScaleDownDelaySeconds  = "scale-down-delay-seconds"
	AttributeUsageScaleDownDelaySeconds = "The minimum amount of time (in seconds) that a replica will remain active after a scale-down decision"

	AttributeNameInitializationTimeoutSeconds  = "initialization-timeout-seconds"
	AttributeUsageInitializationTimeoutSeconds = "The maximum amount of time (in seconds) to wait for the container to become ready"

	AttributeNameRequestTimeoutSeconds  = "request-timeout-seconds"
	AttributeUsageRequestTimeoutSeconds = "The maximum time (in seconds) allowed to process an end-user request. If no response is returned within this time, the request will be ignored"

	AttributeNameMetric  = "metric"
	AttributeUsageMetric = "Autoscaling metric is required if minReplicas < maxReplicas, except when minReplicas = 0 and maxReplicas = 1. Use 'throughput', 'concurrency', 'latency', or custom metrics."

	AttributeNameMetricThreshold  = "metric-threshold"
	AttributeUsageMetricThreshold = "The threshold to use with the specified metric for autoscaling. Mandatory if metric is specified"

	AttributeNameMetricThresholdPercentage  = "metric-threshold-percentage"
	AttributeUsageMetricThresholdPercentage = "The percentage of metric threshold value to use for autoscaling. Defaults to 70. Applicable only with the 'throughput' and 'concurrency' metrics"

	AttributeNameScaleWindowSeconds  = "scale-window-seconds"
	AttributeUsageScaleWindowSeconds = "The duration (in seconds) for which the autoscaler considers past metrics when making scaling decisions."

	AttributeNameNgcSecret       = "ngc-auth-secret"
	AttributeUsageNgcSecret      = "Name of the secret containing NGC API key"
	AttributeNameModelPvc        = "model-existing-pvc"
	AttributeUsageModelPvc       = "Mount an existing PVC for the model store. Format: claimname=CLAIM_NAME"
	AttributeNameModelNewPvc     = "model-new-pvc"
	AttributeUsageModelNewPvc    = "Create and mount a new PVC for the model store. Format: [claimname=CLAIM_NAME],size=SIZE,[storageclass=STORAGE_CLASS],[accessmode-rwo|accessmode-rwm|accessmode-rom],[ro]"
	AttributeNameModelCache      = "model-nim-cache"
	AttributeUsageModelCache     = "NIM Cache configuration. Format: name=CACHE_NAME,[profile=PROFILE]"
	AttributeNameAutoscale       = "autoscale"
	AttributeUsageAutoscale      = "Enable autoscaling (mutually exclusive with --replicas)"
	AttributeNameReadinessProbe  = "readiness-probe"
	AttributeUsageReadinessProbe = "Readiness probe configuration. Format: port=PORT,[path=PATH],[initial-delay=SECONDS],[period=SECONDS]," +
		"[timeout=SECONDS],[success=THRESHOLD],[failure=THRESHOLD],[host=HOST],[scheme=HTTP|HTTPS]"
	AttributeUsageReadinessProbeNim = "Readiness probe. Format: port=PORT,[path=PATH],[host=HOST],[scheme=HTTP|HTTPS],[initial-delay=SECONDS]," +
		"[period=SECONDS],[timeout=SECONDS],[success=THRESHOLD],[failure=THRESHOLD]. Port is required"
	AttributeNameImagePullSecrets  = "image-pull-secret"
	AttributeUsageImagePullSecrets = "Image pull secrets"
	AttributeNameCategory          = "category"
	AttributeUsageCategory         = "Workload category"
)

// Distributed Inference specific flags
const (
	// serving-port for Distributed Inference
	AttributeUsageServingPortDistributed = "Defines various attributes for the serving port. " +
		"Usage formats: " +
		"(1) Simplified format: --serving-port=CONTAINER_PORT " +
		"(2) Full format: --serving-port=container=CONTAINER_PORT," +
		"[authorization-type=public|authenticatedUsers|authorizedUsers|authorizedGroups],[authorized-users=USER1:USER2:APP1...]," +
		"[authorized-groups=GROUP1:GROUP2...],[expose-externally=true|false],[exposed-url=URL],[protocol=http]"

	AttributeUsageServingPortNim = "Serving port options. " +
		"Simplified: --serving-port=PORT. " +
		"Full: --serving-port=port=PORT,[grpc-port=GRPC_PORT],[metrics-port=METRICS_PORT],[expose-externally],[exposed-url=URL],[exposed-protocol=http|grpc],[service-type=ClusterIP|NodePort|LoadBalancer|ExternalName]"

	// Distributed Inference base/leader args and command
	AttributeNameLeaderArgs     = "arguments"
	AttributeUsageLeaderArgs    = "Specifies the arguments to pass to the container command"
	AttributeNameLeaderCommand  = "command"
	AttributeUsageLeaderCommand = "Specifies the command to run in the container, overriding the image's default entrypoint"

	AttributeNameReplicas           = "replicas"
	AttributeUsageReplicas          = "The number of replicas (sets of leader and workers) to run. Defaults to 1"
	AttributeUsageReplicasNim       = "Number of replicas to run"
	AttributeUsageReplicasNimUpdate = "Number of replicas. When switching from autoscaling, use --remove autoscaling"

	AttributeUsageMinReplicasNim   = "Minimum number of replicas for autoscaling"
	AttributeUsageMaxReplicasNim   = "Maximum number of replicas for autoscaling. Defaults to min-replicas or 1"
	AttributeUsageMetricNim        = "Autoscaling metric (e.g. cpu, http_requests_total). Required when min-replicas < max-replicas"
	AttributeNameWorkerAsLeader    = "worker-as-leader"
	AttributeUsageWorkerAsLeader   = "Use leader configuration for workers unless overridden with --worker-* flags. Set to false when workers require explicit configuration."
	AttributeNameStartupPolicy     = "startup-policy"
	AttributeUsageStartupPolicy    = "Controls when worker pods start. Valid values: LeaderCreated (workers start when leader is created), LeaderReady (workers start when leader is ready). Defaults to LeaderCreated"
	AttributeNameMnnvl             = "mnnvl"
	AttributeUsageMnnvl            = "Multi-Node NVLINK (MNNVL) mode controls GPU network acceleration for distributed workloads. Valid values: None, Preferred, Required"
	AttributeNameWorkerAnnotation  = "worker-annotation"
	AttributeUsageWorkerAnnotation = "Set of annotations to populate into worker pods running the workload"
	AttributeNameWorkerLabel       = "worker-label"
	AttributeUsageWorkerLabel      = "Set of labels to populate into worker pods running the workload"

	// Worker Scheduling Configuration
	AttributeNameWorkerNodeType                 = "worker-node-type"
	AttributeUsageWorkerNodeType                = "Enforce node type affinity for worker pods by setting a node-type label."
	AttributeNameWorkerExcludeNode              = "worker-exclude-node"
	AttributeUsageWorkerExcludeNode             = "Nodes that will be excluded from use by worker pods. Format: --worker-exclude-node node-a --worker-exclude-node node-b"
	AttributeNameWorkerToleration               = "worker-toleration"
	AttributeUsageWorkerToleration              = "Add Kubernetes tolerations for worker pods. Format: operator=Equal|Exists,key=KEY,[value=VALUE],[effect=NoSchedule|NoExecute|PreferNoSchedule],[seconds=SECONDS]."
	AttributeNameWorkerRequiredPodTopologyKey   = "worker-required-pod-topology-key"
	AttributeUsageWorkerRequiredPodTopologyKey  = "Require scheduling worker pods within the same topology domain defined by a Kubernetes node label key. Format: KEY."
	AttributeNameWorkerPreferredPodTopologyKey  = "worker-preferred-pod-topology-key"
	AttributeUsageWorkerPreferredPodTopologyKey = "If possible, schedule all worker pods within the same topology domain defined by a Kubernetes node label key. Format: KEY."

	// Worker Container Configuration
	AttributeNameWorkerImage            = "worker-image"
	AttributeUsageWorkerImage           = "The container image to use for worker pods."
	AttributeNameWorkerImagePullPolicy  = "worker-image-pull-policy"
	AttributeUsageWorkerImagePullPolicy = "Image pull policy for worker pods. Valid values: Always, IfNotPresent, Never."
	AttributeNameWorkerCommand          = "worker-command"
	AttributeUsageWorkerCommand         = "Specifies the command to run in worker pods, overriding the image's default entrypoint"
	AttributeNameWorkerArgs             = "worker-arguments"
	AttributeUsageWorkerArgs            = "Specifies the arguments to pass to worker pod command"
	AttributeNameWorkerWorkingDir       = "worker-working-dir"
	AttributeUsageWorkerWorkingDir      = "Working directory inside worker pods. Overrides the default working directory set in the image."
	AttributeNameWorkerEnvironment      = "worker-environment"
	AttributeUsageWorkerEnvironment     = "Set environment variables in worker pods. Format: --worker-environment name=value --worker-environment name-b=value-b."
	AttributeNameWorkerEnvPodFieldRef   = "worker-env-pod-field-ref"
	AttributeUsageWorkerEnvPodFieldRef  = "Set an environment variable from a pod field reference for worker pods. Format: ENV_VARIABLE=FIELD_REFERENCE."
	AttributeNameWorkerEnvSecret        = "worker-env-secret"                                                                                                      // #nosec G101 -- This is a flag name, not a credential
	AttributeUsageWorkerEnvSecret       = "Set an environment variable from a Kubernetes secret for worker pods. Format: ENV_VARIABLE=secret-name,key=secret-key." // #nosec G101 -- flag description
	AttributeNameWorkerCreateHomeDir    = "worker-create-home-dir"
	AttributeUsageWorkerCreateHomeDir   = "Create a temporary home directory for worker pods. Defaults to true when --worker-run-as-uid is set, false otherwise."

	// Worker Compute Configuration
	AttributeNameWorkerGpuDevicesRequest        = "worker-gpu-devices-request"
	AttributeUsageWorkerGpuDevicesRequest       = "Number of GPU devices to allocate per worker pod (e.g. 1, 2)."
	AttributeNameWorkerGpuPortionRequest        = "worker-gpu-portion-request"
	AttributeUsageWorkerGpuPortionRequest       = "Fraction of a GPU to allocate per worker pod (between 0 and 1, e.g. 0.5)."
	AttributeNameWorkerGpuPortionLimit          = "worker-gpu-portion-limit"
	AttributeNameWorkerGpuMemoryPortionRequest  = "worker-gpu-memory-portion-request"
	AttributeNameWorkerGpuMemoryPortionLimit    = "worker-gpu-memory-portion-limit"
	AttributeNameWorkerGpuComputeMode           = "worker-gpu-compute-mode"
	AttributeUsageWorkerGpuMemoryPortionRequest = "GPU memory fraction to allocate per device for each worker pod (between 0 and 1, e.g. 0.5)."
	AttributeUsageWorkerGpuMemoryPortionLimit   = "Maximum GPU memory fraction allowed per device for each worker pod (between 0 and 1)."
	AttributeUsageWorkerGpuPortionLimit         = "Maximum GPU fraction allowed per worker pod (between 0 and 1)."
	AttributeNameWorkerGpuMemoryRequest         = "worker-gpu-memory-request"
	AttributeUsageWorkerGpuMemoryRequest        = "Amount of GPU memory to allocate per worker pod (e.g. 1G, 500M)."
	AttributeNameWorkerGpuMemoryLimit           = "worker-gpu-memory-limit"
	AttributeUsageWorkerGpuMemoryLimit          = "Maximum GPU memory to allocate per worker pod (e.g. 1G, 500M)."
	AttributeNameWorkerCpuCoreRequest           = "worker-cpu-core-request"
	AttributeUsageWorkerCpuCoreRequest          = "Number of CPU cores to request per worker pod (e.g. 0.5, 1)."
	AttributeNameWorkerCpuCoreLimit             = "worker-cpu-core-limit"
	AttributeUsageWorkerCpuCoreLimit            = "Maximum number of CPU cores allowed per worker pod (e.g. 0.5, 1)."
	AttributeNameWorkerCpuMemoryRequest         = "worker-cpu-memory-request"
	AttributeUsageWorkerCpuMemoryRequest        = "Amount of memory to request per worker pod (e.g. 1G, 500M)."
	AttributeNameWorkerCpuMemoryLimit           = "worker-cpu-memory-limit"
	AttributeUsageWorkerCpuMemoryLimit          = "Maximum memory allowed per worker pod (e.g. 1G, 500M)."
	AttributeNameWorkerExtendedResource         = "worker-extended-resource"
	AttributeUsageWorkerExtendedResource        = "Request access to a Kubernetes extended resource per worker pod. Format: resource_name=quantity."

	// Worker Storage Configuration
	AttributeNameWorkerExistingPvc  = "worker-existing-pvc"
	AttributeUsageWorkerExistingPvc = "Mount an existing PersistentVolumeClaim for worker pods. Format: claimname=CLAIM_NAME,path=PATH. Auto-complete supported."
	AttributeNameWorkerNewPvc       = "worker-new-pvc"
	AttributeUsageWorkerNewPvc      = "Create and mount a new volume for worker pods. This volume is used only for the duration of the workload's lifecycle. Format: claimname=CLAIM_NAME," +
		"storageclass=STORAGE_CLASS,size=SIZE,path=PATH,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral."
	AttributeNameWorkerSecretVolume        = "worker-secret-volume"
	AttributeUsageWorkerSecretVolume       = "Mount a Kubernetes Secret as a volume for worker pods. Format: path=PATH,name=SECRET_RESOURCE_NAME." // #nosec G101
	AttributeNameWorkerConfigmapMapVolume  = "worker-configmap-map-volume"
	AttributeUsageWorkerConfigmapMapVolume = "Mount a ConfigMap as a volume for worker pods. Format: name=CONFIGMAP_NAME,path=PATH,subpath=SUBPATH,default-mode=DEFAULT_MODE."
	AttributeNameWorkerEmptyDirVolume      = "worker-empty-dir-volume"
	AttributeUsageWorkerEmptyDirVolume     = "Mount an empty directory as a volume for worker pods. Format: path=PATH,medium=MEDIUM,size-limit=SIZE."

	// Worker Security Configuration
	AttributeNameWorkerRunAsUid                  = "worker-run-as-uid"
	AttributeUsageWorkerRunAsUid                 = "User ID to run worker pods as."
	AttributeNameWorkerRunAsGid                  = "worker-run-as-gid"
	AttributeUsageWorkerRunAsGid                 = "Group ID to run worker pods as."
	AttributeNameWorkerRunAsNonRoot              = "worker-run-as-non-root"
	AttributeUsageWorkerRunAsNonRoot             = "Force worker pods to run as a non-root user."
	AttributeNameWorkerSupplementalGroups        = "worker-supplemental-groups"
	AttributeUsageWorkerSupplementalGroups       = "Comma-separated list of group IDs for worker pods."
	AttributeNameWorkerCapability                = "worker-capability"
	AttributeUsageWorkerCapability               = "Add POSIX capabilities to worker pods. Defaults to the runtime's default set."
	AttributeNameWorkerPrivileged                = "worker-privileged"
	AttributeUsageWorkerPrivileged               = "Grants worker pods full access to the host, bypassing almost all container isolation; the container acts like root."
	AttributeNameWorkerAllowPrivilegeEscalation  = "worker-allow-privilege-escalation"
	AttributeUsageWorkerAllowPrivilegeEscalation = "Allow worker pods to gain additional privileges after starting."
	AttributeNameWorkerHostIpc                   = "worker-host-ipc"
	AttributeUsageWorkerHostIpc                  = "Enable host IPC for worker pods. Default: false."
	AttributeNameWorkerHostNetwork               = "worker-host-network"
	AttributeUsageWorkerHostNetwork              = "Enable host networking for worker pods. Default: false."
)
