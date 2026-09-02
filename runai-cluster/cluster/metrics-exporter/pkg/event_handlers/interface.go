/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package event_handlers

import "context"

type ObjectIdentifier struct {
	Name      string
	Namespace string
}

type EventHandler interface {
	ResetMetrics(identifier ObjectIdentifier)
	SetMetrics(ctx context.Context, eventObject interface{}) error
}
