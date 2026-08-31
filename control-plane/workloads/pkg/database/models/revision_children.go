/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/google/uuid"
)

type RevisionChild struct {
	RevisionID uuid.UUID
	Revision   Revision `gorm:"foreignKey:RevisionID"`
	ChildID    uuid.UUID
	ChildType  string
}

func (RevisionChild) TableName() string {
	return "revisions_children"
}
