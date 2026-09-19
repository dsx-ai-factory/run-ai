/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package preflight

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func credentialedSpec() *spec.Spec {
	return &spec.Spec{
		InstallMode: spec.InstallModeClusterOnly,
		TLS:         &spec.TLS{CAChainPath: "/ca.pem"},
		Cluster: &spec.Cluster{
			Name:                    "c1",
			Domain:                  "cluster.example.com",
			ControlPlaneURL:         "https://cp.example.com",
			ControlPlaneCredentials: &spec.ControlPlaneCredentials{Token: "bearer"},
		},
	}
}

func credentialedDeps() orchestrator2.Deps {
	return orchestrator2.Deps{RunAIClientWithCredentials: func(string, string, step.ControlPlaneCredentials) (step.RunAIAPIClient, error) { return nil, nil }}
}

func TestControlPlaneAccessBuilder_AppliesToCredentialedInstall(t *testing.T) {
	st := controlPlaneAccessBuilder(credentialedSpec(), credentialedDeps())
	require.NotNil(t, st)
	assert.Equal(t, "preflight-control-plane-access", st.Name())

	acts, err := st.BuildActions(nil)
	require.NoError(t, err)
	require.Len(t, acts, 1)
	check := acts[0].(actions.CheckControlPlaneAccess)
	assert.Equal(t, "c1", check.ClusterName)
	assert.Equal(t, "https://cp.example.com", check.ControlPlane.URL)
	assert.Equal(t, "/ca.pem", check.ControlPlane.CACertPath, "the control plane is trusted via the cluster's CA chain")
	require.NotNil(t, check.ControlPlane.Credentials)
	assert.Equal(t, "bearer", check.ControlPlane.Credentials.Token)
}

func TestControlPlaneAccessBuilder_SkipsOtherInstalls(t *testing.T) {
	supplied := credentialedSpec()
	supplied.Cluster.ControlPlaneCredentials = nil
	supplied.Cluster.UID = "uid"
	supplied.Cluster.BootstrapToken = "tok"
	assert.Nil(t, controlPlaneAccessBuilder(supplied, credentialedDeps()), "supplied identity never contacts the control plane")

	sameCluster := &spec.Spec{ControlPlane: &spec.ControlPlane{Domain: "runai.example.com"}, Cluster: &spec.Cluster{Name: "c1"}}
	assert.Nil(t, controlPlaneAccessBuilder(sameCluster, credentialedDeps()))

	assert.Nil(t, controlPlaneAccessBuilder(credentialedSpec(), orchestrator2.Deps{}), "no provider, no check")
}
