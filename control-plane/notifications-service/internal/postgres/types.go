/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgres

type DataBaseConfig struct {
	Host              string `mapstructure:"host"`
	Port              string `mapstructure:"port"`
	Database          string `mapstructure:"database"`
	Schema            string `mapstructure:"schema"`
	User              string `mapstructure:"user"`
	Password          string `mapstructure:"password"`
	SslMode           string `mapstructure:"ssl_mode"`
	EncryptDecryptKey string `mapstructure:"encrypt_decrypt_key"`

	MaxOpenConns          int `mapstructure:"max_open_connections"`
	MaxIdleConns          int `mapstructure:"max_idle_connections"`
	ConnectionMaxLifetime int `mapstructure:"connection_max_lifetime_minutes"`
	ConnectionMaxIdleTime int `mapstructure:"connection_max_idle_time_minutes"`
}
