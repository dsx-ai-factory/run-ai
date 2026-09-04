/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jax

import (
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
)

// ValidateWorkload validates the configuration of a JAX job for unsupported fields.
func ValidateWorkload(workload *v2alpha1.DistributedWorkload) error {
	// Ensure the workload spec does not use a master configuration
	if workload.Spec.NoMaster.IsValueInitialized() && !workload.Spec.DistributedFields.NoMaster.Deref() {
		return MasterNotSupportedError
	}
	if workload.MasterSpec != nil {
		return MasterNotSupportedError
	}
	// Validate unsupported fields in the workload spec
	if err := validateUnsupportedField(workload.Spec.MpiJob.Deref().SlotsPerWorker); err != nil {
		return err
	}
	if err := validateUnsupportedField(workload.Spec.MpiJob.Deref().SshAuthMountPath); err != nil {
		return err
	}
	if err := validateUnsupportedField(workload.Spec.PyTorchJob.Deref().MaxReplicas); err != nil {
		return err
	}
	if err := validateUnsupportedField(workload.Spec.PyTorchJob.Deref().MinReplicas); err != nil {
		return err
	}

	return nil
}

// validateUnsupportedField checks if a specific field in the workload spec is unsupported.
func validateUnsupportedField(field field_types.Field) error {
	if field.IsValueInitialized() {
		return fmt.Errorf("invalid workload fields %v, %w", field, FieldsNotSupportedError)
	}
	return nil
}
