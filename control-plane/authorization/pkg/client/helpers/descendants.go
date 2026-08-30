package helpers

import "fmt"

const (
	FieldName = "descendants"
)

type Descendants struct {
	ProjectIds []string `json:"project_ids"`
}

func GetDescendantsPlural(resourceType string) string {
	switch resourceType {
	case "project":
		return "projects"
	}

	return resourceType
}

func GetDescendantsKey(resourceId string, resourceType string) string {
	pluralResourceType := GetDescendantsPlural(resourceType)
	return fmt.Sprintf("%s-%s", pluralResourceType, resourceId)
}
