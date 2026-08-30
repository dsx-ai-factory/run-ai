package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var NetworkTopologyEdit = models.PermissionSet{
	Name:        "Network topology edit",
	Description: "Grants full permissions to create, read, update, and delete network topologies, and to access all related data and resources required for their management.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_NETWORK_TOPOLOGIES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTER,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
	},
}
