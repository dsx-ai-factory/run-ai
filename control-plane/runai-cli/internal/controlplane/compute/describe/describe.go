package describe

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/compute/client"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var computeClient *client.Client

func Describe(txn *transaction.Txn, cpUrl, token, name, projectId, clusterId string) (*ComputeFormatter, error) {
	var err error
	var clusterIDPtr *string

	if computeClient == nil {
		computeClient, err = client.NewClient(cpUrl, token)
		if err != nil {
			return nil, err
		}
	}

	projectIDPtr, err := projectservice.ParseProjectIdInt32(projectId)
	if err != nil {
		return nil, err
	}

	if clusterId != "" {
		clusterIDPtr = &clusterId
	}

	compute, err := FindByName(txn.GetCtx(), computeClient, name, projectIDPtr, clusterIDPtr)
	if err != nil {
		return nil, err
	}

	return NewComputeFormatter(compute)
}
