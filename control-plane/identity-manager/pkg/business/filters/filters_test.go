/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package business

import (
	"fmt"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/stretchr/testify/assert"
)

func TestParseFilters(t *testing.T) {
	tests := []struct {
		name       string
		filterBy   []string
		sortBy     api.UsersFilterSortFields
		sortOrder  string
		offset     int32
		limit      int32
		expected   *iamModels.UserFilters
		shouldFail bool
	}{
		{name: "Valid Username Filter", filterBy: []string{"username=@john"}, sortBy: "username", sortOrder: "asc", expected: &iamModels.UserFilters{Username: strPtr("john"), SortBy: strPtr("username"), SortOrder: strPtr("asc"), Search: strPtr("test")}, shouldFail: false},
		{name: "Invalid Filter Name", filterBy: []string{"username123=@john"}, sortBy: "username", sortOrder: "asc", expected: nil, shouldFail: true},
		{name: "Valid Creation Time Filter", filterBy: []string{"creationTime>=2023-09-01T00:00:00Z"}, sortBy: "creationTime", sortOrder: "asc", expected: &iamModels.UserFilters{CreationFilter: strPtr(">1693526400"), SortBy: strPtr("creationTime"), SortOrder: strPtr("asc"), Search: strPtr("test")}, shouldFail: false},
		{name: "Valid Operator for Date Field ==", filterBy: []string{"creationTime==2023-09-01T00:00:00Z"}, sortBy: "creationTime", sortOrder: "asc", expected: &iamModels.UserFilters{CreationFilter: strPtr("=1693526400"), SortBy: strPtr("creationTime"), SortOrder: strPtr("asc"), Search: strPtr("test")}, shouldFail: false},
		{name: "Valid Operator for Date Field >=", filterBy: []string{"creationTime>=2023-09-01T00:00:00Z"}, sortBy: "creationTime", sortOrder: "asc", expected: &iamModels.UserFilters{CreationFilter: strPtr(">1693526400"), SortBy: strPtr("creationTime"), SortOrder: strPtr("asc"), Search: strPtr("test")}, shouldFail: false},
		{name: "Invalid Operator for Date Field =>", filterBy: []string{"creationTime=>2023-09-01T00:00:00Z"}, sortBy: "creationTime", sortOrder: "asc", expected: nil, shouldFail: true},
		{name: "Invalid Field in Filter", filterBy: []string{"unknownField==value"}, sortBy: "unknownField", sortOrder: "asc", expected: nil, shouldFail: true},
		{name: "Valid Multiple Filters", filterBy: []string{"username=@john", "createdBy=@admin", "creationTime>=2023-09-01T00:00:00Z"}, sortBy: "username", sortOrder: "asc", expected: &iamModels.UserFilters{Username: strPtr("john"), CreatedBy: strPtr("admin"), CreationFilter: strPtr(">1693526400"), SortBy: strPtr("username"), SortOrder: strPtr("asc"), Search: strPtr("test")}, shouldFail: false},
		{name: "Empty Filter Array", filterBy: []string{}, sortBy: "username", sortOrder: "asc", expected: &iamModels.UserFilters{SortBy: strPtr("username"), SortOrder: strPtr("asc"), Search: strPtr("test")}, shouldFail: false},
		{name: "Invalid Date Format", filterBy: []string{"creationTime>=2023-13-01T00:00:00Z"}, sortBy: "creationTime", sortOrder: "asc", expected: nil, shouldFail: true},
		{name: "Unsupported Operator for Date Field", filterBy: []string{"creationTime=@2023-09-01T00:00:00Z"}, sortBy: "creationTime", sortOrder: "asc", expected: nil, shouldFail: true},
		{name: "Valid and Invalid Combined Filters", filterBy: []string{"username=@john", "unknownField==value"}, sortBy: "username", sortOrder: "asc", expected: nil, shouldFail: true},
		{name: "Valid boolean(false) isLocal", filterBy: []string{"isLocal==true"}, sortBy: "username", sortOrder: "asc", expected: &iamModels.UserFilters{Q: strPtr("runai.is_local:true"), SortBy: strPtr("username"), SortOrder: strPtr("asc"), Search: strPtr("test")}, shouldFail: false},
		{name: "Valid boolean(true) isLocal", filterBy: []string{"isLocal==false"}, sortBy: "username", sortOrder: "asc", expected: &iamModels.UserFilters{Q: strPtr("runai.is_local:false"), SortBy: strPtr("username"), SortOrder: strPtr("asc"), Search: strPtr("test")}, shouldFail: false},
		{name: "Invalid isLocal boolean", filterBy: []string{"isLocal==false123"}, sortBy: "username", sortOrder: "asc", expected: nil, shouldFail: true},
		{name: "Valid Sort By with Limit", filterBy: []string{"username=@john"}, sortBy: "username", sortOrder: "asc", offset: 0, limit: 10, expected: &iamModels.UserFilters{Username: strPtr("john"), SortBy: strPtr("username"), SortOrder: strPtr("asc"), Max: int32ToIntPtr(10), Search: strPtr("test")}, shouldFail: false},
		{name: "Valid Sort By with Offset", filterBy: []string{"createdBy=@admin"}, sortBy: "createdBy", sortOrder: "desc", offset: 5, limit: 0, expected: &iamModels.UserFilters{CreatedBy: strPtr("admin"), SortBy: strPtr("createdBy"), SortOrder: strPtr("desc"), Search: strPtr("test"), Offset: int32ToIntPtr(5)}, shouldFail: false},
		{name: "Invalid Sort Order", filterBy: []string{"username=@john"}, sortBy: "username", sortOrder: "invalid", offset: 0, limit: 10, expected: nil, shouldFail: true},
		{name: "Valid Limit Without Offset", filterBy: []string{"username=@jane"}, sortBy: "username", sortOrder: "asc", offset: 0, limit: 15, expected: &iamModels.UserFilters{Username: strPtr("jane"), SortBy: strPtr("username"), SortOrder: strPtr("asc"), Max: int32ToIntPtr(15), Search: strPtr("test")}, shouldFail: false},
		{name: "Zero Limit with Filters", filterBy: []string{"username=@john"}, sortBy: "username", sortOrder: "asc", offset: 0, limit: 0, expected: &iamModels.UserFilters{Username: strPtr("john"), SortBy: strPtr("username"), SortOrder: strPtr("asc"), Search: strPtr("test")}, shouldFail: false},
		{name: "Large Offset", filterBy: []string{"username=@john"}, sortBy: "username", sortOrder: "asc", offset: 500, limit: 10, expected: &iamModels.UserFilters{Username: strPtr("john"), SortBy: strPtr("username"), SortOrder: strPtr("asc"), Max: int32ToIntPtr(10), Offset: int32ToIntPtr(500), Search: strPtr("test")}, shouldFail: false},
		{name: "No Filters but Valid Sorting", filterBy: []string{}, sortBy: "creationTime", sortOrder: "asc", offset: 0, limit: 20, expected: &iamModels.UserFilters{SortBy: strPtr("creationTime"), SortOrder: strPtr("asc"), Max: int32ToIntPtr(20), Search: strPtr("test")}, shouldFail: false},
		{name: "Multiple Valid Filters with Sorting", filterBy: []string{"username=@john", "isLocal==true"}, sortBy: "creationTime", sortOrder: "asc", offset: 0, limit: 5, expected: &iamModels.UserFilters{Username: strPtr("john"), Q: strPtr("runai.is_local:true"), SortBy: strPtr("creationTime"), SortOrder: strPtr("asc"), Max: int32ToIntPtr(5), Search: strPtr("test")}, shouldFail: false},
		{name: "Invalid Sort Order with No Sorting", filterBy: []string{"username=@john"}, sortBy: "", sortOrder: "random_order", expected: nil, shouldFail: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseFilters(test.filterBy, test.sortBy, test.sortOrder, test.offset, test.limit, "test")
			if test.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result, "Expected: %v, got: %v", test.expected, result)
			}
		})
	}
}

