package dbenrich

import (
	"fmt"

	"github.com/run-ai/runai/runai-application-packages/queryapi"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
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
