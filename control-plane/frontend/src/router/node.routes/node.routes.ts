import { ResourceType, Action } from "@/swagger-models/authorization-client";

import type { RunaiRouteRecord } from "vue-router";
import { NODE_ROUTE_NAMES } from "@/router/node.routes/node.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const nodeRoutes: RunaiRouteRecord = {
  path: "/nodes",
  // route level code-splitting
  // this generates a separate chunk (node.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/node/node-app.vue"),
  children: [
    {
      path: "",
      name: NODE_ROUTE_NAMES.NODE_INDEX,
      component: () => import("@/views/node/node-app.vue"),
      meta: {
        pageTitle: "Nodes",
        resourceType: ResourceType.Nodes,
        minPermittedActions: [Action.Read],
        parentSubject: EMenuSubjectNames.RESOURCES,
        requiresAuth: true,
      },
    },
  ],
};
