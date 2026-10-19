/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package training

// Master GPU flags that not supported for MPI
const (
	SlotPerWorkerFlag = "slots-per-worker"
	NoMaterFlag       = "no-master"
	MaxReplicaFlag    = "max-replicas"
	MinReplicaFlag    = "min-replicas"
)

// General flags
const (
	Annotation = "--annotation"
	LabelFlag  = "--label"
	NamePrefix = "--name-prefix"
	Project    = "--project"
)

// Container flags
const (
	Command             = "--command"
	CreateHomeDir       = "--create-home-dir"
	Environment         = "--environment-variable"
	Image               = "--image"
	ImagePullPolicy     = "--image-pull-policy"
	MasterArgs          = "--master-args"
	MasterEnvironment   = "--master-environment-variable"
	PodTopologyKey      = "--preferred-pod-topology-key"
	RequiredTopologyKey = "--required-pod-topology-key"
	Stdin               = "--stdin"
	Tty                 = "--tty"
	WorkingDir          = "--working-dir"
)

// Compute flags
const (
	CpuCoreLimit            = "--cpu-core-limit"
	CpuCoreRequest          = "--cpu-core-request"
	CpuMemoryLimit          = "--cpu-memory-limit"
	CpuMemoryRequest        = "--cpu-memory-request"
	ExtendedResource        = "--extended-resource"
	GpuDevicesRequest       = "--gpu-devices-request"
	GpuMemoryLimit          = "--gpu-memory-limit"
	GpuMemoryRequest        = "--gpu-memory-request"
	GpuMemoryPortionLimit   = "--gpu-memory-portion-limit"
	GpuMemoryPortionRequest = "--gpu-memory-portion-request"
	GpuComputeMode          = "--gpu-compute-mode"
	GpuRequestType          = "--gpu-request-type"
	LargeShm                = "--large-shm"
	MasterExtendedResource  = "--master-extended-resource"
)

// Storage flags
const (
	ConfigMapVolume = "--configmap-map-volume"
	ExistingPvc     = "--existing-pvc"
	GitSync         = "--git-sync"
	HostPath        = "--host-path"
	MasterNoPvcs    = "--master-no-pvcs"
	NewPvc          = "--new-pvc"
	Nfs             = "--nfs"
	S3              = "--s3"
)

// Network flags
const (
	ExternalUrl = "--external-url"
	Port        = "--port"
)

// Lifecycle flags
const (
	AutoDeletionTime = "--auto-deletion-time-after-completion"
	BackoffLimit     = "--backoff-limit"
)

// Security flags
const (
	AllowPrivilegeEscalation = "--allow-privilege-escalation"
	Capability               = "--capability"
	HostIpc                  = "--host-ipc"
	HostNetwork              = "--host-network"
	RunAsGid                 = "--run-as-gid"
	RunAsUid                 = "--run-as-uid"
	RunAsUser                = "--run-as-uid"
	SeccompProfile           = "--seccomp-profile"
	SupplementalGroups       = "--supplemental-groups"
	UserGroupSource          = "--user-group-source"
)

// Scheduling flags
const (
	NodePools  = "--node-pools"
	NodeType   = "--node-type"
	Toleration = "--toleration"
)

// Distributed flags
const (
	SlotsPerWorker = "--slots-per-worker"
	Workers        = "--workers"
	SegmentSize    = "--segment-size"
)

const (
	ExpectedListWithProjectThatDoesNotExist = "project name: p300, not found\n"
	ExpectedSubmitMissingFlags              = "required flag(s) \"image\" not set"
	ExpectDistSuccessMsg                    = "Creating distributed %s..."
	MissingProject                          = "Error: project is missing. value is empty.\n"
	ProjectAllMutuallyExclusive             = "Error: cannot set project and all flags together. more than 1 of the attributes are set\n"
)
