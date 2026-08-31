/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { useClusterStore } from "@/stores/cluster.store";
import { MIN_CLUSTER_VERSION_FOR_IMAGE_PULL_SECRETS } from "@/common/version.constant";
import { computed } from "vue";
export function useImagePullSecrets(clusterId: string | undefined) {
  const clusterStore = useClusterStore();

  const supportImagePullSecrets = computed(() => {
    if (!clusterId) return false;
    return clusterStore.isClusterVersionSufficient(clusterId, MIN_CLUSTER_VERSION_FOR_IMAGE_PULL_SECRETS);
  });

  return {
    supportImagePullSecrets,
  };
}
