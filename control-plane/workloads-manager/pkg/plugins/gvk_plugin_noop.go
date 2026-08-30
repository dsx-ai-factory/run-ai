package plugins

import (
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

// NoOpPlugin is the default plugin that does nothing
type NoOpPlugin struct{}

func (n *NoOpPlugin) ProcessManifest(_ map[string]any, _ *orgUnitApi.Project) error {
	return nil
}
