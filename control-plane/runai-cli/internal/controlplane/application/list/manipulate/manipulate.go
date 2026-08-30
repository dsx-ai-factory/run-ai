package manipulate

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
)

func GetColumns(columns []AIApplicationColumn) []AIApplicationColumn {
	if len(columns) != 0 {
		return columns
	}
	return defaultColumns
}

func ParseAIApplicationColumns(columns []string) ([]AIApplicationColumn, error) {
	if len(columns) == 0 {
		return nil, nil
	}

	result := make([]AIApplicationColumn, 0, len(columns))
	var invalidColumns []string

	for _, columnName := range columns {
		trimmedKey := AIApplicationColumn(strings.ToLower(strings.TrimSpace(columnName)))
		if _, exists := ColumnMap[trimmedKey]; exists {
			result = append(result, trimmedKey)
		} else {
			invalidColumns = append(invalidColumns, fmt.Sprintf("%q", columnName))
		}
	}

	if len(invalidColumns) > 0 {
		return nil, fmt.Errorf(consts.ErrorInvalidColumns, strings.Join(invalidColumns, ", "))
	}

	return result, nil
}

func CreateAlignedApplications(data Data, columns []AIApplicationColumn) AIApplicationAlignedOutput {
	rows := make([]AIApplicationAlignedRow, 0, len(data.AIApplications))
	for i := range data.AIApplications {
		app := &data.AIApplications[i]
		var row AIApplicationAlignedRow
		for _, col := range columns {
			if fi, ok := ColumnMap[col]; ok {
				fi.extractFunc(app, &row)
			}
		}
		rows = append(rows, row)
	}
	return AIApplicationAlignedOutput{
		AIApplications: rows,
		NextToken:      data.NextToken,
	}
}

func CreateTableData(rows []AIApplicationAlignedRow, columns []AIApplicationColumn) [][]string {
	result := make([][]string, 0, len(rows))
	for i := range rows {
		row := make([]string, 0, len(columns))
		for _, col := range columns {
			if fi, ok := ColumnMap[col]; ok {
				row = append(row, fi.formatFunc(&rows[i]))
			}
		}
		result = append(result, row)
	}
	return result
}
