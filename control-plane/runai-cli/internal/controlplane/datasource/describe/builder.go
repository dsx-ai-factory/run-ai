/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"time"

	assetsgenerated "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

type DatasourceBuilder struct {
	datasource Datasource
}

func NewDatasourceBuilder() *DatasourceBuilder {
	return &DatasourceBuilder{
		datasource: Datasource{
			Meta:                 DatasourceMeta{},
			Spec:                 &DatasourceSpec{},
			Status:               &DatasourceStatus{},
			AdditionalAttributes: make(map[string]string),
		},
	}
}

func (b *DatasourceBuilder) SetName(name string) *DatasourceBuilder {
	b.datasource.Meta.Name = name
	return b
}

func (b *DatasourceBuilder) SetType(dsType string) *DatasourceBuilder {
	b.datasource.Meta.Type = dsType
	return b
}

func (b *DatasourceBuilder) SetDescription(description *string) *DatasourceBuilder {
	b.datasource.Meta.Description = description
	return b
}

func (b *DatasourceBuilder) SetScope(scope string) *DatasourceBuilder {
	b.datasource.Meta.Scope = scope
	return b
}

func (b *DatasourceBuilder) SetScopeID(scopeId string) *DatasourceBuilder {
	b.datasource.Meta.ScopeID = scopeId
	return b
}

func (b *DatasourceBuilder) SetProjectName(projectName *string) *DatasourceBuilder {
	b.datasource.Meta.ProjectName = projectName
	return b
}

func (b *DatasourceBuilder) SetClusterID(clusterId *string) *DatasourceBuilder {
	b.datasource.Meta.ClusterID = clusterId
	return b
}

func (b *DatasourceBuilder) SetProjectID(projectId *int32) *DatasourceBuilder {
	b.datasource.Meta.ProjectID = projectId
	return b
}

func (b *DatasourceBuilder) SetDepartmentID(departmentId *string) *DatasourceBuilder {
	b.datasource.Meta.DepartmentID = departmentId
	return b
}

func (b *DatasourceBuilder) SetCreatedAt(createdAt time.Time) *DatasourceBuilder {
	b.datasource.Meta.CreatedAt = createdAt
	return b
}

func (b *DatasourceBuilder) SetCreatedBy(createdBy string) *DatasourceBuilder {
	b.datasource.Meta.CreatedBy = createdBy
	return b
}

func (b *DatasourceBuilder) SetUpdatedAt(updatedAt time.Time) *DatasourceBuilder {
	b.datasource.Meta.UpdatedAt = updatedAt
	return b
}

func (b *DatasourceBuilder) SetUpdatedBy(updatedBy string) *DatasourceBuilder {
	b.datasource.Meta.UpdatedBy = updatedBy
	return b
}

func (b *DatasourceBuilder) SetStatus(status *string) *DatasourceBuilder {
	if b.datasource.Status == nil {
		b.datasource.Status = &DatasourceStatus{}
	}
	b.datasource.Status.Status = status
	return b
}

func (b *DatasourceBuilder) SetStatusMessage(message *string) *DatasourceBuilder {
	if b.datasource.Status == nil {
		b.datasource.Status = &DatasourceStatus{}
	}
	b.datasource.Status.Message = message
	return b
}

func (b *DatasourceBuilder) cleanupEmptyStatus() {
	if b.datasource.Status != nil && b.datasource.Status.Status == nil && b.datasource.Status.Message == nil {
		b.datasource.Status = nil
	}
}

func (b *DatasourceBuilder) AddAttribute(key, value string) *DatasourceBuilder {
	if b.datasource.AdditionalAttributes == nil {
		b.datasource.AdditionalAttributes = make(map[string]string)
	}
	b.datasource.AdditionalAttributes[key] = value
	return b
}

func (b *DatasourceBuilder) SetNFSSpec(spec *NFSSpec) *DatasourceBuilder {
	if b.datasource.Spec == nil {
		b.datasource.Spec = &DatasourceSpec{}
	}
	b.datasource.Spec.NFS = spec
	return b
}

func (b *DatasourceBuilder) SetPVCSpec(spec *PVCSpec) *DatasourceBuilder {
	if b.datasource.Spec == nil {
		b.datasource.Spec = &DatasourceSpec{}
	}
	b.datasource.Spec.PVC = spec
	return b
}

