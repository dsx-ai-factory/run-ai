package list

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/datasource/client"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var listClient Lister

func SetClient(lc Lister) {
	listClient = lc
}

func List(txn *transaction.Txn, cpUrl, token, projectName, clusterId, dsType string) (*Output, error) {
	var err error
	var projectID *int32
	var clusterIDPtr *string
	var dsTypePtr *string

	if listClient == nil {
		listClient, err = client.NewClient(cpUrl, token)
		if err != nil {
			return nil, err
		}
	}

	// Always set cluster ID if available
	if clusterId != "" {
		clusterIDPtr = &clusterId
	}

	// Set type filter if provided
	if dsType != "" {
		dsTypePtr = &dsType
	}

	// Set project ID if project is specified (not using -A flag)
	if projectName != "" {
		projectIdStr, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, projectName)
		if err != nil {
			return nil, err
		}
		projectID, err = projectservice.ParseProjectIdInt32(projectIdStr)
		if err != nil {
			return nil, err
		}
	}

	datasourceService := NewService(listClient)
	output, err := datasourceService.List(txn.GetCtx(), projectID, clusterIDPtr, dsTypePtr)
	if err != nil {
		return nil, err
	}

	return output, nil
}
