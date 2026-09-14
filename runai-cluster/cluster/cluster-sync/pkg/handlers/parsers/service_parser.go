/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	swagger "github.com/dsx-ai-factory/run-ai/runai-cluster/common/clients/control-plane/generated"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func ParseServiceToServerScheme(service *knv1.Service) swagger.Service {
	// URL is only created after routes are create
	var host = ""
	if service.Status.URL != nil {
		host = service.Status.URL.Host
	}
	return swagger.Service{
		ClusterUuid: config.Config().ClusterUuid,
		Url:         &host,
		Id:          string(service.UID),
		Name:        service.Name,
		Namespace:   service.Namespace,
		CreatedAt:   service.CreationTimestamp.Time,
	}
}
