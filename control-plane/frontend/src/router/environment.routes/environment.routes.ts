import { ResourceType, Action } from "@/swagger-models/authorization-client";

import type { RunaiRouteRecord } from "vue-router";
import { ENVIRONMENT_ROUTE_NAMES } from "@/router/environment.routes/environment.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const environmentRoutes: RunaiRouteRecord = {
  path: "/environments",
  // route level code-splitting
  // this generates a separate chunk (department.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/environment/environment-app.vue"),
  children: [
    {
      path: "",
      name: ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_INDEX,
      component: () => import("@/views/environment/environment-index.vue"),
      meta: {
        pageTitle: "Environments",
        resourceType: ResourceType.Environments,
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
    },
    {
      path: "new",
      name: ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_NEW,
      component: () => import("@/views/environment/environment-new.vue"),
      meta: {
        pageTitle: "New environment",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Environments,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_INDEX;
      },
    },
    {
      path: "edit/:id",
      name: ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_EDIT,
      component: () => import("@/views/environment/environment-new.vue"),
      meta: {
        pageTitle: "Edit environment",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Environments,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_INDEX;
      },
    },
  ],
};
