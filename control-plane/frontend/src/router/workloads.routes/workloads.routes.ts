/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Action, ResourceType } from "@/swagger-models/authorization-client";
import type { NavigationGuardNext, RouteLocationNormalized, RunaiRouteRecord } from "vue-router";
import { EWorkloadViewType, WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";
import { useSettingStore } from "@/stores/setting.store";

export const workloadRoutes: RunaiRouteRecord = {
  path: "/workloads",
  component: () => import("@/views/workload/workload-app.vue"),
  children: [
    {
      path: "",
      component: () => import("@/views/workload/workload-index.vue"),
      name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      meta: {
        pageTitle: "Workloads",
        resourceType: [ResourceType.Workloads, ResourceType.Inferences],
        requiresAuth: true,
        minPermittedActions: [Action.Read],
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      props: () => ({ viewType: EWorkloadViewType.Active }),
    },
    {
      path: "deleted",
      component: () => import("@/views/workload/workload-index.vue"),
      name: WORKLOAD_ROUTE_NAMES.WORKLOAD_DELETED_INDEX,
      meta: {
        pageTitle: "Deleted workloads",
        resourceType: [ResourceType.Workloads, ResourceType.Inferences],
        requiresAuth: true,
        minPermittedActions: [Action.Read],
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      props: () => ({ viewType: EWorkloadViewType.Deleted }),
    },
    {
      path: "new/yaml",
      component: () => import("@/views/workload/workload-yaml-new/workload-yaml-new.vue"),
      name: WORKLOAD_ROUTE_NAMES.WORKLOAD_YAML_NEW,
      meta: {
        pageTitle: "New workload via YAML",
        fullWidth: true,
        requestToLeave: true,
        backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
        resourceType: ResourceType.Workloads,
        requiresAuth: true,
        minPermittedActions: [Action.Create],
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (_: RouteLocationNormalized, __: RouteLocationNormalized, next: NavigationGuardNext) => {
        const settingStore = useSettingStore();
        if (!settingStore.isYamlSubmissionEnabled) {
          // Redirect to workloads page if feature is not available
          next({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX });
        } else {
          next();
        }
      },
    },
    {
      path: ":workloadId/structure",
      component: () => import("@/views/workload/workload-structure/workload-structure.vue"),
      name: WORKLOAD_ROUTE_NAMES.WORKLOAD_STRUCTURE,
      meta: {
        pageTitle: "Workload Structure",
        fullWidth: true,
        fullHeight: true,
        hideVersion: true,
        customHeaderLeft: true,
        backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
        resourceType: [ResourceType.Workloads, ResourceType.Inferences],
        requiresAuth: true,
        minPermittedActions: [Action.Read],
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
    },
  ],
};
