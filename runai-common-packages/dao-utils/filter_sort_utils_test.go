package dao_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSearchSanitize(t *testing.T) {
	search := sanitizeSearch("  %^ &3.. $ very-123-complex__-__str!ing @%#@!")
	assert.Equal(t, search, "---3..--very-123-complex__-__string-@@")
}

func TestValidateSortByField(t *testing.T) {
	fieldToColumn := map[string]string{
		"name":      "name",
		"age":       "age",
		"email":     "email",
		"createdAt": "created_at",
		"clusterId": "cluster_id",
	}

	tests := []struct {
		name          string
		field         string
		fieldToColumn map[string]string
		expected      string
		expectError   bool
	}{
		{
			name:          "valid field",
			field:         "name",
			fieldToColumn: fieldToColumn,
			expected:      "name",
			expectError:   false,
		},
		{
			name:          "valid field with different column name",
			field:         "createdAt",
			fieldToColumn: fieldToColumn,
			expected:      "created_at",
			expectError:   false,
		},
		{
			name:          "invalid field",
			field:         "invalid",
			fieldToColumn: fieldToColumn,
			expected:      "",
			expectError:   true,
		},
		{
			name:          "empty field",
			field:         "",
			fieldToColumn: fieldToColumn,
			expected:      "",
			expectError:   false,
		},
		{
			name:          "case sensitive field",
			field:         "Name",
			fieldToColumn: fieldToColumn,
			expected:      "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateSortByField(tt.field, tt.fieldToColumn)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyTSSearchParamWithSpecialCharacters(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Test with special characters that should be sanitized
	search := "test@example.com!@#$%"
	columnNames := []string{"email"}

	query := db.Table("test")
	result := ApplyTSSearchParam(query, search, columnNames...)

	assert.NotNil(t, result)
	// The search should be sanitized to remove special characters
	sanitizedSearch := sanitizeSearch(search)
	assert.NotEqual(t, search, sanitizedSearch)
}

func TestApplySortByParamWithNullHandling(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	tests := []struct {
		name   string
		column string
		order  string
	}{
		{"asc with nulls first", "name", "ASC"},
		{"desc with nulls last", "created_at", "DESC"},
		{"mixed case order", "name", "asc"},
		{"mixed case order desc", "created_at", "desc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := db.Table("test")
			result, err := ApplySortByParam(query, tt.column, tt.order)

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

type testType struct{}

func TestApplyParameterizedFilters_OneFilter(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	simpleFilter := []string{"name==test"}
	fieldToColumn := map[string]string{
		"name": "name",
	}
	parameterizedFilters, err := ValidateAndParameterizeApiFilters(simpleFilter, fieldToColumn)
	assert.NoError(t, err)
	result := ApplyParameterizedFilters(db, parameterizedFilters)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	stmt := result.Find(&testType{}).Statement.SQL.String()
	assert.Equal(t, "SELECT * FROM `test_types` WHERE name = ?", stmt)
}

func TestApplyParameterizedFilters_MultipleFilters(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	simpleFilter := []string{"name==test", "age>=18"}
	fieldToColumn := map[string]string{
		"name": "name",
		"age":  "age",
	}
	parameterizedFilters, err := ValidateAndParameterizeApiFilters(simpleFilter, fieldToColumn)
	assert.NoError(t, err)
	result := ApplyParameterizedFilters(db, parameterizedFilters)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	stmt := result.Find(&testType{}).Statement.SQL.String()
	assert.Equal(t, "SELECT * FROM `test_types` WHERE name = ? AND age >= ?", stmt)
}

func TestApplyParameterizedFilters_OneFilter_WithOr(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	complexFilters := []string{
		"name==test||admin",
	}

	fieldToColumn := map[string]string{
		"name": "name",
	}
	parameterizedFilters, err := ValidateAndParameterizeApiFilters(complexFilters, fieldToColumn)
	assert.NoError(t, err)
	result := ApplyParameterizedFilters(db, parameterizedFilters)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	stmt := result.Find(&testType{}).Statement.SQL.String()
	assert.Equal(t, "SELECT * FROM `test_types` WHERE (name = ? OR name = ?)", stmt)
}

func TestApplyParameterizedFilters_MultipleFilters_WithOr(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	complexFilters := []string{
		"name==test||admin",
		"created_at>=2023-01-01",
	}

	fieldToColumn := map[string]string{
		"name":       "name",
		"created_at": "created_at",
	}
	parameterizedFilters, err := ValidateAndParameterizeApiFilters(complexFilters, fieldToColumn)
	assert.NoError(t, err)
	result := ApplyParameterizedFilters(db, parameterizedFilters)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	stmt := result.Find(&testType{}).Statement.SQL.String()
	assert.Equal(t, "SELECT * FROM `test_types` WHERE (name = ? OR name = ?) AND created_at >= ?", stmt)
}

func TestApplyParameterizedFilters_MultipleFilters_WithOr_QueryWithExistingWhere(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)
	db = db.Where("tenant_id = 2")
	complexFilters := []string{
		"name==test||admin",
		"created_at>=2023-01-01",
	}

	fieldToColumn := map[string]string{
		"name":       "name",
		"created_at": "created_at",
	}
	parameterizedFilters, err := ValidateAndParameterizeApiFilters(complexFilters, fieldToColumn)
	assert.NoError(t, err)
	result := ApplyParameterizedFilters(db, parameterizedFilters)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	stmt := result.Find(&testType{}).Statement.SQL.String()
	assert.Equal(t, "SELECT * FROM `test_types` WHERE tenant_id = 2 AND (name = ? OR name = ?) AND created_at >= ?", stmt)
}

func TestApplyParameterizedFilters_EmptyFilters(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	emptyFilters := []string{}
	fieldToColumn := map[string]string{}
	parameterizedFilters, err := ValidateAndParameterizeApiFilters(emptyFilters, fieldToColumn)
	assert.NoError(t, err)
	result := ApplyParameterizedFilters(db, parameterizedFilters)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	stmt := result.Find(&testType{}).Statement.SQL.String()
	assert.Equal(t, "SELECT * FROM `test_types`", stmt)
}

func TestApplyParameterizedFilters_FilterWithEmptyValue(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	emptyFilters := []string{"name=="}
	fieldToColumn := map[string]string{
		"name": "name",
	}
	parameterizedFilters, err := ValidateAndParameterizeApiFilters(emptyFilters, fieldToColumn)
	assert.NoError(t, err)
	result := ApplyParameterizedFilters(db, parameterizedFilters)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	stmt := result.Find(&testType{}).Statement.SQL.String()
	assert.Equal(t, "SELECT * FROM `test_types` WHERE name = ?", stmt)
}

func TestApplyParameterizedFilters_FilterWithEmptyValue_Or(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	emptyFilters := []string{"name==test||"}
	fieldToColumn := map[string]string{
		"name": "name",
	}
	parameterizedFilters, err := ValidateAndParameterizeApiFilters(emptyFilters, fieldToColumn)
	assert.NoError(t, err)
	result := ApplyParameterizedFilters(db, parameterizedFilters)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	stmt := result.Find(&testType{}).Statement.SQL.String()
	assert.Equal(t, "SELECT * FROM `test_types` WHERE (name = ? OR name = ?)", stmt)
}
