/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"errors"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/runaipublic"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KnativeValidator struct {
	k8sClient client.Client
}

func NewKnativeValidator(k8s client.Client) *KnativeValidator {
	return &KnativeValidator{k8sClient: k8s}
}

func (v *KnativeValidator) Validate(ctx context.Context, workload *jobs.JobWorkload) error {
	if v.isInferenceWorkload(workload) {
		if isAvailable, err := v.isKnativeAvailable(ctx, workload); !isAvailable {
			return err
		}
	}
	return nil
}

func (v *KnativeValidator) isInferenceWorkload(workload *jobs.JobWorkload) bool {
	return strings.Compare(workload.Kind, constants.KindInferenceWorkload) == 0
}

func (v *KnativeValidator) isKnativeAvailable(ctx context.Context, workload *jobs.JobWorkload) (bool, error) {
	runaiPublic := runaipublic.NewRunaiPublicHandler(ctx, v.k8sClient)
	if v.isInferenceWorkload(workload) {
		isAvailable, reason, err := runaiPublic.IsKnativeAvailable()
		if err != nil {
			return false, err
		}
		if !isAvailable {
			msg := "Knative is not available"
			if len(reason) > 0 {
				msg = msg + ", reason: " + reason
			}
			return false, errors.New(msg)
		}
	}

	return true, nil
}
