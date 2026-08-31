/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodels

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestBaseModelBeforeCreateAssignsUUIDWhenUnset(t *testing.T) {
	m := &BaseModel{}

	err := m.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.UUID{}, m.ID)
}

func TestBaseModelBeforeCreatePreservesExistingID(t *testing.T) {
	id := uuid.New()
	m := &BaseModel{ID: id}

	err := m.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.Equal(t, id, m.ID)
}

func TestDeleteBaseModelBeforeCreateAssignsUUIDWhenUnset(t *testing.T) {
	m := &DeleteBaseModel{}

	err := m.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.UUID{}, m.ID)
}

func TestDeleteBaseModelBeforeCreatePreservesExistingID(t *testing.T) {
	id := uuid.New()
	m := &DeleteBaseModel{ID: id}

	err := m.BeforeCreate(nil)

	assert.NoError(t, err)
	assert.Equal(t, id, m.ID)
}

func TestGormDeletedAtToNullableTimeReturnsNilWhenNotDeleted(t *testing.T) {
	assert.Nil(t, GormDeletedAtToNullableTime(gorm.DeletedAt{}))
}

func TestGormDeletedAtToNullableTimeReturnsTimestampWhenDeleted(t *testing.T) {
	now := time.Now()
	deletedAt := gorm.DeletedAt{Time: now, Valid: true}

	result := GormDeletedAtToNullableTime(deletedAt)

	assert.NotNil(t, result)
	assert.Equal(t, now, *result)
}
