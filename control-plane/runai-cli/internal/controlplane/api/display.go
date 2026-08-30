package api

import "github.com/charmbracelet/bubbles/table"

const (
	DefaultWidth = 20
)

type MultiWorkloadsErrorTable struct {
	Workloads []MinimalWorkloadInfo
}

func NewMultiWorkloadsErrorTable(error MultiWorkloadsError) *MultiWorkloadsErrorTable {
	return &MultiWorkloadsErrorTable{Workloads: error.Workloads}
}

func (mwd *MultiWorkloadsErrorTable) Subject() string {
	return "Multiple Workloads"
}

func (mwd *MultiWorkloadsErrorTable) Columns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: DefaultWidth},
		{Title: "Project", Width: DefaultWidth},
		{Title: "Type", Width: DefaultWidth},
		{Title: "ID", Width: DefaultWidth},
	}
}

func (mwd *MultiWorkloadsErrorTable) Data() [][]string {
	var data [][]string
	for _, workload := range mwd.Workloads {
		data = append(data, []string{workload.Name, workload.Project, workload.Kind, workload.ID})
	}
	return data
}
