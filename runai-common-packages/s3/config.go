/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package s3

import (
	"fmt"

	"github.com/run-ai/runai/runai-common-packages/config"
)

type S3Config struct {
	Endpoint    string `mapstructure:"endpoint"`
	AccessKeyId string `mapstructure:"access_key_id"`
	AccessKey   string `mapstructure:"access_key" config:"secret"`
	UseSSL      bool   `mapstructure:"use_ssl"`
}

func (sc *S3Config) SetDefaults() {
	sc.UseSSL = false
}

func InitConfig() error {
	var err error
	if err = config.SetKbSecretsFolder(""); err != nil {
		return fmt.Errorf("failed to set secrets folder: %w", err)
	}
	if err = config.RegisterRoot[*S3Config](); err != nil {
		return fmt.Errorf("failed to register s3 config: %w", err)
	}

	return nil
}
