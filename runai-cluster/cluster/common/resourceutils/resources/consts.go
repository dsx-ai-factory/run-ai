/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

const (
	migExtractionRegex      = `^nvidia.com/mig-(\d+)g\.(\d+)gb$`
	migRegexNumOfOutputs    = 3
	migDeviceAnnotationName = "runai-mig-device"
	deviceNameKey           = "name"

	A100                   = "A100"
	H100                   = "H100"
	defaultComputePortions = float64(7)
)
