package describe

import (
	"time"
)

// Datasource represents the display model for a datasource
type Datasource struct {
	Meta                 DatasourceMeta    `json:"meta" yaml:"meta"`
	Spec                 *DatasourceSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status               *DatasourceStatus `json:"status,omitempty" yaml:"status,omitempty"`
	AdditionalAttributes map[string]string `json:"additionalAttributes,omitempty" yaml:"additionalAttributes,omitempty"`
}

// DatasourceMeta contains metadata about the datasource
type DatasourceMeta struct {
	Name         string    `json:"name" yaml:"name"`
	Type         string    `json:"type" yaml:"type"`
	Description  *string   `json:"description,omitempty" yaml:"description,omitempty"`
	Scope        string    `json:"scope" yaml:"scope"`
	ScopeID      string    `json:"scopeId,omitempty" yaml:"scopeId,omitempty"`
	ProjectName  *string   `json:"projectName,omitempty" yaml:"projectName,omitempty"`
	ClusterID    *string   `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	ProjectID    *int32    `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	DepartmentID *string   `json:"departmentId,omitempty" yaml:"departmentId,omitempty"`
	CreatedAt    time.Time `json:"createdAt" yaml:"createdAt"`
	CreatedBy    string    `json:"createdBy" yaml:"createdBy"`
	UpdatedAt    time.Time `json:"updatedAt" yaml:"updatedAt"`
	UpdatedBy    string    `json:"updatedBy" yaml:"updatedBy"`
}

// DatasourceStatus contains status information
type DatasourceStatus struct {
	Status  *string `json:"status,omitempty" yaml:"status,omitempty"`
	Message *string `json:"message,omitempty" yaml:"message,omitempty"`
}

// DatasourceSpec contains type-specific specifications
type DatasourceSpec struct {
	NFS       *NFSSpec       `json:"nfs,omitempty" yaml:"nfs,omitempty"`
	PVC       *PVCSpec       `json:"pvc,omitempty" yaml:"pvc,omitempty"`
	S3        *S3Spec        `json:"s3,omitempty" yaml:"s3,omitempty"`
	Git       *GitSpec       `json:"git,omitempty" yaml:"git,omitempty"`
	HostPath  *HostPathSpec  `json:"hostPath,omitempty" yaml:"hostPath,omitempty"`
	ConfigMap *ConfigMapSpec `json:"configMap,omitempty" yaml:"configMap,omitempty"`
	Secret    *SecretSpec    `json:"secret,omitempty" yaml:"secret,omitempty"`
}

// NFSSpec contains NFS-specific fields
type NFSSpec struct {
	Server    *string `json:"server,omitempty" yaml:"server,omitempty"`
	Path      *string `json:"path,omitempty" yaml:"path,omitempty"`
	MountPath *string `json:"mountPath,omitempty" yaml:"mountPath,omitempty"`
	ReadOnly  *bool   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
}

// PVCSpec contains PVC-specific fields
type PVCSpec struct {
	Path        *string    `json:"path,omitempty" yaml:"path,omitempty"`
	ExistingPVC *bool      `json:"existingPvc,omitempty" yaml:"existingPvc,omitempty"`
	ClaimName   *string    `json:"claimName,omitempty" yaml:"claimName,omitempty"`
	ReadOnly    *bool      `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	Ephemeral   *bool      `json:"ephemeral,omitempty" yaml:"ephemeral,omitempty"`
	DataSharing *bool      `json:"dataSharing,omitempty" yaml:"dataSharing,omitempty"`
	ClaimInfo   *ClaimInfo `json:"claimInfo,omitempty" yaml:"claimInfo,omitempty"`
}

// ClaimInfo contains PVC claim information
type ClaimInfo struct {
	Size         *string      `json:"size,omitempty" yaml:"size,omitempty"`
	StorageClass *string      `json:"storageClass,omitempty" yaml:"storageClass,omitempty"`
	AccessModes  *AccessModes `json:"accessModes,omitempty" yaml:"accessModes,omitempty"`
	VolumeMode   *string      `json:"volumeMode,omitempty" yaml:"volumeMode,omitempty"`
}

// AccessModes contains PVC access mode settings
type AccessModes struct {
	ReadWriteOnce *bool `json:"readWriteOnce,omitempty" yaml:"readWriteOnce,omitempty"`
	ReadOnlyMany  *bool `json:"readOnlyMany,omitempty" yaml:"readOnlyMany,omitempty"`
	ReadWriteMany *bool `json:"readWriteMany,omitempty" yaml:"readWriteMany,omitempty"`
}

// S3Spec contains S3-specific fields
type S3Spec struct {
	Bucket           *string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Path             *string `json:"path,omitempty" yaml:"path,omitempty"`
	URL              *string `json:"url,omitempty" yaml:"url,omitempty"`
	AccessKeyAssetID *string `json:"accessKeyAssetId,omitempty" yaml:"accessKeyAssetId,omitempty"`
}

// GitSpec contains Git-specific fields
type GitSpec struct {
	Repository   *string `json:"repository,omitempty" yaml:"repository,omitempty"`
	Branch       *string `json:"branch,omitempty" yaml:"branch,omitempty"`
	Revision     *string `json:"revision,omitempty" yaml:"revision,omitempty"`
	Path         *string `json:"path,omitempty" yaml:"path,omitempty"`
	CredentialID *string `json:"credentialId,omitempty" yaml:"credentialId,omitempty"`
}

// HostPathSpec contains HostPath-specific fields
type HostPathSpec struct {
	Path             *string `json:"path,omitempty" yaml:"path,omitempty"`
	MountPath        *string `json:"mountPath,omitempty" yaml:"mountPath,omitempty"`
	ReadOnly         *bool   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	MountPropagation *string `json:"mountPropagation,omitempty" yaml:"mountPropagation,omitempty"`
}

// ConfigMapSpec contains ConfigMap-specific fields
type ConfigMapSpec struct {
	ConfigMapName *string `json:"configMapName,omitempty" yaml:"configMapName,omitempty"`
	MountPath     *string `json:"mountPath,omitempty" yaml:"mountPath,omitempty"`
	SubPath       *string `json:"subPath,omitempty" yaml:"subPath,omitempty"`
	DefaultMode   *string `json:"defaultMode,omitempty" yaml:"defaultMode,omitempty"`
}

// SecretSpec contains Secret-specific fields
type SecretSpec struct {
	CredentialAssetID *string `json:"credentialAssetId,omitempty" yaml:"credentialAssetId,omitempty"`
	MountPath         *string `json:"mountPath,omitempty" yaml:"mountPath,omitempty"`
	DefaultMode       *string `json:"defaultMode,omitempty" yaml:"defaultMode,omitempty"`
}
