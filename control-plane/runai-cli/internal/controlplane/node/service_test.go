/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package node

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/node/list"
	nodesManipulate "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/node/list/manipulate"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/structs"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"

	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node Service")
}

var _ = Describe("Node Service Test", func() {
	var (
		ctrl           *gomock.Controller
		mocknc         *cpmocks.MockNodeClient
		nbs            *BackendNodeService
		nodeListResult openapi.Nodes
	)

	const (
		ClusterUuid  = "one-uuid"
		GpuCount     = 321
		FreeGpu      = "100"
		AllocatedGpu = "221"
		NodeName     = "my name is dudu"
		Status       = "i'm fine"
		Nodepool     = "pool party"
		GpuType      = "jensen"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mocknc = cpmocks.NewMockNodeClient(ctrl)
		nbs = NewBackendNodeService(mocknc)

		nodeList := []openapi.Node{{
			Status:      Status,
			GpuInfo:     *openapi.NewNullableGpuInfo(openapi.NewGpuInfo(GpuType, GpuCount)),
			Name:        NodeName,
			ClusterUuid: ClusterUuid,
			NodePool:    Nodepool,
		}}
		nodeListResult = openapi.Nodes{Nodes: nodeList}
	})

	AfterEach(func() {
		// We add this in order to check that all the registered mocks ware really called
		ctrl.Finish()
	})

	Context("Service get nodes", func() {
		BeforeEach(func() {
			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, gomock.Any(), gomock.Any()).Return(&nodeListResult, nil).AnyTimes()
		})

		It("Check brief match", func() {
			nodeTelemetryResponse := getMockTelemetryResponse(map[string]string{NodeName: FreeGpu})
			mocknc.EXPECT().GetNodesTelemetry(gomock.Any(), ClusterUuid, RequiredNodeTelemetryTypes).Return(nodeTelemetryResponse, nil).AnyTimes()

			handler, err := nbs.List(context.Background(), ClusterUuid, apilist.PageOptions{})
			Expect(err).To(BeNil())
			response, err := handler.GenerateJson()
			Expect(err).To(BeNil())
			var data nodesManipulate.BriefData
			err = json.Unmarshal([]byte(response), &data)
			Expect(err).To(BeNil())

			briefNodes := data.Nodes
			Expect(briefNodes).To(HaveLen(1))
			Expect(briefNodes[0].Name).To(Equal(NodeName))
			Expect(briefNodes[0].Status).To(Equal(Status))
			Expect(briefNodes[0].GpuInfo.GetGpuType()).To(Equal(GpuType))
			Expect(briefNodes[0].GpuInfo.GetGpuCount()).To(Equal(int32(GpuCount)))
			Expect(briefNodes[0].Nodepool).To(Equal(Nodepool))
			Expect(briefNodes[0].FreeGpu).To(Equal(FreeGpu))
			Expect(briefNodes[0].GpuAllocation).To(Equal(AllocatedGpu + "/" + strconv.Itoa(GpuCount)))
		})

		It("Brief match when node telemetry does not exist", func() {
			nodeTelemetryResponse := getMockTelemetryResponse(map[string]string{"wrong node name": "1"})
			mocknc.EXPECT().GetNodesTelemetry(gomock.Any(), ClusterUuid, RequiredNodeTelemetryTypes).Return(nodeTelemetryResponse, nil).AnyTimes()

			handler, err := nbs.List(context.Background(), ClusterUuid, apilist.PageOptions{})
			Expect(err).To(BeNil())
			nodes := handler.GetList()
			Expect(nodes).To(HaveLen(1))

			brief := nodes[0]
			Expect(brief.FreeGpu).To(Equal(nodesManipulate.NotAvailableString))
			Expect(brief.GpuAllocation).To(Equal(nodesManipulate.NotAvailableString))
		})

	})

	Context("CheckIfNodesExist", func() {
		const (
			NonExistingNodeName = "bla"
		)

		BeforeEach(func() {
			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, gomock.Any(), gomock.Any()).Return(&nodeListResult, nil).AnyTimes()
		})

		It("returns existing and non-existing, without duplicates", func() {
			toCheck := []string{NodeName, NonExistingNodeName, NonExistingNodeName, NonExistingNodeName, NodeName}
			existing, nonExisting, err := nbs.CheckIfNodesExist(context.Background(), ClusterUuid, toCheck)
			Expect(err).NotTo(HaveOccurred())

			Expect(existing).To(ContainElements(NodeName))
			Expect(existing).To(HaveLen(1))

			Expect(nonExisting).To(ContainElements(NonExistingNodeName))
			Expect(nonExisting).To(HaveLen(1))
		})

		It("nodes to check is empty", func() {
			existing, nonExisting, err := nbs.CheckIfNodesExist(context.Background(), ClusterUuid, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(existing).To(BeNil())
			Expect(nonExisting).To(BeNil())
		})
	})

	Context("List pagination", func() {
		makeNode := func(name string) openapi.Node {
			return openapi.Node{Name: name, ClusterUuid: ClusterUuid}
		}

		BeforeEach(func() {
			mocknc.EXPECT().GetNodesTelemetry(gomock.Any(), ClusterUuid, RequiredNodeTelemetryTypes).Return(api.NodeTelemetriesByType{}, nil).AnyTimes()
		})

		It("returns the full list by default and emits no next token when the data is exhausted", func() {
			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, int32(0), int32(101)).
				Return(&openapi.Nodes{Nodes: []openapi.Node{makeNode("a"), makeNode("b")}}, nil)

			handler, err := nbs.List(context.Background(), ClusterUuid, apilist.PageOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(handler.GetList()).To(HaveLen(2))
		})

		It("stops after a single page when NoPagination is set and emits a next token", func() {
			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, int32(0), int32(3)).
				Return(&openapi.Nodes{Nodes: []openapi.Node{makeNode("a"), makeNode("b"), makeNode("c")}}, nil)

			handler, err := nbs.List(context.Background(), ClusterUuid, apilist.PageOptions{PageSize: 2, NoPagination: true})
			Expect(err).NotTo(HaveOccurred())
			Expect(handler.GetList()).To(HaveLen(2))

			resumeOffset, err := resumeOffsetFromOutput(handler)
			Expect(err).NotTo(HaveOccurred())
			Expect(resumeOffset).NotTo(BeNil())
			Expect(*resumeOffset).To(Equal(int32(2)))
		})

		It("loops across pages until MaxItems is reached and emits a next token", func() {
			gomock.InOrder(
				mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, int32(0), int32(3)).
					Return(&openapi.Nodes{Nodes: []openapi.Node{makeNode("a"), makeNode("b"), makeNode("c")}}, nil),
				mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, int32(2), int32(2)).
					Return(&openapi.Nodes{Nodes: []openapi.Node{makeNode("c"), makeNode("d")}}, nil),
			)

			handler, err := nbs.List(context.Background(), ClusterUuid, apilist.PageOptions{PageSize: 2, MaxItems: 3})
			Expect(err).NotTo(HaveOccurred())
			Expect(handler.GetList()).To(HaveLen(3))

			resumeOffset, err := resumeOffsetFromOutput(handler)
			Expect(err).NotTo(HaveOccurred())
			Expect(resumeOffset).NotTo(BeNil())
			Expect(*resumeOffset).To(Equal(int32(3)))
		})

		It("never requests more than the API's max limit when --page-size is 500", func() {
			fullPage := make([]openapi.Node, maxNodePageSize)
			for i := range fullPage {
				fullPage[i] = makeNode(fmt.Sprintf("node-%d", i))
			}

			// GetNodeList must be called with exactly maxNodePageSize (500),
			// never maxNodePageSize+1 — the over-fetch-by-one trick must not
			// push the requested limit past the API's documented maximum.
			// gomock's exact-args matching means a call with limit=501 would
			// fail this expectation outright.
			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, int32(0), maxNodePageSize).
				Return(&openapi.Nodes{Nodes: fullPage}, nil)

			handler, err := nbs.List(context.Background(), ClusterUuid, apilist.PageOptions{PageSize: maxNodePageSize, NoPagination: true})
			Expect(err).NotTo(HaveOccurred())
			Expect(handler.GetList()).To(HaveLen(int(maxNodePageSize)))

			// A full page at the API's cap can't be over-fetched to prove
			// there's more, so it must conservatively assume there is.
			resumeOffset, err := resumeOffsetFromOutput(handler)
			Expect(err).NotTo(HaveOccurred())
			Expect(resumeOffset).NotTo(BeNil())
			Expect(*resumeOffset).To(Equal(maxNodePageSize))
		})

		It("starts from the offset encoded in NextToken", func() {
			token, err := structs.OffsetToNextToken(5)
			Expect(err).NotTo(HaveOccurred())

			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, int32(5), int32(3)).
				Return(&openapi.Nodes{Nodes: []openapi.Node{makeNode("a")}}, nil)

			handler, err := nbs.List(context.Background(), ClusterUuid, apilist.PageOptions{PageSize: 2, NoPagination: true, NextToken: *token})
			Expect(err).NotTo(HaveOccurred())
			Expect(handler.GetList()).To(HaveLen(1))
		})

		It("starts from Offset when no NextToken is set", func() {
			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, int32(5), int32(3)).
				Return(&openapi.Nodes{Nodes: []openapi.Node{makeNode("a")}}, nil)

			handler, err := nbs.List(context.Background(), ClusterUuid, apilist.PageOptions{PageSize: 2, NoPagination: true, Offset: 5})
			Expect(err).NotTo(HaveOccurred())
			Expect(handler.GetList()).To(HaveLen(1))
		})

		It("propagates errors from the client", func() {
			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, int32(0), int32(3)).
				Return(nil, fmt.Errorf("boom"))

			_, err := nbs.List(context.Background(), ClusterUuid, apilist.PageOptions{PageSize: 2, NoPagination: true})
			Expect(err).To(HaveOccurred())
		})

		It("gives up instead of looping forever when the server ignores offset", func() {
			// The iteration cap is derived from a fixed total-node budget
			// divided by the page size, so use a large page size here (the
			// max) to keep the derived cap—and thus this test—fast: the
			// same maxNodePages call count as before.
			fullPage := make([]openapi.Node, maxNodePageSize)
			for i := range fullPage {
				fullPage[i] = makeNode(fmt.Sprintf("node-%d", i))
			}
			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, gomock.Any(), maxNodePageSize).
				Return(&openapi.Nodes{Nodes: fullPage}, nil).Times(maxNodePages)

			_, err := nbs.List(context.Background(), ClusterUuid, apilist.PageOptions{PageSize: maxNodePageSize})
			Expect(err).To(MatchError(ContainSubstring("keeps returning full pages")))
		})
	})

	Context("FetchAllNodes", func() {
		makeNode := func(name string) openapi.Node {
			return openapi.Node{Name: name, ClusterUuid: ClusterUuid}
		}

		It("collects every page when the cluster has more nodes than one page holds", func() {
			firstPage := make([]openapi.Node, maxNodePageSize)
			for i := range firstPage {
				firstPage[i] = makeNode(fmt.Sprintf("node-%d", i))
			}
			secondPage := []openapi.Node{makeNode("extra-1"), makeNode("extra-2")}

			gomock.InOrder(
				mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, int32(0), maxNodePageSize).Return(&openapi.Nodes{Nodes: firstPage}, nil),
				mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, maxNodePageSize, maxNodePageSize).Return(&openapi.Nodes{Nodes: secondPage}, nil),
			)

			nodes, err := FetchAllNodes(context.Background(), mocknc, ClusterUuid)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodes).To(HaveLen(int(maxNodePageSize) + 2))
		})

		It("stops at the first partial page", func() {
			page := []openapi.Node{makeNode("a"), makeNode("b")}
			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, int32(0), maxNodePageSize).Return(&openapi.Nodes{Nodes: page}, nil)

			nodes, err := FetchAllNodes(context.Background(), mocknc, ClusterUuid)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodes).To(HaveLen(2))
		})

		It("returns no nodes for an empty cluster", func() {
			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, int32(0), maxNodePageSize).Return(&openapi.Nodes{Nodes: []openapi.Node{}}, nil)

			nodes, err := FetchAllNodes(context.Background(), mocknc, ClusterUuid)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodes).To(BeEmpty())
		})

		It("gives up instead of looping forever when the server ignores offset", func() {
			fullPage := make([]openapi.Node, maxNodePageSize)
			for i := range fullPage {
				fullPage[i] = makeNode(fmt.Sprintf("node-%d", i))
			}
			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, gomock.Any(), maxNodePageSize).Return(&openapi.Nodes{Nodes: fullPage}, nil).Times(maxNodePages)

			_, err := FetchAllNodes(context.Background(), mocknc, ClusterUuid)
			Expect(err).To(MatchError(ContainSubstring("keeps returning full pages")))
		})

		It("propagates a failure from any page", func() {
			mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid, int32(0), maxNodePageSize).Return(nil, fmt.Errorf("boom"))

			_, err := FetchAllNodes(context.Background(), mocknc, ClusterUuid)
			Expect(err).To(HaveOccurred())
		})
	})
})

func resumeOffsetFromOutput(handler *list.Output) (*int32, error) {
	response, err := handler.GenerateJson()
	if err != nil {
		return nil, err
	}

	var data nodesManipulate.BriefData
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return nil, err
	}
	if data.NextToken == nil {
		return nil, nil
	}

	token, err := structs.NextTokenFromValue(*data.NextToken)
	if err != nil {
		return nil, err
	}
	return structs.NextTokenToOffset(token)
}

func getMockTelemetryResponse(nodeNameToFreeGpu map[string]string) api.NodeTelemetriesByType {
	values := make([]openapi.TelemetryResponseValuesInner, 0)
	for nodeName, freeGpu := range nodeNameToFreeGpu {
		values = append(values, openapi.TelemetryResponseValuesInner{
			Value: freeGpu,
			Groups: []openapi.TelemetryResponseValuesInnerGroupsInner{
				{
					Key:   "Node",
					Value: nodeName,
				},
			},
		})
	}

	telemetryResponse := &openapi.TelemetryResponse{
		Type:   string(openapi.NODETELEMETRYTYPE_FREE_GPUS),
		Values: values,
	}

	return api.NodeTelemetriesByType{openapi.NODETELEMETRYTYPE_FREE_GPUS: telemetryResponse}
}
