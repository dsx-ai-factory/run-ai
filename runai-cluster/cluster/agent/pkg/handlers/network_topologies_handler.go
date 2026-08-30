package handlers

import (
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/adapters"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"

	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	runaiBackendClient "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/metrics"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NetworkTopologiesHandler struct {
	*CommonHandler
	lastUpdated time.Time
}

type NetworkTopologyObjectInspector bool

func (p NetworkTopologyObjectInspector) DeepCopySpec(from, to client.Object) {
	fromNetworkTopology, ok := from.(*kaiv1alpha1.Topology)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return
	}
	toNetworkTopology, ok := to.(*kaiv1alpha1.Topology)
	if !ok {
		return
	}

	fromNetworkTopology.Spec.DeepCopyInto(&toNetworkTopology.Spec)
}

func (p NetworkTopologyObjectInspector) IsUpToDate(left, right client.Object) bool {
	return true
}

func NewNetworkTopologiesHandler(runaiBackendClient *runaiBackendClient.BackendClient, runaiClient client.Client,
	cachedClient client.Reader) (ResourceHandler, error) {
	commonHandler, err := NewCommonHandler(runaiBackendClient, runaiClient, cachedClient,
		&kaiv1alpha1.Topology{}, NetworkTopologyObjectInspector(true), HandlingOptions{})
	if err != nil {
		return nil, err
	}
	// This deletionPredicate ensures that Cluster Source topologies will not be deleted
	commonHandler.deletionPredicate = hasRunAiSource
	return &NetworkTopologiesHandler{
		CommonHandler: commonHandler,
		lastUpdated:   time.Time{},
	}, nil
}

func (h *NetworkTopologiesHandler) Handle() error {
	log.Debug().Msgf("Starting to sync network topologies")

	backendNetworkTopologies, err := h.getBackendNetworkTopologies()
	if err != nil {
		return err
	}

	clusterTopologies := adapters.NewNetworkTopologiesAdapter(backendNetworkTopologies).
		ToClusterTopologiesBySource()

	metrics.GetMetricsExporter().AddResourceAction(h.MetricLabel(), metrics.ActionCreateOrUpdate, len(clusterTopologies))

	return h.ReconcileObjects(clusterTopologies, false)
}

func (h *NetworkTopologiesHandler) MetricLabel() string {
	return "network_topologies"
}

func (h *NetworkTopologiesHandler) getBackendNetworkTopologies() ([]openapi.NetworkTopology, error) {
	log.Debug().Msgf("Getting network topologies from Cluster Service")
	response, err := h.backendClient.ClusterServiceClient.
		GetNetworkTopologies(h.backendClient.Context(), []string{fmt.Sprintf("clusterId==%s", config.Config().ClusterUuid)})

	if err != nil {
		log.Error().Msgf("Failed to get NetworkTopologies: err: %v", err)
		return []openapi.NetworkTopology{}, err
	}
	networkTopologies := response.GetNetworkTopologies()

	return networkTopologies, nil
}

func hasRunAiSource(object client.Object) bool {
	label, ok := object.GetLabels()[adapters.NetworkTopologySourceLabel]
	if !ok {
		return false
	}
	return label == adapters.NetworkTopologyRunAiSource
}
