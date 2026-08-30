import { ResourceType, Action } from "@/swagger-models/authorization-client";

import type { RunaiRouteRecord } from "vue-router";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";
import { DATA_VOLUME_ROUTE_NAMES } from "./data-volume.routes.names";

export const dataVolumeRoutes: RunaiRouteRecord = {
  path: "/data-volumes",
  // route level code-splitting
  // this generates a separate chunk (department.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/data-volume/data-volume-app.vue"),
  children: [
    {
      path: "",
      name: DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_INDEX,
      component: () => import("@/views/data-volume/data-volume-index.vue"),
      meta: {
        pageTitle: "Data volumes",
        resourceType: [ResourceType.Datavolumes],
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
    },
    {
      path: "new",
      name: DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_NEW,
      component: () => import("@/views/data-volume/data-volume-new.vue"),
      meta: {
        pageTitle: "New data volume",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Datavolumes,
        minPermittedActions: [Action.Create, Action.Update],
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
        requiresAuth: true,
      },
      beforeEnter: (to, from, next) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_INDEX;
        } else {
          to.meta.backPageName = (from.name as string) || DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_INDEX;
        }
        next();
      },
    },
    {
      path: "edit/:id",
      name: DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_EDIT,
      component: () => import("@/views/data-volume/data-volume-edit.vue"),
      meta: {
        pageTitle: "Edit data volume",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Datavolumes,
        minPermittedActions: [Action.Update],
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
        requiresAuth: true,
      },
      beforeEnter: (to, from, next) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_INDEX;
        } else {
          to.meta.backPageName = (from.name as string) || DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_INDEX;
        }
        next();
      },
    },
  ],
};
