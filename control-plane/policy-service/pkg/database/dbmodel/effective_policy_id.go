package dbmodel

import (
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
)

type EffectivePolicyId struct {
	PolicyId *uuid.UUID
}

func (EffectivePolicyId) TableName() string {
	return dbtables.EffectivePoliciesTable
}
