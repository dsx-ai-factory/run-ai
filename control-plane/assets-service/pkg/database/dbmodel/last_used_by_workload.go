package dbmodel

import (
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"gorm.io/gorm"
)

// LastUsedByWorkloadRecord stores usage info for assets
type LastUsedByWorkloadRecord struct {
	ID                   uuid.UUID
	TenantId             int32
	ProjectId            int32  `gorm:"uniqueIndex:byprojectworloadtypeandassetid"`
	AssetId              string `gorm:"uniqueIndex:byprojectworloadtypeandassetid"`
	WorkloadType         string `gorm:"uniqueIndex:byprojectworloadtypeandassetid"`
	WorkloadId           string
	WorkloadCreationTime time.Time
}

func (LastUsedByWorkloadRecord) TableName() string {
	return dbtables.TableNameLastUsedByWorkload
}

func (LastUsedByWorkloadRecord) AssetIdColumn() string {
	return dbtables.LastUsedByWorkloadColumnAssetId
}

func (l *LastUsedByWorkloadRecord) BeforeCreate(_ *gorm.DB) (err error) {
	l.ID = uuid.New()
	return nil
}
