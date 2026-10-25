// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package configmaps

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestS3MountCheckImage(t *testing.T) {
	t.Setenv(common.DefaultRegistryEnvVarName, "bundled.example/runai")
	t.Setenv(common.DefaultVersionEnvVarName, "current-build")
	for _, tc := range []struct {
		name     string
		registry string
		image    string
		tag      string
		enabled  *bool
		want     bool
	}{
		{name: "bundled image", registry: "bundled.example/runai", image: "goofys", tag: "current-build", want: true},
		{name: "older image", registry: "bundled.example/runai", image: "goofys", tag: "older-build"},
		{name: "custom registry", registry: "custom.example", image: "goofys", tag: "current-build"},
		{name: "custom image", registry: "bundled.example/runai", image: "custom-goofys", tag: "current-build"},
		{name: "custom opt in", registry: "custom.example", image: "custom-goofys", tag: "custom-build", enabled: ptr.To(true), want: true},
		{name: "bundled opt out", registry: "bundled.example/runai", image: "goofys", tag: "current-build", enabled: ptr.To(false)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			image := &common.Image{Registry: &tc.registry, Name: &tc.image, Tag: &tc.tag}
			got := s3MountCheckImage(image, tc.enabled)
			if tc.want {
				require.Equal(t, formatImageRef(image), got)
			} else {
				require.Nil(t, got)
			}
		})
	}
	require.Nil(t, s3MountCheckImage(nil, ptr.To(true)))
	t.Setenv(common.DefaultVersionEnvVarName, "")
	require.Nil(t, s3MountCheckImage(&common.Image{Registry: ptr.To("bundled.example/runai"), Name: ptr.To("goofys"), Tag: ptr.To("")}, nil))
}
