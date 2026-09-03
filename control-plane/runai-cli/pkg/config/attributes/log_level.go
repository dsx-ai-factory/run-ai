/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attributes

import (
	"errors"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CliLogLevel string

const (
	EmptyString                    = ""
	CliLogLevelDebug   CliLogLevel = "debug"
	CliLogLevelInfo    CliLogLevel = "info"
	CliLogLevelWarning CliLogLevel = "warn"
	CliLogLevelError   CliLogLevel = "error"

	ErrNoMatchLogLevelMsg = "log level must be one of: debug|info|warn|error"
)

func (cll *CliLogLevel) String() string {
	return string(*cll)
}

// Type is only used in help text
func (cll *CliLogLevel) Type() string {
	return "CliLogLevel"
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (cll *CliLogLevel) Set(v string) error {
	switch v {
	case string(CliLogLevelInfo), string(CliLogLevelDebug),
		string(CliLogLevelWarning), string(CliLogLevelError):
		*cll = CliLogLevel(v)
		return nil
	default:
		return errors.New(ErrNoMatchLogLevelMsg)
	}
}

func (cll *CliLogLevel) ZapLogLevel() (zapcore.Level, error) {
	switch *cll {
	case CliLogLevelError:
		return zap.ErrorLevel, nil
	case CliLogLevelWarning:
		return zap.WarnLevel, nil
	case CliLogLevelInfo, EmptyString:
		return zap.InfoLevel, nil
	case CliLogLevelDebug:
		return zap.DebugLevel, nil
	}
	return zap.InfoLevel, errors.New(ErrNoMatchLogLevelMsg)
}

// CliLogLevelCompletion auto completion support for log level
func CliLogLevelCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"debug\tset debug level",
		"info\tset info level",
		"warn\tset warning level",
		"warn\tset error level",
	}, cobra.ShellCompDirectiveNoFileComp
}
