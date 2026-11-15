/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package consts

// Training consts
const (
	TrainingShortMsg     = "training management"
	TrainingListShortMsg = "list all training frameworks"
	TrainingListExample  = `# List all training workloads
runai training list -A

# List training workloads in a specific project
runai training list -p <project-name>

# List training workloads filtered by status
runai training list --status <status>

# List training workloads filtered by framework
runai training list --framework <framework>

# List training workloads with a specific output format
runai training list --yaml

# List training workloads with pagination
runai training list --limit 20 --offset 40
`
	TrainingStandardShortMsg = "standard training management"
)

// Distributed training consts
const (
	MPIShortMsg        = "mpi management"
	XgboostShortMsg    = "xgboost management"
	TensorflowShortMsg = "tensorflow management"
	PytorchShortMsg    = "pytorch management"

	FollowingDistributedTraining = "To track the workload's status, run 'runai training %s describe %s'"
)

// Submit Distributed Training const's
const (
	AttributeNameNoMaster     = "no-master"
	AttributeUsageNoMaster    = "Do not create a separate pod for the master"
	AttributeNameWorkers      = "workers"
	AttributeUsageWorkers     = "the number of workers that will be allocated for running the workload"
	AttributeNameSegmentSize  = "segment-size"
	AttributeUsageSegmentSize = "Number of worker replicas per topology-aware scheduling segment. " +
		"The final segment may contain fewer replicas"
	AttributeNameDistributedFramework          = "framework"
	AttributeShortDistributedFramework         = "f"
	AttributeUsageDistributedFramework         = "The distributed training framework used in the workload."
	AttributeNameMinReplica                    = "min-replicas"
	AttributeUsageMinReplica                   = "Minimum number of replicas for an elastic PyTorch job"
	AttributeNameMaxReplica                    = "max-replicas"
	AttributeUsageMaxReplica                   = "Maximum number of replicas for an elastic PyTorch job"
	AttributeNameSlotsPerWorker                = "slots-per-worker"
	AttributeUsageSlotsPerWorker               = "Number of slots to allocate for each worker"
	AttributeNameSshAuthMountPath              = "ssh-auth-mount-path"
	AttributeUsageSshAuthMountPath             = "Specifies the directory where SSH keys are mounted."
	AttributeNameMasterExtendedResource        = "master-extended-resource"
	AttributeNameMasterGpuDeviceRequest        = "master-gpu-devices-request"
	AttributeNameMasterGpuPortionRequest       = "master-gpu-portion-request"
	AttributeNameMasterGpuPortionLimit         = "master-gpu-portion-limit"
	AttributeNameMasterGpuMemoryPortionRequest = "master-gpu-memory-portion-request"
	AttributeNameMasterGpuMemoryPortionLimit   = "master-gpu-memory-portion-limit"
	AttributeNameMasterGpuComputeMode          = "master-gpu-compute-mode"
	AttributeNameMasterCpuCoreRequest          = "master-cpu-core-request"
	AttributeNameMasterCpuCoreLimit            = "master-cpu-core-limit"
	AttributeNameMasterCpuMemoryRequest        = "master-cpu-memory-request"
	AttributeNameMasterCpuMemoryLimit          = "master-cpu-memory-limit"
	AttributeNameMasterNoPvc                   = "master-no-pvcs"
	AttributeUsageMasterNoPvc                  = "Do not mount any persistent volumes in the master pod"
	AttributeNameMasterArgs                    = "master-args"
	AttributeUsageMasterArgs                   = "Specifies the arguments to pass to the master pod container command"
	AttributeNameMasterCommand                 = "master-command"
	AttributeUsageMasterCommand                = "Specifies the command to run in the master pod container, overriding the image's default entrypoint. The command can include arguments following it."
	// Deprecated: use AttributeNameMasterEnvVariable instead
	AttributeNameMasterEnvDeprecated = "master-environment"
	AttributeNameMasterEnvVariable   = "master-environment-variable"

	AttributeNameCleanPodPolicy  = "clean-pod-policy"
	AttributeUsageCleanPodPolicy = `Specifies which pods will be deleted when the workload reaches a terminal state (completed/failed). Valid values are None, All, Running.
* Running - Only pods still running when a job completes (for example, parameter servers) will be deleted immediately. Completed pods will not be deleted so that the logs will be preserved.
* All     - All (including completed) pods will be deleted immediately when the job finishes.
* None    - No pods will be deleted when the job completes. It will keep running pods that consume GPU, CPU and memory over time. 
            It is recommended to set to None only for debugging and obtaining logs from running pods. (Default)`
	AttributeNameLauncherCreationPolicy  = "launcher-creation-policy"
	AttributeUsageLauncherCreationPolicy = "Launcher creation policy. Valid values: AtStartup, WaitForWorkersReady"

	ErrorSubmitMpiWithGpu   = "GPU resources are not allowed on MPI master pod. Please resubmit without master GPU flags."
	WarningSubmitMpuWithGpu = "Unnecessary request for master GPU resources, GPU resources are not supported on MPI master pod."

	AttributeUsageMasterEnvironmentVariables    = "Set environment variables in the master container. --master-environment-variable name=value -master-environment-variable name-b=value-b."
	AttributeUsageMasterCpuCoreRequest          = "Number of CPU cores to request for the master pod (e.g., 0.5, 1)."
	AttributeUsageMasterCpuCoreLimit            = "Maximum number of CPU cores allowed for the master pod (e.g., 0.5, 1)."
	AttributeUsageMasterCpuMemoryRequest        = "Amount of memory to request for the master pod (e.g., 1G, 500M)."
	AttributeUsageMasterCpuMemoryLimit          = "Maximum memory allowed for the master pod (e.g., 1G, 500M)."
	AttributeUsageMasterGpuDeviceRequest        = "Number of GPU devices to allocate for the master pod (e.g., 1, 2)."
	AttributeUsageMasterGpuPortionRequest       = "Fraction of a GPU to allocate to the master pod (between 0 and 1, e.g., 0.5)."
	AttributeUsageMasterGpuPortionLimit         = "Maximum GPU fraction allowed for the master pod (between 0 and 1)."
	AttributeUsageMasterGpuMemoryPortionRequest = "GPU memory fraction to allocate per device for the master pod (between 0 and 1, e.g. 0.5)."
	AttributeUsageMasterGpuMemoryPortionLimit   = "Maximum GPU memory fraction allowed per device for the master pod (between 0 and 1)."
)

// Backward compatibility flags
const (
	AttributeBCMasterNameGpu = "master-gpu"
)
