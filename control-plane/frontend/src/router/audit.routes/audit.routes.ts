import { ResourceType, Action } from "@/swagger-models/authorization-client";
import type { RunaiRouteRecord } from "vue-router";
import { AUDIT_ROUTE_NAMES } from "@/router/audit.routes/audit.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const auditRoutes: RunaiRouteRecord = {
  path: "/audit",
  component: () => import("@/views/audit/audit-app.vue"),
  children: [
    {
      path: "",
      name: AUDIT_ROUTE_NAMES.AUDIT_INDEX,
      component: () => import("@/views/audit/audit-index.vue"),
      meta: {
        pageTitle: "Event history",
        resourceType: ResourceType.EventsHistory,
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.ADMIN,
      },
    },
  ],
};
