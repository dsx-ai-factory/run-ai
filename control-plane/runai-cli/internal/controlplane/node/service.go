/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package node

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/node/list"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/structs"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/clients"
)

var (
	lister Client
)
var RequiredNodeTelemetryTypes = []openapi.NodeTelemetryType{
	openapi.NODETELEMETRYTYPE_FREE_GPUS,
}

type Client interface {
	GetNodeList(ctx context.Context, clusterUuid string, offset, limit int32) (*openapi.Nodes, error)
	GetNodesTelemetry(ctx context.Context, clusterUuid string, telemetryTypes []openapi.NodeTelemetryType) (api.NodeTelemetriesByType, error)
}

const (
	defaultNodePageSize int32 = 100
	maxNodePageSize     int32 = 500 // the list endpoint's maximum accepted limit
	maxNodePages              = 500 // bounds the loop if the server ignores offset
)

// FetchAllNodes loops over Client.GetNodeList in maxNodePageSize-sized
// chunks until the complete, unpaginated node list for clusterUuid has been
// retrieved. Shared by every caller that needs the full list (List,
// CheckIfNodesExist, and nodepool's node lookup), since Client only exposes
// a single paginated primitive (matching workload/application/project's
// single-List-method pattern).
func FetchAllNodes(ctx context.Context, client Client, clusterUuid string) ([]openapi.Node, error) {
	var nodes []openapi.Node
	offset := int32(0)

	for page := 0; page < maxNodePages; page++ {
		result, err := client.GetNodeList(ctx, clusterUuid, offset, maxNodePageSize)
		if err != nil {
			return nil, err
		}

		items := result.GetNodes()
		nodes = append(nodes, items...)

		if int32(len(items)) < maxNodePageSize {
			return nodes, nil
		}
		offset += maxNodePageSize
	}

	return nil, fmt.Errorf("stopped after reading %d pages: the server keeps returning full pages", maxNodePages)
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

// List fetches nodes according to opts, auto-paging internally until either
// the data is exhausted (the default, when opts is the zero value), opts.MaxItems
// is reached, or opts.NoPagination stops it after a single page. When the
// result was truncated (MaxItems or NoPagination), the returned Output
// carries a --next-token the caller can pass back in opts.NextToken to
// resume from where this call left off.
func (bns *BackendNodeService) List(ctx context.Context, clusterUuid string, opts apilist.PageOptions) (*list.Output, error) {
	nodes, nextToken, err := bns.fetchNodesPaged(ctx, clusterUuid, opts)
	if err != nil {
		return nil, err
	}

	nodesTelemetry, err := bns.nodeClient.GetNodesTelemetry(ctx, clusterUuid, RequiredNodeTelemetryTypes)
	if err != nil {
		return nil, err
	}

	return list.NewOutput(nodes, nodesTelemetry, nextToken)
}

// fetchNodesPaged pages through GetNodeList in pageSize-sized chunks,
// over-fetching one extra item per request to detect whether more data
// exists beyond what's being returned (the same trick workload/application
// list use, adapted for nodes since the Nodes API has no pagination
// metadata in its response). The Nodes API only supports a plain numeric
// offset (no createdAt cursor, see nodeColumnMapping() in cluster-service),
// so opts.NextToken/opts.Offset are used purely as that starting offset
// (structs.OffsetToNextToken/NextTokenToOffset), not as a real timestamp
// cursor.
func (bns *BackendNodeService) fetchNodesPaged(ctx context.Context, clusterUuid string, opts apilist.PageOptions) ([]openapi.Node, *structs.NextToken, error) {
	pageSize := clampPageSize(opts.PageSize)

	offset, err := resolveStartOffset(&opts)
	if err != nil {
		return nil, nil, err
	}

	var nodes []openapi.Node

	// Bounded like FetchAllNodes, as a safety net against a server that
	// ignores offset and keeps returning full pages forever. FetchAllNodes's
	// effective budget is maxNodePages pages of maxNodePageSize items each;
	// derive an equivalent total-node budget here so a smaller --page-size
	// doesn't shrink how many real nodes we're willing to read before
	// giving up.
	maxIterations := (maxNodePages * maxNodePageSize) / pageSize
	if maxIterations < 1 {
		maxIterations = 1
	}

	for page := int32(0); page < maxIterations; page++ {
		limit := pageSize
		if opts.MaxItems > 0 {
			if remaining := opts.MaxItems - int32(len(nodes)); remaining < limit {
				limit = remaining
			}
		}

		// Over-fetch by one to detect whether more data exists, but never
		// exceed the Nodes API's documented limit maximum (500): when limit
		// is already at that cap, limit+1 would be rejected, so request
		// exactly limit instead and treat a full page as "assume more".
		requestLimit := limit + 1
		cappedAtMax := requestLimit > maxNodePageSize
		if cappedAtMax {
			requestLimit = maxNodePageSize
		}

		result, err := bns.nodeClient.GetNodeList(ctx, clusterUuid, offset, requestLimit)
		if err != nil {
			return nil, nil, err
		}

		items := result.GetNodes()
		var hasMore bool
		if cappedAtMax {
			hasMore = int32(len(items)) >= limit
		} else {
			hasMore = int32(len(items)) > limit
		}
		if int32(len(items)) > limit {
			items = items[:limit]
		}

		nodes = append(nodes, items...)
		offset += int32(len(items))

		if !hasMore {
			return nodes, nil, nil
		}

		reachedMax := opts.MaxItems > 0 && int32(len(nodes)) >= opts.MaxItems
		if opts.NoPagination || reachedMax {
			nextToken, err := structs.OffsetToNextToken(offset)
			if err != nil {
				return nil, nil, err
			}
			return nodes, nextToken, nil
		}
	}

	return nil, nil, fmt.Errorf("stopped after reading %d pages: the server keeps returning full pages", maxIterations)
}

func resolveStartOffset(opts *apilist.PageOptions) (int32, error) {
	if opts.NextToken.IsSet() {
		offset, err := structs.NextTokenToOffset(&opts.NextToken)
		if err != nil {
			return 0, err
		}
		if offset != nil {
			return *offset, nil
		}
	}
	return opts.Offset, nil
}

func clampPageSize(pageSize int32) int32 {
	if pageSize <= 0 {
		return defaultNodePageSize
	}
	if pageSize > maxNodePageSize {
		return maxNodePageSize
	}
	return pageSize
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

func GetNodeList(ctx context.Context, domain, token, clusterUuid string, opts apilist.PageOptions) (*list.Output, error) {
	nodeService, err := GetNodeLister(domain, token)
	if err != nil {
		return nil, err
	}
	return nodeService.List(ctx, clusterUuid, opts)
}

func (bns *BackendNodeService) CheckIfNodesExist(ctx context.Context, clusterUuid string, nodesToCheck []string) (existing []string, notExisting []string, err error) {
	if len(nodesToCheck) == 0 {
		return existing, notExisting, nil
	}

	// get list of existing nodes
	actualNodes, err := FetchAllNodes(ctx, bns.nodeClient, clusterUuid)
	if err != nil {
		return existing, notExisting, err
	}

	// turn existing node list to map for efficiency
	actualNodesMap := make(map[string]bool)
	for _, actual := range actualNodes {
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
