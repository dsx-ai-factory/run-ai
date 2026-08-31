/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_groups

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

// +k8s:deepcopy-gen=true
type StorageFields struct {

	// Specifies persistent volume claims to mount into a container running the created workload.
	Pvcs *flds.PvcsField `json:"pvcs,omitempty"`

	// Specifies volumes to mount into a container running the created workload.
	Volumes *flds.VolumesField `json:"volumes,omitempty"`

	// Specifies nfs volumes to mount into a container running the created workload.
	Nfs *flds.NfsField `json:"nfs,omitempty"`

	// Specifies a large /dev/shm device to mount into a container running the created workload.
	// SHM is a shared file system mounted on RAM.
	LargeShm *flds.BoolField `json:"largeShm,omitempty"`

	// Specifies git repositories to mount into the container running the workload.
	GitSync *flds.GitSyncField `json:"gitSync,omitempty"`

	// Specifies S3 buckets to mount into the container running the workload
	S3buckets *flds.S3bucketField `json:"s3,omitempty"`

	// Allows for sharing volumes mounted by a container to other containers in the same pod, or
	// even to other pods on the same node. The volume mount will receive all subsequent mounts that
	// are mounted to this volume or any of its subdirectories.
	MountPropagation *flds.BoolField `json:"mountPropagation,omitempty"`

	// Specifies ConfigMaps to mount as volumes into a container running the created workload.
	ConfigMapVolumes *flds.ConfigMapVolumeField `json:"configMapVolumes,omitempty"`

	// Specifies Secrets to mount as volumes into a container running the created workload.
	SecretVolumes *flds.SecretVolumeField `json:"secretVolumes,omitempty"`

	// Specifies data volumes to mount into the workload
	DataVolumes *flds.DataVolumeField `json:"dataVolumes,omitempty"`

	// Specifies emptyDir volumes to mount into the container running the workload
	EmptyDirVolumes *flds.EmptyDirVolumeField `json:"emptyDirVolumes,omitempty"`
}

var StorageRunaiRules = StorageFields{
	Pvcs: &flds.PvcsField{
		ItemRules: &flds.PvcsItemRules{
			Size: &flds.StringRules{
				Format: ptr.To(workload_fields.FormatMemory),
			},
		},
	},

	GitSync: &flds.GitSyncField{
		ItemRules: &flds.GitSyncItemRules{
			Repository: &flds.StringRules{
				Required: ptr.To(true),
			},
			Directory: &flds.StringRules{
				Required: ptr.To(true),
			},
		},
	},

	S3buckets: &flds.S3bucketField{
		ItemRules: &flds.S3bucketItemRules{
			Bucket: &flds.StringRules{
				Required: ptr.To(true),
			},
			Key: &flds.StringRules{
				Format: ptr.To(workload_fields.FormatPassword),
			},
			Secret: &flds.StringRules{
				Format: ptr.To(workload_fields.FormatPassword),
			},
			Url: &flds.StringRules{
				Format: ptr.To(workload_fields.FormatRequestUrl),
			},
			Path: &flds.StringRules{
				Required: ptr.To(true),
			},
		},
	},

	Volumes: &flds.VolumesField{
		ItemRules: &flds.VolumeItemRules{
			SourcePath: &flds.StringRules{
				Required: ptr.To(true),
			},
			TargetPath: &flds.StringRules{
				Required: ptr.To(true),
			},
		},
	},

	Nfs: &flds.NfsField{
		ItemRules: &flds.NfsItemRules{
			SourcePath: &flds.StringRules{
				Required: ptr.To(true),
			},
			TargetPath: &flds.StringRules{
				Required: ptr.To(true),
			},
			NfsServer: &flds.StringRules{
				Required: ptr.To(true),
			},
		},
	},

	EmptyDirVolumes: &flds.EmptyDirVolumeField{
		ItemRules: &flds.EmptyDirVolumeItemRules{
			Path: &flds.StringRules{
				Required: ptr.To(true),
			},
			SizeLimit: &flds.StringRules{
				Format: ptr.To(workload_fields.FormatMemory),
			},
		},
	},
}
