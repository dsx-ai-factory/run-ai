/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package oauth2proxy defines the shared OAuth2 Proxy template identifiers.
package oauth2proxy

const (
	ConfigMapName          = "runai-oauth2-proxy-templates"
	VolumeName             = "runai-oauth2-proxy-templates"
	TemplateKey            = "error.html"
	TemplatesPath          = "/usr/share/oauth2-proxy"
	TemplatesEnabledEnvVar = "RUNAI_OAUTH2_PROXY_TEMPLATES_ENABLED"
)
