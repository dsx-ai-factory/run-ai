package workloadtype

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workloadtype/client"
	workloadtypelist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workloadtype/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

func List(txn *transaction.Txn, cpUrl, token string, fb *filter.Builder, externalTypesOnly *bool) (*workloadtypelist.Output, error) {
	listClient, err := client.NewClient(cpUrl, token)
	if err != nil {
		return nil, err
	}
	service := workloadtypelist.NewService(listClient)
	return service.List(txn.GetCtx(), cpUrl, token, fb, externalTypesOnly)
}
