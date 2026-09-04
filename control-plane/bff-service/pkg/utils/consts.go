/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import "time"

type CategoryKey string

const (
	NotificationsCategory CategoryKey = "Notifications"
	UIWorkloadsCategory   CategoryKey = "UI-Workloads"
	UIAppearanceCategory  CategoryKey = "UI-Appearance"
	BrandingCategory      CategoryKey = "Branding"
)

func GetCurrentTimestampMs() int64 {
	return time.Now().UnixMilli()
}
