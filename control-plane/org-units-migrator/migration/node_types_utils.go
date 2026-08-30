package migration

import (
	"context"
	"fmt"
	"time"

	orgUnitsServiceDao "github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	orgUnitServiceModels "github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

func CreateNodeTypes(nodeTypes []models.NodeType) error {
	nodeTypesToCreate := convutils.ConvertSliceTypes(nodeTypes, mapNodeTypeToV2)
	return orgUnitsServiceDao.CreateNodeTypes(context.Background(), nodeTypesToCreate)
}

func mapNodeTypeToV2(nodeType models.NodeType) *orgUnitServiceModels.NodeType {
	return &orgUnitServiceModels.NodeType{
		IntBaseModel: orgUnitServiceModels.IntBaseModel{
			ID:        int64(nodeType.ID),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: "",
			UpdatedBy: "",
		},
		Name:      nodeType.Name,
		ClusterId: nodeType.ClusterUUID,
		TenantId:  fmt.Sprintf("%d", nodeType.TenantID),
	}
}
