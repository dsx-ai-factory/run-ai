package client_utils

import (
	"fmt"
	"strings"
)

type FilterOperator string

// Operators --> (==|!=|<=|>=|=@|!@|=\\^|=\\$)
const (
	EqualOp           FilterOperator = "=="
	NotEqualOp        FilterOperator = "!="
	SmallerThanOrEqOp FilterOperator = "<="
	GreaterThanOrEqOp FilterOperator = ">="
	ContainsOp        FilterOperator = "=@"
	NotContainsOp     FilterOperator = "!@"
	StartsWithOp      FilterOperator = "=^"
	EndsWithOp        FilterOperator = "=$"
)

type Filter struct {
	filterBy string
	operator FilterOperator
	value    string
}

func (f *Filter) Get() string {
	return fmt.Sprintf("%s%s%s", f.filterBy, f.operator, f.value)
}

type FilterBuilder struct {
	filters []string
}

func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{}
}

func (f *FilterBuilder) AddFilter(filterBy string, op FilterOperator, value string) *FilterBuilder {
	return f.AddAFilter(&Filter{filterBy, op, value})
}

func (f *FilterBuilder) AddFilters(filterBy string, op FilterOperator, values []string) *FilterBuilder {
	return f.AddAFilter(&Filter{filterBy, op, strings.Join(values, "||")})
}

func (f *FilterBuilder) AddAFilter(filter *Filter) *FilterBuilder {
	f.filters = append(f.filters, filter.Get())
	return f
}

func (f *FilterBuilder) Get() []string {
	return f.filters
}
