package models

import (
	"time"

	"github.com/google/uuid"
)

type WorkloadEventView struct {
	WorkloadID              uuid.UUID
	TenantID                uint
	EventID                 uuid.UUID
	ClusterID               uuid.UUID
	Type                    string
	Message                 string
	Reason                  string
	CreatedAt               time.Time
	Source                  string
	InvolvedObjectUid       string
	InvolvedObjectKind      string
	InvolvedObjectName      string
	InvolvedObjectNamespace string
}

func (WorkloadEventView) TableName() string {
	return "workloads_events_view"
}
