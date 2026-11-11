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
	H200                   = "H200"
	B200                   = "B200" // also matches GB200
	B300                   = "B300"
	RTXPro5000             = "RTX-PRO-5000"
	RTXPro6000             = "RTX-PRO-6000"
	rtxPro5000Portions     = float64(2)
	rtxPro6000Portions     = float64(4)
)
