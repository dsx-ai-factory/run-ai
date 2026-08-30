import type { IStatusColOptions } from "@/models/table.model";

export enum EWorkloadStatus {
  Activating = "Activating...",
  Stopping = "Stopping...",
  Deleting = "Deleting...",
  Creating = "",
  Active = "Active",
  Running = "Running",
  Pending = "Pending",
  Stopped = "Stopped",
  Deleted = "Deleted",

  Suspended = "Suspended",
  Completed = "Completed",
  Failed = "Failed",
  Initializing = "Initializing",
  Submitted = "Submitted",
  Terminating = "Terminating",
  Unknown = "Unknown",
  // used to detect status that came from the "workspace table" and not from the "job table"
}

export enum EWorkloadStatusColor {
  loading = "loading-clr",
  active = "active-clr",
  pending = "pending-clr",
  failed = "failed-clr",
  default = "default-clr",
  initializing = "initializing-clr",
}

export type TUiStatusConfig = {
  status: EWorkloadStatus;
  color: string;
  displayAnimation: boolean;
  showTimeUpdated: boolean;
};

export const UI_WORKSPACE_STATUS: Record<string, IStatusColOptions> = {
  [EWorkloadStatus.Creating]: {
    status: EWorkloadStatus.Creating,
    displayAnimation: true,
  },
  [EWorkloadStatus.Submitted]: {
    status: EWorkloadStatus.Creating,
    displayAnimation: true,
  },
  [EWorkloadStatus.Activating]: {
    status: EWorkloadStatus.Activating,
    displayAnimation: true,
  },
  [EWorkloadStatus.Stopping]: {
    status: EWorkloadStatus.Stopping,
    displayAnimation: true,
  },
  [EWorkloadStatus.Deleting]: {
    status: EWorkloadStatus.Deleting,
    displayAnimation: true,
  },
  [EWorkloadStatus.Deleted]: {
    status: EWorkloadStatus.Deleted,
    displayAnimation: false,
  },
  [EWorkloadStatus.Active]: {
    status: EWorkloadStatus.Active,
    color: "success",
    displayAnimation: false,
  },
  [EWorkloadStatus.Running]: {
    status: EWorkloadStatus.Running,
    color: "success",
    displayAnimation: false,
  },
  [EWorkloadStatus.Initializing]: {
    status: EWorkloadStatus.Initializing,
    displayAnimation: false,
  },
  [EWorkloadStatus.Failed]: {
    status: EWorkloadStatus.Failed,
    color: "negative",
    displayAnimation: false,
  },
  [EWorkloadStatus.Pending]: {
    status: EWorkloadStatus.Pending,
    color: "warning",
    displayAnimation: false,
  },
  [EWorkloadStatus.Stopped]: {
    status: EWorkloadStatus.Stopped,
    color: "warning",
    displayAnimation: false,
  },
  [EWorkloadStatus.Completed]: {
    status: EWorkloadStatus.Completed,
    displayAnimation: false,
  },
  [EWorkloadStatus.Terminating]: {
    status: EWorkloadStatus.Terminating,
    displayAnimation: false,
  },
  [EWorkloadStatus.Unknown]: {
    status: EWorkloadStatus.Unknown,
    displayAnimation: false,
  },
};
