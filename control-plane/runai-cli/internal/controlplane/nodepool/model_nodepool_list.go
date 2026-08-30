package nodepool

import (
	"fmt"
	"strconv"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"
	nodes "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node/list/manipulate"
	"k8s.io/utils/ptr"
)

type BriefNodepool struct {
	Name                string   `json:"name"`
	Status              string   `json:"status"`
	GpuAllocation       string   `json:"gpuAllocation"`
	Nodes               []string `json:"nodes"`
	MNNVLDetected       bool     `json:"MNNVLDetected"`
	NetworkTopologyName string   `json:"networkTopologyName"`
}

type briefNodepoolAccumulator struct {
	Brief                 *BriefNodepool
	TotalGpus             int32
	FreeGpus              int
	TelemetryNotAvailable bool
}

func BriefListFromResponse(nodepools *openapi.GetNodepools200Response, nodes *openapi.Nodes, telemetry api.NodeTelemetriesByType) ([]*BriefNodepool, error) {
	err := validateAttributes(nodepools, nodes, telemetry)
	if err != nil {
		return nil, err
	}

	return createBriefs(nodepools, nodes, telemetry)
}

func validateAttributes(nodepools *openapi.GetNodepools200Response, nodes *openapi.Nodes, telemetry api.NodeTelemetriesByType) error {
	if nodes == nil {
		return fmt.Errorf("node list attribute is missing")
	}
	if telemetry == nil {
		return fmt.Errorf("telemetry list attribute is missing")
	}
	if nodepools == nil {
		return fmt.Errorf("nodepool list attribute is missing")
	}
	return nil
}

func createBriefs(nodepools *openapi.GetNodepools200Response, nodeList *openapi.Nodes, telemetry api.NodeTelemetriesByType) ([]*BriefNodepool, error) {
	var briefList []*BriefNodepool
	for _, nodepool := range nodepools.Nodepools {
		brief := &BriefNodepool{Name: nodepool.Name, Status: string(ptr.Deref(nodepool.Phase, "")), NetworkTopologyName: nodepool.GetNetworkTopologyName()}
		if nodepool.Status != nil {
			brief.MNNVLDetected = ptr.Deref(nodepool.Status.GpuNetworkAccelerationDetected, false)
		}
		briefList = append(briefList, brief)
	}

	nodepoolAccumulatorMap, err := accumulateNodepoolsFromNodes(briefList, nodeList, telemetry)
	if err != nil {
		return nil, err
	}

	for _, accumulator := range nodepoolAccumulatorMap {
		if accumulator.TelemetryNotAvailable {
			accumulator.Brief.GpuAllocation = nodes.NotAvailableString
		} else {
			accumulator.Brief.GpuAllocation, err = nodes.ConstructGpuAllocation(accumulator.TotalGpus, strconv.Itoa(accumulator.FreeGpus))
			if err != nil {
				return nil, err
			}
		}
	}

	return briefList, nil
}

func accumulateNodepoolsFromNodes(briefList []*BriefNodepool, nodeList *openapi.Nodes, telemetry api.NodeTelemetriesByType) (map[string]*briefNodepoolAccumulator, error) {
	nodepoolAccumulatorMap := make(map[string]*briefNodepoolAccumulator)
	for _, brief := range briefList {
		nodepoolAccumulatorMap[brief.Name] = &briefNodepoolAccumulator{Brief: brief}
	}

	freeGpuByNodeTelemetry, err := nodes.ParseTelemetryResponse(telemetry[openapi.NODETELEMETRYTYPE_FREE_GPUS])
	if err != nil {
		return nil, err
	}

	for _, node := range nodeList.GetNodes() {
		accumulator, okMap := nodepoolAccumulatorMap[node.NodePool]
		if !okMap {
			return nil, fmt.Errorf("node %s has an unknown nodepool %s", node.Name, node.NodePool)
		}

		accumulator.Brief.Nodes = append(accumulator.Brief.Nodes, node.Name)

		if node.HasGpuInfo() {
			info := node.GetGpuInfo()
			accumulator.TotalGpus += info.GetGpuCount()

			if freeGpus, ok := freeGpuByNodeTelemetry[node.Name]; ok {
				numericFreeGpus, errAtoi := strconv.Atoi(freeGpus)
				if errAtoi != nil {
					return nil, fmt.Errorf("failed to calculate GPU info for node %s. Got error: %s", node.Name, errAtoi.Error())
				}
				accumulator.FreeGpus += numericFreeGpus
			} else {
				accumulator.TelemetryNotAvailable = true
			}
		}
	}

	return nodepoolAccumulatorMap, nil
}
