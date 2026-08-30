package plugins

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/constants"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

// DynamoGraphDeploymentPlugin handles DynamoGraphDeployment-specific logic
type DynamoGraphDeploymentPlugin struct{}

func (d *DynamoGraphDeploymentPlugin) ProcessManifest(manifest map[string]any, project *orgUnitApi.Project) error {
	if project == nil {
		return fmt.Errorf("project cannot be nil")
	}

	metadata, ok := manifest[constants.FieldMetadata].(map[string]any)
	if !ok {
		return fmt.Errorf("manifest metadata is missing or not a valid map")
	}

	annotations, ok := metadata[constants.FieldAnnotations].(map[string]any)
	if !ok {
		annotations = make(map[string]any)
		metadata[constants.FieldAnnotations] = annotations
	}

	if _, exists := annotations[KubeAnnotationKaiSchedulerQueue]; exists {
		return nil
	}

	annotations[KubeAnnotationKaiSchedulerQueue] = d.deriveQueueName(project)

	return nil
}

// deriveQueueName isolates the logic for naming
func (d *DynamoGraphDeploymentPlugin) deriveQueueName(project *orgUnitApi.Project) string {
	queueName := project.Name
	pools := project.GetDefaultNodePools()

	if len(pools) > 0 {
		firstPool := pools[0]
		if firstPool != DefaultNodepoolName {
			queueName = fmt.Sprintf("%s-%s", queueName, firstPool)
		}
	}

	return queueName
}
