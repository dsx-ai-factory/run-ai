/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { ref, watch, onBeforeUnmount, type Ref } from "vue";
import { onBeforeRouteLeave, onBeforeRouteUpdate, useRouter } from "vue-router";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import { useAppStore } from "@/stores/app.store";
import { requestToLeave } from "@/services/infra/router.service/router.service";

export function useNodePoolWizardNavigation(pool: Ref<Nodepool | undefined>) {
  const app = useAppStore();
  const router = useRouter();
  const saving = ref(false);
  const leaveApproved = ref(false);

  watch(
    pool,
    (value) => {
      app.setPageLoading(false);
      app.setPageTitleOverride(value ? `Edit node pool: ${value.name}` : null);
    },
    { immediate: true },
  );
  watch(
    () => pool.value?.id,
    () => {
      leaveApproved.value = false;
    },
    { flush: "sync" },
  );
  onBeforeUnmount(() => app.setPageTitleOverride(null));

  function canLeave() {
    return !saving.value && (leaveApproved.value || !pool.value || requestToLeave());
  }
  onBeforeRouteLeave(canLeave);
  onBeforeRouteUpdate(canLeave);

  async function openStep(stepName: "scheduling" | "topology", id = pool.value?.id) {
    if (id && !saving.value)
      await router.push({ name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT, params: { id }, query: { stepName } });
  }
  async function finish(saved: Nodepool, destination: "close" | "continue" = "close") {
    leaveApproved.value = true;
    if (destination === "continue") await openStep("topology", saved.id);
    else await router.push({ name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX });
  }
  return { saving, leaveApproved, openStep, finish };
}
