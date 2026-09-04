<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <notification-alert-bar v-if="shouldShowClusterWarning" :variant="statusWarningLevel" :message="clusterWarningMessage">
    <template #inline-actions>
      <router-link to="/clusters" v-if="shouldShowClustersLink">Clusters.</router-link>
    </template>
  </notification-alert-bar>
</template>
<script lang="ts">
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";
import { adminWarningStatuses, nonAdminWarningStatuses } from "@/models/cluster.model";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { defineComponent } from "vue";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { ClusterState, type DisplayedCluster } from "@/swagger-models/cluster-service-client";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";

import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";
import { DASHBOARD_ROUTE_NAMES } from "@/router/dashboard.routes/dashboard.routes.names";
import { SETTINGS_ROUTE_NAMES } from "@/router/settings.routes/settings.routes.names";
import { USER_ROUTES_NAMES } from "@/router/users.routes/users.routes.names";
import { SERVICE_ACCOUNTS_ROUTE_NAMES } from "@/router/service-accounts.routes/service-accounts.routes.names";
import { ROLE_ROUTE_NAMES } from "@/router/role.routes/role.routes.names";
import { AUDIT_ROUTE_NAMES } from "@/router/audit.routes/audit.routes.names";
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import { ENVIRONMENT_ROUTE_NAMES } from "@/router/environment.routes/environment.routes.names";
import { DATA_SOURCE_ROUTE_NAMES } from "@/router/data-source.routes/data-source.routes.names";
import { COMPUTE_RESOURCE_ROUTE_NAMES } from "@/router/compute-resource.routes/compute-resource.routes.names";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import { CREDENTIAL_ROUTE_NAMES } from "@/router/credential.routes/credential.routes.names";
import { ReportRoutesName } from "@/router/report.routes/report.routes.names";
import { ERROR_ROUTE_NAMES } from "@/router/error.routes/error.routes.names";

const NO_CLUSTER_STATUS_BAR_PAGES: Array<string> = [
  DASHBOARD_ROUTE_NAMES.OVERVIEW_INDEX,
  DASHBOARD_ROUTE_NAMES.RESOURCES_INDEX,
  CLUSTER_ROUTE_NAMES.CLUSTER_INDEX,
  CLUSTER_ROUTE_NAMES.CLUSTER_NEW,
  CLUSTER_ROUTE_NAMES.CLUSTER_INSTALL,
  SETTINGS_ROUTE_NAMES.SETTINGS_INDEX,
  USER_ROUTES_NAMES.USER_INDEX,
  SERVICE_ACCOUNTS_ROUTE_NAMES.SERVICE_ACCOUNTS_INDEX,
  ROLE_ROUTE_NAMES.ACCESS_RULES_INDEX,
  ROLE_ROUTE_NAMES.ACCESS_RULES_ROLE_VIEW,
  ROLE_ROUTE_NAMES.ROLE_INDEX,
  ROLE_ROUTE_NAMES.ROLE_VIEW,
  AUDIT_ROUTE_NAMES.AUDIT_INDEX,
  DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX,
  DEPARTMENT_ROUTE_NAMES.DEPARTMENT_NEW,
  DEPARTMENT_ROUTE_NAMES.DEPARTMENT_EDIT,
  DEPARTMENT_ROUTE_NAMES.DEPARTMENT_VIEW,
  PROJECT_ROUTE_NAMES.PROJECT_INDEX,
  PROJECT_ROUTE_NAMES.PROJECT_NEW,
  PROJECT_ROUTE_NAMES.PROJECT_EDIT,
  ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_INDEX,
  DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX,
  COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_INDEX,
  WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
  CREDENTIAL_ROUTE_NAMES.CREDENTIAL_INDEX,
  ReportRoutesName.ReportIndex,
  ...Object.values(ERROR_ROUTE_NAMES),
];

