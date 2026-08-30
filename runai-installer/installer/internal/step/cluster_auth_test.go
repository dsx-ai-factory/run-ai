package step

import (
	"testing"

	"github.com/run-ai/runai/runai-installer/clients/runai/cluster"
	"github.com/stretchr/testify/assert"
)

func TestClusterAuthValue(t *testing.T) {
	assert.Equal(t,
		"controlPlane.bootstrapToken=token",
		ClusterAuthValue(&cluster.ClusterInstallationInfoResponse{
			BootstrapToken: "token",
			ClientSecret:   "secret",
		}),
	)
	assert.Equal(t,
		"controlPlane.clientSecret=secret",
		ClusterAuthValue(&cluster.ClusterInstallationInfoResponse{
			ClientSecret: "secret",
		}),
	)
}
