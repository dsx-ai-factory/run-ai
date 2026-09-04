/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

// NetworkingSubmissionProfileAnnotation is written to a WorkloadRunner by the
// agent. It records which trusted submission path produced endpoint-array
// networking and is not a caller-controlled workload setting.
const NetworkingSubmissionProfileAnnotation = "run.ai/networking-submission-profile"
