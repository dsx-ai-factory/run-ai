/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package business

import (
	"context"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type ClusterInstaller interface {
	GetInstallationInfo(ctx context.Context, clusterId string) (string, error)
}

func GenerateClusterUrl(ctx context.Context, clusterHostTemplate string, clusterId string, tenantId string) (string, error) {
	clusterUrl, err := utils.TemplateBaseUrl(clusterHostTemplate, getUniqueString(utils.ClusterUrlNamePrefix, clusterId, tenantId))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to template cluster url. err: %v", err)
		return "", err
	}
	return clusterUrl, nil
}

func GetClusterUrlDomainWithEnv(clusterUrl string, domainSuffix string) string {
	strNoPreix := strings.ReplaceAll(clusterUrl, "https://", "")
	strNoSuffix := strings.ReplaceAll(strNoPreix, domainSuffix, "")
	return strNoSuffix
}

func getUniqueString(prefix string, clusterId string, tenantId string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, string(clusterId[:4]), tenantId)
}
