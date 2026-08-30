export enum ELogsSource {
  CurrentInstance = "current",
  PreviousInstance = "previous",
}

export interface ILogParams {
  project: string;
  jobName: string;
  podName: string;
  tail: number | null;
  type: string;
  distributedFramework?: string;
  container?: string;
  externalWorkloadName?: string;
  previous?: boolean;
}

export interface IDownloadLogParams {
  project: string;
  jobName: string;
  podName: string;
}

export interface IDownloadLogClusterApiParams {
  project: string;
  jobName: string;
  type: string;
  distributedFramework?: string;
  container?: string;
  podName?: string;
  externalWorkloadName?: string;
  previous?: boolean;
}

export type LogDownloadQueryParams = {
  container?: string;
  pod?: string;
  previous?: boolean;
};
