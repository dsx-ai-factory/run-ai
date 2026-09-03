/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const UrlParamRegex = "&%s=%s"

func GetTenantNameFromHost(host string) string {
	return strings.ToLower(strings.Split(host, ".")[0])
}

func GetStringFromCtx(ctx context.Context, value any) string {
	val, ok := ctx.Value(value).(string)
	if !ok {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("missing %s in ctx", value)
	}
	return val
}

func FormatString(regex string, str1 string, str2 string) string {
	return fmt.Sprintf(regex, str1, str2)
}

func TemplateBaseUrl(baseUrlTemplate string, domain string) (string, error) {
	t, err := template.New("baseUrlTemplate").Parse(baseUrlTemplate)
	if err != nil {
		return "", err
	}
	var value bytes.Buffer
	err = t.Execute(&value, strings.ToLower(domain))
	if err != nil {
		return "", err
	}
	return value.String(), nil
}

// DerefOrZero returns the value pointed to by p, or the zero value of T if p is nil.
func DerefOrZero[T any](p *T) T {
	if p == nil {
		instance := new(T)
		return *instance
	}
	return *p
}

func ParseTenantId(ctx context.Context) (int32, error) {
	tenantId, ok := ctx.Value(runaictx.CfTenantId).(string)
	if !ok {
		return 0, fmt.Errorf("no tenant id provided")
	}
	tenantIdInt32, parseErr := strconv.Atoi(tenantId)
	if parseErr != nil {
		return 0, fmt.Errorf("invalid tenant id")
	}

	return int32(tenantIdInt32), nil
}
