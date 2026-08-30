package dbmodel

import (
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

// PolicyComplianceRecord stores compliance information of a given asset to the effective policy of a given project.
// we keep it here in order not to re-validate the policy each time compliance info is requested
type PolicyComplianceRecord struct {
	TenantId     int32  `gorm:"uniqueIndex:by_asset_workload_and_project"`
	ProjectId    int32  `gorm:"uniqueIndex:by_asset_workload_and_project"`
	AssetId      string `gorm:"uniqueIndex:by_asset_workload_and_project"`
	WorkloadType string `gorm:"uniqueIndex:by_asset_workload_and_project"`
	PolicyTime   time.Time
	Aligned      bool                       // is the asset aligned with the policy
	Reason       []api.ComplianceInfoReason `json:"usedBy,omitempty" gorm:"serializer:json; type:jsonb;"`
	ImposedAsset bool                       // asset is imposed on all workloads by the administrator
}

func (PolicyComplianceRecord) TableName() string {
	return dbtables.TableNamePolicyCompliance
}

func (PolicyComplianceRecord) AssetIdColumn() string {
	return dbtables.PolicyComplianceColumnAssetId
}
