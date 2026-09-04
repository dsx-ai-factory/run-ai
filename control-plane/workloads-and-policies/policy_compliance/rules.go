/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

var (
	PolicyCanEdit  = workloadsApi.POLICYRULEENUM_CAN_EDIT
	PolicyMax      = workloadsApi.POLICYRULEENUM_MAX
	PolicyMin      = workloadsApi.POLICYRULEENUM_MIN
	PolicyStep     = workloadsApi.POLICYRULEENUM_STEP
	PolicyOptions  = workloadsApi.POLICYRULEENUM_OPTIONS
	PolicyRequired = workloadsApi.POLICYRULEENUM_REQUIRED
	PolicyCanAdd   = workloadsApi.POLICYRULEENUM_CAN_ADD
	PolicyLocked   = workloadsApi.POLICYRULEENUM_LOCKED
	PolicyBlocked  = workloadsApi.POLICYRULEENUM_BLOCKED

	PolicyDefaultFrom = "defaultFrom"

	PolicySourceOfRule = "sourceOfRule"
)

var RuleNames = datastructs.NewSet[string](
	itemized.PolicyInstances,
	itemized.PolicyAttributes,
	string(PolicyCanEdit),
	string(PolicyMin),
	string(PolicyMax),
	string(PolicyStep),
	string(PolicyOptions),
	string(PolicyRequired),
	string(PolicyCanAdd),
	string(PolicyLocked),
	string(PolicyBlocked),

	PolicySourceOfRule,
	PolicyDefaultFrom)
