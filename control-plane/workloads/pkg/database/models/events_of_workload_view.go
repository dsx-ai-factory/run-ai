/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
)

type TimeoutEventsOfWorkloadView struct {
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
	ProjectName             string
	WorkloadType            string
	SubmittedBy             string
}

func (TimeoutEventsOfWorkloadView) TableName() string {
	return "timeout_events_of_workload_view"
}
