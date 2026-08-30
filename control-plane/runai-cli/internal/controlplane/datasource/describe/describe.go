package describe

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/datasource/client"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var datasourceClient Lister

func Describe(txn *transaction.Txn, cpUrl, token, name, projectName, clusterId, dsType string) (*Datasource, error) {
	var err error

	if datasourceClient == nil {
		datasourceClient, err = client.NewClient(cpUrl, token)
		if err != nil {
			return nil, err
		}
	}

	// Get project ID from project name if provided
	var projectIDPtr *int32
	if projectName != "" {
		projectIdStr, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, projectName)
		if err != nil {
			return nil, err
		}

		projectIDPtr, err = projectservice.ParseProjectIdInt32(projectIdStr)
		if err != nil {
			return nil, err
		}
	}

	// Set type filter if provided
	var dsTypePtr *string
	if dsType != "" {
		dsTypePtr = &dsType
	}

	return ListByName(txn.GetCtx(), datasourceClient, clusterId, name, projectIDPtr, dsTypePtr)
}
