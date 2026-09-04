/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validators

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
)

func TestUUIDValidator(t *testing.T) {
	validator := &UUIDValidator{}

	// Valid UUIDs
	validUUIDs := []string{
		uuid.New().String(),
		uuid.New().String(),
		uuid.New().String(),
	}

	for _, uuidStr := range validUUIDs {
		if err := validator.Validate(uuidStr); err != nil {
			t.Errorf("Expected valid UUID %s, got error: %v", uuidStr, err)
		}
	}

	// Invalid UUIDs
	invalidUUIDs := []string{
		"invalid-uuid",
		"123e4567-e89b-12d3-a456",
		"123e4567-e89b-12d3-a456-42661417400g", // invalid character
		"",
	}

	for _, uuidStr := range invalidUUIDs {
		if err := validator.Validate(uuidStr); err == nil {
			t.Errorf("Expected invalid UUID %s to return error", uuidStr)
		}
	}
}

func TestDateValidator(t *testing.T) {
	validator := &DateValidator{}

	// Valid dates
	validDates := []string{
		"2023-01-15",
		"2023-01-15T10:30:00Z",
		"2023-01-15T10:30:00.123Z",
		"2023-01-15 10:30:00",
		time.Now().Format(time.RFC3339),
	}

	for _, date := range validDates {
		if err := validator.Validate(date); err != nil {
			t.Errorf("Expected valid date %s, got error: %v", date, err)
		}
	}

	// Invalid dates
	invalidDates := []string{
		"invalid-date",
		"2023-13-45", // invalid month and day
		"",
	}

	for _, date := range invalidDates {
		if err := validator.Validate(date); err == nil {
			t.Errorf("Expected invalid date %s to return error", date)
		}
	}
}

func TestFieldConfiguration(t *testing.T) {
	// Create field configurations
	userIDConfig := NewUUIDFieldValidation("cluster_id")
	createdAtConfig := NewDateFieldValidation("created_at")

	// Test field configuration methods
	if userIDConfig.GetColumnsName() != "cluster_id" {
		t.Errorf("Expected column name 'cluster_id', got '%s'", userIDConfig.GetColumnsName())
	}

	if userIDConfig.GetValidator() == nil {
		t.Error("Expected validator to be set")
	}

	// Test validation through field configuration
	if err := userIDConfig.GetValidator().Validate("123e4567-e89b-12d3-a456-426614174000"); err != nil {
		t.Errorf("Expected valid UUID, got error: %v", err)
	}

	if err := createdAtConfig.GetValidator().Validate("2023-01-15"); err != nil {
		t.Errorf("Expected valid date, got error: %v", err)
	}
}

