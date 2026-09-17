/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck_test

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputcheck"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/stretchr/testify/assert"
)

func externalControlPlaneFindings(s *spec.Spec) []inputcheck.Finding {
	var got []inputcheck.Finding
	for _, f := range inputcheck.Validate(s).Findings {
		if f.Check == "external-control-plane" {
			got = append(got, f)
		}
	}
	return got
}

func TestExternalControlPlaneCheck_AcceptsEitherIdentitySource(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cluster *spec.Cluster
	}{
		{
			name: "supplied identity",
			cluster: &spec.Cluster{
				Name: "c", Domain: "cluster.example.com", ControlPlaneURL: "https://cp.example.com",
				UID: "uid", BootstrapToken: "tok",
			},
		},
		{
			name: "control-plane credentials",
			cluster: &spec.Cluster{
				Name: "c", Domain: "cluster.example.com", ControlPlaneURL: "https://cp.example.com",
				ControlPlaneCredentials: &spec.ControlPlaneCredentials{Token: "bearer"},
			},
		},
		{
			name:    "same-cluster install",
			cluster: &spec.Cluster{Name: "c"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &spec.Spec{InstallMode: spec.InstallModeClusterOnly, Cluster: tc.cluster}
			assert.Empty(t, externalControlPlaneFindings(s))
		})
	}
}

func TestExternalControlPlaneCheck_ReportsEveryProblem(t *testing.T) {
	s := &spec.Spec{
		InstallMode: spec.InstallModeClusterOnly,
		Cluster: &spec.Cluster{
			Name:                    "c",
			ControlPlaneURL:         "http://cp.example.com",
			UID:                     "uid",
			ControlPlaneCredentials: &spec.ControlPlaneCredentials{Username: "admin@example.com"},
		},
	}

	var fields []string
	for _, f := range externalControlPlaneFindings(s) {
		fields = append(fields, f.Field)
	}
	assert.ElementsMatch(t, []string{
		"cluster.controlPlaneUrl",
		"cluster.domain",
		"cluster.controlPlaneCredentials",
		"cluster.controlPlaneCredentials.password",
	}, fields)
}
