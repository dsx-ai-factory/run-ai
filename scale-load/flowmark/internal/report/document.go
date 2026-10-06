/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"encoding/json"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

const (
	SchemaVersion = "v1"
	fileTimestamp = "20060102T150405Z"
)

type RunInfo struct {
	StartedAt           time.Time
	FinishedAt          time.Time
	Provider            string
	ClusterType         string
	ControlPlaneVersion string
	ClusterVersion      string
}

type Metadata struct {
	Date                time.Time `json:"date"`
	FinishedAt          time.Time `json:"finished_at"`
	Provider            string    `json:"provider"`
	ClusterType         string    `json:"cluster_type"`
	ControlPlaneVersion string    `json:"control_plane_version,omitempty"`
	ClusterVersion      string    `json:"cluster_version,omitempty"`
}

type Document struct {
	Schema             string                  `json:"schema"`
	RegressionDetected bool                    `json:"regression_detected"`
	Metadata           Metadata                `json:"metadata"`
	Sections           map[string]core.Section `json:"sections"`
}

func Build(sections map[string]core.Section, info RunInfo) Document {
	return Document{
		Schema: SchemaVersion,
		Metadata: Metadata{
			Date:                info.StartedAt.UTC(),
			FinishedAt:          info.FinishedAt.UTC(),
			Provider:            info.Provider,
			ClusterType:         info.ClusterType,
			ControlPlaneVersion: info.ControlPlaneVersion,
			ClusterVersion:      info.ClusterVersion,
		},
		Sections: sections,
	}
}

func (d Document) FileName() string {
	return "results-" + d.Metadata.Date.UTC().Format(fileTimestamp) + ".json"
}

func (d Document) Encode() ([]byte, error) {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