func TestValidateFilterByParamWithValidators(t *testing.T) {
	// Create field configurations
	fieldConfigs := FieldsValidation{
		"clusterId":      NewUUIDFieldValidation("cluster_id"),
		"createdAt":      NewDateFieldValidation("created_at"),
		"name":           NewAlwaysValidFieldValidation("name"),
		"workloadsCount": NewAlwaysValidFieldValidation("workloads_count"),
	}

	// Test valid filters
	validFilters := []string{
		"clusterId==123e4567-e89b-12d3-a456-426614174000",
		"createdAt==2023-01-15",
	}

	filters, err := fieldConfigs.ValidateAndParameterizeFilters(validFilters)
	if err != nil {
		t.Errorf("Expected valid filters, got error: %v", err)
	}
	if len(filters) != 2 {
		t.Errorf("Expected 2 filters, got %d", len(filters))
	}

	// Test invalid filters
	invalidFilters := []string{
		"clusterId==invalid-uuid",
		"createdAt==invalid-date",
	}

	_, err = fieldConfigs.ValidateAndParameterizeFilters(invalidFilters)
	if err == nil {
		t.Error("Expected validation error for invalid filters")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test helper functions
	userIDConfig := NewUUIDFieldValidation("cluster_id")
	if userIDConfig.GetColumnsName() != "cluster_id" {
		t.Errorf("Expected column name 'cluster_id', got '%s'", userIDConfig.GetColumnsName())
	}

	createdAtConfig := NewDateFieldValidation("created_at")
	if createdAtConfig.GetColumnsName() != "created_at" {
		t.Errorf("Expected column name 'created_at', got '%s'", createdAtConfig.GetColumnsName())
	}

	// Test CreateFieldConfigMap
	configMap := CreateFieldConfigMap(map[string]string{
		"clusterId": "cluster_id",
		"createdAt": "created_at",
		"name":      "name",
	})
	if len(configMap) != 3 {
		t.Errorf("Expected 2 configurations, got %d", len(configMap))
	}
}

func TestParseQueryRegexPattern(t *testing.T) {
	// Test that the regex pattern correctly matches valid field names
	fieldConfigs := FieldsValidation{
		"clusterId":      NewUUIDFieldValidation("cluster_id"),
		"createdAt":      NewDateFieldValidation("created_at"),
		"name":           NewAlwaysValidFieldValidation("name"),
		"workloadsCount": NewAlwaysValidFieldValidation("workloads_count"),
	}

	// Test valid field names that should match
	validFilters := []string{
		"clusterId==123e4567-e89b-12d3-a456-426614174000",
		"createdAt==2023-01-15",
		"name==test",
		"workloadsCount>=5",
	}

	for _, filter := range validFilters {
		_, _, _, err := (&fieldConfigs).parseQueryParameterized(filter)
		if err != nil {
			t.Errorf("Expected valid filter '%s', got error: %v", filter, err)
		}
	}

	// Test invalid field names that should not match
	invalidFilters := []string{
		"123field==value",   // starts with number
		"field-name==value", // contains hyphen
		"field.name==value", // contains dot
		"field name==value", // contains space
	}

	for _, filter := range invalidFilters {
		_, _, _, err := (&fieldConfigs).parseQueryParameterized(filter)
		if err == nil {
			t.Errorf("Expected invalid filter '%s' to return error", filter)
		}
	}
}

func TestParseQuery_notSupported(t *testing.T) {
	fieldsConfig := CreateFieldConfigMap(map[string]string{
		"type": "type",
	})
	_, _, _, err := fieldsConfig.parseQueryParameterized("name==x")
	assert.NotNil(t, err)
}

func TestParseQuery_single(t *testing.T) {
	fieldsConfig := CreateFieldConfigMap(map[string]string{
		"type": "type",
	})
	column, operator, values, err := fieldsConfig.parseQueryParameterized("type==x")
	assert.Nil(t, err)
	assert.Equal(t, column, "type")
	assert.Equal(t, operator, operatorsInfoMap["=="].Operator)
	assert.Len(t, values, 1)
}

func TestParseQuery_multiple(t *testing.T) {
	fieldsConfig := CreateFieldConfigMap(map[string]string{
		"type": "type",
	})
	column, operator, values, err := fieldsConfig.parseQueryParameterized("type==x||y||z")
	assert.Nil(t, err)
	assert.Equal(t, column, "type")
	assert.Equal(t, operator, operatorsInfoMap["=="].Operator)
	assert.Len(t, values, 3)
}

func TestParseQuery_Empty(t *testing.T) {
	fieldsConfig := CreateFieldConfigMap(map[string]string{
		"type": "type",
	})
	column, operator, values, err := fieldsConfig.parseQueryParameterized("type==")
	assert.Nil(t, err)
	assert.Equal(t, column, "type")
	assert.Equal(t, operator, operatorsInfoMap["=="].Operator)
	assert.Len(t, values, 1)
	assert.Equal(t, values[0], "")
}

func TestParseQuery_Empty_Or(t *testing.T) {
	fieldsConfig := CreateFieldConfigMap(map[string]string{
		"type": "type",
	})
	column, operator, values, err := fieldsConfig.parseQueryParameterized("type==x||y||")
	assert.Nil(t, err)
	assert.Equal(t, column, "type")
	assert.Equal(t, operator, operatorsInfoMap["=="].Operator)
	assert.Len(t, values, 3)
	assert.Equal(t, values[0], "x")
	assert.Equal(t, values[1], "y")
	assert.Equal(t, values[2], "")
}

func TestUserInputValidator_Valid(t *testing.T) {
	validator := &UserInputValidator{}
	assert.NoError(t, validator.Validate("test"))
}

func TestUserInputValidator_Invalid(t *testing.T) {
	validator := &UserInputValidator{}
	longString := strings.Repeat("A", 1001)
	err := validator.Validate(longString)
	assert.EqualError(t, err, "input too long")
}
