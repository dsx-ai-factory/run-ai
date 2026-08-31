/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client_utils

import "strings"

const (
	HttpProtocol  = "http://"
	HttpsProtocol = "https://"
)

func GetInternalUrl(baseURL string) string {
	if !strings.HasPrefix(baseURL, HttpProtocol) && !strings.HasPrefix(baseURL, HttpsProtocol) {
		baseURL = HttpProtocol + baseURL
	}
	return baseURL
}

func GetExternalUrl(baseURL string) string {
	if !strings.HasPrefix(baseURL, HttpProtocol) && !strings.HasPrefix(baseURL, HttpsProtocol) {
		baseURL = HttpsProtocol + baseURL
	}
	return baseURL
}
