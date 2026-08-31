/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package enforcer

import (
	"context"
	"encoding/json"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/errors"
)

// SubresourceApplier  is internal interface mainly for unit testing
// it will run the apply function on the merge stack and return the result
type SubresourceApplier interface {
	Apply(ctx context.Context, mergeStack *MergeStack) ([]byte, error)
}

type SubresourceApplierImpl struct {
}

func (self *SubresourceApplierImpl) Apply(ctx context.Context, mergeStack *MergeStack) ([]byte, error) {
	err := mergeStack.MergeAndValidate(ctx, MergeAndValidateOptions{})
	if err != nil {
		return nil, errors.NewWorkloadPolicyValidationError(err.Error())
	}

	mergeStack.CleanTheWorkload(mergeStack.MergedValuesLevel(), CleanOptions{})
	marshalData, err := json.Marshal(mergeStack.MergedValues().Get())
	if err != nil {
		return nil, err
	}
	return marshalData, nil
}
