/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	"fmt"
	"slices"
	"strings"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/samber/lo"
)

func FormatMapAsKeyValue(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	keys := lo.Keys(m)
	slices.Sort(keys)

	pairs := make([]string, 0, len(m))
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, m[k]))
	}
	return strings.Join(pairs, ", ")
}

func FormatToleration(t *openapi.Toleration) string {
	if t == nil {
		return ""
	}

	key := ""
	if t.Key.IsSet() && t.Key.Get() != nil {
		key = *t.Key.Get()
	}
	if key == "" {
		return ""
	}

	value := ""
	if t.Value.IsSet() && t.Value.Get() != nil {
		value = *t.Value.Get()
	}

	effect := ""
	if t.Effect.IsSet() {
		effect = string(t.GetEffect())
	}

	if value != "" && effect != "" {
		return key + "=" + value + ":" + effect
	} else if effect != "" {
		return key + ":" + effect
	} else if value != "" {
		return key + "=" + value
	}
	return key
}

func FormatProbe(p *openapi.Probe) string {
	if p == nil {
		return ""
	}

	if !p.Handler.IsSet() {
		return ""
	}

	handler := p.GetHandler()
	if handler.HttpGet != nil {
		httpGet := handler.HttpGet
		port := ""
		path := ""
		if httpGet.Port.IsSet() && httpGet.Port.Get() != nil {
			port = fmt.Sprintf("%d", *httpGet.Port.Get())
		}
		if httpGet.Path.IsSet() && httpGet.Path.Get() != nil {
			path = *httpGet.Path.Get()
		}
		return "HTTP GET :" + port + path
	}

	return ""
}
