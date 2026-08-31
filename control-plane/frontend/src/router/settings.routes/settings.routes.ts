/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ResourceType, Action } from "@/swagger-models/authorization-client";

import SettingsIndex from "@/views/settings/settings-index.vue";
import type { RunaiRouteRecord } from "vue-router";
import { SETTINGS_ROUTE_NAMES } from "@/router/settings.routes/settings.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const settingsRoutes: RunaiRouteRecord = {
  path: "/general-settings",
  component: () => import("@/views/settings/settings-app.vue"),
  children: [
    {
      path: "",
      name: SETTINGS_ROUTE_NAMES.SETTINGS_INDEX,
      component: SettingsIndex,
      meta: {
        pageTitle: "General settings",
        resourceType: [
          ResourceType.Settings,
          ResourceType.SecuritySettings,
          ResourceType.BrandingSettings,
          ResourceType.NotificationSettings,
          ResourceType.Registries,
        ],
        minPermittedActions: [Action.Create, Action.Update, Action.Delete],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.ADMIN,
      },
      beforeEnter: (to, from, next) => {
        to.meta.fullWidth = false;
        to.meta.supportBack = false;
        next();
      },
    },
  ],
};
