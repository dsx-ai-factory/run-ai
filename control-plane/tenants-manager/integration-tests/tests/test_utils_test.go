/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"

	"github.com/google/uuid"
)

func GenerateRandomString(prefix string) string {
	return prefix + uuid.New().String()
}

func GenerateRandomTenantName() string {
	return GenerateRandomString("test-tenant-")
}

func GenerateRandomEmail(tenantName string) string {
	return fmt.Sprintf("%s@test.run.ai", tenantName)
}

func StrPtr(s string) *string { return &s }
