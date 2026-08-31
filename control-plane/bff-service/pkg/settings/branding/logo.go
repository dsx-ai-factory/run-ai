/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package branding

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/setting"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

type Logo struct {
	Logo []byte `json:"logo"`
}

func NewLogo() setting_pkg.Settinger {
	return &Logo{}
}

func (l *Logo) Category() string {
	return string(utils.BrandingCategory)
}

func (l *Logo) Description() string {
	return "Retrieve the base64 logo file."
}

func (l *Logo) DisplayName() string {
	return "Logo"
}

func (l *Logo) Key() string {
	return string(api.BRANDINGSETTINGSKEYENUM_LOGO)
}

func (l *Logo) SetDefault() {}

func (l *Logo) Stage() setting_pkg.Stage {
	return setting_pkg.GA
}

func (l *Logo) Validate(_ interface{}) bool {
	return true
}

func (l *Logo) TransformToResponse() interface{} {
	return api.GetLogo{
		Value:       api.GetLogoValue{Logo: reconstructDataURL(l.Logo)},
		DisplayName: l.DisplayName(),
		Description: l.Description(),
	}
}

// reconstructDataURL detects the MIME type of the raw bytes and returns a data URL,
// identical to tenants-manager's reconstructDataUrl behavior.
func reconstructDataURL(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	mimeType := http.DetectContentType(data)
	if (strings.HasPrefix(mimeType, "text/xml") || strings.HasPrefix(mimeType, "text/plain")) && isSVGBytes(data) {
		mimeType = "image/svg+xml"
	}
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data)
}

func isSVGBytes(data []byte) bool {
	content := strings.ToLower(string(data))
	if len(content) > 512 {
		content = content[:512]
	}
	return strings.Contains(content, "<svg") && strings.Contains(content, "xmlns")
}

func (l *Logo) TransformForMigration(value interface{}) interface{} {
	logoURL, ok := value.(string)
	if !ok {
		return value
	}
	decodedBytes, err := getLogoIfValid(newLogoValidate(), logoURL)
	if err != nil || decodedBytes == nil {
		return value
	}
	return &Logo{Logo: *decodedBytes}
}

func GetLogo(tenantID string) (setting_pkg.Settinger, error) {
	return setting.Get[*Logo](tenantID)
}

func UpdateLogo(tenantID string, updateRequest any) error {
	req, ok := updateRequest.(api.UpdateLogo)
	if !ok {
		return utils.ErrInvalidInput
	}

	if req.Logo == "" {
		return setting.Update[*Logo](tenantID, Logo{}).UpdatedByUser().Execute()
	}

	decodedBytes, validationErr := getLogoIfValid(newLogoValidate(), req.Logo)
	if validationErr != nil {
		return utils.Error{Code: validationErr.Code, Message: validationErr.Message}
	}

	return setting.Update[*Logo](tenantID, Logo{Logo: *decodedBytes}).UpdatedByUser().Execute()
}
