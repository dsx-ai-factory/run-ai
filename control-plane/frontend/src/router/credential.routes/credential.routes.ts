import { ResourceType, Action } from "@/swagger-models/authorization-client";
import type { RunaiRouteRecord } from "vue-router";
import { CREDENTIAL_ROUTE_NAMES } from "@/router/credential.routes/credential.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const credentialRoutes: RunaiRouteRecord = {
  path: "/credentials",
  // route level code-splitting
  // this generates a separate chunk (department.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/credential/credential-app.vue"),
  children: [
    {
      path: "",
      name: CREDENTIAL_ROUTE_NAMES.CREDENTIAL_INDEX,
      component: () => import("@/views/credential/credential-index.vue"),
      meta: {
        pageTitle: "Credentials",
        resourceType: ResourceType.Credentials,
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
    },
    {
      path: "docker-registry/new",
      name: CREDENTIAL_ROUTE_NAMES.CREDENTIAL_DOCKER_REGISTRY_NEW,
      component: () => import("@/views/credential/docker-registry/docker-registry-new.vue"),
      meta: {
        pageTitle: "New credential",
        fullWidth: true,
        requestToLeave: true,
        backPageName: CREDENTIAL_ROUTE_NAMES.CREDENTIAL_INDEX,
        resourceType: ResourceType.Credentials,
        minPermittedActions: [Action.Create],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
    },
    {
      path: "generic-secret/new",
      name: CREDENTIAL_ROUTE_NAMES.CREDENTIAL_GENERIC_SECRET_NEW,
      component: () => import("@/views/credential/generic-format/generic-format.vue"),
      meta: {
        pageTitle: "New credential",
        fullWidth: true,
        requestToLeave: true,
        backPageName: CREDENTIAL_ROUTE_NAMES.CREDENTIAL_INDEX,
        resourceType: ResourceType.Credentials,
        minPermittedActions: [Action.Create],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = CREDENTIAL_ROUTE_NAMES.CREDENTIAL_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || CREDENTIAL_ROUTE_NAMES.CREDENTIAL_INDEX;
      },
    },
    {
      path: "ngc-api-key/new",
      name: CREDENTIAL_ROUTE_NAMES.CREDENTIAL_NGC_API_KEY_NEW,
      component: () => import("@/views/credential/ngc-api-key/ngc-api-key-new.vue"),
      meta: {
        pageTitle: "New credential",
        fullWidth: true,
        requestToLeave: true,
        backPageName: CREDENTIAL_ROUTE_NAMES.CREDENTIAL_INDEX,
        resourceType: ResourceType.Credentials,
        minPermittedActions: [Action.Create],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = CREDENTIAL_ROUTE_NAMES.CREDENTIAL_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || CREDENTIAL_ROUTE_NAMES.CREDENTIAL_INDEX;
      },
    },
  ],
};
