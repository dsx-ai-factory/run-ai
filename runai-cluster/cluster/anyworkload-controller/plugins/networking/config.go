/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

type Config struct {
	Enabled bool `mapstructure:"enabled"`
}

func (c *Config) SetDefaults() {
	c.Enabled = true
}
