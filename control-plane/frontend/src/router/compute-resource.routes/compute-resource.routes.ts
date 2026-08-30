import { ResourceType, Action } from "@/swagger-models/authorization-client";
import type { RunaiRouteRecord } from "vue-router";
import { COMPUTE_RESOURCE_ROUTE_NAMES } from "@/router/compute-resource.routes/compute-resource.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const computeResourceRoutes: RunaiRouteRecord = {
  path: "/compute-resources",
  // route level code-splitting
  // this generates a separate chunk (department.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/compute-resource/compute-resource-app.vue"),
  children: [
    {
      path: "",
      name: COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_INDEX,
      component: () => import("@/views/compute-resource/compute-resource-index.vue"),
      meta: {
        pageTitle: "Compute resources",
        resourceType: ResourceType.ComputeResources,
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
    },
    {
      path: "new",
      name: COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_NEW,
      component: () => import("@/views/compute-resource/compute-resource-new.vue"),
      meta: {
        pageTitle: "New compute resource",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.ComputeResources,
        minPermittedActions: [Action.Create],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_INDEX;
      },
    },
    {
      path: "edit/:id",
      name: COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_EDIT,
      component: () => import("@/views/compute-resource/compute-resource-new.vue"),
      meta: {
        pageTitle: "Edit compute resource",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.ComputeResources,
        minPermittedActions: [Action.Create],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_INDEX;
      },
    },
  ],
};
