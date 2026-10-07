/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	"github.com/stretchr/testify/require"
)

func TestDeploymentInjectsADSEncryptionKey(t *testing.T) {
	env := handleEnvVars(runutils.GetDefaultRunaiConfig())
	for _, variable := range env {
		if variable.Name != constants.RunaiAdsEncryptionKeyEnvVar {
			continue
		}
		require.NotNil(t, variable.ValueFrom)
		require.NotNil(t, variable.ValueFrom.SecretKeyRef)
		require.Equal(t, constants.RunaiAdsSecretName, variable.ValueFrom.SecretKeyRef.Name)
		require.Equal(t, constants.RunaiAdsEncryptionKeyName, variable.ValueFrom.SecretKeyRef.Key)
		return
	}
	t.Fatalf("%s is not configured", constants.RunaiAdsEncryptionKeyEnvVar)
}

func TestApplicationAllowedServiceAccounts(t *testing.T) {
	for _, tc := range []struct {
		name                string
		agentServiceAccount string
		want                string
	}{
		{name: "default agent", want: "system:serviceaccount:custom-runai:application-controller,system:serviceaccount:custom-runai:runai-agent"},
		{name: "NMC agent", agentServiceAccount: "nmc-agent", want: "system:serviceaccount:custom-runai:application-controller,system:serviceaccount:custom-runai:nmc-agent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := runutils.GetDefaultRunaiConfig()
			config.Namespace = "custom-runai"
			config.Spec.Internal.Global.NMC.ServiceAccountName = tc.agentServiceAccount
			for _, variable := range handleEnvVars(config) {
				if variable.Name == "ALLOWED_SERVICE_ACCOUNTS" {
					require.Equal(t, tc.want, variable.Value)
					return
				}
			}
			t.Fatal("application writers must be explicitly allowed")
		})
	}
}
