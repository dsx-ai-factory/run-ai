import { ResourceType, Action } from "@/swagger-models/authorization-client";
import type { RunaiRouteRecord } from "vue-router";
import { SERVICE_ACCOUNTS_ROUTE_NAMES } from "@/router/service-accounts.routes/service-accounts.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const serviceAccountsRoutes: RunaiRouteRecord = {
  path: "/service-accounts",
  // route level code-splitting
  // this generates a separate chunk (service-accounts.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/service-accounts/service-accounts-app.vue"),
  children: [
    {
      path: "",
      name: SERVICE_ACCOUNTS_ROUTE_NAMES.SERVICE_ACCOUNTS_INDEX,
      component: () => import("@/views/service-accounts/service-accounts-index.vue"),
      meta: {
        pageTitle: "Service accounts",
        subtitle: "(formerly applications)",
        resourceType: [ResourceType.ServiceAccount, ResourceType.Apps],
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.ACCESS,
      },
      beforeEnter: (to, from, next) => {
        to.meta.fullWidth = false;
        to.meta.supportBack = false;
        next();
      },
    },
  ],
};
