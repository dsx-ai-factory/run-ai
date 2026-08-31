/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"testing"
	"time"

	auditconfig "github.com/run-ai/runai/control-plane/audit-service/pkg/client/config"
	"github.com/run-ai/runai/control-plane/audit-service/pkg/client/write"
	"github.com/stretchr/testify/assert"
)

func TestInit_EnabledFalse(t *testing.T) {
	err := Init(&auditconfig.Config{
		Enabled: false,
	})
	assert.Nil(t, err)
	assert.IsType(t, &write.EmptyWriter{}, singleInstance)
}

func TestAuditFromAudit_CopiesAllFields(t *testing.T) {
	al := NewAuditLog()
	al = al.
		WithTenantId("tenant-123").
		WithSubject("user-abc").
		WithSubjectType("User").
		WithSourceIp("10.0.0.1").
		WithAction("Create").
		WithHttpMethod("POST").
		WithUrl("/api/v1/resource").
		WithEntityType("resource").
		WithEntityName("my-resource").
		WithEntityId("entity-777").
		WithClusterName("cluster-a").
		WithClusterId("cluster-id-1").
		WithRequestId("req-999").
		WithMetadata("k1", "v1").WithMetadata("k2", "v2").
		WithStatusCode(200)

	ts := time.Date(2025, 11, 24, 12, 0, 0, 0, time.UTC)
	al = al.WithTimestamp(ts)

	got := AuditFromAudit(al)

	assert.NotSame(t, &al.Log, &got.Log, "inner Log structs should not share the same address")
	assert.Equal(t, &al.Log, &got.Log, "audit logs does not have identical values")
}
