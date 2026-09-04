/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logger

import "fmt"

type Loggerer interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type Logger struct {
	Loggerer
	errorHandler func(err error)
}

func NewLogger(logger Loggerer, errorHandler func(e error)) *Logger {
	return &Logger{
		Loggerer:     logger,
		errorHandler: errorHandler,
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Loggerer.Debugf(format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Loggerer.Infof(format, args...)

}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Loggerer.Warnf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.errorHandler(fmt.Errorf(format, args...))
	l.Loggerer.Errorf(format, args...)
}
