/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"
	"strings"

	managerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

const (
	legacyWorkspacePrefix   = "iw"
	legacyTrainingPrefix    = "tw"
	legacyDistributedPrefix = "dw"
)

// Legacy connections refer to existing cluster resources. Preserve their original
// global port indexes and service-type ordering rather than applying WLR naming.
func legacyPortResourceNames(initials, workloadName string, ports *field_types.PortsField) map[string]ServiceAndPortName {
	result := make(map[string]ServiceAndPortName)
	nicksByType := make(map[managerApi.ServiceType][]string)
	portIndex := 0
	for _, nick := range field_types.SortedPortsMapKeys(ports.Deref()) {
		item := ports.Deref()[nick].Value
		if item == nil {
			continue
		}
		serviceType, err := portItemServiceType(item.ServiceType)
		if err != nil {
			continue
		}
		nicksByType[serviceType] = append(nicksByType[serviceType], nick)
		result[nick] = ServiceAndPortName{
			PortName: fmt.Sprintf("port-%s-%d-%d", workloadName, ptr.Deref(item.ContainerPort, 0), portIndex),
		}
		portIndex++
	}

	serviceIndex := 0
	for _, serviceType := range []managerApi.ServiceType{
		managerApi.SERVICETYPE_CLUSTER_IP,
		managerApi.SERVICETYPE_LOAD_BALANCER,
		managerApi.SERVICETYPE_NODE_PORT,
	} {
		nicks := nicksByType[serviceType]
		if len(nicks) == 0 {
			continue
		}
		serviceName := fmt.Sprintf("%s-%s-%d-%s", initials, workloadName, serviceIndex, strings.ToLower(string(serviceType)))
		for _, nick := range nicks {
			names := result[nick]
			names.ServiceName = serviceName
			result[nick] = names
		}
		serviceIndex++
	}
	return result
}
