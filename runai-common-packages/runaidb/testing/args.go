package testing

import (
	"database/sql/driver"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// AnyArg returns a matcher that accepts any value for a query argument,
func AnyArg() sqlmock.Argument {
	return sqlmock.AnyArg()
}

// AnyTime returns a matcher that accepts any time.Time argument.
func AnyTime() sqlmock.Argument {
	return anyTimeMatcher{}
}

type anyTimeMatcher struct{}

func (anyTimeMatcher) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}
