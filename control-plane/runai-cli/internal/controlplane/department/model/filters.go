package model

import "fmt"

type DepartmentFilter string
type FilterOperator string

// Filters keys --> (name|clusterId|departmentId|createdAt)
// Operators --> (==|!=|<=|>=|=@|!@|=\\^|=\\$)
const (
	ClusterID DepartmentFilter = "clusterId"

	EqualOp FilterOperator = "=="
)

type Filter struct {
	filter   DepartmentFilter
	value    string
	operator FilterOperator
}

func NewClusterIDFilter(clusterID string) *Filter {
	return &Filter{ClusterID, clusterID, EqualOp}
}

func NewDepartmentFilter(departmentName string) *Filter {
	return &Filter{"name", departmentName, EqualOp}
}

func (f *Filter) Get() string {
	return fmt.Sprintf("%s%s%s", f.filter, f.operator, f.value)
}

type FilterCollection struct {
	filters []*Filter
}

func NewFilterCollection() *FilterCollection {
	return &FilterCollection{
		filters: []*Filter{},
	}
}

func (f *FilterCollection) AddFilter(filter *Filter) *FilterCollection {
	f.filters = append(f.filters, filter)
	return f
}

func (f *FilterCollection) Get() []string {
	var filters []string
	for _, filter := range f.filters {
		filters = append(filters, filter.Get())
	}
	return filters
}
