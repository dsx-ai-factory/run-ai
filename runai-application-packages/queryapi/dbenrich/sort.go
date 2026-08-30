package dbenrich

import (
	"strings"

	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

func SortBy(column, order string) orm.Enricher {
	return func(db orm.DB) orm.DB {
		if column == "" {
			return db
		}

		return db.Order(FormatSortClause(column, order))
	}
}

func FormatSortClause(column, order string) string {
	sortOrder := formatSortOrder(order)
	if strings.Contains(column, "%s") {
		return strings.ReplaceAll(column, "%s", sortOrder)
	}

	return column + " " + sortOrder
}

func formatSortOrder(order string) string {
	if strings.ToUpper(order) == "DESC" {
		return "DESC NULLS LAST"
	}

	return "ASC NULLS FIRST"
}
