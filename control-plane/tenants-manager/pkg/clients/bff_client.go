/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	// "strconv"
	"time"

	bff_openapi "github.com/run-ai/runai/control-plane/bff-service/pkg/client/generated"
)

type IBffClient interface {
	GetLogo(ctx context.Context, tenantID string) (string, error)
	UpdateLogo(ctx context.Context, tenantID string, logoDataURL string) error
}

type BffClient struct {
	SettingsAPI bff_openapi.SettingsAPI
}

func NewBffClient(bffURL string) IBffClient {
	bffClient := bff_openapi.NewAPIClient(&bff_openapi.Configuration{
		Servers:    bff_openapi.ServerConfigurations{{URL: bffURL}},
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	})
	return &BffClient{
		SettingsAPI: bffClient.SettingsAPI,
	}
}

func (a *BffClient) GetLogo(ctx context.Context, tenantID string) (string, error) {
	result, _, err := a.SettingsAPI.
		GetBrandingSettingByKey(ctx, bff_openapi.BRANDINGSETTINGSKEYENUM_LOGO).
		TenantId(tenantID).
		Execute()
	if err != nil {
		return "", fmt.Errorf("GetLogo from BFF: %w", err)
	}
	return extractLogoFromSettingResponse(result)
}

func (a *BffClient) UpdateLogo(ctx context.Context, tenantID string, logoDataURL string) error {
	resp, err := a.SettingsAPI.
		UpdateBrandingSettingByKey(ctx, bff_openapi.BRANDINGSETTINGSKEYENUM_LOGO).
		UpdateLogo(*bff_openapi.NewUpdateLogo(logoDataURL)).
		TenantId(tenantID).
		Execute()
	if err != nil {
		return fmt.Errorf("UpdateLogo in BFF: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("UpdateLogo in BFF returned status %d", resp.StatusCode)
	}
	return nil
}

type bffLogoSettingResponse struct {
	Value struct {
		Logo string `json:"logo"`
	} `json:"value"`
}

func extractLogoFromSettingResponse(result map[string]interface{}) (string, error) {
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal BFF logo response: %w", err)
	}
	var resp bffLogoSettingResponse
	if err := json.Unmarshal(jsonBytes, &resp); err != nil {
		return "", fmt.Errorf("failed to parse BFF logo response: %w", err)
	}
	return resp.Value.Logo, nil
}