export default defineComponent({
  components: {
    NotificationAlertBar,
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      authStore: useAuthStore(),
      permissionStore: usePermissionStore(),
      currCluster: null as DisplayedCluster | null, // Full cluster data for operands access
    };
  },
  async created() {
    // Start cluster polling (minimal only - full cluster loaded via watcher)
    await this.loadClusters();
    intervalUtil.startInterval(EIntervalLabels.ClusterStatusBar, this.loadClusters);
  },
  computed: {
    hasClusters(): boolean {
      return this.clusterStore.clusters.length > 0;
    },
    clusterName(): string {
      return this.clusterStore.selectedCluster?.name || "";
    },
    statusWarningLevel(): ENotificationBarVariant {
      return ENotificationBarVariant.Warning;
    },
    // Use minimal data path for state
    statusState(): ClusterState | undefined {
      return this.clusterStore.selectedCluster?.state;
    },
    isAdminWarningStatus(): boolean {
      return this.statusState != undefined && adminWarningStatuses.includes(this.statusState);
    },
    isNonAdminWarningStatus(): boolean {
      return this.statusState != undefined && nonAdminWarningStatuses.includes(this.statusState);
    },
    isDashboardWarningStatus(): boolean {
      // Use full cluster data for operands access (requires Cluster Read permission)
      const prometheusServiceStatus = this.currCluster?.status?.operands?.prometheus?.unreadyThresholdCrossed ?? false;
      return (
        this.$route.path.startsWith("/dashboards") &&
        this.statusState === ClusterState.ServiceIssues &&
        prometheusServiceStatus
      );
    },
    shouldShowClusterWarning(): boolean {
      if (!this.hasClusters || NO_CLUSTER_STATUS_BAR_PAGES.includes(this.$route.name?.toString() || "")) {
        return false;
      }
      return this.authStore.isAdmin
        ? this.isAdminWarningStatus
        : this.isNonAdminWarningStatus || this.isDashboardWarningStatus;
    },
    clusterWarningMessage(): string {
      const state = this.statusState;
      if (!state || !this.clusterName) return "";
      if (this.authStore.isAdmin) {
        return this.showAdminWarning(state, this.currCluster?.status?.operands);
      } else {
        return this.showNonAdminWarning(state);
      }
    },
    shouldShowClustersLink() {
      return this.permissionStore.hasSomePermittedActions(ResourceType.Cluster, [Action.Create]);
    },
    // Check if user has permission to load full cluster data
    hasClusterReadPermission(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Cluster, Action.Read);
    },
  },

  methods: {
    async loadClusters(): Promise<void> {
      try {
        await this.clusterStore.loadClusters();
      } catch (error) {
        console.warn("Failed to load clusters", error);
      }
    },
    async loadCurrClusterDetails(clusterId: string): Promise<void> {
      try {
        this.currCluster = await clusterService.getClusterByUuid(clusterId);
      } catch (error) {
        console.warn("Failed to load full cluster data", error);
        this.currCluster = null;
      }
    },
    showAdminWarning(state: ClusterState, operands?: Record<string, { unreadyThresholdCrossed?: boolean }>): string {
      switch (state) {
        case ClusterState.ServiceIssues:
          if (operands == undefined) {
            return "";
          } else {
            return `*Some Run:ai services in cluster ${this.clusterName} are not functioning properly*. Actions may be affected, and the data may be stale. For more information, see`; //cluster router link comes here
          }
        case ClusterState.MissingPrerequisites:
          return `*Prerequisites are missing from cluster ${this.clusterName}.* Actions may be affected, and the data may be stale. For more information, see`; //cluster router link comes here
        case ClusterState.Disconnected:
        case ClusterState.WaitingToConnect:
          return `*There's no communication from cluster ${this.clusterName}.* Actions may be affected, and the data may be stale. For more information, see`; //cluster router link comes here
        default:
          return "";
      }
    },
    showNonAdminWarning(state: ClusterState): string {
      switch (state) {
        case ClusterState.ServiceIssues:
          return `*Issues were found with cluster ${this.clusterName} and some metrics aren't available right now*. You can continue using the platform as usual. If this persists, contact your administrator.`;
        case ClusterState.MissingPrerequisites:
          return `*Issues were found with cluster ${this.clusterName}*. Actions may be affected, and the data may be stale. If this persists, contact your administrator."`;
        case ClusterState.Disconnected:
        case ClusterState.WaitingToConnect:
          return `*Issues were found with cluster ${this.clusterName}*. Actions may be affected, and the data may be stale. If this persists, contact your administrator.`;
        default:
          return "";
      }
    },
  },
  watch: {
    // Watch for selected cluster changes and load full data when needed
    "clusterStore.selectedCluster": {
      async handler(newCluster) {
        if (newCluster?.uuid && this.hasClusterReadPermission) {
          await this.loadCurrClusterDetails(newCluster.uuid);
        } else {
          this.currCluster = null;
        }
      },
      immediate: true,
    },
  },
  unmounted() {
    intervalUtil.stopInterval(EIntervalLabels.ClusterStatusBar);
  },
});
</script>
