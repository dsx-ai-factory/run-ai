/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaipubsub

type dummyLogger struct{}

func (*dummyLogger) Debugf(format string, args ...interface{}) {
}

func (*dummyLogger) Infof(format string, args ...interface{}) {
}

func (*dummyLogger) Warnf(format string, args ...interface{}) {
}

func (*dummyLogger) Errorf(format string, args ...interface{}) {
}
