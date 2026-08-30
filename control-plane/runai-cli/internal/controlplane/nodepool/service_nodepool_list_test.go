package nodepool

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	nodesList "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"

	"gopkg.in/yaml.v3"
	"k8s.io/utils/ptr"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"
	nodesManipulate "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node/list/manipulate"

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
		ctrl               *gomock.Controller
		mockNodepoolClient *cpmocks.MockNodepoolClient
		mockNodeClient     *cpmocks.MockNodeClient
		nbs                *BackendNodepoolService
	)

	const (
		ClusterUuid   = "one-uuid"
		GpuCount      = 321
		FreeGpu       = "100"
		AllocatedGpu  = "221"
		NodeName      = "my name is dudu"
		OtherNodeName = "my name is yaakov"
		Status        = "Ready"
		Nodepool      = "pool party"
		GpuType       = "jensen"
		TopologyName  = "nt-1"

		EmptyPoolName   = "empty"
		EmptyPoolStatus = "Empty"

		NilPoolName   = "nil"
		NilPoolStatus = "Unschedulable"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockNodepoolClient = cpmocks.NewMockNodepoolClient(ctrl)
		mockNodeClient = cpmocks.NewMockNodeClient(ctrl)
		nbs = NewBackendNodepoolService(mockNodepoolClient, mockNodeClient)

		nodeList := []openapi.Node{
			{
				Status:      Status,
				GpuInfo:     *openapi.NewNullableGpuInfo(openapi.NewGpuInfo(GpuType, GpuCount)),
				Name:        NodeName,
				ClusterUuid: ClusterUuid,
				NodePool:    Nodepool,
			},
			{
				Status:      Status,
				GpuInfo:     *openapi.NewNullableGpuInfo(openapi.NewGpuInfo(GpuType, GpuCount)),
				Name:        OtherNodeName,
				ClusterUuid: ClusterUuid,
				NodePool:    Nodepool,
			},
		}
		nodes := openapi.Nodes{Nodes: nodeList}
		mockNodeClient.EXPECT().GetNodeList(gomock.Any(), ClusterUuid).Return(&nodes, nil).AnyTimes()

		poolStatus := openapi.NODEPOOLPHASE_READY
		emptyPoolStatus := openapi.NODEPOOLPHASE_EMPTY
		nilPoolStatus := openapi.NODEPOOLPHASE_UNSCHEDULABLE

		poolWithTopology := openapi.Nodepool{
			Name:  Nodepool,
			Phase: &poolStatus,
			Status: &openapi.NodepoolSyncFieldsStatus{
				GpuNetworkAccelerationDetected: ptr.To(true),
			},
		}
		poolWithTopology.SetNetworkTopologyName(TopologyName)

		nodepoolList := &openapi.GetNodepools200Response{Nodepools: []openapi.Nodepool{
			poolWithTopology,
			{
				Name:  EmptyPoolName,
				Phase: &emptyPoolStatus,
				Status: &openapi.NodepoolSyncFieldsStatus{
					GpuNetworkAccelerationDetected: ptr.To(false),
				},
			},
			{
				Name:  NilPoolName,
				Phase: &nilPoolStatus,
			},
		}}
		mockNodepoolClient.EXPECT().GetNodepoolList(gomock.Any(), filter.NewBuilder().WithClusterID(ClusterUuid).StringList()).Return(nodepoolList, nil).AnyTimes()
	})

	AfterEach(func() {
		// We add this in order to check that all the registered mocks ware really called
		ctrl.Finish()
	})

	Context("Service get nodepool", func() {
		Context("Output validation", func() {
			var (
				briefOut *Output
			)

			BeforeEach(func() {
				nodeTelemetryResponse := getMockTelemetryResponse(map[string]string{NodeName: FreeGpu, OtherNodeName: strconv.Itoa(GpuCount)})
				mockNodeClient.EXPECT().GetNodesTelemetry(gomock.Any(), ClusterUuid, nodesList.RequiredNodeTelemetryTypes).Return(nodeTelemetryResponse, nil).AnyTimes()

				nodepoolList, nodeList, telemetry, err := nbs.GetNodepoolList(context.Background(), ClusterUuid)
				Expect(err).To(BeNil())

				briefOut, err = NewOutput(nodepoolList, nodeList, telemetry)
			})

			It("Check brief match - json", func() {
				response, err := briefOut.GenerateJson()
				Expect(err).To(BeNil())
				var data Data
				err = json.Unmarshal([]byte(response), &data)
				Expect(err).To(BeNil())

				// non-empty pool
				briefNodepools := data.Nodepools
				Expect(briefNodepools[0].Name).To(Equal(Nodepool))
				Expect(briefNodepools[0].Status).To(Equal(Status))
				Expect(briefNodepools[0].GpuAllocation).To(Equal(AllocatedGpu + "/" + strconv.Itoa(GpuCount*2)))
				Expect(briefNodepools[0].Nodes).To(Equal([]string{NodeName, OtherNodeName}))
				Expect(briefNodepools[0].MNNVLDetected).To(Equal(true))
				Expect(briefNodepools[0].NetworkTopologyName).To(Equal(TopologyName))

				// empty pool
				Expect(briefNodepools[1].Name).To(Equal(EmptyPoolName))
				Expect(briefNodepools[1].Status).To(Equal(EmptyPoolStatus))
				Expect(briefNodepools[1].GpuAllocation).To(Equal(nodesManipulate.NotAvailableString))
				Expect(briefNodepools[1].MNNVLDetected).To(Equal(false))

				// nil pool
				Expect(briefNodepools[2].Name).To(Equal(NilPoolName))
				Expect(briefNodepools[2].Status).To(Equal(NilPoolStatus))
				Expect(briefNodepools[2].GpuAllocation).To(Equal(nodesManipulate.NotAvailableString))
				Expect(briefNodepools[2].MNNVLDetected).To(Equal(false))

				var emptyNodes []string
				Expect(briefNodepools[1].Nodes).To(Equal(emptyNodes))
			})

			It("Check brief match - yaml", func() {
				response, err := briefOut.GenerateYaml()
				Expect(err).To(BeNil())
				var data Data
				err = yaml.Unmarshal([]byte(response), &data)
				Expect(err).To(BeNil())

				// non-empty pool
				briefNodepools := data.Nodepools
				Expect(briefNodepools[0].Name).To(Equal(Nodepool))
				Expect(briefNodepools[0].Status).To(Equal(Status))
				Expect(briefNodepools[0].GpuAllocation).To(Equal(AllocatedGpu + "/" + strconv.Itoa(GpuCount*2)))
				Expect(briefNodepools[0].Nodes).To(Equal([]string{NodeName, OtherNodeName}))
				Expect(briefNodepools[0].MNNVLDetected).To(Equal(true))
				Expect(briefNodepools[0].NetworkTopologyName).To(Equal(TopologyName))

				// empty pool
				Expect(briefNodepools[1].Name).To(Equal(EmptyPoolName))
				Expect(briefNodepools[1].Status).To(Equal(EmptyPoolStatus))
				Expect(briefNodepools[1].GpuAllocation).To(Equal(nodesManipulate.NotAvailableString))
				Expect(briefNodepools[1].MNNVLDetected).To(Equal(false))

				// nil pool
				Expect(briefNodepools[2].Name).To(Equal(NilPoolName))
				Expect(briefNodepools[2].Status).To(Equal(NilPoolStatus))
				Expect(briefNodepools[2].GpuAllocation).To(Equal(nodesManipulate.NotAvailableString))
				Expect(briefNodepools[2].MNNVLDetected).To(Equal(false))

				// empty nodes
				Expect(briefNodepools[1].Nodes).To(HaveLen(0))
			})

			It("Check brief match - table", func() {
				printOut, err := briefOut.GenerateTable()
				Expect(err).To(BeNil())

				Expect(printOut).To(ContainSubstring("pool party  Ready          221/642               2           Detected         nt-1"))
				Expect(printOut).To(ContainSubstring("empty       Empty          N/A                   0           Not Detected"))
				Expect(printOut).To(ContainSubstring("nil         Unschedulable  N/A                   0           Not Detected"))
			})
		})

		It("Brief match when node telemetry does not exist", func() {
			nodeTelemetryResponse := getMockTelemetryResponse(map[string]string{"wrong node name": "1"})
			mockNodeClient.EXPECT().GetNodesTelemetry(gomock.Any(), ClusterUuid, nodesList.RequiredNodeTelemetryTypes).Return(nodeTelemetryResponse, nil).AnyTimes()

			nodepoolList, nodeList, telemetry, err := nbs.GetNodepoolList(context.Background(), ClusterUuid)
			Expect(err).To(BeNil())

			briefList, err := BriefListFromResponse(nodepoolList, nodeList, telemetry)
			Expect(err).To(BeNil())

			brief := briefList[0]
			Expect(brief.GpuAllocation).To(Equal(nodesManipulate.NotAvailableString))
		})

		It("Expect error when can't parse telemetry", func() {
			nodeTelemetryResponse := getMockTelemetryResponse(map[string]string{NodeName: "blabla"})
			mockNodeClient.EXPECT().GetNodesTelemetry(gomock.Any(), ClusterUuid, nodesList.RequiredNodeTelemetryTypes).Return(nodeTelemetryResponse, nil).AnyTimes()

			nodepoolList, nodeList, telemetry, err := nbs.GetNodepoolList(context.Background(), ClusterUuid)
			briefList, err := BriefListFromResponse(nodepoolList, nodeList, telemetry)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to calculate GPU info for node my name is dudu. Got error: strconv.Atoi: parsing \"blabla\": invalid syntax"))
			Expect(briefList).To(BeNil())
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
