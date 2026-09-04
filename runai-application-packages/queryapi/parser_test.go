/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package queryapi

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUUIDValidator(t *testing.T) {
	validator := UUIDValidator{}

	for range 3 {
		require.NoError(t, validator.Validate(uuid.New().String()))
	}

	invalidUUIDs := []string{
		"invalid-uuid",
		"123e4567-e89b-12d3-a456",
		"123e4567-e89b-12d3-a456-42661417400g",
		"",
	}
	for _, value := range invalidUUIDs {
		require.Error(t, validator.Validate(value))
	}
}

func TestDateValidator(t *testing.T) {
	validator := DateValidator{}

	validDates := []string{
		"2023-01-15",
		"2023-01-15T10:30:00Z",
		"2023-01-15T10:30:00.123Z",
		"2023-01-15 10:30:00",
	}
	for _, value := range validDates {
		require.NoError(t, validator.Validate(value))
	}

	invalidDates := []string{"invalid-date", "2023-13-45", ""}
	for _, value := range invalidDates {
		require.Error(t, validator.Validate(value))
	}
}

func TestParser_Filters(t *testing.T) {
	parser := NewParser(APIFieldsMapping{
		"clusterId": {Column: "cluster_id", Validators: []Validator{UUIDValidator{}}},
		"createdAt": {Column: "created_at", Validators: []Validator{DateValidator{}}},
		"name":      {Column: "name"},
	})

	t.Run("valid filters", func(t *testing.T) {
		filters, err := parser.Filters([]string{
			"clusterId==" + uuid.New().String(),
			"createdAt==2023-01-15",
			"name==alice",
		})
		require.NoError(t, err)
		require.Len(t, filters, 3)
		require.Equal(t, "cluster_id", filters[0].Field)
		require.Equal(t, "=", filters[0].Operator)
		require.Equal(t, "%s", filters[0].ValueFormat)
		require.Equal(t, "created_at", filters[1].Field)
		require.Equal(t, "name", filters[2].Field)
		require.Equal(t, []string{"alice"}, filters[2].Values)
		require.Equal(t, "%s", filters[2].ValueFormat)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		_, err := parser.Filters([]string{"clusterId==invalid-uuid"})
		require.Error(t, err)
	})

	t.Run("unknown field", func(t *testing.T) {
		_, err := parser.Filters([]string{"unknown==value"})
		require.Error(t, err)
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := parser.Filters([]string{"name"})
		require.Error(t, err)
	})

	t.Run("invalid operator", func(t *testing.T) {
		_, err := parser.Filters([]string{"name>>alice"})
		require.Error(t, err)
	})

	t.Run("or values", func(t *testing.T) {
		filters, err := parser.Filters([]string{"name==alice||bob"})
		require.NoError(t, err)
		require.Len(t, filters, 1)
		require.Equal(t, []string{"alice", "bob"}, filters[0].Values)
	})

	t.Run("empty value", func(t *testing.T) {
		filters, err := parser.Filters([]string{"name=="})
		require.NoError(t, err)
		require.Len(t, filters, 1)
		require.Equal(t, []string{""}, filters[0].Values)
	})

	t.Run("empty value with or", func(t *testing.T) {
		filters, err := parser.Filters([]string{"name==alice||"})
		require.NoError(t, err)
		require.Len(t, filters, 1)
		require.Equal(t, []string{"alice", ""}, filters[0].Values)
	})

	t.Run("too many or values", func(t *testing.T) {
		_, err := parser.Filters([]string{
			"name==a||b||c||d||e||f||g||h||i||j||k",
		})
		require.Error(t, err)
	})

	t.Run("like operators", func(t *testing.T) {
		filters, err := parser.Filters([]string{"name=@alice", "name!@alice", "name=^alice", "name=$alice"})
		require.NoError(t, err)
		require.Len(t, filters, 4)
		require.Equal(t, "ILIKE", filters[0].Operator)
		require.Equal(t, "%%%s%%", filters[0].ValueFormat)
		require.Equal(t, "NOT ILIKE", filters[1].Operator)
		require.Equal(t, "%%%s%%", filters[1].ValueFormat)
		require.Equal(t, "ILIKE", filters[2].Operator)
		require.Equal(t, "%s%%", filters[2].ValueFormat)
		require.Equal(t, "ILIKE", filters[3].Operator)
		require.Equal(t, "%%%s", filters[3].ValueFormat)
	})

	t.Run("comparison operators", func(t *testing.T) {
		filters, err := parser.Filters([]string{"createdAt>=2023-01-01", "createdAt<=2023-12-31"})
		require.NoError(t, err)
		require.Len(t, filters, 2)
		require.Equal(t, ">=", filters[0].Operator)
		require.Equal(t, "<=", filters[1].Operator)
	})

	t.Run("not equal operator", func(t *testing.T) {
		filters, err := parser.Filters([]string{"name!=alice"})
		require.NoError(t, err)
		require.Len(t, filters, 1)
		require.Equal(t, "is distinct from", filters[0].Operator)
	})
}

func TestParser_GetMapping(t *testing.T) {
	parser := NewParser(APIFieldsMapping{
		"name":      {Column: "name"},
		"createdAt": {Column: "created_at"},
	})

	t.Run("valid field", func(t *testing.T) {
		column, err := parser.GetMapping("name")
		require.NoError(t, err)
		require.Equal(t, "name", column)
	})

	t.Run("mapped column name", func(t *testing.T) {
		column, err := parser.GetMapping("createdAt")
		require.NoError(t, err)
		require.Equal(t, "created_at", column)
	})

	t.Run("invalid field", func(t *testing.T) {
		_, err := parser.GetMapping("invalid")
		require.Error(t, err)
	})

	t.Run("empty field", func(t *testing.T) {
		column, err := parser.GetMapping("")
		require.NoError(t, err)
		require.Equal(t, "", column)
	})

	t.Run("case sensitive field", func(t *testing.T) {
		_, err := parser.GetMapping("Name")
		require.Error(t, err)
	})
}
