package helpers

import (
	"fmt"
)

const (
	Field = "ancestors"
)

type Ancestors struct {
	TenantId     string  `json:"tenant_id"` // Any resource should have a tenant ancestor
	ClusterUuid  *string `json:"cluster_uuid"`
	DepartmentId *string `json:"department_id"` // will be an array once we have multi hierarchy
	ProjectId    *string `json:"project_id"`
}

func GetPlural(resourceType string) string {
	switch resourceType {
	case "project":
		return "projects"
	case "department":
		return "departments"
	case "cluster":
		return "clusters"
	}

	return resourceType
}

func ConvertSubResource(resourceType string) string {
	// for sub resources, the ancestors are calculated by the parent resource
	switch resourceType {
	case "datavolumes/sharedscopes":
		return "datavolumes"
	case "nodepools-minimal":
		return "nodepools"
	case "clusters-minimal":
		return "cluster"
	}

	return resourceType
}

func GetKey(resourceId string, resourceType string) string {
	pluralResourceType := ConvertSubResource(GetPlural(resourceType))
	return fmt.Sprintf("%s-%s", pluralResourceType, resourceId)
}
