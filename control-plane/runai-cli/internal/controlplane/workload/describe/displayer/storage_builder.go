/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

type StorageBuilder struct {
	storage Storage
}

func NewStorageBuilder() *StorageBuilder {
	return &StorageBuilder{storage: Storage{}}
}

func (b *StorageBuilder) ConfigMaps(configMaps []StorageVolume) *StorageBuilder {
	b.storage.ConfigMaps = configMaps
	return b
}

func (b *StorageBuilder) AddConfigMap(configMap StorageVolume) *StorageBuilder {
	b.storage.ConfigMaps = append(b.storage.ConfigMaps, configMap)
	return b
}

func (b *StorageBuilder) Secrets(secrets []StorageVolume) *StorageBuilder {
	b.storage.Secrets = secrets
	return b
}

func (b *StorageBuilder) AddSecret(secret StorageVolume) *StorageBuilder {
	b.storage.Secrets = append(b.storage.Secrets, secret)
	return b
}

func (b *StorageBuilder) PVCs(pvcs []StorageVolume) *StorageBuilder {
	b.storage.PVCs = pvcs
	return b
}

func (b *StorageBuilder) AddPVC(pvc StorageVolume) *StorageBuilder {
	b.storage.PVCs = append(b.storage.PVCs, pvc)
	return b
}

func (b *StorageBuilder) EmptyDirs(emptyDirs []StorageVolume) *StorageBuilder {
	b.storage.EmptyDirs = emptyDirs
	return b
}

func (b *StorageBuilder) AddEmptyDir(emptyDir StorageVolume) *StorageBuilder {
	b.storage.EmptyDirs = append(b.storage.EmptyDirs, emptyDir)
	return b
}

func (b *StorageBuilder) Gits(gits []StorageVolume) *StorageBuilder {
	b.storage.Gits = gits
	return b
}

func (b *StorageBuilder) AddGit(git StorageVolume) *StorageBuilder {
	b.storage.Gits = append(b.storage.Gits, git)
	return b
}

func (b *StorageBuilder) HostPaths(hostPaths []StorageVolume) *StorageBuilder {
	b.storage.HostPaths = hostPaths
	return b
}

func (b *StorageBuilder) AddHostPath(hostPath StorageVolume) *StorageBuilder {
	b.storage.HostPaths = append(b.storage.HostPaths, hostPath)
	return b
}

func (b *StorageBuilder) S3s(s3s []StorageVolume) *StorageBuilder {
	b.storage.S3s = s3s
	return b
}

func (b *StorageBuilder) AddS3(s3 StorageVolume) *StorageBuilder {
	b.storage.S3s = append(b.storage.S3s, s3)
	return b
}

func (b *StorageBuilder) Build() *Storage {
	return &b.storage
}

type StorageVolumeBuilder struct {
	volume StorageVolume
}

func NewStorageVolumeBuilder() *StorageVolumeBuilder {
	return &StorageVolumeBuilder{volume: StorageVolume{}}
}

func (b *StorageVolumeBuilder) Name(name string) *StorageVolumeBuilder {
	b.volume.Name = name
	return b
}

func (b *StorageVolumeBuilder) MountPath(mountPath string) *StorageVolumeBuilder {
	b.volume.MountPath = mountPath
	return b
}

func (b *StorageVolumeBuilder) Size(size string) *StorageVolumeBuilder {
	b.volume.Size = size
	return b
}

func (b *StorageVolumeBuilder) ReadOnly(readOnly bool) *StorageVolumeBuilder {
	b.volume.ReadOnly = &readOnly
	return b
}

func (b *StorageVolumeBuilder) Ephemeral(ephemeral bool) *StorageVolumeBuilder {
	b.volume.Ephemeral = &ephemeral
	return b
}

func (b *StorageVolumeBuilder) DataSharing(dataSharing bool) *StorageVolumeBuilder {
	b.volume.DataSharing = &dataSharing
	return b
}

func (b *StorageVolumeBuilder) StorageClass(storageClass string) *StorageVolumeBuilder {
	b.volume.StorageClass = &storageClass
	return b
}

