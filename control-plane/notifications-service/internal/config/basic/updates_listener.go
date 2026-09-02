/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package basic

import "context"

type ConfigurationUpdatesListener struct{}

func NewConfigurationUpdatesListener() *ConfigurationUpdatesListener {
	return &ConfigurationUpdatesListener{}
}

func (b *ConfigurationUpdatesListener) Listen(context.Context, func(listenerId string) error) {
}
