/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbenrich

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"
)

func Filters(filters []queryapi.Filter) orm.Enricher {
	return func(db orm.DB) orm.DB {
		for _, filter := range filters {
			if len(filter.Values) == 0 {
				continue
			}
			db = applyFilter(db, filter)
		}
		return db
	}
}

func applyFilter(db orm.DB, filter queryapi.Filter) orm.DB {
	orConditions := clause.OrConditions{}

	for _, value := range filter.Values {
		orConditions.Exprs = append(orConditions.Exprs, clause.Expr{
			SQL:  fmt.Sprintf("%s %s ?", filter.Field, filter.Operator),
			Vars: []interface{}{formatValue(filter, value)},
		})
	}

	return db.Where(clause.Where{Exprs: []clause.Expression{orConditions}})
}

func formatValue(filter queryapi.Filter, value string) string {
	if filter.ValueFormat == "" {
		return value
	}

	return fmt.Sprintf(filter.ValueFormat, value)
}
