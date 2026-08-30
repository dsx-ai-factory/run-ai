package dbmodel

import (
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
)

type ApprovedDeletions struct {
	ApprovedDeletionsKey `gorm:"embedded"`

	AvoidDeletionsOfObsoletePolicies bool
}

type ApprovedDeletionsKey struct {
	TenantId  int32
	ClusterId *string
}

func (ApprovedDeletions) TableName() string {
	return dbtables.ApprovedDeletionsTable
}
