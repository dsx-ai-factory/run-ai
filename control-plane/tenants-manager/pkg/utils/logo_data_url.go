/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"encoding/base64"
	"net/http"
	"strings"
)

const (
	mimeTypePNG       = "image/png"
	mimeTypeSVG       = "image/svg+xml"
	mimeTypeTextXML   = "text/xml"
	mimeTypeTextPlain = "text/plain"
)

func ReconstructDataUrl(data []byte) string {
	fileType := http.DetectContentType(data)
	base64Data := base64.StdEncoding.EncodeToString(data)
	switch {
	case fileType == mimeTypePNG:
		return constructDataURL(mimeTypePNG, base64Data)
	case fileType == mimeTypeSVG:
		return constructDataURL(mimeTypeSVG, base64Data)
	case strings.HasPrefix(fileType, mimeTypeTextXML) && isSVGContent(data):
		return constructDataURL(mimeTypeSVG, base64Data)
	case strings.HasPrefix(fileType, mimeTypeTextPlain) && isSVGContent(data):
		return constructDataURL(mimeTypeSVG, base64Data)
	default:
		return constructDataURL(mimeTypePNG, base64Data)
	}
}

func constructDataURL(mimeType, base64Data string) string {
	return "data:" + mimeType + ";base64," + base64Data
}

func isSVGContent(data []byte) bool {
	content := strings.ToLower(string(data))
	if len(content) > 512 {
		content = content[:512]
	}
	return strings.Contains(content, "<svg") && strings.Contains(content, "xmlns")
}
