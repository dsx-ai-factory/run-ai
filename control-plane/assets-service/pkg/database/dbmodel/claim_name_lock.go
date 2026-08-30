package dbmodel

import (
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"gorm.io/gorm"
)

type ClaimNameLockRecord struct {
	// Unique identifier of the lock.
	Id string `json:"id" gorm:"not null;"`
	// The claim name of the PVC.
	ClaimName string `json:"claimName" gorm:"not null;"`
	// The id of the tenant.
	TenantId int32 `json:"tenantId" gorm:"not null;"`
	// The user who created the lock.
	CreatedBy string `json:"createdBy" gorm:"not null;"`
	// The time at which the lock was created
	CreatedAt time.Time `json:"createdAt" gorm:"not null;"`
}

func (ClaimNameLockRecord) TableName() string { return dbtables.TableNamePvcClaimNameLocks }

func (ClaimNameLockRecord) BeforeCreate(tx *gorm.DB) (err error) {
	// Remove all rows created more than 1 minute ago
	cutoffTime := time.Now().Add(-1 * time.Minute)
	tx.Where("created_at < ?", cutoffTime).Delete(&ClaimNameLockRecord{})
	return nil
}
