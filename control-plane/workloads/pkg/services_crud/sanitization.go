/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"

	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func sanitizeWorkload(ctx context.Context, workload *models_crud.Workload) {
	currentUser := runaictx.GetString(ctx, runaictx.CfSubjectID)
	if currentUser == nil || *currentUser != workload.CreatedBy {
		sanitizeWorkloadCredentials(workload)
	}
}

func sanitizeWorkloadResponse(ctx context.Context, workload *workloadsApi.SupersetWorkload) {
	currentUser := runaictx.GetString(ctx, runaictx.CfSubjectID)
	if currentUser == nil || *currentUser != workload.CreatedBy {
		sanitizeWorkloadCredentialsResponse(workload)
	}
}

// sanitizeEnvironmentVariables removes environment variables that contain user credentials
func sanitizeEnvironmentVariables(envVars []workloadsApi.EnvironmentVariable) []workloadsApi.EnvironmentVariable {
	if len(envVars) == 0 {
		return nil
	}

	sanitized := make([]workloadsApi.EnvironmentVariable, 0, len(envVars))
	for _, envVar := range envVars {
		if !envVar.UserCredential.IsSet() {
			sanitized = append(sanitized, envVar)
		}
	}
	return sanitized
}

// sanitizeWorkloadCredentials removes user credentials from the workload's database model
func sanitizeWorkloadCredentials(workload *models_crud.Workload) {
	if workload.Spec != nil {
		workload.Spec.EnvironmentVariables = sanitizeEnvironmentVariables(workload.Spec.EnvironmentVariables)
		workload.Spec.ImagePullSecrets = sanitizeImagePullSecrets(workload.Spec.ImagePullSecrets)
	}

	if workload.MasterSpec != nil {
		workload.MasterSpec.EnvironmentVariables = sanitizeEnvironmentVariables(workload.MasterSpec.EnvironmentVariables)
		workload.MasterSpec.ImagePullSecrets = sanitizeImagePullSecrets(workload.MasterSpec.ImagePullSecrets)
	}
}

// sanitizeWorkloadCredentialsResponse removes user credentials from the workload's API response
func sanitizeWorkloadCredentialsResponse(workload *workloadsApi.SupersetWorkload) {
	if workload.Spec != nil {
		workload.Spec.EnvironmentVariables = sanitizeEnvironmentVariables(workload.Spec.EnvironmentVariables)
		workload.Spec.ImagePullSecrets = sanitizeImagePullSecrets(workload.Spec.ImagePullSecrets)
	}

	if workload.MasterSpec.IsSet() && workload.MasterSpec.Get() != nil {
		workload.MasterSpec.Get().EnvironmentVariables = sanitizeEnvironmentVariables(workload.MasterSpec.Get().EnvironmentVariables)
		workload.MasterSpec.Get().ImagePullSecrets = sanitizeImagePullSecrets(workload.MasterSpec.Get().ImagePullSecrets)
	}
}

// sanitizeImagePullSecrets removes image pull secrets that contain user credentials
func sanitizeImagePullSecrets(imagePullSecrets []workloadsApi.ImagePullSecret) []workloadsApi.ImagePullSecret {
	if len(imagePullSecrets) == 0 {
		return nil
	}

	sanitized := make([]workloadsApi.ImagePullSecret, 0, len(imagePullSecrets))

	for _, secret := range imagePullSecrets {
		if !secret.UserCredential.IsSet() {
			sanitized = append(sanitized, secret)
		}
	}
	return sanitized
}
