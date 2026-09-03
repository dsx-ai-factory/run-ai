/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scale

import (
	"context"
)

type Entity interface {
	Create(ctx context.Context) error
	Monitor() error
	NumberOfSuccessfulCreatedObjects() int
	GetName() string
	Delete(forceDelete bool)
	Count(filter []string) (int64, error)
}

type SampledEntity interface {
	Entity
	GenerateSamplesReport() (*SamplesReport, error)
	GetSamplesReportFileName() string
}
