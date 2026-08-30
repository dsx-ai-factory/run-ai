package describe

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/environment/client"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var environmentClient *client.Client

func Describe(txn *transaction.Txn, cpUrl, token, name, projectId, clusterId string) (*EnvironmentFormatter, error) {
	var err error
	var clusterIDPtr *string

	if environmentClient == nil {
		environmentClient, err = client.NewClient(cpUrl, token)
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

	env, err := FindByName(txn.GetCtx(), environmentClient, name, projectIDPtr, clusterIDPtr, "")
	if err != nil {
		return nil, err
	}

	return NewEnvironmentFormatter(env)
}
