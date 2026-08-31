/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { notificationServiceApi } from "@/services/infra/client-apis/notification-service-api/notification-service-api";
import {
  type NotificationChannel,
  type Subscription,
  type DetailedNotificationStateSlack,
  type NotificationChannelMeta,
  type NotificationType,
  type EmailChannel,
  type EmailChannelConfig,
  type NotificationStateStatus,
} from "@/swagger-models/notifications-proxy-client";
import { httpService } from "@/services/infra/https.service/http.service";

export const notificationService = {
  getEmail,
  updateEmail,
  createEmail,
  isSmtpEnabledForTenant,
  getDetailedSlackNotificationState,
  getSlackNotificationState,
  getEmailNotificationState,
  updateSlackNotificationState,
  updateEmailNotificationState,
  validateEmail,
  deleteEmail,
  getUserNotificationTypes,
  getUserSubscriptions,
  updateUserSubscription,
  createUserSubscription,
  connectSlackWorkspace,
  disconnectSlackWorkspace,
  createSlackApp,
  deleteSlackApp,
  getNotificationChannels,
  deleteNotificationChannel,
  createNotificationChannel,
};

async function getEmail(): Promise<EmailChannel> {
  try {
    const response = await notificationServiceApi.emailApi.getEmail();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateEmail(emailChannel: EmailChannel): Promise<void> {
  try {
    await notificationServiceApi.emailApi.updateEmail(emailChannel);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createEmail(emailChannel: EmailChannel): Promise<EmailChannel> {
  try {
    const response = await notificationServiceApi.emailApi.createEmail(emailChannel);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteEmail(): Promise<void> {
  try {
    await notificationServiceApi.emailApi.deleteEmail();
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createNotificationChannel(channel: NotificationChannel): Promise<void> {
  try {
    await notificationServiceApi.notificationChannelsApi.createNotificationChannel(channel);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function validateEmail(emailConfig: EmailChannelConfig): Promise<boolean> {
  try {
    const response = await notificationServiceApi.emailApi.validateEmail(emailConfig);
    return response.data.status;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function isSmtpEnabledForTenant(): Promise<boolean> {
  try {
    const response = await notificationServiceApi.emailApi.getEmailNotificationState();
    return response.data?.enabled || false;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getEmailNotificationState(): Promise<NotificationStateStatus> {
  try {
    const response = await notificationServiceApi.emailApi.getEmailNotificationState();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getSlackNotificationState(): Promise<NotificationStateStatus> {
  try {
    const response = await notificationServiceApi.slackApi.getSlackNotificationState();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getDetailedSlackNotificationState(): Promise<DetailedNotificationStateSlack> {
  try {
    const response = await notificationServiceApi.slackApi.getDetailedSlackNotificationState();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateEmailNotificationState(state: NotificationStateStatus): Promise<NotificationStateStatus> {
  try {
    const response = await notificationServiceApi.emailApi.updateEmailNotificationState(state);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateSlackNotificationState(state: NotificationStateStatus): Promise<NotificationStateStatus> {
  try {
    const response = await notificationServiceApi.slackApi.updateSlackNotificationState(state);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getUserNotificationTypes(): Promise<Array<NotificationType>> {
  try {
    const response = notificationServiceApi.notificationTypesApi.getNotificationTypes();
    return (await response).data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getUserSubscriptions(): Promise<Array<Subscription>> {
  try {
    const response = notificationServiceApi.subscriptionApi.getUserSubscriptions();
    return (await response).data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateUserSubscription(subscriptionID: string, subscription: Subscription): Promise<Subscription> {
  try {
    const response = await notificationServiceApi.subscriptionApi.updateSubscription(subscriptionID, subscription);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createUserSubscription(subscription: Subscription): Promise<Subscription> {
  try {
    const response = await notificationServiceApi.subscriptionApi.createSubscription(subscription);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function connectSlackWorkspace(): Promise<string> {
  try {
    const {
      data: { redirectUrl },
    } = await notificationServiceApi.slackApi.connectSlackWorkspace();
    if (!redirectUrl) {
      console.warn("No redirect URL received for Slack app installation.");
      throw new Error("Failed to get redirect URL for Slack app installation.");
    }

    return redirectUrl;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function disconnectSlackWorkspace(): Promise<void> {
  try {
    await notificationServiceApi.slackApi.disconnectSlackWorkspace();
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createSlackApp(accessToken: string): Promise<void> {
  try {
    await notificationServiceApi.slackApi.createSlackApp({ accessToken });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteSlackApp(): Promise<void> {
  try {
    await notificationServiceApi.slackApi.deleteSlackApp();
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getNotificationChannels(): Promise<Array<NotificationChannelMeta>> {
  try {
    const response = await notificationServiceApi.notificationChannelsApi.getNotificationChannels();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteNotificationChannel(channelName: string): Promise<void> {
  try {
    await notificationServiceApi.notificationChannelsApi.deleteNotificationChannel(channelName);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
