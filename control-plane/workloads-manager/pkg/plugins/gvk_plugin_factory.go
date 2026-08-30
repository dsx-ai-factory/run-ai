package plugins

import (
	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

const (
	DynamoGraphDeploymentKind  = "DynamoGraphDeployment"
	DynamoGraphDeploymentGroup = "nvidia.com"
)

// Factory creates GVK-specific plugins based on GVK
type Factory struct{}

// NewFactory creates a new instance of the factory
func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) CreatePlugin(gvk api.Gvk) GVKPlugin {
	if gvk.Kind == DynamoGraphDeploymentKind &&
		convutils.Deref(gvk.Group, "") == DynamoGraphDeploymentGroup {
		return &DynamoGraphDeploymentPlugin{}
	}

	return &NoOpPlugin{}
}
