package distributedinference

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	apilist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributedinference/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributedinference/service"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var wsClient workload.WithUpdateAction

func SetClient(wsClientSet workload.WithUpdateAction) {
	wsClient = wsClientSet
}

func GetByNameAndProj(txn *transaction.Txn, domain, token, name, project, cluster string, fb *filter.Builder) (*list.Output, error) {
	fb.AppendName(name).WithProject(project).WithClusterID(cluster)
	return GetByFilter(txn, domain, token, fb)
}

func GetByFilter(txn *transaction.Txn, domain, token string, fb *filter.Builder) (*list.Output, error) {
	fb.WithWorkloadType(api.DistributedInference.QueryParam())
	return workload.GetWorkloadByFilter(txn, domain, token, fb)
}

func List(txn *transaction.Txn, domain, token string, pagingOptions *apilist.PageOptions, fb *filter.Builder, deleted bool) (*list.Output, error) {
	fb.WithWorkloadType(api.DistributedInference.QueryParam())
	return workload.GetWorkloadPagination(txn, domain, token, pagingOptions, fb, deleted)
}

func Delete(txn *transaction.Txn, domain string, token string, uuid string) error {
	distInferenceClient, err := getDistributedInferenceClient(domain, token)
	if err != nil {
		return err
	}
	return distInferenceClient.Delete(txn.GetCtx(), uuid)
}

func Submit(txn *transaction.Txn, domain, token, name, projectId, clusterId string, generalSpec *spec.GeneralSpec) (string, error) {
	distInferenceClient, err := getDistributedInferenceClient(domain, token)
	if err != nil {
		return "", err
	}
	return distInferenceClient.Submit(txn.GetCtx(), name, projectId, clusterId, generalSpec)
}

func Update(txn *transaction.Txn, domain, token, uuid, clusterId string, spec *spec.GeneralSpec) (string, error) {
	distInferenceService, err := getDistributedInferenceService(domain, token)
	if err != nil {
		return "", err
	}
	return distInferenceService.Update(txn.GetCtx(), uuid, clusterId, spec)
}

func getDistributedInferenceClient(domain string, token string) (workload.WithUpdateAction, error) {
	if wsClient != nil {
		return wsClient, nil
	}
	return client.NewDistributedInferenceClient(domain, token)
}

func getDistributedInferenceService(domain string, token string) (*service.BackendDistributedInferenceService, error) {
	var err error
	if wsClient == nil {
		wsClient, err = client.NewDistributedInferenceClient(domain, token)
		if err != nil {
			return nil, err
		}
		return service.NewBackendDistributedInferenceService(wsClient), nil
	}
	return service.NewBackendDistributedInferenceService(wsClient), nil
}
