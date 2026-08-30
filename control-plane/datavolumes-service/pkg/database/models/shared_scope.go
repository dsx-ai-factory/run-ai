package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
)

const (
	SharedScopeTableName = "shared_scopes"
)

type SharedScope struct {
	DatavolumeId uuid.UUID
	ScopeType    api.ScopeType
	ScopeId      string
	ScopeName    string
	TenantId     uint
	CreatedAt    time.Time
	CreatedBy    string

	ProjectId    string
	DepartmentId string
}

func (SharedScope) TableName() string {
	return SharedScopeTableName
}
