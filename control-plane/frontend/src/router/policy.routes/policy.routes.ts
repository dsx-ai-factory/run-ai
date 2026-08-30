import { ResourceType, Action } from "@/swagger-models/authorization-client";

import type { RunaiRouteRecord } from "vue-router";
import { POLICIES_ROUTE_NAMES } from "@/router/policy.routes/policy.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const policiesRoutes: RunaiRouteRecord = {
  path: "/policies",
  // route level code-splitting
  // this generates a separate chunk (department.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/policy/policy-app.vue"),
  children: [
    {
      path: "",
      name: POLICIES_ROUTE_NAMES.POLICIES_INDEX,
      component: () => import("@/views/policy/policy-index.vue"),
      meta: {
        pageTitle: "Workload policies",
        resourceType: ResourceType.Policies,
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.POLICIES,
      },
      beforeEnter: (to, from, next) => {
        to.meta.fullWidth = false;
        to.meta.supportBack = false;
        next();
      },
    },
    {
      path: "new",
      name: POLICIES_ROUTE_NAMES.POLICY_NEW,
      component: () => import("@/views/policy/policy-edit.vue"),
      meta: {
        pageTitle: "New policy",
        fullWidth: true,
        requestToLeave: true,
        backPageName: POLICIES_ROUTE_NAMES.POLICIES_INDEX,
        resourceType: ResourceType.Policies,
        minPermittedActions: [Action.Create],
        closeIcon: true,
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.POLICIES,
      },
    },
    {
      path: "edit/:id",
      name: POLICIES_ROUTE_NAMES.POLICY_EDIT,
      component: () => import("@/views/policy/policy-edit.vue"),
      meta: {
        pageTitle: "Edit policy",
        fullWidth: true,
        requestToLeave: true,
        backPageName: POLICIES_ROUTE_NAMES.POLICIES_INDEX,
        resourceType: ResourceType.Policies,
        minPermittedActions: [Action.Update],
        closeIcon: true,
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.POLICIES,
      },
    },
    {
      path: "view/:id",
      name: POLICIES_ROUTE_NAMES.POLICY_VIEW,
      component: () => import("@/views/policy/policy-view.vue"),
      meta: {
        pageTitle: "View policy",
        fullWidth: true,
        requestToLeave: false,
        backPageName: POLICIES_ROUTE_NAMES.POLICIES_INDEX,
        resourceType: ResourceType.Policies,
        minPermittedActions: [Action.Read],
        closeIcon: true,
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.POLICIES,
        showViewOnlyBadge: true,
      },
    },
  ],
};
