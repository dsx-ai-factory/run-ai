/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/audit-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

var SuccessStatusOptions = map[string]bool{
	"success":   true,
	"succeeded": true,
	"succeed":   true,
}

func toApi(logs []*models.AuditLog) []api.AuditLog {
	apiLogs := []api.AuditLog{}
	for _, log := range logs {
		apiLogs = append(apiLogs, getApiLog(log))
	}
	return apiLogs
}

func getApiLog(log *models.AuditLog) api.AuditLog {
	return api.AuditLog{
		Id:             log.LogId.String(),
		Timestamp:      log.Timestamp,
		TenantId:       fmt.Sprintf("%d", log.TenantId),
		Subject:        log.Subject,
		SubjectType:    log.SubjectType,
		SourceIp:       log.SourceIp,
		Action:         log.Action,
		HttpMethod:     log.HttpMethod,
		Url:            log.Url,
		EntityType:     log.EntityType,
		EntityId:       log.EntityId,
		EntityName:     log.EntityName,
		Result:         parseResult(log.Result),
		HttpStatusCode: int32(log.HttpStatusCode),
		ClusterName:    log.ClusterName,
		ClusterId:      log.ClusterId.String(),
		RequestId:      log.RequestId.String(),
		Metadata:       convertToStringMap(log.Metadata),
	}
}

func convertToStringMap(source map[string]interface{}) map[string]string {
	target := make(map[string]string)
	for key, value := range source {
		strValue, ok := value.(string)
		if ok {
			target[key] = strValue
		} else {
			logging.Warnf("Failed to parse value in metadata, value is not a string. Skipping key %s: \n", key)
		}
	}
	return target
}

func parseTenantId(ctx context.Context) (int, error) {
	tenantId, ok := ctx.Value(runaictx.CfTenantId).(string)
	if !ok {
		return 0, fmt.Errorf("no tenant id provided")
	}

	tenantIdInt, parseErr := strconv.Atoi(tenantId)
	if parseErr != nil {
		return 0, fmt.Errorf("invalid tenant id")
	}

	return tenantIdInt, nil
}

func parseResult(result string) string {
	if SuccessStatusOptions[strings.ToLower(result)] {
		return "Succeeded"
	}
	return "Failed"
}
