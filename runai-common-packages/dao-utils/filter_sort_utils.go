/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils/validators"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const LimitFilterOrCount = 10

func ApplyTSSearchParam(query *gorm.DB, search string, columnNames ...string) *gorm.DB {
	if len(columnNames) == 0 {
		return query
	}

	search = sanitizeSearch(search)

	if search == "" {
		return query
	}

	whereClause := ""
	args := []interface{}{}

	for i, columnName := range columnNames {
		tsCondition := columnName + " @@ to_tsquery('english', ? || ':*')"
		if i > 0 {
			whereClause += " OR "
		}
		whereClause += tsCondition
		args = append(args, search)
	}

	if whereClause != "" {
		query = query.Where("("+whereClause+")", args...)
	}

	return query
}

func sanitizeSearch(search string) string {
	search = strings.ReplaceAll(search, " ", "-")
	re := regexp.MustCompile("[^a-zA-Z0-9_@.-]")
	search = re.ReplaceAllString(search, "")

	return search
}

func ValidateAndParameterizeApiFilters(filterQueries []string, fieldToColumn map[string]string) ([]validators.ParameterizedFilter, error) {
	config := validators.CreateFieldConfigMap(fieldToColumn)
	fieldFilters, err := config.ValidateAndParameterizeFilters(filterQueries)
	if err != nil {
		return nil, err
	}

	return fieldFilters, nil
}

func ApplyParameterizedFilters(query *gorm.DB, filters []validators.ParameterizedFilter) *gorm.DB {
	for _, filter := range filters {
		query = query.Where(buildWhereClause(filter))
	}

	return query
}

func buildWhereClause(filter validators.ParameterizedFilter) clause.Where {
	return clause.Where{
		Exprs: []clause.Expression{buildOrConditions(filter)},
	}
}

func buildOrConditions(filter validators.ParameterizedFilter) clause.OrConditions {
	orConditions := clause.OrConditions{}
	for _, value := range filter.FormattedValues {
		orConditions.Exprs = append(orConditions.Exprs, clause.Expr{
			SQL:  fmt.Sprintf("%s %s ?", filter.Field, filter.Operator),
			Vars: []interface{}{value},
		})
	}
	return orConditions
}

func ApplySortByParam(query *gorm.DB, column string, order string) (*gorm.DB, error) {
	if strings.ToUpper(order) == "DESC" {
		order = "DESC NULLS LAST"
	} else {
		order = "ASC NULLS FIRST"
	}

	sort := fmt.Sprintf("%s %s", column, order)
	if strings.Contains(column, "%s") {
		sort = strings.ReplaceAll(column, "%s", order)
	}

	query = query.Order(sort)

	return query, nil
}

func ValidateSortByField(field string, fieldToColumn map[string]string) (string, error) {
	config := validators.CreateFieldConfigMap(fieldToColumn)
	return config.ValidateSortByField(field)
}
