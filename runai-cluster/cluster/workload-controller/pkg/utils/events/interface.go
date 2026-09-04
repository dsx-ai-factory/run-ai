/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package events

type Recorder interface {
	Event(eventtype, reason, message string)
	Eventf(eventtype, reason, messageFmt string, args ...interface{})
}
