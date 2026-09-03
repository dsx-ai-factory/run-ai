/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package events

import "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"

type sink interface {
	Handle(e orchestrator2.Event) error
	Close() error
}
