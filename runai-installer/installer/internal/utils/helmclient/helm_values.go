package helmclient

import (
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"helm.sh/helm/v3/pkg/action"
)

func GetReleaseValues(releaseName, namespace string) (map[string]interface{}, error) {
	actionConfig, _, err := createActionConfig(namespace, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create action config: %w", err)
	}
	getValues := action.NewGetValues(actionConfig)
	getValues.AllValues = true
	vals, err := getValues.Run(releaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get values for release '%s': %w", releaseName, err)
	}
	logging.Infof("[Helm] Retrieved values for release '%s' in namespace '%s'", releaseName, namespace)
	return vals, nil
}
