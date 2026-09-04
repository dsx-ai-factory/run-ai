/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// services
import { workloadService } from "./workload.service";

// types
import type { ICanSubmitWorkload } from "@/models/workload.model";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";

// utils
import { clusterUtil } from "@/utils/cluster.util";

// constant
import { errorMessages } from "@/common/error-message.constant";

export const workloadCreateService = {
  isSupportedKnative,
  isClusterConnected,
  getDisabledClusterReasons,
};

/**
 * Checks whether Knative is installed in the given cluster.
 *
 * @param {ClustersMinimalData} cluster - The cluster to check for Knative support.
 * @returns {string | undefined} - Returns an error message if Knative is not installed,
 *                                 otherwise returns `undefined` if Knative is installed.
 */
function isSupportedKnative(cluster: ClustersMinimalData): string | undefined {
  if (clusterUtil.isKnativeInstalledInCluster(cluster)) return undefined;

  return errorMessages.KNATIVE_IS_NOT_INSTALLED_SELECTED_CLUSTER;
}

/**
 * Checks if the given cluster is connected and ready for workload submission.
 *
 * @param {ClustersMinimalData} cluster - The cluster to check for connectivity.
 * @returns {undefined | string} - Returns an error message if the cluster is not connected,
 *                                 otherwise returns `undefined` if the cluster is connected.
 */
function isClusterConnected(cluster: ClustersMinimalData): undefined | string {
  const canSubmitWorkload: ICanSubmitWorkload | null = workloadService.canSubmitWorkload(cluster);
  return canSubmitWorkload ? errorMessages.CLUSTER_NOT_CONNECTED : undefined;
}

/**
 * Retrieves reasons why specific clusters are disabled.
 *
 * @param {Array<ClustersMinimalData>} clusters - List of clusters to check.
 * @param {boolean} [knativeRequired=false] - If `true`, checks if Knative is supported in the cluster.
 * @returns {Record<string, string>}
 *   - A record where:
 *     - The key is the cluster UUID.
 *     - The value is a string explaining why the cluster is disabled.
 */
function getDisabledClusterReasons(
  clusters: Array<ClustersMinimalData>,
  knativeRequired = false,
): Record<string, string> {
  const disabledClusterOptions: Record<string, string> = {};

  for (const cluster of clusters) {
    if (knativeRequired) {
      const knativeNotSupportedDescribe = isSupportedKnative(cluster);
      if (knativeNotSupportedDescribe) {
        disabledClusterOptions[cluster.uuid] = knativeNotSupportedDescribe;
      }
    }
  }

  return disabledClusterOptions;
}
