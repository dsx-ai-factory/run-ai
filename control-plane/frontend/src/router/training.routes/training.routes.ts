import TrainingNewApp from "@/views/training/creation/training-new-app.vue";
import TrainingNew from "@/views/training/creation/training-new.vue";

import TrainingAssetsEdit from "@/views/training/creation/training-assets-edit.vue";
import DistributedAssetsEdit from "@/views/training/creation/distributed-assets-edit.vue";
import TrainingAssetsApp from "@/views/training/creation/training-assets-app.vue";

import { ResourceType, Action } from "@/swagger-models/authorization-client";

import { TRAINING_ROUTE_NAMES } from "./training.routes.names";
import { START_LOCATION, type RunaiRouteRecord } from "vue-router";
import { WORKLOAD_ROUTE_NAMES } from "../workloads.routes/workloads.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";
import TrainingHybrid from "@/views/training/creation/training-hybrid.vue";
import DistributedHybrid from "@/views/training/distributed-hybrid/distributed-hybrid.vue";
import { USER_SETTINGS_ROUTE_NAMES } from "../user-settings.routes/user-settings.routes.names";

export const trainingRoutes: RunaiRouteRecord = {
  path: "/trainings",
  component: () => import("@/views/training/training-app.vue"),
  children: [
    {
      path: "new",
      component: TrainingNewApp, // router-view
      children: [
        {
          path: "",
          name: TRAINING_ROUTE_NAMES.TRAINING_NEW,
          component: TrainingNew,
          meta: {
            pageTitle: "New training",
            fullWidth: true,
            requestToLeave: true,
            backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
            resourceType: ResourceType.Trainings,
            minPermittedActions: [Action.Create, Action.Update],
            requiresAuth: true,
            parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
          },
        },
        {
          path: "assets",
          name: "training-assets-app",
          component: TrainingAssetsApp, // router-view
          children: [
            {
              path: "",
              name: TRAINING_ROUTE_NAMES.TRAINING_ASSETS_EDIT,
              component: TrainingAssetsEdit,
              meta: {
                pageTitle: "New training",
                fullWidth: true,
                requestToLeave: true,
                backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
                resourceType: ResourceType.Trainings,
                minPermittedActions: [Action.Create, Action.Update],
                requiresAuth: true,
                parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
              },
              beforeEnter: (to, from, next) => {
                // If the user tries to access the assets page directly, redirect to the new training page
                if (from === START_LOCATION) next({ name: TRAINING_ROUTE_NAMES.TRAINING_NEW });
                else next();
              },
            },
            {
              path: "distributed",
              name: TRAINING_ROUTE_NAMES.DISTRIBUTED_ASSETS_EDIT,
              component: DistributedAssetsEdit,
              meta: {
                pageTitle: "New training",
                fullWidth: true,
                requestToLeave: true,
                backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
                resourceType: ResourceType.Trainings,
                minPermittedActions: [Action.Create, Action.Update],
                requiresAuth: true,
                parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
              },
              beforeEnter: (to, from, next) => {
                // If the user tries to access the assets page directly, redirect to the new training page
                if (from === START_LOCATION) next({ name: TRAINING_ROUTE_NAMES.TRAINING_NEW });
                else next();
              },
            },
          ],
        },
        {
          path: "hybrid",
          name: "training-hybrid",
          component: TrainingHybrid,
          meta: {
            pageTitle: "New training",
            fullWidth: true,
            requestToLeave: true,
            backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
            resourceType: ResourceType.Trainings,
            minPermittedActions: [Action.Create, Action.Update],
            parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
            requiresAuth: true,
          },
          beforeEnter: (to, from, next) => {
            // If the user tries to access the assets page directly, redirect to the new workspace page
            if (from === START_LOCATION || from.name === USER_SETTINGS_ROUTE_NAMES.USER_SETTINGS_INDEX)
              next({ name: TRAINING_ROUTE_NAMES.TRAINING_NEW });
            else next();
          },
        },
        {
          path: "hybrid/distributed",
          name: TRAINING_ROUTE_NAMES.DISTRIBUTED_HYBRID,
          component: DistributedHybrid,
          meta: {
            pageTitle: "New training",
            fullWidth: true,
            requestToLeave: true,
            backPageName: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
            resourceType: ResourceType.Trainings,
            minPermittedActions: [Action.Create, Action.Update],
            requiresAuth: true,
            parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
          },
          beforeEnter: (to, from, next) => {
            // If the user tries to access the assets page directly, redirect to the new training page
            if (from === START_LOCATION || from.name === USER_SETTINGS_ROUTE_NAMES.USER_SETTINGS_INDEX)
              next({ name: TRAINING_ROUTE_NAMES.TRAINING_NEW });
            else next();
          },
        },
      ],
    },
  ],
};
