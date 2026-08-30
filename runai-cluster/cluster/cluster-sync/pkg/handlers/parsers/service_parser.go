package parsers

import (
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/generated/swagger-runai-backend-client"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func ParseServiceToServerScheme(service *knv1.Service) swagger.Service {
	// URL is only created after routes are create
	var host = ""
	if service.Status.URL != nil {
		host = service.Status.URL.Host
	}
	return swagger.Service{
		ClusterUuid: config.Config().ClusterUuid,
		Url:         host,
		Id:          string(service.UID),
		Name:        service.Name,
		Namespace:   service.Namespace,
		CreatedAt:   service.CreationTimestamp.Time,
	}
}
