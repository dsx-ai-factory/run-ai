/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/common/topology"
	apijobs "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type TopologyValidator struct {
	enabled bool
}

func NewTopologyValidator(enabled bool) *TopologyValidator {
	return &TopologyValidator{
		enabled: enabled,
	}
}

func (tv *TopologyValidator) Validate(ctx context.Context, job *apijobs.JobWorkload) error {
	log := logf.FromContext(ctx).
		WithValues("workload", job.Name).
		WithValues("framework", job.Spec.JobType)

	if !tv.enabled {
		log.Info("workload topology is disabled, feature disabled")
		return nil
	}

	return topology.ValidateTopologyAnnotation(ctx, job.Annotations)
}
