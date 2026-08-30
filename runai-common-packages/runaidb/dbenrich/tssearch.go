package dbenrich

import (
	"regexp"
	"strings"

	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

var nonSearchCharsRegex = regexp.MustCompile("[^a-zA-Z0-9_@.-]")

func TSSearch(search string, columns ...string) orm.Enricher {
	return func(db orm.DB) orm.DB {
		if len(columns) == 0 {
			return db
		}

		search = sanitizeSearch(search)
		if search == "" {
			return db
		}
		stmt := db.Actual().Statement
		conditions := make([]string, len(columns))
		args := make([]interface{}, len(columns))
		for i, column := range columns {
			conditions[i] = stmt.Quote(column) + " @@ to_tsquery('english', ? || ':*')"
			args[i] = search
		}

		return db.Where("("+strings.Join(conditions, " OR ")+")", args...)
	}
}

func sanitizeSearch(search string) string {
	search = strings.ReplaceAll(search, " ", "-")
	return nonSearchCharsRegex.ReplaceAllString(search, "")
}
