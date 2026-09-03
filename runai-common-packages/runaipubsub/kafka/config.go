/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kafka

type KafkaAuthType string

const (
	KafkaAuthTypePlain KafkaAuthType = "plain"
	KafkaAuthTypeScram KafkaAuthType = "scram"
)

type Config struct {
	AuthType KafkaAuthType `mapstructure:"pubsub_kafka_auth_type"`
}

func (c *Config) SetDefaults() {
	c.AuthType = KafkaAuthTypeScram
}
