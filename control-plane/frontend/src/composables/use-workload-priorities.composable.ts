/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ref, computed, type Ref } from "vue";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import type { ClusterPriority } from "@/swagger-models/workloads-client";
import type { EWorkloadName } from "@/models/workload.model";
import { useClusterStore } from "@/stores/cluster.store";
import { MIN_CLUSTER_VERSION_FOR_PRIORITY_CLASS } from "@/common/version.constant";

export function useWorkloadPriorities(clusterId: string | undefined) {
  const clusterStore = useClusterStore();
  const workloadPriorities: Ref<ClusterPriority[]> = ref([]);
  const clusterIdToCheck: Ref<string | undefined> = ref(undefined);

  const targetClusterId = computed(() => clusterIdToCheck.value || clusterId);

  const supportPriorityClass = computed((): boolean => {
    if (!targetClusterId.value) return false;
    return clusterStore.isClusterVersionSufficient(targetClusterId.value, MIN_CLUSTER_VERSION_FOR_PRIORITY_CLASS);
  });

  const loadWorkloadPriorities = async (clusterId?: string): Promise<void> => {
    if (clusterId) {
      clusterIdToCheck.value = clusterId;
    }

    if (!targetClusterId.value || !supportPriorityClass.value) {
      return;
    }

    try {
      workloadPriorities.value = await workloadService.getWorkloadClusterPriorities(targetClusterId.value);
    } catch (error: unknown) {
      console.error("Failed to load workload priorities:\n", error);
      workloadPriorities.value = [];
      throw error;
    }
  };

  const loadDefaultWorkloadPriority = async (workloadName: EWorkloadName): Promise<string | null> => {
    try {
      const workloadTypeConfig = await workloadService.getDefaultWorkloadPriorityByType(workloadName);
      return workloadTypeConfig.priorityName || null;
    } catch (error: unknown) {
      console.error("Failed to load default workload priority:\n", error);
      return workloadService.getFallbackWorkloadPriorityByType(workloadName) || null;
    }
  };

  const resetPriorities = (): void => {
    workloadPriorities.value = [];
  };

  return {
    // State
    workloadPriorities,

    // Computed
    supportPriorityClass,

    // Methods
    loadWorkloadPriorities,
    loadDefaultWorkloadPriority,
    resetPriorities,
  };
}