func (b *StorageVolumeBuilder) Build() StorageVolume {
	return b.volume
}

func buildConfigMapVolume(cm openapi.ConfigMapInstance) StorageVolume {
	builder := NewStorageVolumeBuilder()
	if cm.ConfigMap.IsSet() && cm.ConfigMap.Get() != nil {
		builder.Name(*cm.ConfigMap.Get())
	} else if cm.Name.IsSet() && cm.Name.Get() != nil {
		builder.Name(*cm.Name.Get())
	}
	if cm.MountPath.IsSet() && cm.MountPath.Get() != nil {
		builder.MountPath(*cm.MountPath.Get())
	}
	return builder.Build()
}

func buildSecretVolume(s openapi.SecretInstance) StorageVolume {
	builder := NewStorageVolumeBuilder()
	if s.Secret.IsSet() && s.Secret.Get() != nil {
		builder.Name(*s.Secret.Get())
	} else if s.Name.IsSet() && s.Name.Get() != nil {
		builder.Name(*s.Name.Get())
	}
	if s.MountPath.IsSet() && s.MountPath.Get() != nil {
		builder.MountPath(*s.MountPath.Get())
	}
	return builder.Build()
}

func buildPvcVolume(pvc openapi.PvcInstance) StorageVolume {
	builder := NewStorageVolumeBuilder()
	if pvc.ClaimName.IsSet() && pvc.ClaimName.Get() != nil {
		builder.Name(*pvc.ClaimName.Get())
	} else if pvc.Name.IsSet() && pvc.Name.Get() != nil {
		builder.Name(*pvc.Name.Get())
	}
	if pvc.Path.IsSet() && pvc.Path.Get() != nil {
		builder.MountPath(*pvc.Path.Get())
	}
	if pvc.ReadOnly.IsSet() && pvc.ReadOnly.Get() != nil {
		builder.ReadOnly(*pvc.ReadOnly.Get())
	}
	if pvc.Ephemeral.IsSet() && pvc.Ephemeral.Get() != nil {
		builder.Ephemeral(*pvc.Ephemeral.Get())
	}
	if pvc.DataSharing.IsSet() && pvc.DataSharing.Get() != nil {
		builder.DataSharing(*pvc.DataSharing.Get())
	}
	if pvc.ClaimInfo.IsSet() && pvc.ClaimInfo.Get() != nil {
		claimInfo := pvc.ClaimInfo.Get()
		if claimInfo.Size.IsSet() && claimInfo.Size.Get() != nil {
			builder.Size(*claimInfo.Size.Get())
		}
		if claimInfo.StorageClass.IsSet() && claimInfo.StorageClass.Get() != nil {
			builder.StorageClass(*claimInfo.StorageClass.Get())
		}
	}
	return builder.Build()
}

func buildEmptyDirVolume(ed openapi.EmptyDirInstance) StorageVolume {
	builder := NewStorageVolumeBuilder()
	if ed.Name.IsSet() && ed.Name.Get() != nil {
		builder.Name(*ed.Name.Get())
	}
	if ed.Path.IsSet() && ed.Path.Get() != nil {
		builder.MountPath(*ed.Path.Get())
	}
	if ed.SizeLimit.IsSet() && ed.SizeLimit.Get() != nil {
		builder.Size(*ed.SizeLimit.Get())
	}
	builder.Ephemeral(true)
	return builder.Build()
}

func buildGitVolume(git openapi.GitInstance) StorageVolume {
	builder := NewStorageVolumeBuilder()
	if git.Repository.IsSet() && git.Repository.Get() != nil {
		builder.Name(*git.Repository.Get())
	} else if git.Name.IsSet() && git.Name.Get() != nil {
		builder.Name(*git.Name.Get())
	}
	if git.Path.IsSet() && git.Path.Get() != nil {
		builder.MountPath(*git.Path.Get())
	}
	return builder.Build()
}

