package models

import (
	"time"

	"github.com/google/uuid"
)

type WorkloadHistoryRecord struct {
	WorkloadID uuid.UUID
	ClusterID  uuid.UUID
	TenantID   uint
	CreatedAt  time.Time

	Phase        *string
	PhaseMessage *string

	EventID                 *uuid.UUID
	Type                    *string
	Message                 *string
	Reason                  *string
	Source                  *string
	InvolvedObjectUid       *string
	InvolvedObjectKind      *string
	InvolvedObjectName      *string
	InvolvedObjectNamespace *string
}
