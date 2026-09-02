/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package queryapi

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	limitFilterOrCount = 10
	nameIndex          = 1
	operatorIndex      = 2
	valuesIndex        = 3
)

var filterPattern = regexp.MustCompile(`^(.*)(==|!=|<=|>=|=@|!@|=\^|=\$)(.*)$`)

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

type Parser struct {
	mapping APIFieldsMapping
}

func NewParser(mapping APIFieldsMapping) *Parser {
	return &Parser{mapping: mapping}
}

func (p *Parser) Filters(apiQueryFilter []string) ([]Filter, error) {
	filters := make([]Filter, 0, len(apiQueryFilter))
	for _, filterQuery := range apiQueryFilter {
		filter, err := p.parseFilter(filterQuery)
		if err != nil {
			return filters, fmt.Errorf("error parsing parameterized filter query %s: %w", filterQuery, err)
		}

		if filter != nil {
			filters = append(filters, *filter)
		}
	}

	return filters, nil
}

func (p *Parser) GetMapping(field string) (string, error) {
	if field == "" {
		return "", nil
	}

	fieldMapping, ok := p.mapping[field]
	if !ok {
		return "", fmt.Errorf("invalid sortBy field %s", field)
	}

	return fieldMapping.Column, nil
}

func (p *Parser) parseFilter(filter string) (*Filter, error) {
	filterParts := filterPattern.FindStringSubmatch(filter)
	if len(filterParts) != 4 {
		return nil, fmt.Errorf("filter '%s' has invalid format. Expected: fieldName==value", filter)
	}

	fieldMapping, ok := p.mapping[filterParts[nameIndex]]
	if !ok {
		return nil, fmt.Errorf("unknown field '%s' in filter '%s'", filterParts[nameIndex], filter)
	}

	operator, ok := operatorsInfoMap[filterParts[operatorIndex]]
	if !ok {
		return nil, fmt.Errorf("invalid operator '%s' in filter '%s'", filterParts[operatorIndex], filter)
	}

	values := strings.Split(filterParts[valuesIndex], "||")
	if len(values) > limitFilterOrCount {
		return nil, fmt.Errorf("filter %s exceeds the limit of %d options", fieldMapping.Column, limitFilterOrCount)
	}

	if err := p.validate(fieldMapping, values, filterParts); err != nil {
		return nil, err
	}

	return &Filter{
		Field:       fieldMapping.Column,
		Operator:    operator.Operator,
		ValueFormat: operator.ValueFormat,
		Values:      values,
	}, nil
}

func (p *Parser) validate(fieldMapping FieldMapping, values []string, filterParts []string) error {
	if len(fieldMapping.Validators) == 0 {
		return nil
	}

	var validationErr error
	for _, value := range values {
		for _, validator := range fieldMapping.Validators {
			if err := validator.Validate(value); err != nil {
				validationErr = errors.Join(validationErr, fmt.Errorf("validation failed for field '%s' with value '%s': %w", filterParts[nameIndex], value, err))
			}
		}
	}

	return validationErr
}
