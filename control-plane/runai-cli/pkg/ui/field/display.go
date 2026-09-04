/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field

import (
	"fmt"
	"strings"
	"time"
)

const dateFormat = time.RFC3339

// StringOrEmpty returns the string value or empty string if nil
func StringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Int32OrEmpty returns the int32 value as string or empty string if nil
func Int32OrEmpty(i *int32) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%d", *i)
}

// Int64OrEmpty returns the int64 value as string or empty string if nil
func Int64OrEmpty(i *int64) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%d", *i)
}

// BoolOrEmpty returns the bool value as string or empty string if nil
func BoolOrEmpty(b *bool) string {
	if b == nil {
		return ""
	}
	return fmt.Sprintf("%t", *b)
}

// BoolIsTrue returns true if the bool pointer is not nil and true
func BoolIsTrue(b *bool) bool {
	return b != nil && *b
}

// Float64OrEmpty returns the float64 value as formatted string or empty string if nil
func Float64OrEmpty(f *float64) string {
	if f == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *f)
}

// Float32OrEmpty returns the float32 value as formatted string or empty string if nil
func Float32OrEmpty(f *float32) string {
	if f == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *f)
}

// TimeOrEmpty returns the time value as RFC3339 formatted string or empty string if nil
func TimeOrEmpty(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(dateFormat)
}

func CapitalizeFirst(s string) string {
	if len(s) > 0 {
		return strings.ToUpper(s[:1]) + s[1:]
	}
	return s
}

// BoolToYesNo returns "Yes", "No", or "-" for table displays
func BoolToYesNo(b *bool) string {
	if b == nil {
		return "-"
	}
	if *b {
		return "Yes"
	}
	return "No"
}

// BoolToEnabled returns "Enabled", "Disabled", or "-" for table displays
func BoolToEnabled(b *bool) string {
	if b == nil {
		return "-"
	}
	if *b {
		return "Enabled"
	}
	return "Disabled"
}

// StringOrDash returns the string value or "-" for table displays
func StringOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// PtrStringOrDash returns the string pointer value or "-" for table displays
func PtrStringOrDash(s *string) string {
	if s == nil || *s == "" {
		return "-"
	}
	return *s
}