func (b *DatasourceBuilder) SetS3Spec(spec *S3Spec) *DatasourceBuilder {
	if b.datasource.Spec == nil {
		b.datasource.Spec = &DatasourceSpec{}
	}
	b.datasource.Spec.S3 = spec
	return b
}

func (b *DatasourceBuilder) SetGitSpec(spec *GitSpec) *DatasourceBuilder {
	if b.datasource.Spec == nil {
		b.datasource.Spec = &DatasourceSpec{}
	}
	b.datasource.Spec.Git = spec
	return b
}

func (b *DatasourceBuilder) SetHostPathSpec(spec *HostPathSpec) *DatasourceBuilder {
	if b.datasource.Spec == nil {
		b.datasource.Spec = &DatasourceSpec{}
	}
	b.datasource.Spec.HostPath = spec
	return b
}

func (b *DatasourceBuilder) SetConfigMapSpec(spec *ConfigMapSpec) *DatasourceBuilder {
	if b.datasource.Spec == nil {
		b.datasource.Spec = &DatasourceSpec{}
	}
	b.datasource.Spec.ConfigMap = spec
	return b
}

func (b *DatasourceBuilder) SetSecretSpec(spec *SecretSpec) *DatasourceBuilder {
	if b.datasource.Spec == nil {
		b.datasource.Spec = &DatasourceSpec{}
	}
	b.datasource.Spec.Secret = spec
	return b
}

func (b *DatasourceBuilder) Build() *Datasource {
	return &b.datasource
}

// buildDatasourceFromApiResponse transforms API response into display structure
func buildDatasourceFromApiResponse(raw *assetsgenerated.DatasourceListResponseEntry) (*Datasource, error) {
	builder := NewDatasourceBuilder()

	setMetaInformation(builder, raw)
	setStatusInformation(builder, raw)

	// Determine datasource type and set appropriate spec
	if err := setSpecInformation(builder, raw); err != nil {
		return nil, err
	}

	return builder.Build(), nil
}

func setMetaInformation(builder *DatasourceBuilder, raw *assetsgenerated.DatasourceListResponseEntry) {
	meta := raw.Meta

	builder.
		SetName(meta.Name).
		SetType(string(meta.Kind)).
		SetScope(string(meta.Scope)).
		SetCreatedAt(meta.CreatedAt).
		SetCreatedBy(meta.CreatedBy).
		SetUpdatedAt(meta.UpdatedAt).
		SetUpdatedBy(meta.UpdatedBy)

	if meta.Description.IsSet() {
		desc := meta.GetDescription()
		if desc != "" {
			builder.SetDescription(&desc)
		}
	}

	if meta.ClusterId.IsSet() {
		clusterId := meta.GetClusterId()
		if clusterId != "" {
			builder.SetClusterID(&clusterId)
		}
	}

	if meta.ProjectId.IsSet() {
		projectId := meta.GetProjectId()
		if projectId != 0 {
			builder.SetProjectID(&projectId)
		}
	}

	if meta.DepartmentId.IsSet() {
		deptId := meta.GetDepartmentId()
		if deptId != "" {
			builder.SetDepartmentID(&deptId)
		}
	}
}

func setStatusInformation(builder *DatasourceBuilder, raw *assetsgenerated.DatasourceListResponseEntry) {
	if raw.Status.IsSet() {
		status := raw.GetStatus()

		if status.Status != nil {
			statusStr := string(*status.Status)
			builder.SetStatus(&statusStr)
		}

		if status.Message != nil && *status.Message != "" {
			builder.SetStatusMessage(status.Message)
		}
	}

	// Remove status object if it's empty after processing
	builder.cleanupEmptyStatus()
}

func setSpecInformation(builder *DatasourceBuilder, raw *assetsgenerated.DatasourceListResponseEntry) error {
	spec := raw.Spec

	// Check which type of datasource this is and build the appropriate spec
	if spec.Nfs.IsSet() {
		nfsSpec := buildNFSSpec(spec.GetNfs())
		builder.SetNFSSpec(nfsSpec)
	} else if spec.Pvc.IsSet() {
		pvcSpec := buildPVCSpec(spec.GetPvc(), raw.Meta)
		builder.SetPVCSpec(pvcSpec)
	} else if spec.S3.IsSet() {
		s3Spec := buildS3Spec(spec.GetS3())
		builder.SetS3Spec(s3Spec)
	} else if spec.Git.IsSet() {
		gitSpec := buildGitSpec(spec.GetGit())
		builder.SetGitSpec(gitSpec)
	} else if spec.HostPath.IsSet() {
		hostPathSpec := buildHostPathSpec(spec.GetHostPath())
		builder.SetHostPathSpec(hostPathSpec)
	} else if spec.ConfigMap.IsSet() {
		configMapSpec := buildConfigMapSpec(spec.GetConfigMap())
		builder.SetConfigMapSpec(configMapSpec)
	} else if spec.Secret.IsSet() {
		secretSpec := buildSecretSpec(spec.GetSecret())
		builder.SetSecretSpec(secretSpec)
	}

	return nil
}

