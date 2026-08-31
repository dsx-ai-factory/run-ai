/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ref, onUnmounted, type Ref } from "vue";
import { useRoute } from "vue-router";

interface IUseTopRowOptions {
  refreshTime?: number;
}

export function useTopRow<T>(loadEntityById: CallableFunction, options: IUseTopRowOptions = {}) {
  const route = useRoute();

  const lastCreatedEntity: Ref<null | T> = ref(null);
  const refreshIntervalId: Ref<null | number> = ref(null);

  const createdEntityId: string | null = route.query.createdEntityId?.toString() || null;

  const executeCall = async () => {
    if (!createdEntityId) {
      lastCreatedEntity.value = null;
      return;
    }
    try {
      lastCreatedEntity.value = await loadEntityById(createdEntityId);
    } catch (error: unknown) {
      console.log("could not load specific entity", error);
      clearLastCreated();
    }
  };

  const clearLastCreated = () => {
    lastCreatedEntity.value = null;
    refreshIntervalId.value && window.clearInterval(refreshIntervalId.value);
  };

  executeCall();

  if (options.refreshTime) refreshIntervalId.value = window.setInterval(executeCall, options.refreshTime);
  onUnmounted(() => {
    clearLastCreated();
  });

  return { lastCreatedEntity, clearLastCreated };
}
