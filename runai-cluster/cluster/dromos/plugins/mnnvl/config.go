/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

type Config struct {
	Enabled              bool `mapstructure:"ENABLE_MNNVL_SUPPORT"`
	ElasticComputeDomain bool `mapstructure:"ENABLE_ELASTIC_COMPUTE_DOMAIN"`
}
