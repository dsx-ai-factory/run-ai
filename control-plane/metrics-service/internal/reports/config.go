/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports

type ReportsConfiguration struct {
	S3Config               *S3Config `mapstructure:"s3_config" validate:"required"`
	FetchLimit             int       `mapstructure:"fetch_limit" validate:"required"`
	ConcurrencyLimit       int       `mapstructure:"processor_concurrency_limit,min=1"`
	RunPeriodSec           int       `mapstructure:"processor_run_interval_sec,min=1"`
	DeleteFetchLimit       int       `mapstructure:"delete_fetch_limit" validate:"min=1"`
	DeleteConcurrencyLimit int       `mapstructure:"delete_concurrency_limit" validate:"min=1"`
}

type S3Config struct {
	Bucket string `mapstructure:"bucket" validate:"required"`
}

func (rc *ReportsConfiguration) SetDefaults() {
	rc.S3Config = &S3Config{
		Bucket: "metric_service_reports",
	}
	rc.FetchLimit = 5
	rc.ConcurrencyLimit = 10
	rc.RunPeriodSec = 60
	rc.DeleteFetchLimit = 5
	rc.DeleteConcurrencyLimit = 10
}
