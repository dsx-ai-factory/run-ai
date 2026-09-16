/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"errors"
	"testing"

	runaicluster "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// credentialedCall records one call to the credentialed provider.
type credentialedCall struct {
	url, caCertPath string
	creds           step.ControlPlaneCredentials
}

func credentialedProviderFor(api step.RunAIAPIClient, calls *[]credentialedCall) step.RunAIAPICredentialedClientProvider {
	return func(url, caCertPath string, creds step.ControlPlaneCredentials) (step.RunAIAPIClient, error) {
		*calls = append(*calls, credentialedCall{url: url, caCertPath: caCertPath, creds: creds})
		return api, nil
	}
}

func externalRef(creds *step.ControlPlaneCredentials) ControlPlaneRef {
	return ControlPlaneRef{
		URL:         "https://cp.example.com",
		Domain:      "cluster.example.com",
		CACertPath:  "/ca.pem",
		Credentials: creds,
	}
}

var userCreds = &step.ControlPlaneCredentials{Username: "admin@example.com", Password: "pw"}

func TestControlPlaneRef_Endpoint(t *testing.T) {
	assert.Equal(t, "https://runai.example.com", ControlPlaneRef{Domain: "runai.example.com"}.Endpoint())
	assert.Equal(t, "https://cp.example.com", externalRef(nil).Endpoint())
}

func TestControlPlaneRef_Client_PicksProviderByCredentials(t *testing.T) {
	local := &fakeAPI{}
	external := &fakeAPI{}
	var calls []credentialedCall
	localCalls := 0
	deps := orchestrator2.Deps{
		RunAIClientProvider: func(url, ca string) (step.RunAIAPIClient, error) {
			localCalls++
			assert.Equal(t, "https://runai.example.com", url)
			return local, nil
		},
		RunAIClientWithCredentials: credentialedProviderFor(external, &calls),
	}

	got, err := ControlPlaneRef{Domain: "runai.example.com"}.Client(deps)
	require.NoError(t, err)
	assert.Same(t, local, got)
	assert.Equal(t, 1, localCalls)

	got, err = externalRef(userCreds).Client(deps)
	require.NoError(t, err)
	assert.Same(t, external, got)
	require.Len(t, calls, 1)
	assert.Equal(t, "https://cp.example.com", calls[0].url)
	assert.Equal(t, "/ca.pem", calls[0].caCertPath)
	assert.Equal(t, *userCreds, calls[0].creds)
}

func TestControlPlaneRef_Client_MissingProviders(t *testing.T) {
	_, err := externalRef(userCreds).Client(orchestrator2.Deps{})
	require.ErrorContains(t, err, "credentials")

	_, err = ControlPlaneRef{Domain: "d"}.Client(orchestrator2.Deps{})
	require.ErrorContains(t, err, "provider")
}

func TestEnsureCluster_ExternalControlPlane_CreatesWithClusterDomain(t *testing.T) {
	api := &fakeAPI{}
	var calls []credentialedCall
	deps := orchestrator2.Deps{RunAIClientWithCredentials: credentialedProviderFor(api, &calls)}
	act := EnsureCluster{ClusterName: "c1", ControlPlane: externalRef(userCreds)}

	res, err := act.Execute(context.Background(), deps)
	require.NoError(t, err)

	require.Len(t, api.created, 1)
	assert.Equal(t, "c1", api.created[0].Name)
	require.NotNil(t, api.created[0].Domain)
	assert.Equal(t, "https://cluster.example.com", *api.created[0].Domain, "the control plane records the cluster's own address, not its own")
	assert.Equal(t, true, res.Data["created"])
	assert.Equal(t, "new-uuid", res.Data["clusterUid"])
	require.Len(t, calls, 1, "the external control plane is reached with the supplied credentials")
	assert.Equal(t, "https://cp.example.com", calls[0].url)
	assert.Contains(t, act.Describe(), "https://cp.example.com")
}

func TestEnsureCluster_ExternalControlPlane_ReusesExisting(t *testing.T) {
	api := &fakeAPI{clusters: []runaicluster.DisplayedCluster{{Name: "c1", Uuid: "u1"}}}
	var calls []credentialedCall
	deps := orchestrator2.Deps{RunAIClientWithCredentials: credentialedProviderFor(api, &calls)}
	act := EnsureCluster{ClusterName: "c1", ControlPlane: externalRef(userCreds)}

	registered, err := act.Detect(context.Background(), deps)
	require.NoError(t, err)
	assert.True(t, registered)

	res, err := act.Execute(context.Background(), deps)
	require.NoError(t, err)
	assert.Empty(t, api.created)
	assert.Equal(t, false, res.Data["created"])
	assert.Equal(t, "u1", res.Data["clusterUid"])
}

func TestInstallClusterChart_ResolvesFromExternalControlPlane(t *testing.T) {
	api := &fakeAPI{clusters: []runaicluster.DisplayedCluster{{Name: "c1", Uuid: "u1"}}, clientSecret: "sekret"}
	var calls []credentialedCall
	fakeHelm := &harness.FakeHelmInstaller{}
	deps := orchestrator2.Deps{RunAIClientWithCredentials: credentialedProviderFor(api, &calls), HelmInstaller: fakeHelm}
	act := InstallClusterChart{
		ClusterName:     "c1",
		Domain:          "cluster.example.com",
		ControlPlaneURL: "https://cp.example.com",
		Credentials:     &step.ControlPlaneCredentials{Token: "bearer"},
		CACertPath:      "/ca.pem",
		Repo:            "https://repo",
		Version:         "2.26.0",
		ReleaseName:     "runai-cluster",
		ChartName:       "runai-cluster",
		Namespace:       "runai",
		HostRouting:     true,
	}

	assert.False(t, act.UsesSuppliedCredentials())
	execOK(t, act, deps)

	calls2 := fakeHelm.CallsByRelease("runai-cluster")
	require.Len(t, calls2, 1)
	assert.Subset(t, calls2[0].Values, []string{
		"controlPlane.url=https://cp.example.com",
		"controlPlane.clientSecret=sekret",
		"cluster.uid=u1",
		"cluster.url=cluster.example.com",
		"global.customCA.enabled=true",
	})
	require.Len(t, calls, 1)
	assert.Equal(t, "bearer", calls[0].creds.Token)
	assert.Equal(t, "/ca.pem", calls[0].caCertPath)
}

