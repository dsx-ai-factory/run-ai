/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"fmt"
	"testing"
	"time"

	api "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/stretchr/testify/assert"
)

func TestClusterVersionHistory_CTOR(t *testing.T) {
	ch := *NewClusterVersionHistory([]api.ClusterHistory{
		{
			Version: "3", UpdateTimestamp: time.Now().Unix(),
		}, {
			Version: "2", UpdateTimestamp: time.Now().Add(-12 * time.Hour).Unix(),
		}, {
			Version: "4", UpdateTimestamp: time.Now().Add(12 * time.Hour).Unix(),
		}, {
			Version: "1", UpdateTimestamp: time.Now().Add(-24 * time.Hour).Unix(),
		}, {
			Version: "5", UpdateTimestamp: time.Now().Add(36 * time.Hour).Unix(),
		},
	})

	for i := 0; i < 5; i += 1 {
		assert.Equal(t, ch.Versions[i].Version.String(), fmt.Sprintf("%d.0.0", i+1))
	}
	assert.Equal(t, ch.CurrentVersion.String(), "5.0.0")
	assert.Equal(t, ch.Versions[0].End, ch.Versions[1].Start)
	assert.Equal(t, ch.Versions[1].End, ch.Versions[2].Start)
	assert.Equal(t, ch.Versions[2].End, ch.Versions[3].Start)
	assert.Equal(t, ch.Versions[3].End, ch.Versions[4].Start)
	assert.Equal(t, ch.Versions[4].End, &max_time)
}

func TestClusterVersionHistory_CTOR_illegalVersion(t *testing.T) {
	ch := *NewClusterVersionHistory([]api.ClusterHistory{
		{
			Version: "3asd", UpdateTimestamp: time.Now().Unix(),
		},
	})
	assert.Equal(t, len(ch.Versions), 0)
}
