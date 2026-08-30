package artifactlock

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageEntriesPreserveGroups(t *testing.T) {
	gpuOperator := testImage("operator", "1")
	kaiOperator := testImage("operator", "2")
	lock := &ArtifactLock{Spec: Spec{ImageLocks: []ImageLock{
		{Name: "gpu-sharing", Version: "v0.1.0", Images: []Image{gpuOperator}},
		{Name: "kai-scheduler", Version: "v0.16.8", Images: []Image{kaiOperator}},
	}}}

	require.NoError(t, lock.ValidateImages())
	assert.Equal(t, []ImageEntry{
		{Group: "gpu-sharing", GroupVersion: "v0.1.0", Image: gpuOperator},
		{Group: "kai-scheduler", GroupVersion: "v0.16.8", Image: kaiOperator},
	}, lock.ImageEntries())
}

func TestValidateImagesRejectsDuplicateGroups(t *testing.T) {
	lock := &ArtifactLock{Spec: Spec{ImageLocks: []ImageLock{
		{Name: "runai", Images: []Image{testImage("one", "1")}},
		{Name: "runai", Images: []Image{testImage("two", "2")}},
	}}}

	require.EqualError(t, lock.ValidateImages(), `duplicate image lock group "runai"`)
}

func TestValidateImagesRejectsDuplicateNamesWithinGroup(t *testing.T) {
	lock := &ArtifactLock{Spec: Spec{ImageLocks: []ImageLock{{
		Name: "runai", Images: []Image{testImage("operator", "1"), testImage("operator", "2")},
	}}}}

	require.EqualError(t, lock.ValidateImages(), `duplicate image "operator" in group "runai"`)
}

func TestValidateImagesRejectsUnpinnedImage(t *testing.T) {
	image := testImage("operator", "1")
	image.Image = "registry.example.com/operator:latest"
	lock := &ArtifactLock{Spec: Spec{ImageLocks: []ImageLock{{
		Name: "runai", Images: []Image{image},
	}}}}

	require.ErrorContains(t, lock.ValidateImages(), "must use an immutable sha256 digest")
}

func testImage(name, digit string) Image {
	return Image{
		Name:        name,
		Image:       "registry.example.com/" + name + "@sha256:" + strings.Repeat(digit, 64),
		Source:      "registry.example.com/" + name + ":latest",
		IndexDigest: "sha256:" + strings.Repeat(digit, 64),
	}
}
