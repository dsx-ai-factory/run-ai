/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"fmt"

	auditconfig "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/config"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/generated"
)

func validate(a *openapi.AuditLog) error {
	if a.GetTenantId() == "" {
		return fmt.Errorf("missing tenantId")
	}

	if a.GetAction() == "" {
		return fmt.Errorf("missing action")
	}

	if len(a.GetMetadata()) > auditconfig.MaxMetadataProperties {
		return fmt.Errorf("metadata cannot have more than %d key value pairs", auditconfig.MaxMetadataProperties)
	}

	return nil
}
