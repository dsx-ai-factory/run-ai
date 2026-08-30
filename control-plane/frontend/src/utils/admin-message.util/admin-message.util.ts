import type { IAdminMessage } from "@/models/setting.model";
import { storageUtil } from "../storage.util";
import { STORAGE_KEYS } from "@/models/storage.model";

export const adminMessageUtil = {
  getAdminMessage,
  dismissAdminMessage,
  getSystemMessage,
  dismissSystemMessage,
};

function getAdminMessage(adminMessage: IAdminMessage): null | IAdminMessage {
  const savedMessageId: string | null | undefined = storageUtil.get(STORAGE_KEYS.ADMIN_MESSAGE_ID);
  const dismissedMessageId: string | null | undefined = storageUtil.get(STORAGE_KEYS.DISMISSED_ADMIN_MESSAGE_ID);
  let msg = null as IAdminMessage | null;
  if ((adminMessage.enableDismiss && adminMessage.id === dismissedMessageId) || !adminMessage.msg) return msg;
  if (dismissedMessageId !== adminMessage.id) storageUtil.remove(STORAGE_KEYS.DISMISSED_ADMIN_MESSAGE_ID);
  if (!savedMessageId || savedMessageId !== adminMessage.id) {
    msg = adminMessage;
    storageUtil.save(STORAGE_KEYS.ADMIN_MESSAGE_ID, adminMessage.id);
  }
  return msg;
}

function dismissAdminMessage(adminMessageId: string): void {
  storageUtil.save(STORAGE_KEYS.DISMISSED_ADMIN_MESSAGE_ID, adminMessageId);
}

function getSystemMessage(message: IAdminMessage): null | IAdminMessage {
  if (!message.msg) return null;
  const dismissedMessageId: string | null | undefined = storageUtil.get(STORAGE_KEYS.DISMISSED_SYSTEM_MESSAGE_ID);
  if (message.id === dismissedMessageId) return null;
  if (dismissedMessageId !== message.id) storageUtil.remove(STORAGE_KEYS.DISMISSED_SYSTEM_MESSAGE_ID);
  return message;
}

function dismissSystemMessage(adminMessageId: string): void {
  storageUtil.save(STORAGE_KEYS.DISMISSED_SYSTEM_MESSAGE_ID, adminMessageId);
}
