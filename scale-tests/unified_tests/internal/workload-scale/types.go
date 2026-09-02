/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_scale

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/workload-scale/aggregated_report/common/stages"
	workloadclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

type FailureCategory string

const (
	FailureCategoryNone          FailureCategory = ""
	FailureCategoryCreation      FailureCategory = "creation_failed"
	FailureCategoryTimeout       FailureCategory = "timeout"
	FailureCategoryPhaseMismatch FailureCategory = "phase_mismatch"
	FailureCategoryMissingData   FailureCategory = "missing_data"
)

type WorkloadCreateSampleResponse struct {
	Name              string               `json:"name"`
	ProjectName       string               `json:"namespace"`
	Id                string               `json:"id"`
	SuccessAtCreation bool                 `json:"success_at_creation"`
	CreateFailMessage string               `json:"create_fail_message"`
	ExpectedEndPhase  workloadclient.Phase `json:"expected_end_phase"`
	CreateTime        *int64               `json:"create_time"`
}

type WorkloadFinalRawDataResponse struct {
	WorkloadName         string                                `json:"workload_name"`
	ProjectName          string                                `json:"project_name"`
	Phase                workloadclient.Phase                  `json:"phase"`
	CPPhaseUpdatedAt     *int64                                `json:"cp_phase_updated_at"`
	CPCreatedAt          *int64                                `json:"cp_created_at"`
	WorkloadK8sData      WorkloadK8sData                       `json:"workload_k8s_data"`
	CreatedSuccessfully  bool                                  `json:"created_successfully"`
	ReachedExpectedPhase bool                                  `json:"reached_expected_phase"`
	Stages               map[stages.WorkloadStage]StageDetails `json:"stages"`
	FailureReason        string                                `json:"failure_reason,omitempty"`
	FailureCategory      FailureCategory                       `json:"failure_category,omitempty"`
}

type WorkloadK8sData struct {
	WorkloadCreationTime *int64               `json:"workload_creation_time"`
	ClusterPhase         workloadclient.Phase `json:"cluster_phase"`
	ClusterPhaseTime     *int64               `json:"cluster_phase_time"`
	Pods                 []PodK8sData         `json:"pods"`
}

type PodK8sData struct {
	WorkloadName  string `json:"workload_name"`
	CreationTime  *int64 `json:"creation_time"`
	ScheduledTime *int64 `json:"scheduled_time"`
}

type StageDetails struct {
	Duration  time.Duration
	Completed bool
}
