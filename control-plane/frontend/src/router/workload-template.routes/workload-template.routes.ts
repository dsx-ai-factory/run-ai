import { ResourceType, Action } from "@/swagger-models/authorization-client";

import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "./workload-template.routes.names";

import WorkloadTemplateNewApp from "@/views/workload-template/workload-template-new-app.vue";
import WorkloadTemplateNew from "@/views/workload-template/workload-template-new.vue";
import WorkloadTemplateHybridNew from "@/views/workload-template/workload-template-hybrid/template-hybrid-new.vue";
import WorkloadTemplateAssetsApp from "@/views/workload-template/workload-template-assets-app.vue";
import WorkloadTemplateTrainingHybrid from "@/views/workload-template/workload-template-hybrid/template-training-hybrid.vue";
import WorkloadTemplateWorkspaceHybrid from "@/views/workload-template/workload-template-hybrid/template-workspace-hybrid.vue";
import WorkloadTemplateDistributedHybrid from "@/views/workload-template/workload-template-hybrid/template-distributed-hybrid.vue";
import { START_LOCATION, type RunaiRouteRecord } from "vue-router";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";
import { USER_SETTINGS_ROUTE_NAMES } from "../user-settings.routes/user-settings.routes.names";
import TemplateIndexWrapper from "@/views/workload-template/template-index-wrapper.vue";
import { inferenceTemplateRoutes } from "./inference-routes/workload-template.routes";

export const workloadTemplateRoutes: RunaiRouteRecord = {
  path: "/workload-templates",
  // route level code-splitting
  // this generates a separate chunk (workload-template.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/workload-template/workload-template-app.vue"),
  children: [
    {
      path: "",
      name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
      component: TemplateIndexWrapper,
      meta: {
        pageTitle: "Templates",
        resourceType: ResourceType.Templates,
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
    },
    {
      path: "new",
      name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW_APP,
      component: WorkloadTemplateNewApp,
      children: [
        {
          path: "",
          name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW,
          component: WorkloadTemplateNew,
          meta: {
            pageTitle: "New template",
            fullWidth: true,
            requestToLeave: true,
            backPageName: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
            resourceType: ResourceType.Templates,
            minPermittedActions: [Action.Create, Action.Update],
            requiresAuth: true,
            parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
          },
        },
        {
          path: "assets",
          name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_ASSETS_APP,
          component: WorkloadTemplateAssetsApp,
        },
      ],
    },
    {
      path: "workspace",
      name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_NEW,
      component: WorkloadTemplateHybridNew,
      meta: {
        pageTitle: "New workspace template",
        fullWidth: true,
        requestToLeave: true,
        backPageName: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
        resourceType: ResourceType.Templates,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
    },
    {
      path: "workspace/hybrid/new",
      name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_HYBRID,
      component: WorkloadTemplateWorkspaceHybrid,
      meta: {
        pageTitle: "New workspace template",
        fullWidth: true,
        requestToLeave: true,
        backPageName: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
        resourceType: ResourceType.Templates,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from, next) => {
        if (from === START_LOCATION || from.name === USER_SETTINGS_ROUTE_NAMES.USER_SETTINGS_INDEX)
          next({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_NEW });
        else next();
      },
    },
    {
      path: "training",
      name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_NEW,
      component: WorkloadTemplateHybridNew,
      meta: {
        pageTitle: "New training template",
        fullWidth: true,
        requestToLeave: true,
        backPageName: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
        resourceType: ResourceType.Templates,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
    },
    {
      path: "training/hybrid/new",
      name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID,
      component: WorkloadTemplateTrainingHybrid,
      meta: {
        pageTitle: "New training template",
        fullWidth: true,
        requestToLeave: true,
        backPageName: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
        resourceType: ResourceType.Templates,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from, next) => {
        if (from === START_LOCATION || from.name === USER_SETTINGS_ROUTE_NAMES.USER_SETTINGS_INDEX)
          next({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_NEW });
        else next();
      },
    },
    {
      path: "distributed/hybrid/new",
      name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID_DISTRIBUTED,
      component: WorkloadTemplateDistributedHybrid,
      meta: {
        pageTitle: "New training template",
        fullWidth: true,
        requestToLeave: true,
        backPageName: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
        resourceType: ResourceType.Templates,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from, next) => {
        if (from === START_LOCATION || from.name === USER_SETTINGS_ROUTE_NAMES.USER_SETTINGS_INDEX)
          next({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_NEW });
        else next();
      },
    },
    {
      path: "training/hybrid/edit/:id",
      name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID_EDIT,
      component: WorkloadTemplateTrainingHybrid,
      meta: {
        pageTitle: "Edit training template",
        fullWidth: true,
        requestToLeave: true,
        backPageName: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
        resourceType: ResourceType.Templates,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from, next) => {
        if (from.name === USER_SETTINGS_ROUTE_NAMES.USER_SETTINGS_INDEX)
          next({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_NEW });
        else next();
      },
    },
    {
      path: "distributed/hybrid/edit/:id",
      name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID_DISTRIBUTED_EDIT,
      component: WorkloadTemplateDistributedHybrid,
      meta: {
        pageTitle: "Edit training template",
        fullWidth: true,
        requestToLeave: true,
        backPageName: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
        resourceType: ResourceType.Templates,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from, next) => {
        if (from.name === USER_SETTINGS_ROUTE_NAMES.USER_SETTINGS_INDEX)
          next({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID_DISTRIBUTED });
        else next();
      },
    },
    {
      path: "workspace/hybrid/edit/:id",
      name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_HYBRID_EDIT,
      component: WorkloadTemplateWorkspaceHybrid,
      meta: {
        pageTitle: "Edit workspace template",
        fullWidth: true,
        requestToLeave: true,
        backPageName: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
        resourceType: ResourceType.Templates,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from, next) => {
        if (from.name === USER_SETTINGS_ROUTE_NAMES.USER_SETTINGS_INDEX)
          next({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_NEW });
        else next();
      },
    },

    ...inferenceTemplateRoutes,
  ],
};
