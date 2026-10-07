/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package connections

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	managerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
)

const (
	legacyWorkspacePrefix   = "iw"
	legacyTrainingPrefix    = "tw"
	legacyDistributedPrefix = "dw"
)

var workloadTypeToInitials = map[workloadsApi.WorkloadType]string{
	workloadsApi.WORKLOADTYPE_WORKSPACE:   legacyWorkspacePrefix,
	workloadsApi.WORKLOADTYPE_TRAINING:    legacyTrainingPrefix,
	workloadsApi.WORKLOADTYPE_DISTRIBUTED: legacyDistributedPrefix,
	workloadsApi.WORKLOADTYPE_INFERENCE:   composer.NetworkingResourcePrefixInference,
}

// PortResourceNames preserves native service ordering and global port indexes
// so connection metadata still identifies existing cluster resources.
func PortResourceNames(workloadType workloadsApi.WorkloadType, workloadName string, ports *field_types.PortsField) map[string]composer.ServiceAndPortName {
	result := make(map[string]composer.ServiceAndPortName)
	initials, found := workloadTypeToInitials[workloadType]
	if !found || ports == nil {
		return result
	}
	nicksByType := make(map[managerApi.ServiceType][]string)
	portIndex := 0
	for _, nick := range field_types.SortedPortsMapKeys(ports.Deref()) {
		item := ports.Deref()[nick].Value
		if item == nil {
			continue
		}
		var portServiceType *workloadsApi.PortServiceType
		if item.ServiceType != nil {
			portServiceType = ptr.To(workloadsApi.PortServiceType(*item.ServiceType))
		}
		serviceType, err := composer.PortServiceTypeToServiceType(portServiceType)
		if err != nil {
			continue
		}
		nicksByType[serviceType] = append(nicksByType[serviceType], nick)
		result[nick] = composer.ServiceAndPortName{
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
