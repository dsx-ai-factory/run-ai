import type { IAdminMessage } from "@/models/setting.model";
import { bffServiceApi } from "@/services/infra/client-apis/bff-service-api/bff-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import { NotificationsSettingsKeyEnum, type GetSystemNotification } from "@/swagger-models/bff-service-client";

export const notificationsSettingsService = {
  getSystemMessage,
  getAdminMessage,
  updateAdminMessage,
};

async function getSystemMessage(): Promise<IAdminMessage | undefined> {
  try {
    const response = await bffServiceApi.systemNotificationApi.getSystemNotification();
    const message = response.data as GetSystemNotification;
    if (!message || !message.msg) return undefined;
    return {
      id: message.id,
      msg: message.msg,
      enableDismiss: true,
      createdAt: message.createdAt ? new Date(message.createdAt) : new Date(),
    } as IAdminMessage;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getAdminMessage(): Promise<IAdminMessage | undefined> {
  try {
    const response = await bffServiceApi.settingsApi.getNotificationsSettingByKey(
      NotificationsSettingsKeyEnum.AdminMessage,
    );
    return response.data as IAdminMessage | undefined;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateAdminMessage(msg: string, enableDismiss: boolean): Promise<void> {
  try {
    await bffServiceApi.settingsApi.updateNotificationsSettingByKey(NotificationsSettingsKeyEnum.AdminMessage, {
      msg,
      enableDismiss,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
