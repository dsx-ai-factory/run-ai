/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"database/sql"
)

type PolicyWithOptionalEffectiveId struct {
	Policy            `gorm:"embedded"`
	EffectivePolicyId `gorm:"foreignKey:ID;references:ID"`
}

type PolicyWithPartialEffective struct {
	Policy `gorm:"embedded"`

	UpdateCount     int64
	ValidationError *string
}

// PolicyWithOptionalEffective is used for join of policy <-> effective_policy
//
//	--- NOTE! ---
//
// This structure is only suitable for reading only non-deleted records
// Do not use it for reading deleted records, because the DeletedAt which appears in both
// Policy and EffectivePolicy confuses GORM.
//
// PolicyDeleted / EffectiveDeleted must not use gorm.DeletedAt: GORM treats that type as a
// soft-delete column and appends WHERE policies.policy_deleted IS NULL even though these
// are only SELECT aliases (see dao.selectForOptionalEffective).
type PolicyWithOptionalEffective struct {
	PolicyDeleted    sql.NullTime `gorm:"column:policy_deleted"`
	EffectiveDeleted sql.NullTime `gorm:"column:effective_deleted"`
	Policy
	EffectivePolicy `gorm:"foreignKey:ID;references:ID"`
}

type PolicyAndEffective struct {
	EffectivePolicy
	Policy `gorm:"foreignKey:PolicyId;references:PolicyId"`
}

// NeedToRecalculateEffective returns indication if the effective policy needs to be re-calculated.
// Re-calculation is needed if the policy itself exist, but its effective policy has been deleted.
func (pae PolicyAndEffective) NeedToRecalculateEffective() bool {
	return !pae.Policy.IsDeleted() && pae.EffectivePolicy.IsDeleted()
}
