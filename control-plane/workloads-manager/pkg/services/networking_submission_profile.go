/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/database/models"
)

// networkingSubmissionProfileForRecord classifies the server-owned creation
// origin. It deliberately does not interpret caller-controlled configuration.
func networkingSubmissionProfileForRecord(record *models.WorkloadsQueueRecord) api.NetworkingSubmissionProfile {
	if isGeneratedNativeNetworkingRecord(record) {
		return api.NETWORKINGSUBMISSIONPROFILE_GENERATED_NATIVE
	}
	if isGeneratedNIMNetworkingRecord(record) {
		return api.NETWORKINGSUBMISSIONPROFILE_GENERATED_NIM
	}
	return api.NETWORKINGSUBMISSIONPROFILE_DIRECT_V2
}

// isGeneratedNativeNetworkingRecord uses the persisted trusted-origin flag and
// route instead of duplicating the composer's workload-to-GVK mapping.
func isGeneratedNativeNetworkingRecord(record *models.WorkloadsQueueRecord) bool {
	return record != nil &&
		record.TranslatedFromV1 &&
		record.InputRequestPath == InternalWorkloadsAPIPath
}

// isGeneratedNIMNetworkingRecord uses the request route captured by the server;
// the NIM API owns the generated workload shape.
func isGeneratedNIMNetworkingRecord(record *models.WorkloadsQueueRecord) bool {
	return record != nil && record.InputRequestPath == NimServicesAPIPath
}

// workloadToClusterUpdateWithNetworkingSubmissionProfile derives the profile
// at cluster egress, after all caller-controlled input has been persisted and
// separated from trusted record metadata. It deliberately does not persist the
// profile, so existing records require no migration.
func workloadToClusterUpdateWithNetworkingSubmissionProfile(record *models.WorkloadsQueueRecord) api.WorkloadV2Update {
	update := record.ToClusterUpdate()
	if update.DesiredPhase == api.DESIREDPHASE_DELETED ||
		update.Configuration.NetworkingEndpoints == nil ||
		len(*update.Configuration.NetworkingEndpoints) == 0 {
		return update
	}

	update.Configuration.NetworkingSubmissionProfile = networkingSubmissionProfileForRecord(record)
	return update
}
