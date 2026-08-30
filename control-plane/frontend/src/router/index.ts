import {
  createRouter,
  createWebHistory,
  type RouteLocationNormalized,
  type RouteRecordName,
  type RouteRecordRaw,
  type NavigationGuardNext,
} from "vue-router";

import { projectRoutes } from "./project.routes/project.routes";
import { environmentRoutes } from "./environment.routes/environment.routes";
import { departmentRoutes } from "./department.routes/department.routes";
import { workspaceRoutes } from "./workspace.routes/workspace.routes";
import { trainingRoutes } from "./training.routes/training.routes";
import { dataSourceRoutes } from "./data-source.routes/data-source.routes";
import { computeResourceRoutes } from "./compute-resource.routes/compute-resource.routes";
import { workloadTemplateRoutes } from "./workload-template.routes/workload-template.routes";
import { auditRoutes } from "./audit.routes/audit.routes";
import { nodeRoutes } from "./node.routes/node.routes";
import { nodePoolRoutes } from "./node-pool.routes/node-pool.routes";
import { credentialRoutes } from "./credential.routes/credential.routes";
import { roleRoutes } from "./role.routes/role.routes";
import { ROLE_ROUTE_NAMES } from "./role.routes/role.routes.names";
import { serviceAccountsRoutes } from "./service-accounts.routes/service-accounts.routes";
import { SERVICE_ACCOUNTS_ROUTE_NAMES } from "./service-accounts.routes/service-accounts.routes.names";
import { clusterRoutes } from "./cluster.routes/cluster.routes";
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { usersRoutes } from "./users.routes/users.routes";
import { workloadRoutes } from "@/router/workloads.routes/workloads.routes";
import { aiApplicationRoutes } from "@/router/ai-application.routes/ai-application.routes";
import { SETTINGS_ROUTE_NAMES } from "@/router/settings.routes/settings.routes.names";
import { settingsRoutes } from "@/router/settings.routes/settings.routes";
import { dashboardRoutes } from "@/router/dashboard.routes/dashboard.routes";
import { policiesRoutes } from "./policy.routes/policy.routes";
import { POLICIES_ROUTE_NAMES } from "./policy.routes/policy.routes.names";
import { inferenceRoutes } from "@/router/inference.routes/inference.routes";
import { USER_ROUTES_NAMES } from "./users.routes/users.routes.names";
import { errorRoutes } from "./error.routes/error.routes";
import { ERROR_ROUTE_NAMES } from "./error.routes/error.routes.names";
import { oidcRoutes } from "./oidc.routes/oidc.routes";
import { usePermissionStore } from "@/stores/permissions.store";
import { loginRoutes } from "@/router/login.routes/login.routes";
import { USER_SETTINGS_ROUTE_NAMES } from "@/router/user-settings.routes/user-settings.routes.names";
import { userSettingsRoutes } from "@/router/user-settings.routes/user-settings.routes";
import { useClusterStore } from "@/stores/cluster.store";
import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";
import { reportRoutes } from "@/router/report.routes/report.routes";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { noPermissionsRoutes } from "@/router/no-permissions.routes/no-permissions.routes";
import { useDataSourceStore } from "@/stores/data-source.store";
import type { ResourceType } from "@/swagger-models/authorization-client";
import {
  DS_RESOURCE_TYPE_TO_ASSET_KIND,
  type TDataSourceKinds,
  type TDataSourceResourceTypes,
} from "@/models/data-source.model";
import { dataVolumeRoutes } from "./data-volume.routes/data-volume.routes";

const supportBackPath: RouteRecordName[] = [
  ROLE_ROUTE_NAMES.ACCESS_RULES_INDEX,
  ROLE_ROUTE_NAMES.ROLE_INDEX,
  SERVICE_ACCOUNTS_ROUTE_NAMES.SERVICE_ACCOUNTS_INDEX,
  USER_ROUTES_NAMES.USER_INDEX,
  ROLE_ROUTE_NAMES.ROLE_VIEW,
  ROLE_ROUTE_NAMES.ACCESS_RULES_ROLE_VIEW,
  SETTINGS_ROUTE_NAMES.SETTINGS_INDEX,
  POLICIES_ROUTE_NAMES.POLICIES_INDEX,
  POLICIES_ROUTE_NAMES.POLICY_EDIT,
  POLICIES_ROUTE_NAMES.POLICY_VIEW,
  POLICIES_ROUTE_NAMES.POLICY_NEW,
  USER_SETTINGS_ROUTE_NAMES.USER_SETTINGS_INDEX,
];

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "",
      component: () => import("@/views/authenticated/authenticated-page.vue"),
      children: [
        projectRoutes,
        departmentRoutes,
        workspaceRoutes,
        trainingRoutes,
        nodeRoutes,
        nodePoolRoutes,
        auditRoutes,
        environmentRoutes,
        dataSourceRoutes,
        computeResourceRoutes,
        workloadTemplateRoutes,
        credentialRoutes,
        roleRoutes,
        serviceAccountsRoutes,
        clusterRoutes,
        usersRoutes,
        workloadRoutes,
        aiApplicationRoutes,
        settingsRoutes,
        dashboardRoutes,
        policiesRoutes,
        inferenceRoutes,
        userSettingsRoutes,
        reportRoutes,
        noPermissionsRoutes,
        dataVolumeRoutes,
      ] as RouteRecordRaw[],
    },
    ...errorRoutes,
    oidcRoutes,
    ...loginRoutes,
    {
      path: "/:catchAll(.*)",
      redirect: "/error/404",
    },
  ] as RouteRecordRaw[],
  scrollBehavior() {
    return { top: 0, behavior: "smooth" };
  },
});

