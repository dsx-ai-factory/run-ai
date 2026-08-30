import { ResourceType, Action } from "@/swagger-models/authorization-client";

import type { RunaiRouteRecord } from "vue-router";
import { ROLE_ROUTE_NAMES } from "@/router/role.routes/role.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const roleRoutes: RunaiRouteRecord = {
  path: "/access-rules",
  component: () => import("@/views/rbac/roles/roles-app.vue"),
  children: [
    {
      path: "",
      name: ROLE_ROUTE_NAMES.ACCESS_RULES_INDEX,
      component: () => import("@/views/rbac/access-rule/access-rule-index.vue"),
      meta: {
        requiresAuth: true,
        resourceType: ResourceType.AccessRules,
        minPermittedActions: [Action.Read],
        parentSubject: EMenuSubjectNames.ACCESS,
      },
      beforeEnter: (to, from, next) => {
        to.meta.pageTitle = "Access rules";
        to.meta.fullWidth = false;
        to.meta.supportBack = false;
        next();
      },
    },
    {
      path: "roles",
      name: ROLE_ROUTE_NAMES.ROLE_INDEX,
      component: () => import("@/views/rbac/roles/roles-index.vue"),
      meta: {
        resourceType: ResourceType.Roles,
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.ACCESS,
      },
      beforeEnter: (to, from, next) => {
        to.meta.pageTitle = "Roles";
        to.meta.fullWidth = false;
        to.meta.supportBack = false;
        next();
      },
    },
    {
      path: "review/:id",
      name: ROLE_ROUTE_NAMES.ROLE_VIEW,
      component: () => import("@/views/rbac/roles/role-review.vue"),
      meta: {
        pageTitle: "Review role",
        fullWidth: true,
        backPageName: ROLE_ROUTE_NAMES.ROLE_INDEX,
        closeIcon: true,
        readOnly: true,
        showViewOnlyBadge: true,
        resourceType: ResourceType.Roles,
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.ACCESS,
      },
    },
    {
      path: "roles/role-review/:id",
      name: ROLE_ROUTE_NAMES.ACCESS_RULES_ROLE_VIEW,
      component: () => import("@/views/rbac/roles/role-review.vue"),
      meta: {
        pageTitle: "Review role",
        fullWidth: true,
        backPageName: ROLE_ROUTE_NAMES.ACCESS_RULES_INDEX,
        closeIcon: true,
        readOnly: true,
        showViewOnlyBadge: true,
        resourceType: ResourceType.Roles,
        minPermittedActions: [Action.Read],
        parentSubject: EMenuSubjectNames.ACCESS,
        requiresAuth: true,
      },
    },
  ],
};
