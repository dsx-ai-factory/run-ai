/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notification_types

import (
	"context"
)

type NotificationType struct {
	Origin      string
	Category    string
	Type        string
	Description string
	Template    string
}

type NotificationKindsClienter interface {
	Get(ctx context.Context) ([]NotificationType, error)
}
