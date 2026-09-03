/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator2

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/apiclient"
)

func newRunAIClientProvider(k8s K8sExecutor, version string) step.RunAIAPIClientProvider {
	return func(fqdn, caCertPath string) (step.RunAIAPIClient, error) {
		credentials, err := apiclient.LoadControlPlaneCredentials(context.Background(), k8s, step.CPNamespace, step.CPReleaseName, version)
		if err != nil {
			return nil, err
		}
		clients, err := apiclient.GetClientsInstanceWithCredentials(fqdn, caCertPath, credentials)
		if err != nil {
			return nil, err
		}
		return clients.ClusterServiceClient, nil
	}
}
