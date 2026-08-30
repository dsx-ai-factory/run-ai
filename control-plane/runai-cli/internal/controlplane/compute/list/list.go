package list

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/compute/client"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var listClient Lister

func SetClient(lc Lister) {
	listClient = lc
}

func List(txn *transaction.Txn, cpUrl, token, projectName, clusterId string) (*Output, error) {
	var err error
	var projectID *int32
	var clusterIDPtr *string

	if listClient == nil {
		listClient, err = client.NewClient(cpUrl, token)
		if err != nil {
			return nil, err
		}
	}

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

	if clusterId != "" {
		clusterIDPtr = &clusterId
	}

	computeService := NewService(listClient)
	return computeService.List(txn.GetCtx(), projectID, clusterIDPtr)
}
