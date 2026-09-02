/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"gorm.io/gorm"
)

const (
	ValidationTable = "validation"
)

type Validation struct {
	gorm.Model
	Data string `gorm:"not null" json:"data"`
}

func (v *Validation) TableName() string {
	return ValidationTable
}
