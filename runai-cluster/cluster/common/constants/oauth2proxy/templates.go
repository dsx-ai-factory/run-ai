/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package oauth2proxy defines the files supplied to the upstream OAuth2 Proxy image.
package oauth2proxy

import _ "embed"

const (
	ConfigMapName            = "runai-oauth2-proxy-templates"
	VolumeName               = "runai-oauth2-proxy-templates"
	TemplateKey              = "error.html"
	TemplatesPath            = "/usr/share/oauth2-proxy"
	TemplatesConfigMapEnvVar = "RUNAI_OAUTH2_PROXY_TEMPLATES_CONFIG_MAP"
)

// ErrorPageHTML is the existing Run:ai error page, distributed as a ConfigMap
// instead of being copied into a custom container image.
//
//go:embed error.html
var ErrorPageHTML string
