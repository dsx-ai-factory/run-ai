package manipulate

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/mycredential/client"
)

func CreateAlignedMyCredentials(entries []client.MyCredentialEntry, columns []MyCredentialColumn) MyCredentialAlignedData {
	var aligned []MyCredentialAlignedRow

	for _, entry := range entries {
		row := MyCredentialAlignedRow{}
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			fieldInfo.extractFunc(&entry, &row)
		}
		aligned = append(aligned, row)
	}

	return MyCredentialAlignedData{Credentials: aligned}
}

func CreateTableData(credentials []MyCredentialAlignedRow, columns []MyCredentialColumn) [][]string {
	var data [][]string

	for _, credRow := range credentials {
		var row []string
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			row = append(row, fieldInfo.formatFunc(&credRow))
		}
		data = append(data, row)
	}

	return data
}
