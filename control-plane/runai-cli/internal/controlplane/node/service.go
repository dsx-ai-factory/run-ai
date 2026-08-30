package node

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node/list"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/clients"
)

var (
	lister Client
)
var RequiredNodeTelemetryTypes = []openapi.NodeTelemetryType{
	openapi.NODETELEMETRYTYPE_FREE_GPUS,
}

type Client interface {
	GetNodeList(ctx context.Context, clusterUuid string) (*openapi.Nodes, error)
	GetNodesTelemetry(ctx context.Context, clusterUuid string, telemetryTypes []openapi.NodeTelemetryType) (api.NodeTelemetriesByType, error)
}

type BackendNodeService struct {
	nodeClient Client
}

func CreateBackendNodeService(urlStr, token string) (*BackendNodeService, error) {
	nodeClient, err := clients.NewClusterService(urlStr, token)
	if err != nil {
		return nil, err
	}
	return NewBackendNodeService(nodeClient), err
}

func NewBackendNodeService(nodeLister Client) *BackendNodeService {
	return &BackendNodeService{nodeClient: nodeLister}
}

func (bns *BackendNodeService) List(ctx context.Context, clusterUuid string) (*list.Output, error) {
	// Get raw node list
	nodeList, err := bns.nodeClient.GetNodeList(ctx, clusterUuid)
	if err != nil {
		return nil, err
	}

	// Get telemetry data
	nodesTelemetry, err := bns.nodeClient.GetNodesTelemetry(ctx, clusterUuid, RequiredNodeTelemetryTypes)
	if err != nil {
		return nil, err
	}

	// Convert to slice for manipulate layer
	nodes := []openapi.Node{}
	if nodeList != nil {
		nodes = nodeList.GetNodes()
	}

	return list.NewOutput(nodes, nodesTelemetry)
}

func GetNodeLister(domain, token string) (*BackendNodeService, error) {
	var err error
	if lister == nil {
		lister, err = clients.NewClusterService(domain, token)
		if err != nil {
			return nil, err
		}
		return NewBackendNodeService(lister), nil
	}
	return NewBackendNodeService(lister), nil
}

func GetNodeList(ctx context.Context, domain, token, clusterUuid string) (*list.Output, error) {
	nodeService, err := GetNodeLister(domain, token)
	if err != nil {
		return nil, err
	}
	return nodeService.List(ctx, clusterUuid)
}

func (bns *BackendNodeService) CheckIfNodesExist(ctx context.Context, clusterUuid string, nodesToCheck []string) (existing []string, notExisting []string, err error) {
	if len(nodesToCheck) == 0 {
		return existing, notExisting, nil
	}

	// get list of existing nodes
	actualNodesList, err := bns.nodeClient.GetNodeList(ctx, clusterUuid)
	if err != nil || actualNodesList == nil {
		return existing, notExisting, err
	}

	// turn existing node list to map for efficiency
	actualNodesMap := make(map[string]bool)
	for _, actual := range actualNodesList.GetNodes() {
		actualNodesMap[actual.Name] = true
	}

	// validate node exists and remove duplicates
	// also track non-existing names
	existingMap := make(map[string]bool)
	nonExistingMap := make(map[string]bool)
	for _, nodeName := range nodesToCheck {
		// if already added, continue
		_, ok1 := existingMap[nodeName]
		_, ok2 := nonExistingMap[nodeName]

		if ok1 || ok2 {
			continue
		}

		// if not added and exists, add it to the final list
		if _, ok := actualNodesMap[nodeName]; ok {
			existingMap[nodeName] = true
			existing = append(existing, nodeName)
		} else {
			nonExistingMap[nodeName] = true
			notExisting = append(notExisting, nodeName)
		}
	}

	return existing, notExisting, nil
}
