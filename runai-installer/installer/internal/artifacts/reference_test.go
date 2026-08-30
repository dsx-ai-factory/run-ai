package artifacts

import (
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMirrorRefPreservesImageGroup(t *testing.T) {
	entry := artifactlock.ImageEntry{
		Group: "kai-scheduler",
		Image: artifactlock.Image{
			Name:   "operator",
			Source: "ghcr.io/kai-scheduler/kai-scheduler/operator:v0.16.8",
		},
	}

	ref, tag, err := MirrorRef("registry.example.com/runai", entry)

	require.NoError(t, err)
	assert.Equal(t, "registry.example.com/runai/kai-scheduler/operator:v0.16.8", ref)
	assert.Equal(t, "v0.16.8", tag)
}
