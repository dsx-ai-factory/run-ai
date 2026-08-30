package plugins

import (
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

// GVKPlugin defines the interface for GVK-specific processing
type GVKPlugin interface {
	ProcessManifest(manifest map[string]any, project *orgUnitApi.Project) error
}
