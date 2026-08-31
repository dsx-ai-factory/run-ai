/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kubernetes

import (
	"fmt"

	"k8s.io/client-go/discovery"
)

func GetKubernetesClusterVersion(dClient discovery.DiscoveryInterface) (string, error) {
	kversion, err := dClient.ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get server version %s", err.Error())
	}
	return kversion.String(), nil
}