func TestInstallClusterChartLocal_ResolvesFromExternalControlPlane(t *testing.T) {
	api := &fakeAPI{clusters: []runaicluster.DisplayedCluster{{Name: "c1", Uuid: "u1"}}, clientSecret: "sekret"}
	var calls []credentialedCall
	fakeHelm := &harness.FakeHelmInstaller{}
	deps := orchestrator2.Deps{RunAIClientWithCredentials: credentialedProviderFor(api, &calls), HelmInstaller: fakeHelm}
	act := InstallClusterChartLocal{
		ClusterName:     "c1",
		Domain:          "cluster.example.com",
		ControlPlaneURL: "https://cp.example.com",
		Credentials:     userCreds,
		ChartPath:       "/pkg/charts/runai-cluster-2.20.0.tgz",
		Version:         "2.26.0",
		ReleaseName:     "runai-cluster",
		Namespace:       "runai",
	}

	execOK(t, act, deps)

	helmCalls := fakeHelm.CallsByRelease("runai-cluster")
	require.Len(t, helmCalls, 1)
	assert.Subset(t, helmCalls[0].Values, []string{
		"controlPlane.url=https://cp.example.com",
		"controlPlane.clientSecret=sekret",
		"cluster.uid=u1",
		"cluster.url=cluster.example.com",
	})
	require.Len(t, calls, 1)
	assert.Equal(t, *userCreds, calls[0].creds)
}

func TestInstallClusterChart_ExternalControlPlane_ClusterNotFound(t *testing.T) {
	var calls []credentialedCall
	deps := orchestrator2.Deps{RunAIClientWithCredentials: credentialedProviderFor(&fakeAPI{}, &calls), HelmInstaller: &harness.FakeHelmInstaller{}}
	act := InstallClusterChart{ClusterName: "missing", Domain: "d", ControlPlaneURL: "https://cp.example.com", Credentials: userCreds, ReleaseName: "runai-cluster"}

	_, err := act.Execute(context.Background(), deps)

	require.ErrorContains(t, err, "not found")
}

func TestCheckControlPlaneAccess_Pass(t *testing.T) {
	api := &fakeAPI{clusters: []runaicluster.DisplayedCluster{{Name: "c1", Uuid: "u1"}}}
	var calls []credentialedCall
	deps := orchestrator2.Deps{RunAIClientWithCredentials: credentialedProviderFor(api, &calls)}

	res, err := CheckControlPlaneAccess{ClusterName: "c1", ControlPlane: externalRef(userCreds)}.Execute(context.Background(), deps)
	require.NoError(t, err)
	assert.Equal(t, orchestrator2.StatusOK, res.Status)
	assert.Equal(t, true, res.Data["clusterRegistered"])
	assert.Equal(t, "u1", res.Data["clusterUid"])
	assert.Equal(t, "https://cp.example.com", res.Data["controlPlaneUrl"])

	res, err = CheckControlPlaneAccess{ClusterName: "other", ControlPlane: externalRef(userCreds)}.Execute(context.Background(), deps)
	require.NoError(t, err)
	assert.Equal(t, orchestrator2.StatusOK, res.Status)
	assert.Equal(t, false, res.Data["clusterRegistered"])
}

func TestCheckControlPlaneAccess_Unreachable(t *testing.T) {
	deps := orchestrator2.Deps{RunAIClientWithCredentials: func(string, string, step.ControlPlaneCredentials) (step.RunAIAPIClient, error) {
		return nil, errors.New("connect: connection refused")
	}}

	res, err := CheckControlPlaneAccess{ClusterName: "c1", ControlPlane: externalRef(userCreds)}.Execute(context.Background(), deps)
	require.NoError(t, err, "a failed check is a verdict, not an error")
	assert.Equal(t, orchestrator2.StatusFail, res.Status)
	assert.Contains(t, res.Reason, "not reachable")
	assert.Contains(t, res.Reason, "connection refused")
	assert.Contains(t, res.Remediation, "cluster.controlPlaneUrl")
}

func TestCheckControlPlaneAccess_Unauthorized(t *testing.T) {
	var calls []credentialedCall
	deps := orchestrator2.Deps{RunAIClientWithCredentials: credentialedProviderFor(&fakeAPI{getErr: errors.New("401 Unauthorized")}, &calls)}

	res, err := CheckControlPlaneAccess{ClusterName: "c1", ControlPlane: externalRef(userCreds)}.Execute(context.Background(), deps)
	require.NoError(t, err)
	assert.Equal(t, orchestrator2.StatusFail, res.Status)
	assert.Contains(t, res.Reason, "rejected the supplied credentials")
	assert.Contains(t, res.Reason, "401 Unauthorized")
	assert.Contains(t, res.Remediation, "cluster.controlPlaneCredentials")
}
