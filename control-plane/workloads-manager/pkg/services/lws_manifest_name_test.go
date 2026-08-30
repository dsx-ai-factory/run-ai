package services

import (
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/stretchr/testify/require"
)

func TestShouldPreserveManifestName(t *testing.T) {
	require.True(t, shouldPreserveManifestName(map[string]interface{}{
		"apiVersion": constants.LwsGroup + "/" + constants.LwsVersion,
		"kind":       constants.LwsKind,
		"metadata": map[string]interface{}{
			"name": "dist-inf-abcd",
			"labels": map[string]interface{}{
				ConvertedFromNativeLabel: TrueAsString,
			},
		},
	}))
	require.False(t, shouldPreserveManifestName(map[string]interface{}{
		"apiVersion": constants.LwsGroup + "/" + constants.LwsVersion,
		"kind":       constants.LwsKind,
		"metadata":   map[string]interface{}{"name": "user-lws"},
	}))
	require.False(t, shouldPreserveManifestName(map[string]interface{}{
		"apiVersion": "batch/v1",
		"kind":       "Job",
	}))
}
