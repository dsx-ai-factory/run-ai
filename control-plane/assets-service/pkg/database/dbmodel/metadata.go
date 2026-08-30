package dbmodel

import (
	"time"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"gorm.io/gorm"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

type AssetMetaRec struct {
	AssetMeta `gorm:"embedded"`
}

func (AssetMetaRec) TableName() string {
	return dbtables.TableNameAssetsMeta
}

func (AssetMetaRec) AssetIdColumn() string {
	return dbtables.AssetMetaColumnId
}

type AssetMeta struct {
	// The name of the asset.
	Name string `json:"name" gorm:"not null;"`

	Scope api.Scope `json:"scope" gorm:"not null;"`

	// The id of the cluster. It is obsolete for asset creation request, left here only for backward compatibility.
	ClusterId *string `json:"clusterId,omitempty" gorm:""`

	// The id of the department. Must be specified for department scoped assets.
	DepartmentId *string `json:"departmentId,omitempty" gorm:""`

	// The id of the project. Must be specified for project scoped assets.
	ProjectId *int32 `json:"projectId,omitempty" gorm:""`

	// The asset will be deleted automatically. This is intended for internal use.
	AutoDelete *bool `json:"autoDelete,omitempty" gorm:""`

	// Unique identifier of the asset.
	Id string `json:"id" gorm:"not null;"`

	Kind api.AssetKind `json:"kind" gorm:"not null;"`

	// The id of the tenant.
	TenantId *int32 `json:"tenantId,omitempty" gorm:""`

	// The user who created the asset.
	CreatedBy string `json:"createdBy" gorm:"not null;"`

	// The time at which the asset were created
	CreatedAt time.Time `json:"createdAt" gorm:"not null;"`

	// The user who updated the asset.
	UpdatedBy string `json:"updatedBy" gorm:"not null;"`

	// The time at which the asset has been updated
	UpdatedAt time.Time `json:"updatedAt" gorm:"not null;"`

	DeletedBy *string `json:"deletedBy"`

	DeletedAt gorm.DeletedAt `json:"deletedAt"`

	// The name of the project that the asset is associated with, for project scoped assets.
	ProjectName *string `json:"projectName,omitempty" gorm:""`

	// The description of the asset.
	Description *string `json:"description,omitempty" gorm:""`

	UpdateCount int64 `json:"updateCount,omitempty"`

	// Internal name for remembering the previous name of the asset, used for internal sync purposes only
	// does not appear in the APIs
	PrevName string `json:"prevName,omitempty"`

	// Private indicates if the asset is private or not
	Private *bool `json:"private,omitempty" gorm:""`
}

// ToApiAssetMeta convert dbmodel.AssetMeta to api.AssetMeta
func (a *AssetMeta) ToApiAssetMeta() api.AssetMeta {
	return api.AssetMeta{
		Name:         a.Name,
		Scope:        a.Scope,
		ClusterId:    a.ClusterId,
		DepartmentId: a.DepartmentId,
		ProjectId:    a.ProjectId,
		AutoDelete:   a.AutoDelete,
		Id:           a.Id,
		Kind:         a.Kind,
		TenantId:     a.TenantId,
		CreatedBy:    a.CreatedBy,
		CreatedAt:    a.CreatedAt,
		UpdatedBy:    a.UpdatedBy,
		UpdatedAt:    a.UpdatedAt,
		DeletedBy:    convutils.Deref(a.DeletedBy, ""),
		DeletedAt:    a.DeletedAt.Time,
		ProjectName:  a.ProjectName,
		Description:  a.Description,
		UpdateCount:  a.UpdateCount,
		Private:      a.Private,
	}
}

func (a *AssetMeta) ToAssetMetaSync() api.AssetMetaForSync {
	return api.AssetMetaForSync{
		Id:           a.Id,
		Name:         a.Name,
		Description:  a.Description,
		Scope:        a.Scope,
		ClusterId:    a.ClusterId,
		DepartmentId: a.DepartmentId,
		ProjectId:    a.ProjectId,
		AutoDelete:   a.AutoDelete,
		Kind:         a.Kind,
		PrevName:     a.PrevName,
		Private:      a.Private,
	}
}

// IsPrivate returns true if the asset is private, false otherwise
func (a *AssetMeta) IsPrivate() bool {
	return convutils.BoolIsTrue(a.Private)
}

// CombinedRec holds result of join assets_metadata and supported types queries
type CombinedRec struct {
	api.AssetMeta              `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (p CombinedRec) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p CombinedRec) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}
