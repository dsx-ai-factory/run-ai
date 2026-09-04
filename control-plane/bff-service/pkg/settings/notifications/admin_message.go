/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notifications

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
)

type AdminMessage struct {
	ID            string `json:"id"`
	Msg           string `json:"msg"`
	CreatedAt     int64  `json:"createdAt"`
	EnableDismiss bool   `json:"enableDismiss"`
}

var allowedTags = map[string]bool{
	"a": true, "b": true, "i": true, "u": true, "p": true, "br": true, "hr": true,
	"ul": true, "ol": true, "li": true, "strong": true, "em": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"div": true, "span": true, "blockquote": true, "pre": true, "code": true, "font": true,
}

var (
	eventHandlerPattern = regexp.MustCompile(`(?i)\bon\w+\s*=`)
	dangerousProtocol   = regexp.MustCompile(`(?i)(javascript|vbscript|data):`)
	tagPattern          = regexp.MustCompile(`(?i)</?([a-zA-Z][a-zA-Z0-9]*)\b[^>]*>`)
	hrefPattern         = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']*)["']`)
	safeURLPattern      = regexp.MustCompile(`(?i)^(https?://|mailto:|#|/)`)
)

func NewAdminMessage() setting_pkg.SettingerWithTransform {
	return &AdminMessage{}
}

func (a *AdminMessage) Category() string {
	return string(utils.NotificationsCategory)
}

func (a *AdminMessage) Description() string {
	return "The message will be displayed every time users sign in and immediately after it is saved to users who are already signed in."
}

func (a *AdminMessage) DisplayName() string {
	return "Admin message"
}

func (a *AdminMessage) Key() string {
	return string(api.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE)
}

func (a *AdminMessage) SetDefault() {
	a.ID = uuid.New().String()
	a.Msg = ""
	a.CreatedAt = 0
	a.EnableDismiss = false
}

func (a *AdminMessage) Stage() setting_pkg.Stage {
	return setting_pkg.GA
}

func (a *AdminMessage) Validate(value interface{}) bool {
	var msg string

	switch v := value.(type) {
	case AdminMessage:
		msg = v.Msg
	case *AdminMessage:
		msg = v.Msg
	// The following cases handle migration from legacy backend where value comes as map or JSON string
	case map[string]interface{}:
		if m, ok := v["msg"].(string); ok {
			msg = m
		}
	case string:
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return false
		}
		if m, ok := parsed["msg"].(string); ok {
			msg = m
		}
	default:
		return false
	}

	return validateHTMLContent(msg)
}

func validateHTMLContent(content string) bool {
	if eventHandlerPattern.MatchString(content) {
		return false
	}

	if dangerousProtocol.MatchString(content) {
		return false
	}

	matches := tagPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			tagName := strings.ToLower(match[1])
			if !allowedTags[tagName] {
				return false
			}
		}
	}

	hrefMatches := hrefPattern.FindAllStringSubmatch(content, -1)
	for _, match := range hrefMatches {
		if len(match) > 1 {
			url := match[1]
			if url != "" && !safeURLPattern.MatchString(url) {
				return false
			}
		}
	}

	return true
}

func (a *AdminMessage) TransformForMigration(value interface{}) interface{} {
	var valueMap map[string]interface{}

	switch v := value.(type) {
	case map[string]interface{}:
		valueMap = v
	case string:
		if err := json.Unmarshal([]byte(v), &valueMap); err != nil {
			return value
		}
	default:
		return value
	}

	adminMsg := AdminMessage{}

	if id, ok := valueMap["id"].(string); ok {
		adminMsg.ID = id
	}
	if msg, ok := valueMap["msg"].(string); ok {
		adminMsg.Msg = msg
	}
	if createdAt, ok := valueMap["createdAt"].(float64); ok {
		adminMsg.CreatedAt = int64(createdAt)
	}
	if enableDismiss, ok := valueMap["enableDismiss"].(bool); ok {
		adminMsg.EnableDismiss = enableDismiss
	}

	return adminMsg
}

func GetAdminMessage(tenantID string) (setting_pkg.Settinger, error) {
	return setting.Get[*AdminMessage](tenantID)
}

func UpdateAdminMessage(tenantID string, updateRequest any) error {
	var msg string
	var enableDismiss bool

	switch req := updateRequest.(type) {
	case api.UpdateAdminMessage:
		if req.Msg != nil {
			msg = *req.Msg
		}
		if req.EnableDismiss != nil {
			enableDismiss = *req.EnableDismiss
		}
	case map[string]interface{}:
		if m, ok := req["msg"].(string); ok {
			msg = m
		}
		if e, ok := req["enableDismiss"].(bool); ok {
			enableDismiss = e
		}
	default:
		return utils.ErrInvalidInput
	}

	settingValue := AdminMessage{
		ID:            uuid.New().String(),
		CreatedAt:     utils.GetCurrentTimestampMs(),
		Msg:           msg,
		EnableDismiss: enableDismiss,
	}

	return setting.Update[*AdminMessage](tenantID, settingValue).UpdatedByUser().Execute()
}
