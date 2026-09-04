/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { RunaiRouteRecord } from "vue-router";
import NoRoles from "@/views/errors/no-roles.vue";
import { ERROR_ROUTE_NAMES } from "@/router/error.routes/error.routes.names";

export const noPermissionsRoutes: RunaiRouteRecord = {
  path: "/no-permission",
  component: NoRoles,
  name: ERROR_ROUTE_NAMES.NO_ROLES,
  meta: {
    requiresAuth: false,
    fullWidth: true,
    minPermittedActions: null,
    resourceType: null,
    children: [],
    hideCloseButton: true,
    hideVersion: true,
  },
};
