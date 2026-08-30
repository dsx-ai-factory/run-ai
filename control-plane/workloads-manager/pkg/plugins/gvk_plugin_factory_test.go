package plugins

import (
	"testing"

	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/stretchr/testify/assert"
)

func TestGVKPluginFactory_CreatePlugin_DynamoGraphDeployment(t *testing.T) {
	factory := NewFactory()

	gvk := api.Gvk{
		Kind:    DynamoGraphDeploymentKind,
		Group:   convutils.Ref(DynamoGraphDeploymentGroup),
		Version: "v1",
	}

	plugin := factory.CreatePlugin(gvk)
	assert.IsType(t, &DynamoGraphDeploymentPlugin{}, plugin)
}

func TestGVKPluginFactory_CreatePlugin_OtherKinds(t *testing.T) {
	factory := NewFactory()

	testCases := []struct {
		name string
		gvk  api.Gvk
	}{
		{
			name: "Pod",
			gvk: api.Gvk{
				Kind:    "Pod",
				Group:   convutils.Ref(""),
				Version: "v1",
			},
		},
		{
			name: "Deployment",
			gvk: api.Gvk{
				Kind:    "Deployment",
				Group:   convutils.Ref("apps"),
				Version: "v1",
			},
		},
		{
			name: "Job",
			gvk: api.Gvk{
				Kind:    "Job",
				Group:   convutils.Ref("batch"),
				Version: "v1",
			},
		},
		{
			name: "NIMService",
			gvk: api.Gvk{
				Kind:    "NIMService",
				Group:   convutils.Ref("apps.nvidia.com"),
				Version: "v1alpha1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plugin := factory.CreatePlugin(tc.gvk)
			assert.IsType(t, &NoOpPlugin{}, plugin, "Expected NoOpPlugin for %s", tc.name)
		})
	}
}

func TestGVKPluginFactory_CreatePlugin_DifferentGroup(t *testing.T) {
	factory := NewFactory()

	// Same kind but different group - should get NoOpPlugin
	gvk := api.Gvk{
		Kind:    DynamoGraphDeploymentKind,
		Group:   convutils.Ref("other.group"),
		Version: "v1",
	}

	plugin := factory.CreatePlugin(gvk)
	assert.IsType(t, &NoOpPlugin{}, plugin)
}
