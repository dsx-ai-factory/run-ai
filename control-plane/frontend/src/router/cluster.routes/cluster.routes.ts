import ClusterIndex from "@/views/cluster/cluster-index.vue";
import ClusterNew from "@/views/cluster/cluster-new.vue";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import type { RunaiRouteRecord } from "vue-router";
import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const clusterRoutes: RunaiRouteRecord = {
  path: "/clusters",
  // route level code-splitting
  // this generates a separate chunk (department.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/cluster/cluster-app.vue"),
  children: [
    {
      path: "",
      name: CLUSTER_ROUTE_NAMES.CLUSTER_INDEX,
      component: ClusterIndex,
      meta: {
        pageTitle: "Clusters",
        resourceType: ResourceType.Cluster,
        minPermittedActions: [Action.Create],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.RESOURCES,
      },
    },
    {
      path: "new",
      name: CLUSTER_ROUTE_NAMES.CLUSTER_NEW,
      component: ClusterNew,
      meta: {
        pageTitle: "New cluster",
        fullWidth: true,
        requestToLeave: true,
        backPageName: CLUSTER_ROUTE_NAMES.CLUSTER_INDEX,
        resourceType: ResourceType.Cluster,
        minPermittedActions: [Action.Create],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.RESOURCES,
      },
    },
    {
      path: "install/:id",
      name: CLUSTER_ROUTE_NAMES.CLUSTER_INSTALL,
      component: ClusterNew,
      meta: {
        pageTitle: "Cluster installation instructions",
        fullWidth: true,
        requestToLeave: true,
        backPageName: CLUSTER_ROUTE_NAMES.CLUSTER_INDEX,
        resourceType: ResourceType.Cluster,
        minPermittedActions: [Action.Create],
        parentSubject: EMenuSubjectNames.RESOURCES,
        requiresAuth: true,
      },
    },
  ],
};
