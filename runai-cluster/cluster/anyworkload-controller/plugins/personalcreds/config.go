/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package personalcreds

import commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"

type Config struct {
	Enabled        bool   `mapstructure:"enabled"`
	RunaiNamespace string `mapstructure:"runai_namespace"`
}

func (c *Config) SetDefaults() {
	c.Enabled = true
	c.RunaiNamespace = commonconstants.NamespaceRunAi
}
