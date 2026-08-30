export const uniqueEmailChannelRunaiId = "runai-email";

export const workloadNotificationCategories = ["Workload status change", "Workload timeout"];

export enum NotificationChannelLabels {
  EMAIL = "Email",
  SLACK = "Slack",
}

export interface CategorizedNotificationSettings {
  [category: string]: NotificationSettings[];
}

export interface NotificationSettings {
  type: string;
  enabled: boolean;
  category: string;
  description?: string;
}
