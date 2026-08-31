/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package branding

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/control-plane/bff-service/pkg/clients"
	tmOpenapi "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

const logoMigrationPageSize int32 = 50

type LogoMigrator struct {
	tmClient clients.TenantsManagerClient
	m        *manager.Manager
}

func NewLogoMigrator(tmClient clients.TenantsManagerClient) *LogoMigrator {
	m := manager.NewManager(nil, nil)
	m.Register(&Logo{})
	return &LogoMigrator{tmClient: tmClient, m: m}
}

func (lm *LogoMigrator) Migrate() error {
	shouldMigrate, err := lm.m.ShouldMigrateSetting(new(Logo).Key())
	if err != nil {
		return fmt.Errorf("failed to check logo migration status: %w", err)
	}
	if !shouldMigrate {
		logging.Infof("logo migration already completed, skipping")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logos, err := lm.getAllLogos(ctx)
	if err != nil {
		return err
	}
	if len(logos) == 0 {
		logging.Infof("no logos to migrate")
		return nil
	}

	inputs := lm.formatForMigration(logos)
	if err := lm.m.UpdateAllTenantsByKey(new(Logo).Key(), inputs); err != nil {
		return fmt.Errorf("logo migration failed: %w", err)
	}

	logging.Infof("logo migration completed successfully")
	return nil
}

func (lm *LogoMigrator) getAllLogos(ctx context.Context) ([]tmOpenapi.TenantLogo, error) {
	var all []tmOpenapi.TenantLogo
	var offset int32
	for {
		response, _, err := lm.tmClient.GetAllLogos(ctx, offset, logoMigrationPageSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch logos from tenants-manager (offset=%d): %w", offset, err)
		}
		if response == nil || len(response.Logos) == 0 {
			break
		}
		all = append(all, response.Logos...)
		if int32(len(response.Logos)) < logoMigrationPageSize {
			break
		}
		offset += logoMigrationPageSize
	}
	return all, nil
}

func (lm *LogoMigrator) formatForMigration(logos []tmOpenapi.TenantLogo) []manager.TenantSettingValueForBackendResponse {
	logo := &Logo{}
	inputs := make([]manager.TenantSettingValueForBackendResponse, 0, len(logos))
	for _, tl := range logos {
		inputs = append(inputs, manager.TenantSettingValueForBackendResponse{
			TenantId: int(tl.TenantId),
			Value:    logo.TransformForMigration(tl.Logo),
		})
	}
	return inputs
}
