/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notification_state

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/pkg/errors"
)

type NotificationStateManager struct {
	stateClient          StateClienter
	channelClient        ChannelClienter
	userSubClient        SubUserClienter
	stateMigrationClient StateMigrationClient
}

func NewNotificationStateManager(stateClient StateClienter, channelClient ChannelClienter, userSubClient SubUserClienter, stateMigrateClient StateMigrationClient) *NotificationStateManager {
	return &NotificationStateManager{
		stateClient:          stateClient,
		stateMigrationClient: stateMigrateClient,
		channelClient:        channelClient,
		userSubClient:        userSubClient,
	}
}

func (nsm *NotificationStateManager) GetEmailNotificationState(ctx context.Context, tenantId string) (Stater, error) {
	if err := nsm.stateMigrationClient.Migrate(ctx, tenantId); err != nil {
		return nil, errors.Wrap(ErrFailedToUpdateNotificationState, err.Error())
	}

	return nsm.stateClient.GetEmail(ctx, tenantId)
}

func (nsm *NotificationStateManager) GetSlackNotificationState(ctx context.Context, tenantId string) (Stater, error) {
	if err := nsm.stateMigrationClient.Migrate(ctx, tenantId); err != nil {
		return nil, errors.Wrap(ErrFailedToUpdateNotificationState, err.Error())
	}

	return nsm.stateClient.GetSlack(ctx, tenantId)
}

func (nsm *NotificationStateManager) GetNotificationState(ctx context.Context, tenantId string) ([]Stater, error) {
	if err := nsm.stateMigrationClient.Migrate(ctx, tenantId); err != nil {
		return nil, errors.Wrap(ErrFailedToUpdateNotificationState, err.Error())
	}

	return nsm.stateClient.Get(ctx, tenantId)
}

func (nsm *NotificationStateManager) GetDetailedNotificationState(ctx context.Context, tenantId string) ([]Stater, error) {
	if err := nsm.stateMigrationClient.Migrate(ctx, tenantId); err != nil {
		return nil, errors.Wrap(ErrFailedToUpdateNotificationState, err.Error())
	}

	return nsm.stateClient.GetDetails(ctx, tenantId)
}

func (nsm *NotificationStateManager) GetDetailedSlackNotificationState(ctx context.Context, tenantId string) (Stater, error) {
	if err := nsm.stateMigrationClient.Migrate(ctx, tenantId); err != nil {
		return nil, errors.Wrap(ErrFailedToUpdateNotificationState, err.Error())
	}

	return nsm.stateClient.GetSlackDetails(ctx, tenantId)
}

func (nsm *NotificationStateManager) UpdateNotificationState(ctx context.Context, tenantId string, notificationState []Stater) ([]Stater, error) {
	if err := nsm.stateMigrationClient.Migrate(ctx, tenantId); err != nil {
		return nil, errors.Wrap(ErrFailedToUpdateNotificationState, err.Error())
	}

	for _, state := range notificationState {
		if err := nsm.validateExistAndDisableChannelIfNeeded(ctx, tenantId, state); err != nil {
			return nil, err
		}
	}

	return nsm.stateClient.UpdateState(ctx, tenantId, notificationState)

}

func (nsm *NotificationStateManager) UpdateEmailNotificationState(ctx context.Context, tenantId string, notificationState Stater) (Stater, error) {
	if err := nsm.stateMigrationClient.Migrate(ctx, tenantId); err != nil {
		return nil, errors.Wrap(ErrFailedToUpdateNotificationState, err.Error())
	}

	err := nsm.validateExistAndDisableChannelIfNeeded(ctx, tenantId, notificationState)
	if err != nil {
		return nil, err
	}

	return nsm.stateClient.UpdateEmail(ctx, tenantId, notificationState)
}

func (nsm *NotificationStateManager) UpdateSlackNotificationState(ctx context.Context, tenantId string, notificationState Stater) (Stater, error) {
	if err := nsm.stateMigrationClient.Migrate(ctx, tenantId); err != nil {
		return nil, errors.Wrap(ErrFailedToUpdateNotificationState, err.Error())
	}

	if err := nsm.validateExistAndDisableChannelIfNeeded(ctx, tenantId, notificationState); err != nil {
		return nil, err
	}

	return nsm.stateClient.UpdateSlack(ctx, tenantId, notificationState)

}

func (nsm *NotificationStateManager) validateExistAndDisableChannelIfNeeded(ctx context.Context, tenantId string, notificationState Stater) error {
	exists, err := nsm.channelClient.IsExists(ctx, tenantId, notificationState)
	if err != nil {
		return err
	}
	if !exists && notificationState.IsEnabled() {
		return errors.Wrap(ErrNoSinkConfigured, "cannot enable misconfigured channel")

	}

	if !notificationState.IsEnabled() {
		err := nsm.userSubClient.DisableChannel(ctx, tenantId, notificationState.GetChanName())
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to disable channel on all of tenant %s subscriptions, error: %v", tenantId, err)
			return err
		}
	}
	return nil
}
