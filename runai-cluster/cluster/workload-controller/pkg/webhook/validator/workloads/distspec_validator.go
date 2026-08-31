/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"errors"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/mpi"
	v1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	jaxbuilder "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/jax"
)

var (
	ErrorRestartPolicyNeverWithBackOf = errors.New("Back of limit cannot be set with restart policy value 'Never'")
)

const (
	RestartPolicyEmptyString     = ""
	RestartPolicyNeverString     = string(v1.RestartPolicyNever)
	RestartPolicyAlwaysString    = string(v1.RestartPolicyAlways)
	RestartPolicyOnFailureString = string(v1.RestartPolicyOnFailure)
)

type DistSpecValidator struct {
}

func NewDistSpecValidator() *DistSpecValidator {
	return &DistSpecValidator{}
}

func (d *DistSpecValidator) Validate(ctx context.Context, workload *jobs.JobWorkload) error {
	var (
		err error
		log = logf.FromContext(ctx).
			WithValues("workload", workload.Name).
			WithValues("framework", workload.Spec.JobType)
	)

	log.V(4).Info("validating dist spec for workload")
	distWorkload := v2alpha1.NewFromJobWorkload(workload)
	switch distWorkload.Spec.JobType {
	case constants.KindMpiJob:
		err = mpi.ValidateWorkload(distWorkload)
	case constants.KindJAXJob:
		err = jaxbuilder.ValidateWorkload(distWorkload)
	}

	if err != nil {
		return err
	}
	return restartPolicyValidation(ctx, distWorkload)
}

func restartPolicyValidation(ctx context.Context, workload *v2alpha1.DistributedWorkload) error {
	log := logf.FromContext(ctx)
	// backofflimit is unset, or set to 0 - no need to check the restart policy
	if workload.Spec.RunPolicy.Deref().BackoffLimit.Deref() == 0 {
		log.V(4).Info("allowed - no back of limit")
		return nil
	}

	// worker check for never restart policy
	restartPolicy, err := retrieveRestartPolicyValue(workload.Spec.RestartPolicy)
	if err != nil {
		return err
	}
	if restartPolicy != RestartPolicyNeverString {
		log.V(4).Info("allowed - worker restart policy not never", "specRestartPolicy", restartPolicy)
		return nil
	}

	// only worker, with restart policy 'Never', and back of limit is not 0.
	if workload.MasterSpec == nil {
		return ErrorRestartPolicyNeverWithBackOf
	}

	_, err = builders.ParseImagePullSecrets(workload.MasterSpec.ImagePullPolicy.Deref(), builders.WorkloadInfo{
		Name:      workload.Name,
		Namespace: workload.Namespace,
		Initials:  common.KindDistributedWorkloadInitial,
	})
	if err != nil {
		return err
	}

	// master check for never restart policy
	restartPolicy, err = retrieveRestartPolicyValue(workload.MasterSpec.RestartPolicy)
	if err != nil {
		return err
	}
	if restartPolicy == RestartPolicyNeverString {
		return ErrorRestartPolicyNeverWithBackOf
	}
	log.V(4).Info("allowed - master restart policy not never", "specRestartPolicy", restartPolicy)
	return nil
}

func retrieveRestartPolicyValue(value *flds.StringField) (string, error) {
	if value == nil || value.Value == nil {
		return RestartPolicyEmptyString, nil
	}
	switch value.Deref() {
	case RestartPolicyEmptyString:
		return RestartPolicyEmptyString, nil
	case RestartPolicyAlwaysString:
		return RestartPolicyAlwaysString, nil
	case RestartPolicyNeverString:
		return RestartPolicyNeverString, nil
	case RestartPolicyOnFailureString:
		return RestartPolicyOnFailureString, nil
	default:
		return "", fmt.Errorf("invalid restartPolicy value '%s'", value.Deref())
	}
}
