/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/generated/pkg/api"
	"github.com/gocarina/gocsv"
)

type SerializedAuditLog struct {
	Id             string    `json:"id" csv:"id"`
	Timestamp      time.Time `json:"timestamp" csv:"timestamp"`
	TenantId       string    `json:"tenant_id" csv:"tenant_id"`
	Subject        string    `json:"subject" csv:"subject"`
	SubjectType    string    `json:"subject_type" csv:"subject_type"`
	SourceIp       string    `json:"source_ip" csv:"source_ip"`
	Action         string    `json:"action" csv:"action"`
	HttpMethod     string    `json:"http_method" csv:"http_method"`
	Url            string    `json:"url" csv:"url"`
	EntityType     string    `json:"entity_type" csv:"entity_type"`
	EntityName     string    `json:"entity_name" csv:"entity_name"`
	EntityId       string    `json:"entity_id" csv:"entity_id"`
	Result         string    `json:"result" csv:"result"`
	HttpStatusCode int32     `json:"http_status_code" csv:"http_status_code"`
	ClusterName    string    `json:"cluster_name" csv:"cluster_name"`
	ClusterId      string    `json:"cluster_id" csv:"cluster_id"`
	RequestId      string    `json:"request_id" csv:"request_id"`
	Metadata       string    `json:"metadata" csv:"metadata"`
}

func writeJSONFile(apiLogs []api.AuditLog, file *os.File) error {
	data, err := json.MarshalIndent(apiLogs, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

func writeCSVFile(apiLogs []api.AuditLog, file *os.File) error {
	serializedAuditLogs := serializeAuditLogs(apiLogs, transformAuditLogForCSV)
	return gocsv.Marshal(serializedAuditLogs, file)
}

func newSerializedAuditLog(log *api.AuditLog, formatMetadata func(map[string]string) string) SerializedAuditLog {
	return SerializedAuditLog{
		Id:             log.Id,
		Timestamp:      log.Timestamp,
		TenantId:       log.TenantId,
		Subject:        log.Subject,
		SubjectType:    log.SubjectType,
		SourceIp:       log.SourceIp,
		Action:         log.Action,
		HttpMethod:     log.HttpMethod,
		Url:            log.Url,
		EntityType:     log.EntityType,
		EntityName:     log.EntityName,
		EntityId:       log.EntityId,
		Result:         log.Result,
		HttpStatusCode: log.HttpStatusCode,
		ClusterName:    log.ClusterName,
		ClusterId:      log.ClusterId,
		RequestId:      log.RequestId,
		Metadata:       formatMetadata(log.Metadata),
	}
}

func transformAuditLogForCSV(log *api.AuditLog) SerializedAuditLog {
	return newSerializedAuditLog(log, formatMetadataForCSV)
}

func formatMetadataForCSV(metadata map[string]string) string {
	if len(metadata) == 0 {
		return ""
	}

	sortedKeys := sortKeys(metadata)

	parts := []string{}
	for _, k := range sortedKeys {
		parts = append(parts, escapeKeyValue(k, metadata[k]))
	}

	return strings.Join(parts, ",")
}

func escapeKeyValue(key, value string) string {
	escapedKey := strings.ReplaceAll(key, `"`, `""`)
	escapedValue := strings.ReplaceAll(value, `"`, `""`)
	return fmt.Sprintf(`%s=%s`, escapedKey, escapedValue)
}

func sortKeys(metadata map[string]string) []string {
	keys := make([]string, 0, len(metadata))
	for k := range metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func serializeAuditLogs(apiLogs []api.AuditLog, transform func(*api.AuditLog) SerializedAuditLog) []SerializedAuditLog {
	result := make([]SerializedAuditLog, len(apiLogs))
	for i, log := range apiLogs {
		result[i] = transform(&log)
	}
	return result
}
