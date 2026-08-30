export enum PREDEFINED_DATES {
  CUSTOM = "Custom",
  WORKLOADS_LIFESPAN = "Workload's lifespan",
  SINCE_FIRST_RUN = "Since first run",
  TODAY = "Today",
  YESTERDAY = "Yesterday",
  LAST_7_DAYS = "Last 7 days",
  LAST_14_DAYS = "Last 14 days",
  LAST_30_DAYS = "Last 30 days",
}

export enum PREDEFINED_HOURS {
  LAST_HOUR = "Last hour",
  LAST_3_HOURS = "Last 3 hours",
  LAST_6_HOURS = "Last 6 hours",
  LAST_24_HOURS = "Last 24 hours",
}

export enum PREDEFINED_MINUTES {
  LAST_MINUTE = "Last minute",
  LAST_5_MINUTES = "Last 5 minutes",
  LAST_15_MINUTES = "Last 15 minutes",
}

export interface PredefinedOption {
  label: string;
  daysCount?: number;
  hoursCount?: number;
  minutesCount?: number;
  endDate?: Date;
  disabled?: boolean;
}

export interface DateTimeFormatOptions {
  includeSeconds?: boolean;
  includeMilliseconds?: boolean;
}