function _updateClusterFallbackState(
  to: RouteLocationNormalized,
  appStore: ReturnType<typeof useAppStore>,
  clusterStore: ReturnType<typeof useClusterStore>,
): void {
  const { isClustersLoaded, clusters } = clusterStore;
  const isNewClusterRoute = to.name === CLUSTER_ROUTE_NAMES.CLUSTER_NEW;
  const isClusterIndexRoute = to.name === CLUSTER_ROUTE_NAMES.CLUSTER_INDEX;

  if (isClustersLoaded && !clusters.length && !isNewClusterRoute && !isClusterIndexRoute) {
    appStore.setFallback(true);
  } else if (isClustersLoaded && (clusters.length || isNewClusterRoute || isClusterIndexRoute)) {
    appStore.setFallback(false);
  }
}

function isRoutePermitted(to: RouteLocationNormalized, next: NavigationGuardNext): boolean {
  const permissionStore = usePermissionStore();
  const appStore = useAppStore();
  const authStore = useAuthStore();

  if (appStore.appLoaded && to.meta.requiresAuth && !authStore.accessToken) {
    authStore.logout();
    return false;
  }

  if (
    appStore.appLoaded &&
    to.meta.requiresAuth &&
    to.meta.minPermittedActions &&
    to.meta.resourceType &&
    !permissionStore.hasSomePermittedActions(to.meta.resourceType, to.meta.minPermittedActions)
  ) {
    next({ name: ERROR_ROUTE_NAMES.NO_ROLES });
    return false;
  }

  // this is a temp solution for branded app until we wil have the custom rbac
  if (appStore.isWhiteLabel) {
    if (to.fullPath.includes("/models")) {
      next({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX });
      return false;
    }
  }

  return true;
}

async function isAvailableByFallbackPolicy(resourceType: ResourceType, next: NavigationGuardNext) {
  const authStore = useAuthStore();

  // this following conidition is here to prevent a particular case:
  // after loading token data from storage if its value is 'falsy' what happens next is a redirect to a login page,
  // therefore in this case the function does not attempt to send the API request as it would result in a redirction to 401 page and prevent the redirect to the login page
  if (!authStore.accessToken) return true;

  const dataSouceKind: TDataSourceKinds = DS_RESOURCE_TYPE_TO_ASSET_KIND[resourceType as TDataSourceResourceTypes];
  if (!dataSouceKind) return true;

  const availableDataSources = await useDataSourceStore().getAvailableDataSources();
  if (useDataSourceStore().fallbackPolicyLoadingError) useAppStore().setFallback(true);

  if (!availableDataSources.has(dataSouceKind)) {
    next({ name: ERROR_ROUTE_NAMES.NOT_FOUND });
    return false;
  }

  return true;
}

router.beforeEach(async (to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
  const appStore = useAppStore();

  if (!isRoutePermitted(to, next)) return;

  // Only check fallback policy for single ResourceType (not arrays or null)
  const resourceType = to.meta.resourceType;
  if (resourceType && !Array.isArray(resourceType)) {
    if (!(await isAvailableByFallbackPolicy(resourceType, next))) return;
  }

  // the condition below to prevent the app from going into infinite loading in cases when we stay in the same page and only change params or query
  // such as when using this.$router.replace({ query: {...} }) or this.$router.replace({ params: {...} })
  if (from.name !== to.name) appStore.setPageLoading(true);

  return next();
});

router.afterEach((to: RouteLocationNormalized) => {
  const appStore = useAppStore();
  const clusterStore = useClusterStore();

  _updateClusterFallbackState(to, appStore, clusterStore);

  if (to.name && !supportBackPath.includes(to.name)) {
    appStore.setLastBackPath(to.path);
  }
});

export default router;
