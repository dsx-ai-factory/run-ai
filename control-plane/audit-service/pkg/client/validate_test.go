/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"fmt"
	"testing"

	auditconfig "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/config"
	"github.com/stretchr/testify/require"
)

func TestValidate_Success(t *testing.T) {
	al := NewAuditLog().WithTenantId("1").WithAction("login")
	err := validate(&al.Log)
	require.Nil(t, err)
}

func TestValidate_Failure_WithoutTenantId(t *testing.T) {
	al := NewAuditLog()
	err := validate(&al.Log)
	require.EqualError(t, err, "missing tenantId")
}

func TestValidate_Failure_WithoutAction(t *testing.T) {
	al := NewAuditLog().WithTenantId("1")
	err := validate(&al.Log)
	require.EqualError(t, err, "missing action")
}

func TestValidate_Failure_TooLargeMetadata(t *testing.T) {
	al := NewAuditLog().WithTenantId("1").WithAction("login")
	for i := range auditconfig.MaxMetadataProperties + 1 {
		key := fmt.Sprintf(`k-%d`, i)
		value := fmt.Sprintf(`v-%d`, i)
		al.WithMetadata(key, value)
	}
	err := validate(&al.Log)
	require.EqualError(t, err, fmt.Sprintf("metadata cannot have more than %d key value pairs", auditconfig.MaxMetadataProperties))
}
