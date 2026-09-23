/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	trackerclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	resourcev1 "k8s.io/api/resource/v1"
)

// ParseDeviceClassToServerScheme converts a cluster DeviceClass into the tracker payload.
// DeviceClasses are cluster-scoped and tracked by name only.
func ParseDeviceClassToServerScheme(deviceClass *resourcev1.DeviceClass) trackerclient.DeviceClass {
	clusterId := config.Config().ClusterUuid
	return trackerclient.DeviceClass{
		ClusterId:       &clusterId,
		DeviceClassName: deviceClass.Name,
	}
}
