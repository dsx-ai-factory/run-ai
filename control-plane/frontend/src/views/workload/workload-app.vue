<template>
  <section class="workload-app">
    <router-view></router-view>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

import { workloadUtil } from "@/utils/workload.util";

import { useAuthStore } from "@/stores/auth.store";
import { usePermissionStore } from "@/stores/permissions.store";

import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { useSettingStore } from "@/stores/setting.store";
import { SettingKeys } from "@/models/setting.model";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

export default defineComponent({
  name: "workload-app",
  data() {
    return {
      authStore: useAuthStore(),
      permissionStore: usePermissionStore(),
      settingStore: useSettingStore(),
    };
  },
  created() {
    this.validateRoute();
    this.applyUserCreatedWorkloadsFilter();
  },
  computed: {
    isDeletedWorkloadsViewEnabled(): boolean {
      return this.settingStore.isFeatureEnabled(SettingKeys.DeletedWorkloadsView);
    },
  },
  methods: {
    /*
     * Apply filter to show only workloads created by the current user.
     * apply this filter only for 'researchers' role - we define researchers as users who can create workloads but not projects.
     */
    applyUserCreatedWorkloadsFilter(): void {
      const shouldApplyUserFilter =
        this.permissionStore.hasPermission(ResourceType.Workloads, Action.Create) &&
        !this.permissionStore.hasPermission(ResourceType.Project, Action.Create);

      workloadUtil.applyUserCreatedWorkloadsFilter(shouldApplyUserFilter, this.authStore.currentUser.username);
    },
    validateRoute(): void {
      if (this.$route.name === WORKLOAD_ROUTE_NAMES.WORKLOAD_DELETED_INDEX && !this.isDeletedWorkloadsViewEnabled) {
        this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX });
      }
    },
  },
});
</script>

<style scoped lang="scss"></style>
