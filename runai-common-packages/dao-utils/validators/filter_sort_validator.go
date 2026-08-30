package validators

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type FieldsValidation map[string]FieldValidator

const LimitFilterOrCount = 10

var filterPattern = "^(.*)(==|!=|<=|>=|=@|!@|=\\^|=\\$)(.*)$"

var operatorsInfoMap = map[string]operatorInfo{
	"==": {
		Operator:    "=",
		ValueFormat: "%s",
	},
	"!=": {
		Operator:    "is distinct from",
		ValueFormat: "%s",
	},
	"<=": {
		Operator:    "<=",
		ValueFormat: "%s",
	},
	">=": {
		Operator:    ">=",
		ValueFormat: "%s",
	},
	"=@": {
		Operator:    "ILIKE",
		ValueFormat: "%%%s%%",
	},
	"!@": {
		Operator:    "NOT ILIKE",
		ValueFormat: "%%%s%%",
	},
	"=^": {
		Operator:    "ILIKE",
		ValueFormat: "%s%%",
	},
	"=$": {
		Operator:    "ILIKE",
		ValueFormat: "%%%s",
	},
}

type operatorInfo struct {
	Operator    string
	ValueFormat string
}

type ParameterizedFilter struct {
	Field           string
	Operator        string
	FormattedValues []string
}

// NewFieldValidation functions to create field configurations
func NewFieldValidation(columnName string, validator Validator) *FieldValidation {
	return &FieldValidation{
		columnsName: columnName,
		validator:   validator,
	}
}

// CreateFieldConfigMap creates a map of field configurations from a slice of configurations
func CreateFieldConfigMap(fieldToColumn map[string]string) FieldsValidation {
	configMap := make(FieldsValidation, len(fieldToColumn))
	for fieldName, columnName := range fieldToColumn {
		configMap[fieldName] = NewAlwaysValidFieldValidation(columnName)
	}
	return configMap
}

func (fc FieldsValidation) parseQueryParameterized(filter string) (string, string, []string, error) {
	re := regexp.MustCompile(filterPattern)

	filterParts := re.FindStringSubmatch(filter)
	if len(filterParts) != 4 {
		return "", "", nil, fmt.Errorf("filter '%s' has invalid format. Expected: fieldName==value", filter)
	}

	var fieldConfiguration FieldValidator
	var ok bool
	fieldName := filterParts[1]
	operatorName := filterParts[2]
	valuesData := filterParts[3]
	if fieldConfiguration, ok = fc[fieldName]; !ok {
		return "", "", nil, fmt.Errorf("unknown field '%s' in filter '%s'", fieldName, filter)
	}

	var operatorInfo operatorInfo
	if operatorInfo, ok = operatorsInfoMap[operatorName]; !ok {
		return "", "", nil, fmt.Errorf("invalid operator '%s' in filter '%s'", operatorName, filter)
	}

	values := strings.Split(valuesData, "||")
	formattedValues := make([]string, 0)
	var errs error
	if fieldConfiguration.GetValidator() != nil {
		for _, value := range values {
			if err := fieldConfiguration.GetValidator().Validate(value); err != nil {
				errs = errors.Join(errs, fmt.Errorf("validation failed for field '%s' with value '%s': %w", fieldName, value, err))
			}
			formattedValues = append(formattedValues, fmt.Sprintf(operatorInfo.ValueFormat, value))
		}
	}
	return fieldConfiguration.GetColumnsName(), string(operatorInfo.Operator), formattedValues, errs
}

func (fc FieldsValidation) ValidateAndParameterizeFilters(filterQueries []string) ([]ParameterizedFilter, error) {
	filters := []ParameterizedFilter{}

	for _, filterQuery := range filterQueries {
		field, operator, formattedValues, err := fc.parseQueryParameterized(filterQuery)
		if err != nil {
			return filters, fmt.Errorf("error parsing parameterized filter query %s: %w", filterQuery, err)
		}

		if len(formattedValues) == 0 {
			continue
		}

		if len(formattedValues) > LimitFilterOrCount {
			return filters, fmt.Errorf("filter %s exceeds the limit of %d options", field, LimitFilterOrCount)
		}

		var filter = ParameterizedFilter{
			Field:           field,
			Operator:        operator,
			FormattedValues: formattedValues,
		}

		filters = append(filters, filter)
	}

	return filters, nil
}

func (fc FieldsValidation) ValidateSortByField(field string) (string, error) {
	if field == "" {
		return "", nil
	}

	fieldConfig, ok := fc[field]
	if !ok {
		return "", fmt.Errorf("invalid sortBy field %s", field)
	}
	return fieldConfig.GetColumnsName(), nil
}