func TestValidateFilter(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedField Field
		expectedOp    Operator
		expectedValue string
		expectedErr   error
	}{
		{"Valid Username Equals", "username=@john_doe", FieldUsername, OperatorContains, "john_doe", nil},
		{"Valid CreatedBy Contains", "createdBy=@admin", FieldCreatedBy, OperatorContains, "admin", nil},
		{"Valid LastLogin GreaterThan", "lastLogin>=2023-01-01T00:00:00Z", FieldLastLogin, ">=", "2023-01-01T00:00:00Z", nil},
		{"Valid CreationTime LessThan", "creationTime<=2023-12-31T23:59:59Z", FieldCreationTime, "<=", "2023-12-31T23:59:59Z", nil},
		{"Valid IsLocal Equals", "isLocal==true", FieldIsLocal, "==", "true", nil},

		{"Invalid Field", "unknownField==value", "", "", "", fmt.Errorf("invalid filter: unknownField==value")},
		{"Invalid Operator", "username!=john_doe", "", "", "", fmt.Errorf("invalid filter: username!=john_doe")},
		{"Invalid Value for Date Field", "lastLogin>invalid_date", "", "", "", fmt.Errorf("invalid filter: lastLogin>invalid_date")},
		{"Invalid Contains Operator on Non-text Field", "lastLogin=@2023-01-01", "", "", "", fmt.Errorf("invalid operator =@ for date field lastLogin, only '>=' or '<=' allowed")},
		{"Missing Operator", "usernamejohn_doe", "", "", "", fmt.Errorf("invalid filter: usernamejohn_doe")},
		{"Missing Value", "username==", "", "", "", fmt.Errorf("invalid filter: username==")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field, op, value, err := ValidateFilter(test.input)
			assert.Equal(t, test.expectedField, field)
			assert.Equal(t, test.expectedOp, op)
			assert.Equal(t, test.expectedValue, value)
			if test.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestConvertOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    Operator
		expected Operator
	}{
		{"Convert Operator >=", ">=", OperatorGreaterThan},
		{"Convert Operator <=", "<=", OperatorLessThan},
		{"Convert Operator ==", "==", OperatorEqual},
		{"Unmapped Operator", "!=", "!="},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ConvertOperator(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsBooleanField(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		expected bool
	}{
		{"Is Local Field", FieldIsLocal, true},
		{"Username Field", FieldUsername, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isBooleanField(test.field)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsTextField(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		expected bool
	}{
		{"Username Field", FieldUsername, true},
		{"Created By Field", FieldCreatedBy, true},
		{"Creation Time Field", FieldCreationTime, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isTextField(test.field)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsDateField(t *testing.T) {
	tests := []struct {
		name           string
		field          Field
		expectedIsDate bool
	}{
		{"Creation Time Field", FieldCreationTime, true},
		{"Last Login Field", FieldLastLogin, true},
		{"Last Updated Field", FieldLastUpdated, true},
		{"Username Field", FieldUsername, false},
		{"Is Local Field", FieldIsLocal, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isDate := isDateField(test.field)

			assert.Equal(t, test.expectedIsDate, isDate)
		})
	}
}

func TestEmptyFilterString(t *testing.T) {
	tests := []struct {
		name       string
		filterBy   []string
		shouldFail bool
	}{
		{"Empty Username Filter", []string{"username=="}, true},
		{"Empty isLocal Filter", []string{"isLocal=="}, true},
		{"Empty Creation Time Filter", []string{"creationTime=="}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseFilters(test.filterBy, "", "", 0, 0, "")
			if test.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseUsersMinimalFilters(t *testing.T) {
	tests := []struct {
		name              string
		filterBy          []string
		limit             int32
		search            string
		ids               []string
		expected          *iamModels.UserFilters
		expectedUsernames []string
		shouldFail        bool
	}{
		{
			name:     "Username exact match sets UsernameExact",
			filterBy: []string{"username==john@run.ai"},
			limit:    10,
			expected: &iamModels.UserFilters{Username: strPtr("john@run.ai"), UsernameExact: boolPtr(true), Max: intPtr(10)},
		},
		{
			name:     "Username contains does not set UsernameExact",
			filterBy: []string{"username=@john"},
			limit:    20,
			expected: &iamModels.UserFilters{Username: strPtr("john"), Max: intPtr(20)},
		},
		{
			name:     "Search filter",
			filterBy: []string{},
			search:   "john",
			expected: &iamModels.UserFilters{Search: strPtr("john"), MinimalSearch: boolPtr(true)},
		},
		{
			name:     "id-only lookup is accepted",
			ids:      []string{"u1", "u2"},
			expected: &iamModels.UserFilters{},
		},
		{
			name:     "search and id together is accepted",
			search:   "john",
			ids:      []string{"u1"},
			expected: &iamModels.UserFilters{Search: strPtr("john"), MinimalSearch: boolPtr(true)},
		},
		{
			name:              "Multiple == filterBy returns bulk usernames",
			filterBy:          []string{"username==a@run.ai", "username==b@run.ai", "username==c@run.ai"},
			expected:          &iamModels.UserFilters{},
			expectedUsernames: []string{"a@run.ai", "b@run.ai", "c@run.ai"},
		},
		{
			name:              "Bulk usernames respects limit",
			filterBy:          []string{"username==a@run.ai", "username==b@run.ai"},
			limit:             5,
			expected:          &iamModels.UserFilters{Max: intPtr(5)},
			expectedUsernames: []string{"a@run.ai", "b@run.ai"},
		},
		{
			name:       "No filterBy, search, or id rejected",
			filterBy:   []string{},
			limit:      0,
			shouldFail: true,
		},
		{
			name:       "Too many ids rejected",
			ids:        []string{"u1", "u2", "u3", "u4", "u5", "u6", "u7", "u8", "u9", "u10", "u11", "u12", "u13", "u14", "u15", "u16", "u17", "u18", "u19", "u20", "u21", "u22", "u23", "u24", "u25", "u26"},
			shouldFail: true,
		},
		{
			name:       "Too many filterBy entries rejected",
			filterBy:   []string{"username==a", "username==b", "username==c", "username==d", "username==e", "username==f", "username==g", "username==h", "username==i", "username==j", "username==k", "username==l", "username==m", "username==n", "username==o", "username==p", "username==q", "username==r", "username==s", "username==t", "username==u", "username==v", "username==w", "username==x", "username==y", "username==z"},
			shouldFail: true,
		},
		{
			name:       "Single-char search rejected",
			search:     "a",
			shouldFail: true,
		},
		{
			name:       "Empty id rejected",
			ids:        []string{"u1", "", "u3"},
			shouldFail: true,
		},
		{
			name:       "Whitespace-only id rejected",
			ids:        []string{"u1", "   ", "u3"},
			shouldFail: true,
		},
		{
			name:       "Multiple filterBy with =@ rejected",
			filterBy:   []string{"username=@alice", "username=@bob"},
			shouldFail: true,
		},
		{
			name:       "Mixed operators in multi-filterBy rejected",
			filterBy:   []string{"username==a@run.ai", "username=@bob"},
			shouldFail: true,
		},
		{
			name:       "Unsupported field rejected",
			filterBy:   []string{"firstName=@john"},
			shouldFail: true,
		},
		{
			name:       "Date field rejected",
			filterBy:   []string{"creationTime>=2023-09-01T00:00:00Z"},
			shouldFail: true,
		},
		{
			name:       "isLocal field rejected",
			filterBy:   []string{"isLocal==true"},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, bulkUsernames, err := ParseUsersMinimalFilters(tt.filterBy, tt.limit, tt.search, tt.ids)
			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
				assert.Equal(t, tt.expectedUsernames, bulkUsernames)
			}
		})
	}
}

func intPtr(v int) *int {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}
