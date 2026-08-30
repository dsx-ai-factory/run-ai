import { ResourceType, Action } from "@/swagger-models/authorization-client";

import type { RunaiRouteRecord } from "vue-router";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

import { NodePoolIndex } from "@/views/node-pools";

export const nodePoolRoutes: RunaiRouteRecord = {
  path: "/node-pools",
  // route level code-splitting
  // this generates a separate chunk (node-pool.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/node-pools/node-pool-app.vue"),
  children: [
    {
      path: "",
      name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX,
      component: NodePoolIndex,
      meta: {
        pageTitle: "Node pools",
        resourceType: ResourceType.Nodepools,
        minPermittedActions: [Action.Read],
        parentSubject: EMenuSubjectNames.RESOURCES,
        requiresAuth: true,
      },
    },
    {
      path: "new",
      name: NODE_POOL_ROUTE_NAMES.NODE_POOL_NEW,
      component: () => import("@/views/node-pools/node-pool-edit/node-pool-edit.vue"),
      meta: {
        location: "node-pools/new",
        backPageName: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX,
        pageTitle: "New node pool",
        resourceType: ResourceType.Nodepools,
        minPermittedActions: [Action.Create],
        parentSubject: EMenuSubjectNames.RESOURCES,
        requiresAuth: true,
        fullWidth: true,
        requestToLeave: true,
      },
    },
    {
      path: "edit/:id",
      name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT,
      component: () => import("@/views/node-pools/node-pool-edit/node-pool-edit.vue"),
      meta: {
        location: "node-pools/edit/:id",
        backPageName: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX,
        pageTitle: "Edit node pool",
        resourceType: ResourceType.Nodepools,
        minPermittedActions: [Action.Update],
        parentSubject: EMenuSubjectNames.RESOURCES,
        requiresAuth: true,
        fullWidth: true,
        requestToLeave: true,
      },
    },
  ],
};