func buildHostPathVolume(hp openapi.HostPathInstance) StorageVolume {
	builder := NewStorageVolumeBuilder()
	if hp.Path.IsSet() && hp.Path.Get() != nil {
		builder.Name(*hp.Path.Get())
	} else if hp.Name.IsSet() && hp.Name.Get() != nil {
		builder.Name(*hp.Name.Get())
	}
	if hp.MountPath.IsSet() && hp.MountPath.Get() != nil {
		builder.MountPath(*hp.MountPath.Get())
	}
	if hp.ReadOnly.IsSet() && hp.ReadOnly.Get() != nil {
		builder.ReadOnly(*hp.ReadOnly.Get())
	}
	return builder.Build()
}

func buildS3Volume(s3 openapi.S3Instance) StorageVolume {
	builder := NewStorageVolumeBuilder()
	if s3.Bucket.IsSet() && s3.Bucket.Get() != nil {
		builder.Name(*s3.Bucket.Get())
	} else if s3.Name.IsSet() && s3.Name.Get() != nil {
		builder.Name(*s3.Name.Get())
	}
	if s3.Path.IsSet() && s3.Path.Get() != nil {
		builder.MountPath(*s3.Path.Get())
	}
	return builder.Build()
}

func finalizeStorage(storage *Storage) *Storage {
	if len(storage.ConfigMaps) == 0 && len(storage.Secrets) == 0 && len(storage.PVCs) == 0 &&
		len(storage.EmptyDirs) == 0 && len(storage.Gits) == 0 && len(storage.HostPaths) == 0 && len(storage.S3s) == 0 {
		return nil
	}
	return storage
}

func BuildStorageFromDistributedInferenceSpec(spec *openapi.DistributedInferenceLeaderWorkerSpecStorage) *Storage {
	if spec == nil {
		return nil
	}

	storageBuilder := NewStorageBuilder()

	for _, cm := range spec.ConfigMapVolume {
		if vol := buildConfigMapVolume(cm); vol.Name != "" {
			storageBuilder.AddConfigMap(vol)
		}
	}

	for _, s := range spec.SecretVolume {
		if vol := buildSecretVolume(s); vol.Name != "" {
			storageBuilder.AddSecret(vol)
		}
	}

	for _, pvc := range spec.Pvc {
		if vol := buildPvcVolume(pvc); vol.Name != "" {
			storageBuilder.AddPVC(vol)
		}
	}

	for _, ed := range spec.EmptyDirVolume {
		if vol := buildEmptyDirVolume(ed); vol.Name != "" {
			storageBuilder.AddEmptyDir(vol)
		}
	}

	return finalizeStorage(storageBuilder.Build())
}

func BuildStorageFromWorkloadSpec(
	configMaps []openapi.ConfigMapInstance,
	secrets []openapi.SecretInstance,
	pvcs []openapi.PvcInstance,
	emptyDirs []openapi.EmptyDirInstance,
	gits []openapi.GitInstance,
	hostPaths []openapi.HostPathInstance,
	s3s []openapi.S3Instance,
) *Storage {
	storageBuilder := NewStorageBuilder()

	for _, cm := range configMaps {
		if vol := buildConfigMapVolume(cm); vol.Name != "" {
			storageBuilder.AddConfigMap(vol)
		}
	}

	for _, s := range secrets {
		if vol := buildSecretVolume(s); vol.Name != "" {
			storageBuilder.AddSecret(vol)
		}
	}

	for _, pvc := range pvcs {
		if vol := buildPvcVolume(pvc); vol.Name != "" {
			storageBuilder.AddPVC(vol)
		}
	}

	for _, ed := range emptyDirs {
		if vol := buildEmptyDirVolume(ed); vol.Name != "" {
			storageBuilder.AddEmptyDir(vol)
		}
	}

	for _, git := range gits {
		if vol := buildGitVolume(git); vol.Name != "" {
			storageBuilder.AddGit(vol)
		}
	}

	for _, hp := range hostPaths {
		if vol := buildHostPathVolume(hp); vol.Name != "" {
			storageBuilder.AddHostPath(vol)
		}
	}

	for _, s3 := range s3s {
		if vol := buildS3Volume(s3); vol.Name != "" {
			storageBuilder.AddS3(vol)
		}
	}

	return finalizeStorage(storageBuilder.Build())
}
