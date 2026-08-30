package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var IntegrationMetricsRead = models.PermissionSet{
	Name:        "Integration metrics read",
	Description: "",
	Internal:    true,
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_WORKLOAD_INTEGRATION_METRICS,
			Read:         true,
		},
	},
}
