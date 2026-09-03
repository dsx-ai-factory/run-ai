/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dns

import "fmt"

// GetServiceDNSNames returns a list of DNS names for a service.
// It returns the service name, the service name with the namespace, and the service name with the namespace and cluster domain.
func GetServiceDNSNames(serviceName, namespace string, optionalClusterDomain *string) []string {
	clusterDomain := "cluster.local"
	if optionalClusterDomain != nil {
		clusterDomain = *optionalClusterDomain
	}

	return []string{
		serviceName,
		fmt.Sprintf("%s.%s", serviceName, namespace),
		fmt.Sprintf("%s.%s.svc", serviceName, namespace),
		fmt.Sprintf("%s.%s.svc.%s", serviceName, namespace, clusterDomain),
	}
}
