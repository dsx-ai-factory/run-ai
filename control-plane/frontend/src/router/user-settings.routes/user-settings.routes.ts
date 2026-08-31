/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { RunaiRouteRecord } from "vue-router";
import UserSettingsIndex from "@/views/user-settings/user-settings-index.vue";
import { USER_SETTINGS_ROUTE_NAMES } from "@/router/user-settings.routes/user-settings.routes.names";

export const userSettingsRoutes: RunaiRouteRecord = {
  path: "/user-settings",
  component: () => import("@/views/user-settings/user-settings-app.vue"),
  children: [
    {
      path: "",
      name: USER_SETTINGS_ROUTE_NAMES.USER_SETTINGS_INDEX,
      component: UserSettingsIndex,
      meta: {
        pageTitle: "User settings",
        resourceType: null,
        minPermittedActions: [],
        fullWidth: true,
        supportBack: true,
        requiresAuth: true,
      },
    },
  ],
};
