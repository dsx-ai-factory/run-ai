package manipulate

import (
	cliutils "github.com/run-ai/runai/control-plane/runai-cli/internal/utils"
	"k8s.io/utils/ptr"
)

func CreateAlignedWorkloadTypes(workloadTypesRaw Data, columns []WorkloadTypeColumn) WorkloadTypesAlignedOutput {
	workloadTypesAligned := make([]WorkloadTypeAlignedRow, 0)
	for _, wt := range workloadTypesRaw.WorkloadTypes {
		row := &WorkloadTypeAlignedRow{}
		for _, column := range columns {
			ColumnMap[column].extractFunc(ptr.To(wt), row)
		}
		workloadTypesAligned = append(workloadTypesAligned, *row)
	}
	return WorkloadTypesAlignedOutput{
		WorkloadTypes: workloadTypesAligned,
	}
}

func CreateTableData(workloadTypesAligned []WorkloadTypeAlignedRow, columns []WorkloadTypeColumn) [][]string {
	var data [][]string

	for _, wt := range workloadTypesAligned {
		var row []string
		for _, column := range columns {
			row = append(row, ColumnMap[column].formatFunc(ptr.To(wt)))
		}
		data = append(data, row)
	}

	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: nameColumnIndex, Order: cliutils.Ascending},
	})

	return data
}
