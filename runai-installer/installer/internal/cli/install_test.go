/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	helmcli "helm.sh/helm/v3/pkg/cli"
)

func TestSetHelmKubeconfig(t *testing.T) {
	t.Setenv("KUBECONFIG", "/tmp/control-plane")
	kubeconfig := filepath.Join(t.TempDir(), "workload")
	require.NoError(t, os.WriteFile(kubeconfig, []byte(`
apiVersion: v1
kind: Config
current-context: workload
clusters:
  - name: workload
    cluster:
      server: https://workload.example.com
contexts:
  - name: workload
    context:
      cluster: workload
      user: workload
users:
  - name: workload
    user: {}
`), 0o600))

	require.NoError(t, setHelmKubeconfig(kubeconfig))

	config, err := helmcli.New().RESTClientGetter().ToRESTConfig()
	require.NoError(t, err)
	assert.Equal(t, "https://workload.example.com", config.Host)
}
