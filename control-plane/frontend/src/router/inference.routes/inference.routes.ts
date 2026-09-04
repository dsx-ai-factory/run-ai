/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ResourceType, Action } from "@/swagger-models/authorization-client";

import {
  type RunaiRouteRecord,
  type RouteLocationNormalized,
  type NavigationGuardNext,
  START_LOCATION,
} from "vue-router";
import { INFERENCE_ROUTE_NAMES } from "./inference.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

import { useSettingStore } from "@/stores/setting.store";
import { SettingKeys } from "@/models/setting.model";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const inferenceRoutes: RunaiRouteRecord = {
  path: "/inferences/new",
  component: () => import("@/views/inference/inference-app.vue"),
  children: [
    {
      path: "",
      name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW,
      component: () => import("@/views/inference/creation/inference-new"),
      meta: {
        pageTitle: "New inference",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Inferences,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
        backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      },
    },
    {
      path: "assets",
      name: INFERENCE_ROUTE_NAMES.INFERENCE_ASSETS_EDIT,
      component: () => import("@/views/inference/creation/inference-assets-based.vue"),
      meta: {
        pageTitle: "New inference",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Inferences,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
        backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      },
      beforeEnter: (to, from, next) => {
        // If the user tries to access the assets page directly, redirect to the new training page
        if (from === START_LOCATION) next({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
        else next();
      },
    },

    {
      path: "hybrid",
      name: INFERENCE_ROUTE_NAMES.INFERENCE_FLAT_HYBRID,
      component: () => import("@/views/inference/creation/inference-hybrid.vue"),
      meta: {
        pageTitle: "New inference",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Inferences,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
        backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      },
      beforeEnter: (to, from, next) => {
        // If the user tries to access the assets page directly, redirect to the new training page
        if (from === START_LOCATION) next({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
        else next();
      },
    },
    {
      path: "update",
      name: INFERENCE_ROUTE_NAMES.INFERENCE_FLAT_UPDATE,
      component: () => import("@/views/inference/creation/inference-update.vue"),
      meta: {
        pageTitle: "Update inference",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Inferences,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
        backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      },
    },
    {
      path: "vllm",
      name: INFERENCE_ROUTE_NAMES.INFERENCE_VLLM_EDIT,
      component: () => import("@/views/inference/creation/inference-hugging-face-based.vue"),
      meta: {
        pageTitle: "New inference",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Inferences,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
        backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      },
      beforeEnter(to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) {
        if (from === START_LOCATION) {
          next({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
          return;
        }
        const settingStore = useSettingStore();
        if (settingStore.isFeatureEnabled(SettingKeys.EnableModelCatalog)) {
          next();
        } else {
          next({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
        }
      },
    },
    {
      path: "tgi",
      name: INFERENCE_ROUTE_NAMES.INFERENCE_TGI_EDIT,
      component: () => import("@/views/inference/creation/inference-hugging-face-based.vue"),
      meta: {
        pageTitle: "New inference",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Inferences,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
        backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      },
      beforeEnter(to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) {
        if (from === START_LOCATION) {
          next({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
          return;
        }
        const settingStore = useSettingStore();
        if (settingStore.isFeatureEnabled(SettingKeys.EnableModelCatalog)) {
          next();
        } else {
          next({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
        }
      },
    },
    {
      path: "nim",
      name: INFERENCE_ROUTE_NAMES.INFERENCE_NIM_EDIT,
      component: () => import("@/views/inference/creation/inference-nim-based.vue"),
      meta: {
        pageTitle: "New inference",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Inferences,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
        backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      },
      beforeEnter(to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) {
        if (from === START_LOCATION) {
          next({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
          return;
        }
        const settingStore = useSettingStore();
        if (settingStore.isFeatureEnabled(SettingKeys.EnableNIM)) {
          next();
        } else {
          next({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
        }
      },
    },
    {
      path: "distributed",
      name: INFERENCE_ROUTE_NAMES.INFERENCE_DISTRIBUTED_EDIT,
      component: () => import("@/views/inference/creation/inference-distributed/inference-distributed.vue"),
      meta: {
        pageTitle: "New inference",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Inferences,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
        backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      },
      beforeEnter(to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) {
        if (from === START_LOCATION) next({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
        else next();
      },
    },
  ],
};
