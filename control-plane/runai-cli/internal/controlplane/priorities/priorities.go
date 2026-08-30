package priorities

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var wpClient Lister

func SetClient(wpc Lister) {
	wpClient = wpc
}

func List(txn *transaction.Txn, cpUrl, token, clusterId string) (*Output, error) {
	var err error

	if wpClient == nil {
		wpClient, err = NewWorkloadPrioritiesServiceClient(cpUrl, token)
		if err != nil {
			return nil, err
		}
	}

	prioritiesService := NewPrioritiesService(wpClient)
	return prioritiesService.GetWorkloadClusterPrioritiesBrief(txn.GetCtx(), clusterId)
}
