/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import ProjectIndex from "@/views/projects/project-index.vue";
import ProjectEdit from "@/views/projects/project-edit.vue";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import { PROJECT_ROUTE_NAMES } from "./project.routes.names";
import type { RunaiRouteRecord } from "vue-router";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const projectRoutes: RunaiRouteRecord = {
  path: "/projects",
  component: () => import("@/views/projects/project-app.vue"),
  children: [
    {
      path: "",
      name: PROJECT_ROUTE_NAMES.PROJECT_INDEX,
      component: ProjectIndex,
      meta: {
        pageTitle: "Projects",
        resourceType: ResourceType.Project,
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.ORGANIZATION,
      },
    },
    {
      path: "new",
      alias: ["new-v2"],
      name: PROJECT_ROUTE_NAMES.PROJECT_NEW,
      component: ProjectEdit,
      meta: {
        pageTitle: "New project",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Project,
        minPermittedActions: [Action.Create],
        parentSubject: EMenuSubjectNames.ORGANIZATION,
        requiresAuth: true,
        hideVersion: true,
      },
      beforeEnter: (to, from) => {
        to.meta.backPageName = (from.name as string) || PROJECT_ROUTE_NAMES.PROJECT_INDEX;
      },
    },
    {
      path: "edit/:id",
      alias: ["edit-v2/:id"],
      name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
      component: ProjectEdit,
      meta: {
        pageTitle: "Edit project",
        fullWidth: true,
        requestToLeave: true,
        backPageName: PROJECT_ROUTE_NAMES.PROJECT_INDEX,
        resourceType: ResourceType.Project,
        minPermittedActions: [Action.Update],
        parentSubject: EMenuSubjectNames.ORGANIZATION,
        requiresAuth: true,
        hideVersion: true,
      },
    },
  ],
};
