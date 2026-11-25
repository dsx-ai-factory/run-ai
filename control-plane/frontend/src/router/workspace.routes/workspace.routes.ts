/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import WorkspaceNewApp from "@/views/workspace/creation/worksapce-new-app.vue";
import WorkspaceNew from "@/views/workspace/creation/workspace-new.vue";
import WorkspaceHybrid from "@/views/workspace/creation/workspace-hybrid.vue";

import { WORKSPACE_ROUTE_NAMES } from "./workspace.routes.names";
import { ResourceType, Action } from "@/swagger-models/authorization-client";

import { type RunaiRouteRecord, START_LOCATION } from "vue-router";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";
import { USER_SETTINGS_ROUTE_NAMES } from "../user-settings.routes/user-settings.routes.names";

export const workspaceRoutes: RunaiRouteRecord = {
  path: "/workspaces",
  // route level code-splitting
  // this generates a separate chunk (workspace.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/workspace/workspace-app.vue"),
  children: [
    {
      path: "new",
      component: WorkspaceNewApp, // router-view
      children: [
        {
          path: "",
          alias: "assets",
          name: WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW,
          component: WorkspaceNew,
          meta: {
            pageTitle: "New workspace",
            fullWidth: true,
            requestToLeave: true,
            backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
            resourceType: ResourceType.Workspaces,
            minPermittedActions: [Action.Create, Action.Update],
            requiresAuth: true,
            parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
          },
        },
        {
          path: "hybrid",
          name: "workspace-hybrid",
          component: WorkspaceHybrid,
          meta: {
            pageTitle: "New workspace",
            fullWidth: true,
            requestToLeave: true,
            backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
            resourceType: ResourceType.Workspaces,
            minPermittedActions: [Action.Create, Action.Update],
            parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
            requiresAuth: true,
          },
          beforeEnter: (to, from, next) => {
            if (from === START_LOCATION || from.name === USER_SETTINGS_ROUTE_NAMES.USER_SETTINGS_INDEX)
              next({ name: WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW });
            else next();
          },
        },
      ],
    },
  ],
};
