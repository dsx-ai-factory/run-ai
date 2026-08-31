/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apis

import "context"

type HardTenantDeletionAPIServicer interface {
	HardDelete(ctx context.Context, tenantId string) error
}
