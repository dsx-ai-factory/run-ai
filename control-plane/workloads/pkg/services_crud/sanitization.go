/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	huggingFaceTokenEnvironmentVariable = "HF_TOKEN"
	ngcAPIKeyEnvironmentVariable        = "NGC_API_KEY"
)

func isSensitivePlaintextEnvironmentVariable(name string) bool {
	return name == huggingFaceTokenEnvironmentVariable || name == ngcAPIKeyEnvironmentVariable
}

func sanitizeWorkload(ctx context.Context, workload *models_crud.Workload) {
	redactWorkloadPlaintextTokenValues(workload)

	currentUser := runaictx.GetString(ctx, runaictx.CfSubjectID)
	if currentUser == nil || *currentUser != workload.CreatedBy {
		sanitizeWorkloadCredentials(workload)
	}
}

func sanitizeWorkloadResponse(ctx context.Context, workload *workloadsApi.SupersetWorkload) {
	redactWorkloadResponsePlaintextTokenValues(workload)

	currentUser := runaictx.GetString(ctx, runaictx.CfSubjectID)
	if currentUser == nil || *currentUser != workload.CreatedBy {
		sanitizeWorkloadCredentialsResponse(workload)
	}
}

func redactWorkloadPlaintextTokenValues(workload *models_crud.Workload) {
	if workload == nil {
		return
	}

	redactSupersetSpecPlaintextTokenValues(workload.Spec)
	redactMasterSpecPlaintextTokenValues(workload.MasterSpec)
}

func redactWorkloadResponsePlaintextTokenValues(workload *workloadsApi.SupersetWorkload) {
	if workload == nil {
		return
	}

	redactSupersetSpecPlaintextTokenValues(workload.Spec)
	if workload.MasterSpec.IsSet() {
		redactMasterSpecPlaintextTokenValues(workload.MasterSpec.Get())
	}
}

func redactUnifiedWorkloadPlaintextTokenValues(workload *workloadsApi.UnifiedWorkload) {
	if workload == nil || !workload.Spec.IsSet() || workload.Spec.Get() == nil {
		return
	}

	spec := workload.Spec.Get()
	if spec.Worker.IsSet() {
		redactSupersetSpecPlaintextTokenValues(spec.Worker.Get())
	}
	if spec.Master.IsSet() {
		redactMasterSpecPlaintextTokenValues(spec.Master.Get())
	}
}

func redactSupersetSpecPlaintextTokenValues(spec *workloadsApi.SupersetSpec) {
	if spec == nil {
		return
	}

	spec.EnvironmentVariables = redactPlaintextTokenValues(spec.EnvironmentVariables)
	if spec.Leader.IsSet() && spec.Leader.Get() != nil {
		spec.Leader.Get().EnvironmentVariables = redactPlaintextTokenValues(spec.Leader.Get().EnvironmentVariables)
	}
	if spec.Worker.IsSet() && spec.Worker.Get() != nil {
		spec.Worker.Get().EnvironmentVariables = redactPlaintextTokenValues(spec.Worker.Get().EnvironmentVariables)
	}
}

func redactMasterSpecPlaintextTokenValues(spec *workloadsApi.MasterSpec) {
	if spec == nil {
		return
	}
	spec.EnvironmentVariables = redactPlaintextTokenValues(spec.EnvironmentVariables)
}

// redactPlaintextTokenValues keeps sensitive variable entries but omits literal values.
// It returns a response-only copy so callers cannot mutate stored environment-variable slices.
func redactPlaintextTokenValues(environmentVariables []workloadsApi.EnvironmentVariable) []workloadsApi.EnvironmentVariable {
	if environmentVariables == nil {
		return nil
	}

	redacted := make([]workloadsApi.EnvironmentVariable, len(environmentVariables))
	copy(redacted, environmentVariables)
	for index := range redacted {
		name := redacted[index].Name.Get()
		if name == nil {
			continue
		}
		if isSensitivePlaintextEnvironmentVariable(*name) {
			redacted[index].UnsetValue()
		}
	}
	return redacted
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
