package list

import (
	"encoding/json"

	"github.com/charmbracelet/bubbles/table"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node/list/manipulate"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"
	uitable "github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	"gopkg.in/yaml.v3"
)

type BriefData struct {
	Nodes     []manipulate.BriefNode `json:"nodes,omitempty" yaml:"nodes"`
	NextToken *string                `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type Output struct {
	data            manipulate.Data
	raw             bool
	columns         []manipulate.NodeColumn
	nodeTelemetries api.NodeTelemetriesByType
}

func NewOutput(nodeList []openapi.Node, nodeTelemetries api.NodeTelemetriesByType) (*Output, error) {
	return &Output{
		data: manipulate.Data{
			Nodes:     nodeList,
			NextToken: nil,
		},
		nodeTelemetries: nodeTelemetries,
	}, nil
}

func (o *Output) Subject() string {
	return consts.NodeStr
}

func (o *Output) Columns() []table.Column {
	columns := manipulate.GetColumns(o.columns)
	var columnsTable []table.Column
	for _, column := range columns {
		columnsTable = append(columnsTable, table.Column{Title: manipulate.ColumnMap[column].GetColumnTitle(), Width: manipulate.ColumnMap[column].GetColumnWidth()})
	}
	return columnsTable
}

func (o *Output) Data() [][]string {
	columns := manipulate.GetColumns(o.columns)
	// Process raw nodes with telemetry enrichment
	nodesAligned := manipulate.CreateAlignedNodesWithTelemetry(o.data, o.nodeTelemetries, columns).Nodes
	return manipulate.CreateTableData(nodesAligned, columns)
}

func (o *Output) GenerateTable() (string, error) {
	return uitable.NewTableDisplay(o).Generate()
}

func (o *Output) getDataStructure() interface{} {
	if o.raw {
		return o.data
	}

	formatAlign := settings.GetInstance().GetListFormatAlign()
	if !formatAlign {
		return BriefData{
			Nodes:     o.GetList(),
			NextToken: o.data.NextToken,
		}
	}

	columns := manipulate.GetColumns(o.columns)
	return manipulate.CreateAlignedNodesWithTelemetry(o.data, o.nodeTelemetries, columns)
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

func (o *Output) SetRaw(raw bool) {
	o.raw = raw
}

func (o *Output) SetColumns(columns []string) error {
	if len(columns) == 0 {
		columns = settings.GetInstance().GetNodeColumns()
	}
	nodeColumns, err := manipulate.ParseNodeColumns(columns)
	if err != nil {
		return err
	}
	o.columns = nodeColumns
	return nil
}

func (o *Output) GetList() []manipulate.BriefNode {
	return manipulate.BriefListFromResponseWithTelemetry(o.data.Nodes, o.nodeTelemetries)
}
