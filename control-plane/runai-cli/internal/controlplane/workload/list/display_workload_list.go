package list

import (
	"encoding/json"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/list/manipulate"

	"github.com/charmbracelet/bubbles/table"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/structs"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	uitable "github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	workloadservice "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"gopkg.in/yaml.v3"
)

// SetRaw sets the raw output preference
func (o *Output) SetRaw(raw bool) {
	o.raw = raw
}

// SetColumns sets the columns to display
func (o *Output) SetColumns(columns []string) error {
	if len(columns) == 0 {
		columns = settings.GetInstance().GetWorkloadColumns()
	}
	workloadColumns, err := manipulate.ParseWorkloadColumns(columns)
	if err != nil {
		return err
	}

	o.columns = workloadColumns
	return nil

}

type briefData struct {
	Workloads []manipulate.Brief `json:"workloads" yaml:"workloads"`
	NextToken *string            `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type Output struct {
	data    manipulate.Data
	deleted bool
	raw     bool
	columns []manipulate.WorkloadColumn
}

func NewOutput(workloadList []workloadservice.Workload, next *structs.NextToken, deleted bool) (*Output, error) {
	var nextToken *string

	if next != nil {
		encoded, errEnc := next.Encode()
		if errEnc != nil {
			return nil, errEnc
		}
		nextToken = &encoded
	}

	return &Output{
		data: manipulate.Data{
			Workloads: workloadList, // Store originals directly
			NextToken: nextToken,
		},
		deleted: deleted,
	}, nil
}

func (o *Output) Subject() string {
	return consts.ClusterStr
}

func (o *Output) Columns() []table.Column {
	// will send the column data flag after this pr , instead of empty list
	columns := manipulate.GetColumns(o.columns, o.deleted)
	var columnsTable []table.Column
	for _, column := range columns {
		columnsTable = append(columnsTable, table.Column{Title: manipulate.ColumnMap[column].GetColumnTitle(), Width: manipulate.ColumnMap[column].GetColumnWidth()})
	}
	return columnsTable
}

// getBriefWorkloads converts original workloads to brief format on-demand
func (o *Output) getBriefWorkloads() []manipulate.Brief {
	if len(o.data.Workloads) == 0 {
		return []manipulate.Brief{}
	}
	briefList, _ := manipulate.BriefListFromResponse(o.data.Workloads, o.deleted)
	return briefList
}

func (o *Output) Data() [][]string {
	columns := manipulate.GetColumns(o.columns, o.deleted)
	workloadsAligned := manipulate.CreateAlignedWorkloads(o.data, columns).Workloads
	return manipulate.CreateTableData(workloadsAligned, columns)
}

func (o *Output) GenerateTable() (string, error) {
	display, err := uitable.NewTableDisplay(o).Generate()
	if o.data.NextToken != nil {
		display += fmt.Sprintf("\r\n\n next token: %s", *o.data.NextToken)
	}
	return display, err
}

// getDataStructure returns the appropriate data structure based on raw and format alignment settings
func (o *Output) getDataStructure() interface{} {
	if o.raw {
		return o.data
	}

	formatAlign := settings.GetInstance().GetListFormatAlign()
	if !formatAlign {
		// Use original brief format when format alignment is disabled
		return briefData{
			Workloads: o.getBriefWorkloads(),
			NextToken: o.data.NextToken,
		}
	}

	// Use aligned format with column titles when format alignment is enabled
	columns := manipulate.GetColumns(o.columns, o.deleted)
	return manipulate.CreateAlignedWorkloads(o.data, columns)
}

func (o *Output) GenerateJson() (string, error) {
	data := o.getDataStructure()
	output, err := json.MarshalIndent(data, "", "  ")
	return string(output), err
}

func (o *Output) GenerateYaml() (string, error) {
	data := o.getDataStructure()
	output, err := yaml.Marshal(data)
	return string(output), err
}

func (o *Output) GetList() []manipulate.Brief {
	return o.getBriefWorkloads()
}
func (o *Output) GetRawList() []workloadservice.Workload {
	return o.data.Workloads
}

func (o *Output) IsEmpty() bool {
	return len(o.data.Workloads) == 0
}
