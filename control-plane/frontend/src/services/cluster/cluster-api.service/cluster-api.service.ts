/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { clusterApiServiceOld } from "@/services/cluster/cluster-api.service/cluster-api-old.service";
import type { IDownloadLogClusterApiParams, LogDownloadQueryParams } from "@/models/logs.model";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import type { AxiosResponse } from "axios";

export const clusterApiService = {
  downloadLogsClusterApi,
};

async function downloadLogsClusterApi(
  downloadLogParams: IDownloadLogClusterApiParams,
): Promise<{ data: Blob; headers: AxiosResponse["headers"] }> {
  let endpoint: string;

  if (downloadLogParams.externalWorkloadName) {
    endpoint = `cluster-api/api/v1/${downloadLogParams.project}/workloads/external/external/${downloadLogParams.externalWorkloadName}/logs/download`;
  } else {
    const framework = workloadService.getFrameworkForUrl(downloadLogParams.type, downloadLogParams.distributedFramework);
    endpoint = `cluster-api/api/v1/${
      downloadLogParams.project
    }/workloads/${downloadLogParams.type.toLowerCase()}/${framework}/${downloadLogParams.jobName}/logs/download`;
  }

  const params: LogDownloadQueryParams = {};
  if (downloadLogParams.container) {
    params.container = downloadLogParams.container;
  }
  if (downloadLogParams.podName) {
    params.pod = downloadLogParams.podName;
  }
  if (downloadLogParams.previous) {
    params.previous = true;
  }

  return await clusterApiServiceOld.getFullData(endpoint, params, { responseType: "blob" });
}
