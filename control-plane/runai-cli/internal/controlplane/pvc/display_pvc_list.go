package pvc

import (
	"encoding/json"

	cliutils "github.com/run-ai/runai/control-plane/runai-cli/internal/utils"

	"github.com/charmbracelet/bubbles/table"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	uitable "github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	openapi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	"gopkg.in/yaml.v3"
)

const (
	DefaultWidth           = 10
	DefaultAssetRunAiWidth = 15

	TitleClaimName    = "Claim Name"
	TitleIsDataSource = "Run:ai Managed"

	ManagedYes = "Yes"
	ManagedNo  = "No"
	ManagedNa  = "-"
)

const (
	nameColumnIndex = 0
)

type Data struct {
	Pvcs      []Brief `json:"pvcs,omitempty" yaml:"pvcs"`
	NextToken *string `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type Output struct {
	data Data
}

func NewOutput(response *openapi.PVCListResponse) (*Output, error) {
	briefList, err := BriefListFromResponse(response)
	if err != nil {
		return nil, err
	}
	return &Output{
		data: Data{Pvcs: briefList},
	}, nil
}

func (o *Output) Subject() string {
	return consts.ClusterStr
}

func (o *Output) Columns() []table.Column {
	return []table.Column{
		{Title: TitleClaimName, Width: DefaultWidth},
		{Title: TitleIsDataSource, Width: DefaultAssetRunAiWidth},
	}
}

func (o *Output) Data() [][]string {
	var (
		data    [][]string
		managed string
	)

	for _, brief := range o.data.Pvcs {
		if brief.Managed == nil {
			// should not reach this point - more of sanity.
			managed = ManagedNa
		} else if *brief.Managed {
			managed = ManagedYes
		} else {
			managed = ManagedNo
		}

		var row []string
		row = append(row, brief.Name)
		row = append(row, managed)
		data = append(data, row)
	}

	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: nameColumnIndex, Order: cliutils.Ascending},
	})

	return data
}

func (o *Output) GenerateTable() (string, error) {
	return uitable.NewTableDisplay(o).Generate()
}

func (o *Output) GenerateJson() (string, error) {
	output, err := json.MarshalIndent(o.data, "", "  ")
	return string(output), err
}

func (o *Output) GenerateYaml() (string, error) {
	output, err := yaml.Marshal(o.data)
	return string(output), err
}

func (o *Output) GetList() []Brief {
	return o.data.Pvcs
}
