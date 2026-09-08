/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaictx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTenantId(t *testing.T) {
	ctx := context.WithValue(context.Background(), CfTenantId, "12345")
	value, err := ParseTenantId(ctx)
	assert.Nil(t, err)
	if value != "12345" {
		t.Errorf("expected '12345', got %v", value)
	}
}
func TestParseSubjectId(t *testing.T) {
	ctx := context.WithValue(context.Background(), CfSubjectID, "test@example.com")
	value, err := ParseSubjectId(ctx)
	assert.Nil(t, err)
	if value != "test@example.com" {
		t.Errorf("expected 'test@example.com', got %v", value)
	}
}
func TestParseGroups(t *testing.T) {
	ctx := context.WithValue(context.Background(), CfGroups, []string{"group-a"})
	values := ParseGroups(ctx)
	assert.Equal(t, len(values), 1)
	if values[0] != "group-a" {
		t.Errorf("expected 'group-a', got %v", values[0])
	}
}
func TestParseAuthData(t *testing.T) {
	ctx := context.WithValue(context.Background(), CfTenantId, "12345")
	ctx = context.WithValue(ctx, CfSubjectID, "test@example.com")
	ctx = context.WithValue(ctx, CfGroups, []string{"group-a"})
	tenantId, subjectId, groups, err := ParseAuthData(ctx)
	assert.Nil(t, err)

	if tenantId != "12345" {
		t.Errorf("expected '12345', got %v", tenantId)
	}
	if subjectId != "test@example.com" {
		t.Errorf("expected 'test@example.com', got %v", subjectId)
	}
	assert.Equal(t, len(groups), 1)
	if groups[0] != "group-a" {
		t.Errorf("expected 'group-a', got %v", groups[0])
	}
}
func TestIsManagementTenant(t *testing.T) {
	ctx := context.WithValue(context.Background(), CfIsManagementTenant, "true")
	value := IsManagementTenant(ctx)

	if !value {
		t.Errorf("expected is management tenant to be true")
	}
}
