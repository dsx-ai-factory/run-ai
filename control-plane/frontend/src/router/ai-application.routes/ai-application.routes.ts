import { Action, ResourceType } from "@/swagger-models/authorization-client";
import type { RunaiRouteRecord, RouteLocationNormalized, NavigationGuardNext } from "vue-router";
import { AI_APPLICATION_ROUTE_NAMES } from "@/router/ai-application.routes/ai-application.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";
import { useClusterStore } from "@/stores/cluster.store";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";
import { EFeatureToggle } from "@/services/infra/unleash.service/unleash.model";
import {
  MIN_CLUSTER_VERSION_FOR_AI_APPLICATIONS,
  MIN_CLUSTER_VERSION_FOR_CREATING_AI_APPLICATIONS,
} from "@/common/version.constant";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { useSettingStore } from "@/stores/setting.store";

export const aiApplicationRoutes: RunaiRouteRecord = {
  path: "/ai-applications",
  component: () => import("@/views/ai-application/ai-application-app.vue"),
  children: [
    {
      path: "",
      component: () => import("@/views/ai-application/ai-application-index.vue"),
      name: AI_APPLICATION_ROUTE_NAMES.AI_APPLICATION_INDEX,
      meta: {
        pageTitle: "AI applications",
        resourceType: ResourceType.Workloads,
        requiresAuth: true,
        minPermittedActions: [Action.Read],
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: async (_to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
        const clusterStore = useClusterStore();

        if (!clusterStore.isClustersLoaded) {
          await clusterStore.loadClusters();
        }

        const isFeatureEnabled = unleashService.isFeatureEnabled(EFeatureToggle.ShowAIApplications);
        const hasCompatibleCluster = clusterStore.isAtleastOneClusterSufficientForVersion(
          MIN_CLUSTER_VERSION_FOR_AI_APPLICATIONS,
        );

        if (!isFeatureEnabled || !hasCompatibleCluster) {
          next({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX });
        } else {
          next();
        }
      },
    },
    {
      path: "new",
      component: () => import("@/views/ai-application/ai-application-new.vue"),
      name: AI_APPLICATION_ROUTE_NAMES.AI_APPLICATION_NEW,
      meta: {
        pageTitle: "New AI application",
        fullWidth: true,
        requestToLeave: true,
        backPageName: AI_APPLICATION_ROUTE_NAMES.AI_APPLICATION_INDEX,
        resourceType: ResourceType.Workloads,
        requiresAuth: true,
        minPermittedActions: [Action.Create],
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: async (_to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
        const clusterStore = useClusterStore();
        const settingStore = useSettingStore();

        const isFeatureEnabled = unleashService.isFeatureEnabled(EFeatureToggle.ShowAIApplications);
        const isSubmitEnabled = unleashService.showSubmitAIApplication();
        let isApplicationSubmissionEnabled = settingStore.isApplicationSubmissionEnabled;

        if (isApplicationSubmissionEnabled === null) {
          try {
            const applicationSubmissionSetting = await settingStore.loadApplicationSubmissionSetting();
            isApplicationSubmissionEnabled = applicationSubmissionSetting.enabled;
          } catch {
            next({ name: AI_APPLICATION_ROUTE_NAMES.AI_APPLICATION_INDEX });
            return;
          }
        }

        const hasCompatibleCluster = clusterStore.isAtleastOneClusterSufficientForVersion(
          MIN_CLUSTER_VERSION_FOR_CREATING_AI_APPLICATIONS,
        );

        if (!isFeatureEnabled || !isSubmitEnabled || isApplicationSubmissionEnabled === false || !hasCompatibleCluster) {
          next({ name: AI_APPLICATION_ROUTE_NAMES.AI_APPLICATION_INDEX });
        } else {
          next();
        }
      },
    },
  ],
};
