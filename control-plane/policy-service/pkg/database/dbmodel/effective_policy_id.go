/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbtables"
)

type EffectivePolicyId struct {
	PolicyId *uuid.UUID
}

func (EffectivePolicyId) TableName() string {
	return dbtables.EffectivePoliciesTable
}
