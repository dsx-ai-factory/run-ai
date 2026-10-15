/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package framework

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PluginDeps struct {
	Client client.Reader
	// APIReader bypasses the cache when resource existence affects immutable templates.
	APIReader client.Reader
	Logger    logr.Logger
}
