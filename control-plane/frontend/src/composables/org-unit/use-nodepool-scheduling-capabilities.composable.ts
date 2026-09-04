/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ref, reactive } from "vue";
import { useQuasar } from "quasar";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { NodepoolSortFilterFields } from "@/swagger-models/cluster-service-client";
import { ErrorAlert } from "@/utils/error-alert.util";

export interface INodepoolSchedulingCapabilities {
  cpuQuota: boolean;
  quota: boolean;
}

/**
 * Composable for loading and caching nodepool scheduling capabilities by cluster.
 * Used for filtering wizard steps based on what the cluster's nodepools support.
 *
 * @example
 * ```ts
 * const { loading, capabilitiesByClusterId, loadCapabilities } = useNodepoolSchedulingCapabilities();
 *
 * // Load capabilities when edit button is clicked
 * const handleEditAction = () => {
 *   if (selectedProject?.clusterId) {
 *     loadCapabilities(selectedProject.clusterId);
 *   }
 * };
 *
 * // Use in computed for filtering steps
 * const availableSteps = computed(() => {
 *   return filterOrgUnitSteps(STEPS, {
 *     nodepoolCapabilities: capabilitiesByClusterId[selectedProject.clusterId],
 *   });
 * });
 * ```
 */
export function useNodepoolSchedulingCapabilities() {
  const $q = useQuasar();
  const loading = ref(false);
  const capabilitiesByClusterId = reactive<Record<string, INodepoolSchedulingCapabilities>>({});

  /**
   * Loads nodepool scheduling capabilities for a given cluster ID.
   * Results are cached - subsequent calls with the same clusterId return immediately.
   *
   * @param clusterId - The cluster ID to load capabilities for
   */
  const loadCapabilities = async (clusterId: string): Promise<void> => {
    if (capabilitiesByClusterId[clusterId]) return;

    try {
      loading.value = true;
      const filterBy = [filterUtil.getEqualsFilterString(NodepoolSortFilterFields.ClusterId, clusterId)];
      const nodepools = await clusterService.getNodepoolsMinimal(filterBy);

      capabilitiesByClusterId[clusterId] = {
        cpuQuota: nodepools.some((nodepool) => nodepool.schedulingCapabilities?.cpuQuota === true),
        quota: nodepools.some((nodepool) => nodepool.schedulingCapabilities?.quota === true),
      };
    } catch (err) {
      const errorAlert = new ErrorAlert({
        generalMessage: ErrorAlert.failedLoadingMessage("nodepool scheduling capabilities"),
      });
      $q.notify(errorAlert.getNotification(err));

      capabilitiesByClusterId[clusterId] = {
        cpuQuota: false,
        quota: false,
      };
    } finally {
      loading.value = false;
    }
  };

  return {
    loading,
    capabilitiesByClusterId,
    loadCapabilities,
  };
}
