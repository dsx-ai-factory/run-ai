/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

type Config struct {
	KeyDelimiter string `validate:"required,excludesall= "`
}

func getConfig() *Config {
	return &Config{
		KeyDelimiter: getEnvOrDefault(envConfigKeyDelimiter, defaultKeyDelimiter),
	}
}
