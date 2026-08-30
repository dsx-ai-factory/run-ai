import type { RunaiRouteRecord } from "vue-router";
import { ReportRoutesName } from "@/router/report.routes/report.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import ReportIndex from "@/views/report/report-index.vue";

export const reportRoutes: RunaiRouteRecord = {
  path: "/reports",
  component: () => import("@/views/report/report-app.vue"),
  children: [
    {
      path: "",
      name: ReportRoutesName.ReportIndex,
      component: ReportIndex,
      meta: {
        pageTitle: "Reports",
        resourceType: ResourceType.Project,
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.ANALYTICS,
      },
    },
  ],
};
