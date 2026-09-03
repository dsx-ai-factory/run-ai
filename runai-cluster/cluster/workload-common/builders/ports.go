/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"

	commonutils "github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

// GetPorts builds the sets of port to be included in the "ports:" section of the container. Note that
// specifying container ports is optional, as the complete set of exposed ports it determined by the list
// of services. But it is still nice to list them.
func GetPorts(_ context.Context, networkingFields *field_groups.NetworkingFields, allowServingPort bool) ([]corev1.ContainerPort, error) {
	var result []corev1.ContainerPort

	ports := networkingFields.Ports.Deref()

	//   keep a set of all the container ports we already added to prevent duplicate ports for the same container
	allContainerPorts := datastructs.NewSet[int32]()

	for _, nick := range flds.SortedPortsMapKeys(ports) {
		port := ports[nick].Value
		if port == nil {
			continue
		}

		portValue := int32(ptr.Deref(port.ContainerPort, 0))
		containerPort := corev1.ContainerPort{
			Name:          fmt.Sprintf("port-%d", portValue),
			ContainerPort: portValue,
			Protocol:      corev1.ProtocolTCP,
		}

		serviceType := ptr.Deref(ports[nick].Value.ServiceType, networkingFields.ServiceType.Deref())

		if serviceType == constants.ServiceTypeServingPort && !allowServingPort {
			return nil, fmt.Errorf("serving port is not allowed for this workload type")
		}

		if serviceType == constants.ServiceTypeNodePort && ports[nick].Value.HostPort != nil {
			if *ports[nick].Value.HostPort < minHostPort || *ports[nick].Value.HostPort > maxHostPort {
				return nil, fmt.Errorf("%s: NodePort must be between %d and %d", nick, minHostPort, maxHostPort)
			}
		}

		if allContainerPorts.Has(containerPort.ContainerPort) {
			continue // no need to add the same container twice
		}

		result = append(result, containerPort)
		allContainerPorts.Add(containerPort.ContainerPort)
	}

	//
	//   add the port of each exposed URL to the list of container exposed ports.
	//   note that we do not add the port of exposed URLs for inference workloads
	//
	exposedUrls := networkingFields.ExposedUrls.Deref()

	for _, nick := range commonutils.SortedKeys(exposedUrls) {
		exposedUrl := exposedUrls[nick].Value
		if exposedUrl == nil {
			continue
		}

		portValue := ptr.Deref(exposedUrl.ContainerPort, DefaultIngressHttpPort)
		containerPort := corev1.ContainerPort{
			ContainerPort: portValue,
			Name:          fmt.Sprintf("port-%d", portValue),
			Protocol:      corev1.ProtocolTCP,
		}

		if allContainerPorts.Has(containerPort.ContainerPort) {
			continue // no need to add the same container twice
		}

		result = append(result, containerPort)
		allContainerPorts.Add(containerPort.ContainerPort)
	}

	return result, nil
}
