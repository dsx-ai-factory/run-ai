package priorities

import (
	"fmt"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type Brief struct {
	Name string `json:"name"`
}

func newBriefFromClusterPriority(priority openapi.ClusterPriority) Brief {
	return Brief{priority.Name}
}

func BriefListFromResponse(prioritiesList []openapi.ClusterPriority) ([]Brief, error) {
	if prioritiesList == nil {
		return nil, fmt.Errorf("no priority suggestions available")
	}
	var briefList []Brief

	for _, priority := range prioritiesList {
		briefList = append(briefList, newBriefFromClusterPriority(priority))
	}

	return briefList, nil
}