func buildNFSSpec(apiSpec assetsgenerated.NFSAssetSpec) *NFSSpec {
	spec := &NFSSpec{}

	if apiSpec.Server.IsSet() {
		server := apiSpec.GetServer()
		if server != "" {
			spec.Server = &server
		}
	}

	if apiSpec.Path.IsSet() {
		path := apiSpec.GetPath()
		if path != "" {
			spec.Path = &path
		}
	}

	if apiSpec.MountPath.IsSet() {
		mountPath := apiSpec.GetMountPath()
		if mountPath != "" {
			spec.MountPath = &mountPath
		}
	}

	if apiSpec.ReadOnly.IsSet() {
		readOnly := apiSpec.GetReadOnly()
		spec.ReadOnly = &readOnly
	}

	return spec
}

func buildPVCSpec(apiSpec assetsgenerated.PVCAssetSpec, meta assetsgenerated.AssetMeta) *PVCSpec {
	spec := &PVCSpec{}

	if apiSpec.Path.IsSet() {
		path := apiSpec.GetPath()
		if path != "" {
			spec.Path = &path
		}
	}

	if apiSpec.ExistingPvc.IsSet() {
		existingPvc := apiSpec.GetExistingPvc()
		spec.ExistingPVC = &existingPvc
	}

	if apiSpec.ClaimName.IsSet() {
		claimName := apiSpec.GetClaimName()
		if claimName != "" {
			spec.ClaimName = &claimName
		}
	}

	if apiSpec.ReadOnly.IsSet() {
		readOnly := apiSpec.GetReadOnly()
		spec.ReadOnly = &readOnly
	}

	if apiSpec.Ephemeral.IsSet() {
		ephemeral := apiSpec.GetEphemeral()
		spec.Ephemeral = &ephemeral
	}

	if apiSpec.DataSharing.IsSet() {
		dataSharing := apiSpec.GetDataSharing()
		spec.DataSharing = &dataSharing
	}

	if apiSpec.ClaimInfo.IsSet() {
		claimInfo := buildClaimInfo(apiSpec.GetClaimInfo())
		spec.ClaimInfo = claimInfo
	}

	return spec
}

func buildClaimInfo(apiClaimInfo assetsgenerated.ClaimInfo) *ClaimInfo {
	claimInfo := &ClaimInfo{}

	if apiClaimInfo.Size.IsSet() {
		size := apiClaimInfo.GetSize()
		if size != "" {
			claimInfo.Size = &size
		}
	}

	if apiClaimInfo.StorageClass.IsSet() {
		storageClass := apiClaimInfo.GetStorageClass()
		if storageClass != "" {
			claimInfo.StorageClass = &storageClass
		}
	}

	if apiClaimInfo.AccessModes.IsSet() {
		accessModes := buildAccessModes(apiClaimInfo.GetAccessModes())
		claimInfo.AccessModes = accessModes
	}

	if apiClaimInfo.VolumeMode.IsSet() {
		volumeMode := string(apiClaimInfo.GetVolumeMode())
		if volumeMode != "" {
			claimInfo.VolumeMode = &volumeMode
		}
	}

	return claimInfo
}

func buildAccessModes(apiAccessModes assetsgenerated.PvcAccessModes) *AccessModes {
	accessModes := &AccessModes{}

	if apiAccessModes.ReadWriteOnce.IsSet() {
		rwo := apiAccessModes.GetReadWriteOnce()
		accessModes.ReadWriteOnce = &rwo
	}

	if apiAccessModes.ReadOnlyMany.IsSet() {
		rom := apiAccessModes.GetReadOnlyMany()
		accessModes.ReadOnlyMany = &rom
	}

	if apiAccessModes.ReadWriteMany.IsSet() {
		rwm := apiAccessModes.GetReadWriteMany()
		accessModes.ReadWriteMany = &rwm
	}

	return accessModes
}

