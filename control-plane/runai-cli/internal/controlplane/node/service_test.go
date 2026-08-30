package node

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	nodesManipulate "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node/list/manipulate"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cpmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node Service")
}

var _ = Describe("Node Service Test", func() {
	var (
		ctrl   *gomock.Controller
		mocknc *cpmocks.MockNodeClient
		nbs    *BackendNodeService
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
		nodes := openapi.Nodes{Nodes: nodeList}
		mocknc.EXPECT().GetNodeList(gomock.Any(), ClusterUuid).Return(&nodes, nil).AnyTimes()
	})

	AfterEach(func() {
		// We add this in order to check that all the registered mocks ware really called
		ctrl.Finish()
	})

	Context("Service get nodes", func() {
		It("Check brief match", func() {
			nodeTelemetryResponse := getMockTelemetryResponse(map[string]string{NodeName: FreeGpu})
			mocknc.EXPECT().GetNodesTelemetry(gomock.Any(), ClusterUuid, RequiredNodeTelemetryTypes).Return(nodeTelemetryResponse, nil).AnyTimes()

			handler, err := nbs.List(context.Background(), ClusterUuid)
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

			handler, err := nbs.List(context.Background(), ClusterUuid)
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
})

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
