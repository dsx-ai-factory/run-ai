/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/apiclient"
)

type RunAIClientsProvider func(fqdn, caCertPath string) (*apiclient.Clients, error)

func NewRunAIClientsProvider(getter apiclient.ObjectGetter, version string) RunAIClientsProvider {
	return func(fqdn, caCertPath string) (*apiclient.Clients, error) {
		credentials, err := apiclient.LoadControlPlaneCredentials(context.Background(), getter, CPNamespace, CPReleaseName, version)
		if err != nil {
			return nil, err
		}
		return apiclient.GetClientsInstanceWithCredentials(fqdn, caCertPath, credentials)
	}
}

func NewRunAIAPIClientProvider(getter apiclient.ObjectGetter, version string) RunAIAPIClientProvider {
	clientsProvider := NewRunAIClientsProvider(getter, version)
	return func(fqdn, caCertPath string) (RunAIAPIClient, error) {
		clients, err := clientsProvider(fqdn, caCertPath)
		if err != nil {
			return nil, err
		}
		return clients.ClusterServiceClient, nil
	}
}
