/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

type Notification struct {
	Category             string
	Type                 string
	Title                string
	Description          string
	Tenant               string
	User                 string
	AdditionalProperties map[string]interface{}
}

type NotificationKind struct {
	Category    string
	Type        string
	Description string
	Template    string
}