func buildS3Spec(apiSpec assetsgenerated.S3AssetSpec) *S3Spec {
	spec := &S3Spec{}

	if apiSpec.Bucket.IsSet() {
		bucket := apiSpec.GetBucket()
		if bucket != "" {
			spec.Bucket = &bucket
		}
	}

	if apiSpec.Path.IsSet() {
		path := apiSpec.GetPath()
		if path != "" {
			spec.Path = &path
		}
	}

	if apiSpec.Url.IsSet() {
		url := apiSpec.GetUrl()
		if url != "" {
			spec.URL = &url
		}
	}

	if apiSpec.AccessKeyAssetId.IsSet() {
		accessKeyAssetId := apiSpec.GetAccessKeyAssetId()
		if accessKeyAssetId != "" {
			spec.AccessKeyAssetID = &accessKeyAssetId
		}
	}

	return spec
}

func buildGitSpec(apiSpec assetsgenerated.GitAssetSpec) *GitSpec {
	spec := &GitSpec{}

	if apiSpec.Repository.IsSet() {
		repository := apiSpec.GetRepository()
		if repository != "" {
			spec.Repository = &repository
		}
	}

	if apiSpec.Branch.IsSet() {
		branch := apiSpec.GetBranch()
		if branch != "" {
			spec.Branch = &branch
		}
	}

	if apiSpec.Revision.IsSet() {
		revision := apiSpec.GetRevision()
		if revision != "" {
			spec.Revision = &revision
		}
	}

	if apiSpec.Path.IsSet() {
		path := apiSpec.GetPath()
		if path != "" {
			spec.Path = &path
		}
	}

	if apiSpec.CredentialId.IsSet() {
		credentialId := apiSpec.GetCredentialId()
		if credentialId != "" {
			spec.CredentialID = &credentialId
		}
	}

	return spec
}

func buildHostPathSpec(apiSpec assetsgenerated.HostPathSpec) *HostPathSpec {
	spec := &HostPathSpec{}

	if apiSpec.Path.IsSet() {
		path := apiSpec.GetPath()
		if path != "" {
			spec.Path = &path
		}
	}

	if apiSpec.MountPath.IsSet() {
		mountPath := apiSpec.GetMountPath()
		if mountPath != "" {
			spec.MountPath = &mountPath
		}
	}

	if apiSpec.ReadOnly.IsSet() {
		readOnly := apiSpec.GetReadOnly()
		spec.ReadOnly = &readOnly
	}

	if apiSpec.MountPropagation.IsSet() {
		mountPropagation := string(apiSpec.GetMountPropagation())
		if mountPropagation != "" {
			spec.MountPropagation = &mountPropagation
		}
	}

	return spec
}

func buildConfigMapSpec(apiSpec assetsgenerated.ConfigMapAssetSpec) *ConfigMapSpec {
	spec := &ConfigMapSpec{}

	if apiSpec.ConfigMap.IsSet() {
		configMap := apiSpec.GetConfigMap()
		if configMap != "" {
			spec.ConfigMapName = &configMap
		}
	}

	if apiSpec.MountPath.IsSet() {
		mountPath := apiSpec.GetMountPath()
		if mountPath != "" {
			spec.MountPath = &mountPath
		}
	}

	if apiSpec.SubPath.IsSet() {
		subPath := apiSpec.GetSubPath()
		if subPath != "" {
			spec.SubPath = &subPath
		}
	}

	if apiSpec.DefaultMode.IsSet() {
		defaultMode := apiSpec.GetDefaultMode()
		if defaultMode != "" {
			spec.DefaultMode = &defaultMode
		}
	}

	return spec
}

func buildSecretSpec(apiSpec assetsgenerated.SecretAssetSpec) *SecretSpec {
	spec := &SecretSpec{}

	if apiSpec.CredentialAssetId.IsSet() {
		credentialAssetId := apiSpec.GetCredentialAssetId()
		if credentialAssetId != "" {
			spec.CredentialAssetID = &credentialAssetId
		}
	}

	if apiSpec.MountPath.IsSet() {
		mountPath := apiSpec.GetMountPath()
		if mountPath != "" {
			spec.MountPath = &mountPath
		}
	}

	if apiSpec.DefaultMode.IsSet() {
		defaultMode := apiSpec.GetDefaultMode()
		if defaultMode != "" {
			spec.DefaultMode = &defaultMode
		}
	}

	return spec
}
