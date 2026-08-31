/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logger

var log Logger = &dummyLogger{}

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
	Debugf(format string, args ...any)
}

func SetLogger(l Logger) {
	if l == nil {
		return
	}
	log = l
}

func GetLogger() Logger {
	return log
}
