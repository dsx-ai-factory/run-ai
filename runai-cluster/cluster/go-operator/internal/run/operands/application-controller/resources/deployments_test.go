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
