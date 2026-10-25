/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actionhistory

import (
	"context"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
)

type WorkloadToLog struct {
	ID           uuid.UUID
	ClusterID    uuid.UUID
	ProjectID    string
	DepartmentID string
	Name         string
}

// NewWorkloadActionRecord builds a history record from the authenticated request context.
func NewWorkloadActionRecord(ctx context.Context, recordID uuid.UUID, action api.WorkloadActionType, workload WorkloadToLog, acceptedAt time.Time) (*models.WorkloadActionHistory, *api.Error) {
	if ctx == nil {
		return nil, api.NewValidationError("missing authenticated context")
	}
	tenantID, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return nil, apiErr
	}
	// PostgreSQL bigint is signed, while ParseTenantId returns uint.
	if uint64(tenantID) > math.MaxInt64 {
		return nil, api.NewValidationError("invalid action tenant")
	}
	subjectID, subjectType, apiErr := subjectFromContext(ctx)
	if apiErr != nil {
		return nil, apiErr
	}
	if recordID == uuid.Nil {
		var err error
		recordID, err = uuid.NewRandom()
		if err != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, "failed to generate workload action ID", err)
		}
	}
	return &models.WorkloadActionHistory{
		ID: recordID,
		WorkloadActionScope: models.WorkloadActionScope{
			TenantID:     tenantID,
			ClusterID:    workload.ClusterID,
			ProjectID:    workload.ProjectID,
			DepartmentID: workload.DepartmentID,
		},
		WorkloadID:   workload.ID,
		WorkloadName: workload.Name,
		Subject:      subjectID,
		// Examples: "User"/"user" for a person; "ServiceAccount"/"service-account" for an application.
		// Older identity labels include "App"/"app". Preserve the authenticated value unchanged.
		SubjectType: subjectType,
		Action:      string(action),
		AcceptedAt:  acceptedAt.UTC().Truncate(time.Microsecond),
	}, nil
}

func subjectFromContext(ctx context.Context) (string, string, *api.Error) {
	subjectID := *runaictx.GetString(ctx, runaictx.CfSubjectID)
	subjectType := runaictx.GetString(ctx, runaictx.CfSubjectType)
	if subjectType == nil || strings.TrimSpace(*subjectType) == "" {
		return "", "", api.NewValidationError("missing action subject type")
	}
	return subjectID, *subjectType, nil
}
