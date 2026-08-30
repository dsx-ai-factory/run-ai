import DepartmentIndex from "@/views/department/department-index.vue";
import DepartmentEdit from "@/views/department/department-edit.vue";
import { ResourceType, Action } from "@/swagger-models/authorization-client";

import type { RunaiRouteRecord } from "vue-router";
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const departmentRoutes: RunaiRouteRecord = {
  path: "/departments",
  // route level code-splitting
  // this generates a separate chunk (department.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/department/department-app.vue"),
  children: [
    {
      path: "",
      name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX,
      component: DepartmentIndex,
      meta: {
        pageTitle: "Departments",
        resourceType: ResourceType.Department,
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.ORGANIZATION,
      },
    },
    {
      path: "new",
      alias: ["new-v2"],
      name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_NEW,
      component: DepartmentEdit,
      meta: {
        pageTitle: "New department",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.Department,
        minPermittedActions: [Action.Create],
        parentSubject: EMenuSubjectNames.ORGANIZATION,
        requiresAuth: true,
        hideVersion: true,
      },
      beforeEnter: (to, from) => {
        to.meta.backPageName = (from.name as string) || DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX;
      },
    },
    {
      path: "edit/:id",
      alias: ["edit-v2/:id"],
      name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_EDIT,
      component: DepartmentEdit,
      meta: {
        pageTitle: "Edit department",
        fullWidth: true,
        requestToLeave: true,
        backPageName: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX,
        resourceType: ResourceType.Department,
        minPermittedActions: [Action.Update],
        parentSubject: EMenuSubjectNames.ORGANIZATION,
        requiresAuth: true,
        hideVersion: true,
      },
    },
    {
      path: "view/:id",
      alias: ["view-v2/:id"],
      name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_VIEW,
      component: DepartmentEdit,
      meta: {
        pageTitle: "Review department",
        fullWidth: true,
        backPageName: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX,
        resourceType: ResourceType.Department,
        minPermittedActions: [Action.Read],
        parentSubject: EMenuSubjectNames.ORGANIZATION,
        requiresAuth: true,
        hideVersion: true,
      },
    },
  ],
};
