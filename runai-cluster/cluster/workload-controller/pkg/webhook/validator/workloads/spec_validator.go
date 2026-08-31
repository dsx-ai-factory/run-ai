/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"errors"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// SpecValidator validates the spec of a workload
type SpecValidator struct {
}

// NewSpecValidator creates a new SpecValidator
func NewSpecValidator() *SpecValidator {
	return &SpecValidator{}
}

// Validate validates the spec of a workload
func (v *SpecValidator) Validate(ctx context.Context, workload *jobs.JobWorkload) error {
	log := logf.FromContext(ctx)

	spec := workload.Spec
	log.Info("checking spec for", "workload", workload.Name)

	// Checking for deprecated fields
	/* TODO: uncomment this
	if spec.Name != nil {
		log.Info("workload spec contains a name")
		return errors.New("spec.name is deprecated, please use metadata.name")
	} */
	if spec.NamePrefix != nil {
		log.Info("workload spec contains a name prefix")
		return errors.New("spec.namePrefix is deprecated, please use metadata.generateName")
	}

	if err := validatePriorityClassName(ctx, spec); err != nil {
		return err
	}

	workloadKind, err := common.GetWorkloadInitials(workload.Kind)
	if err != nil {
		return err
	}
	workloadInfo := builders.WorkloadInfo{
		Name:      workload.Name,
		Namespace: workload.Namespace,
		Initials:  string(workloadKind),
	}
	if err := validateImagePullSecrets(spec, workloadInfo); err != nil {
		return err
	}

	return nil
}

func validatePriorityClassName(ctx context.Context, spec jobs.JobSpec) error {
	log := logf.FromContext(ctx)

	preemtibleInitialized := spec.Preemptible.IsValueInitialized()
	nonPreemptibleInitialized := spec.NonPreemptible.IsValueInitialized()
	priorityClassNameInitialized := spec.PriorityClassName.IsValueInitialized()

	if preemtibleInitialized && priorityClassNameInitialized {
		return errors.New("Both Preemtible and PriorityClassName are set. Please use only PriorityClassName.")
	}

	if nonPreemptibleInitialized && priorityClassNameInitialized {
		return errors.New("Both NonPreemtible and PriorityClassName are set. Please use only PriorityClassName.")
	}

	log.Info("priorityClassName validation passed")
	return nil
}

func validateImagePullSecrets(spec jobs.JobSpec, workloadInfo builders.WorkloadInfo) error {
	_, err := builders.ParseImagePullSecrets(spec.ImagePullSecrets.Deref(), workloadInfo)
	if err != nil {
		return err
	}

	return nil
}
