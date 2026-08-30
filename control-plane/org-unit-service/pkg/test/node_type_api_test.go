package testing

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
)

var _ = Describe("Node type API tests", Ordered, func() {
	ctx := SetContext(TestTenant, TestUser)

	var nodeTypeService api.NodeTypesAPIServicer

	BeforeEach(func() {
		nodeTypeService = services.NewNodeTypeApiService()
	})

	Context("test CRUD node type", func() {
		It("test node type", func() {
			clusterId := uuid.New().String()
			nodeTypeName := fmt.Sprint("nodeType-", uuid.New())
			request := api.NodeTypeCreateRequest{Name: nodeTypeName, ClusterId: clusterId}
			createResponse, err := nodeTypeService.CreateNodeType(ctx, request)
			if err != nil {
				Fail("Failed creating node type")
			}

			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			created := createResponse.Body.(api.NodeType)
			nodeTypeId := created.Id
			Expect(created.Name).To(Equal(nodeTypeName))
			Expect(created.ClusterId).To(Equal(clusterId))

			// create with same name - expect 409
			createResponse, err = nodeTypeService.CreateNodeType(ctx, request)
			if err != nil {
				Fail("Failed creating node type")
			}
			Expect(createResponse.Code).To(Equal(http.StatusConflict))

			getResponse, err := nodeTypeService.GetNodeTypes(ctx, clusterId)
			if err != nil {
				Fail("Failed getting node type")
			}
			getNodeType := getResponse.Body.([]api.NodeType)
			Expect((getNodeType)[0].Id).To(Equal(nodeTypeId))
			Expect((getNodeType)[0].ClusterId).To(Equal(created.ClusterId))
			Expect((getNodeType)[0].Name).To(Equal(created.Name))

			newName := "new-name"
			updateRequest := api.NodeTypeUpdateRequest{Name: newName, ClusterId: clusterId}
			updatedResponse, err := nodeTypeService.UpdateNodeType(ctx, nodeTypeId, updateRequest)
			if err != nil {
				Fail("Failed getting node type")
			}
			Expect(updatedResponse.Code).To(Equal(http.StatusOK))
			updated := updatedResponse.Body.(api.NodeType)
			Expect(updated.Name).To(Equal(newName))
			Expect(updated.ClusterId).To(Equal(clusterId))

			getResponse, err = nodeTypeService.GetNodeTypes(ctx, clusterId)
			if err != nil {
				Fail("Failed getting node type")
			}
			getNodeType = getResponse.Body.([]api.NodeType)
			Expect((getNodeType)[0].Id).To(Equal(nodeTypeId))
			Expect((getNodeType)[0].ClusterId).To(Equal(created.ClusterId))
			Expect((getNodeType)[0].Name).To(Equal(newName))

			deleteResponse, err := nodeTypeService.DeleteNodeType(ctx, nodeTypeId)
			if err != nil {
				Fail("Failed deleting node type")
			}
			Expect(deleteResponse.Code).To(Equal(http.StatusNoContent))

			getResponse, err = nodeTypeService.GetNodeTypes(ctx, clusterId)
			if err != nil {
				Fail("Failed getting node type")
			}
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			getNodeType = getResponse.Body.([]api.NodeType)
			Expect(len(getNodeType)).To(Equal(0))
		})
	})
})
