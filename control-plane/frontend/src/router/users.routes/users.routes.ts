/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ResourceType, Action } from "@/swagger-models/authorization-client";

import type { RunaiRouteRecord } from "vue-router";
import { USER_ROUTES_NAMES } from "@/router/users.routes/users.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const usersRoutes: RunaiRouteRecord = {
  path: "/users",
  component: () => import("@/views/users/users-index.vue"),
  name: USER_ROUTES_NAMES.USER_INDEX,
  meta: {
    pageTitle: "Users",
    resourceType: ResourceType.Users,
    minPermittedActions: [Action.Read],
    requiresAuth: true,
    parentSubject: EMenuSubjectNames.ACCESS,
  },
  beforeEnter: (to, from, next) => {
    to.meta.fullWidth = false;
    to.meta.supportBack = false;
    next();
  },
};
