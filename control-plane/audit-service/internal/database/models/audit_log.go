/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"gorm.io/gorm"
)

const (
	AuditLogsTable = "audit_logs"
)

var QueryApiParser = queryapi.NewParser(auditLogColumnMapping())

type AuditLog struct {
	gorm.Model

	LogId          uuid.UUID         `gorm:"not null,primaryKey" json:"log_id"`
	TenantId       int               `gorm:"not nul" json:"tenant_id"`
	ClusterId      uuid.UUID         `gorm:"not null" json:"cluster_id"`
	ClusterName    string            `gorm:"not null" json:"cluster_name"`
	RequestId      uuid.UUID         `gorm:"not null" json:"request_id"`
	Timestamp      time.Time         `gorm:"not null" json:"timestamp"`
	Subject        string            `gorm:"not null" json:"subject"`
	SubjectType    string            `gorm:"not null" json:"subject_type"`
	SourceIp       string            `gorm:"not null" json:"source_ip"`
	Action         string            `gorm:"not null" json:"action"`
	HttpMethod     string            `gorm:"not null" json:"http_method"`
	Url            string            `gorm:"not null" json:"url"`
	EntityType     string            `gorm:"not null" json:"entity_type"`
	EntityName     string            `gorm:"not null" json:"entity_name"`
	EntityId       string            `gorm:"not null" json:"entity_id"`
	HttpStatusCode int               `gorm:"not null" json:"http_status_code"`
	Result         string            `gorm:"not null" json:"result"`
	Origin         string            `gorm:"not null" json:"origin"`
	Metadata       dbmodels.JsonBMap `gorm:"not null" json:"metadata"`
}

func (al *AuditLog) TableName() string {
	return AuditLogsTable
}

func auditLogColumnMapping() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"created_at":       {Column: "created_at"},
		"origin":           {Column: "origin"},
		"subject":          {Column: "subject"},
		"subject_type":     {Column: "subject_type"},
		"source_ip":        {Column: "source_ip"},
		"action":           {Column: "action"},
		"result":           {Column: "result"},
		"url":              {Column: "url"},
		"entity_type":      {Column: "entity_type"},
		"entity_name":      {Column: "entity_name"},
		"cluster_name":     {Column: "cluster_name"},
		"http_method":      {Column: "http_method"},
		"http_status_code": {Column: "http_status_code"},
		"timestamp":        {Column: "timestamp"},
		"id":               {Column: "log_id::text"},
	}
}
