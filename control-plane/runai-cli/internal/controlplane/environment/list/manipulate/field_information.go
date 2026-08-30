package manipulate

import (
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/field"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

var ColumnMap = map[EnvironmentColumn]FieldInformation{
	ColumnKeyName:          NameField,
	ColumnKeyScope:         ScopeField,
	ColumnKeyImage:         ImageField,
	ColumnKeyTools:         ToolsField,
	ColumnKeyWorkloadTypes: WorkloadTypesField,
}

var defaultColumns = []EnvironmentColumn{
	ColumnKeyName,
	ColumnKeyScope,
	ColumnKeyImage,
	ColumnKeyTools,
	ColumnKeyWorkloadTypes,
}

func GetDefaultColumns() []EnvironmentColumn {
	return defaultColumns
}

func GetColumns(columns []EnvironmentColumn) []EnvironmentColumn {
	if len(columns) > 0 {
		return columns
	}
	return GetDefaultColumns()
}

type FieldInformation struct {
	columnTitle string
	extractFunc func(env *openapi.EnvironmentAsset, envRow *EnvironmentsAlignedRow)
	formatFunc  func(envRow *EnvironmentsAlignedRow) string
	columnWidth int
}

func (f FieldInformation) GetColumnWidth() int {
	return f.columnWidth
}

func (f FieldInformation) GetColumnTitle() string {
	return f.columnTitle
}

// Field definitions for environment columns
var (
	NameField = FieldInformation{
		columnTitle: "Name",
		columnWidth: DefaultWidth,
		extractFunc: extractNameData,
		formatFunc:  formatNameValue,
	}
	ScopeField = FieldInformation{
		columnTitle: "Scope",
		columnWidth: DefaultWidth,
		extractFunc: extractScopeData,
		formatFunc:  formatScopeValue,
	}
	ImageField = FieldInformation{
		columnTitle: "Image",
		columnWidth: DefaultWidthImage,
		extractFunc: extractImageData,
		formatFunc:  formatImageValue,
	}
	ToolsField = FieldInformation{
		columnTitle: "Tools",
		columnWidth: DefaultWidthTools,
		extractFunc: extractToolsData,
		formatFunc:  formatToolsValue,
	}
	WorkloadTypesField = FieldInformation{
		columnTitle: "Workload Types",
		columnWidth: DefaultWidthWorkloadTypes,
		extractFunc: extractWorkloadTypesData,
		formatFunc:  formatWorkloadTypesValue,
	}
)

func extractNameData(env *openapi.EnvironmentAsset, envRow *EnvironmentsAlignedRow) {
	envRow.Name = env.Meta.Name
}

func extractScopeData(env *openapi.EnvironmentAsset, envRow *EnvironmentsAlignedRow) {
	scope := string(env.Meta.Scope)
	if scope == "" {
		envRow.Scope = "Unknown"
	} else {
		envRow.Scope = scope
	}
}

func extractImageData(env *openapi.EnvironmentAsset, envRow *EnvironmentsAlignedRow) {
	if env.Spec.HasImage() {
		envRow.Image = env.Spec.GetImage()
	} else {
		envRow.Image = noDataFormat
	}
}

func extractToolsData(env *openapi.EnvironmentAsset, envRow *EnvironmentsAlignedRow) {
	envRow.RawConnections = env.Spec.GetConnections()
}

func extractWorkloadTypesData(env *openapi.EnvironmentAsset, envRow *EnvironmentsAlignedRow) {
	envRow.RawWorkloadTypes = env.Meta.WorkloadSupportedTypes
}

func formatNameValue(envRow *EnvironmentsAlignedRow) string {
	return envRow.Name
}

func formatScopeValue(envRow *EnvironmentsAlignedRow) string {
	return envRow.Scope
}

func formatImageValue(envRow *EnvironmentsAlignedRow) string {
	return envRow.Image
}

func formatToolsValue(envRow *EnvironmentsAlignedRow) string {
	return formatTools(envRow.RawConnections)
}

func formatWorkloadTypesValue(envRow *EnvironmentsAlignedRow) string {
	return formatWorkloadTypes(envRow.RawWorkloadTypes)
}

func formatTools(connections []openapi.Connection) string {
	if len(connections) == 0 {
		return noDataFormat
	}

	var tools []string
	for _, conn := range connections {
		if conn.HasInternalToolInfo() {
			toolInfo := conn.GetInternalToolInfo()
			toolType := string(toolInfo.ToolType)
			switch toolType {
			case "JupyterNotebook":
				tools = append(tools, "jupyter-notebook")
			case "Tensorboard":
				tools = append(tools, "tensorboard")
			case "VisualStudioCode":
				tools = append(tools, "visual-studio-code")
			case "MLFlow":
				tools = append(tools, "mlflow")
			case "Wandb":
				tools = append(tools, "wandb")
			case "PyCharm":
				tools = append(tools, "pycharm")
			default:
				tools = append(tools, strings.ToLower(toolType))
			}
		} else if conn.HasExternalToolInfo() {
			toolInfo := conn.GetExternalToolInfo()
			toolType := string(toolInfo.ToolType)
			tools = append(tools, strings.ToLower(toolType))
		}
	}

	if len(tools) == 0 {
		return noDataFormat
	}
	return strings.Join(tools, ", ")
}

func formatWorkloadTypes(supportedTypes openapi.NullableWorkloadSupportedTypes) string {
	if !supportedTypes.IsSet() || supportedTypes.Get() == nil {
		return noDataFormat
	}

	types := supportedTypes.Get()
	var workloadTypes []string

	if field.BoolIsTrue(types.Workspace.Get()) {
		workloadTypes = append(workloadTypes, "Workspace")
	}
	if field.BoolIsTrue(types.Training.Get()) {
		workloadTypes = append(workloadTypes, "Training")
	}
	if field.BoolIsTrue(types.Distributed.Get()) {
		workloadTypes = append(workloadTypes, "Distributed")
	}
	if field.BoolIsTrue(types.Inference.Get()) {
		workloadTypes = append(workloadTypes, "Inference")
	}

	if len(workloadTypes) == 0 {
		return noDataFormat
	}
	return strings.Join(workloadTypes, ", ")
}
