/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant_lifecycle

import (
	"context"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

type TenantLifecycleSubscriber struct {
	tenantDao          dao.Tenant
	provisioningRunDao dao.ProvisioningRun
	builder            builders.Interface
}

func NewTenantLifecycleSubscriber(
	tenantDao dao.Tenant,
	provisioningRunDao dao.ProvisioningRun,
	builder builders.Interface,
) *TenantLifecycleSubscriber {
	return &TenantLifecycleSubscriber{
		tenantDao:          tenantDao,
		provisioningRunDao: provisioningRunDao,
		builder:            builder,
	}
}

func (s *TenantLifecycleSubscriber) Register() error {
	if err := runaipubsub.Subscribe[TenantProvision](s.handleTenantProvision); err != nil {
		return err
	}
	return runaipubsub.Subscribe[TenantDestruct](s.handleTenantDestruct)
}

func (s *TenantLifecycleSubscriber) handleTenantProvision(ctx context.Context, event *TenantProvision) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Received TenantProvision for tenant %s (id: %d)", event.TenantName, event.TenantId)

	tenant, daoErr := s.tenantDao.GetById(ctx, event.TenantId)
	if daoErr != nil {
		if daoErr.Reason == dao.NotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Tenant %d not found during TenantProvision, acking", event.TenantId)
			return nil
		}
		return daoErr
	}
	if tenant.Status == models.TenantReady || tenant.Status == models.TenantDeleting {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Tenant %s already %s, skipping TenantProvision", event.TenantName, tenant.Status)
		return nil
	}
	if tenant.Status == models.TenantFailed {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Tenant %s is Failed (redelivery), resetting to Creating", event.TenantName)
		if _, daoErr := s.tenantDao.UpdateStatus(ctx, event.TenantId, models.TenantCreating); daoErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to reset tenant %s to Creating: %v — nacking", event.TenantName, daoErr.Error())
			return daoErr
		}
	}

	run, runErr := s.provisioningRunDao.Create(ctx, event.TenantId)
	if runErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to create provisioning run for tenant %s: %v", event.TenantName, runErr.Error())
	}

	info := &builders.TenantInfo{
		Name:         event.TenantName,
		Id:           event.TenantId,
		MimirId:      event.MimirId,
		Email:        event.Email,
		Password:     event.Password,
		ContractType: event.ContractType,
		Role:         event.Role,
		MetricStore:  event.MetricStore,
	}

	builderResponse := make(map[string]interface{})
	if err := s.builder.Construct(ctx, info, builderResponse); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("builder.Construct failed for tenant %s: %v — persisting Failed state", event.TenantName, err)
		if run != nil {
			if completeErr := s.provisioningRunDao.Complete(ctx, run.ID, models.ProvisioningRunFailed, err.Error()); completeErr != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to record provisioning run failure for tenant %s: %v", event.TenantName, completeErr.Error())
			}
		}
		if _, daoErr := s.tenantDao.UpdateStatus(ctx, event.TenantId, models.TenantFailed); daoErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to set tenant %s to Failed: %v", event.TenantName, daoErr.Error())
		}
		return err
	}

	if run != nil {
		if completeErr := s.provisioningRunDao.Complete(ctx, run.ID, models.ProvisioningRunSuccess, ""); completeErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to record provisioning run success for tenant %s: %v", event.TenantName, completeErr.Error())
		}
	}
	if _, daoErr := s.tenantDao.UpdateStatus(ctx, event.TenantId, models.TenantReady); daoErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("UpdateStatus(Ready) failed for tenant %s: %v — nacking", event.TenantName, daoErr.Error())
		return daoErr
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Tenant %s provisioned successfully", event.TenantName)
	return nil
}

func (s *TenantLifecycleSubscriber) handleTenantDestruct(ctx context.Context, event *TenantDestruct) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Received TenantDestruct for tenant %s (id: %d)", event.TenantName, event.TenantId)

	info := &builders.TenantInfo{
		Name:    event.TenantName,
		Id:      event.TenantId,
		MimirId: event.MimirId,
		Email:   event.Email,
	}

	if err := s.builder.Destruct(ctx, info); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("builder.Destruct failed for tenant %s: %v — nacking for redelivery", event.TenantName, err)
		return err
	}

	if daoErr := s.tenantDao.HardDeleteById(ctx, event.TenantId); daoErr != nil {
		if daoErr.Reason == dao.NotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Tenant %d already deleted from DB, treating as success", event.TenantId)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("HardDeleteById failed for tenant %d: %v — nacking", event.TenantId, daoErr.Error())
			return daoErr
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Tenant %s destructed successfully", event.TenantName)
	return nil
}
