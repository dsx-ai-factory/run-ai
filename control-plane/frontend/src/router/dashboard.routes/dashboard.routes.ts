/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ResourceType, Action } from "@/swagger-models/authorization-client";
import type { RunaiRouteRecord } from "vue-router";
import OverviewApp from "@/views/dashboard-v2/overview/overview-app.vue";
import QuotaManagementApp from "@/views/dashboard-v2/quota-management/quota-management-app.vue";
import { DASHBOARD_ROUTE_NAMES } from "@/router/dashboard.routes/dashboard.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const dashboardRoutes: RunaiRouteRecord = {
  path: "/dashboards",
  // route level code-splitting
  // this generates a separate chunk (department.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/dashboard/dashboard-app.vue"),
  children: [
    {
      path: "now",
      name: DASHBOARD_ROUTE_NAMES.OVERVIEW_INDEX,
      component: OverviewApp,
      meta: {
        pageTitle: "Overview",
        anyOfResourceTypes: [
          ResourceType.Cluster,
          ResourceType.Nodes,
          ResourceType.Department,
          ResourceType.Project,
          ResourceType.Workloads,
          ResourceType.Nodepools,
        ],
        anyOfResourceTypesAction: Action.Read,
        hideVersion: false,
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.ANALYTICS,
      },
    },
    {
      path: "resources",
      name: DASHBOARD_ROUTE_NAMES.RESOURCES_INDEX,
      component: QuotaManagementApp,
      meta: {
        pageTitle: "Quota management",
        resourceType: ResourceType.Cluster,
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.ANALYTICS,
      },
    },
  ],
};
