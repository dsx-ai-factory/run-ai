// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"regexp"
	"time"
)

const (
	GpuPciDeviceIdMaxAge        = 5 * time.Minute
	GpuPciDeviceIdMaxFutureSkew = time.Minute
	GpuPciDeviceIdPattern       = `^[0-9a-f]{4}:10de$`
)

var gpuPciDeviceIdPattern = regexp.MustCompile(GpuPciDeviceIdPattern)

// NormalizeGpuPciIdentity expires collector evidence even when cluster-sync is stopped.
// The original observation time must survive ingestion; receiving an update is not discovery.
func NormalizeGpuPciIdentity(id string, observedAt *time.Time, now time.Time) (string, *time.Time) {
	if !gpuPciDeviceIdPattern.MatchString(id) || id == "0000:10de" || observedAt == nil ||
		observedAt.Before(now.Add(-GpuPciDeviceIdMaxAge)) || observedAt.After(now.Add(GpuPciDeviceIdMaxFutureSkew)) {
		return "", nil
	}
	return id, observedAt
}
