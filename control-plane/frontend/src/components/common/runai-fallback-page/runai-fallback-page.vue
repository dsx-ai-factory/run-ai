<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-fallback-page text-center">
    <div class="text-subtitle1 q-pt-xl">
      <div>For some reason this page can't be displayed right now.</div>
      <div class="q-mb-lg">
        Try <a href="#" class="refresh-anchor" @click.prevent="refresh">refreshing the page</a> or contact your admin.
      </div>
      <section class="go-back-section">
        <div class="q-mb-sm">You can also</div>
        <q-btn color="primary" @click="back">{{ btnText }}</q-btn>
      </section>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// stores
import { useAppStore } from "@/stores/app.store";

// routes
import { ENVIRONMENT_ROUTE_NAMES } from "@/router/environment.routes/environment.routes.names";
import { DATA_SOURCE_ROUTE_NAMES } from "@/router/data-source.routes/data-source.routes.names";
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import { COMPUTE_RESOURCE_ROUTE_NAMES } from "@/router/compute-resource.routes/compute-resource.routes.names";
import { CREDENTIAL_ROUTE_NAMES } from "@/router/credential.routes/credential.routes.names";
import { NODE_ROUTE_NAMES } from "@/router/node.routes/node.routes.names";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import { ROLE_ROUTE_NAMES } from "@/router/role.routes/role.routes.names";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import { USER_ROUTES_NAMES } from "@/router/users.routes/users.routes.names";
import { SERVICE_ACCOUNTS_ROUTE_NAMES } from "@/router/service-accounts.routes/service-accounts.routes.names";

const buttonRouteTexts: Record<string, string> = {
  [ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_INDEX]: "ENVIRONMENTS",
  [DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX]: "DATA SOURCES",
  [DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX]: "DEPARTMENTS",
  [PROJECT_ROUTE_NAMES.PROJECT_INDEX]: "PROJECTS",
  [WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX]: "TEMPLATES",
  [COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_INDEX]: "COMPUTE RESOURCES",
  [CREDENTIAL_ROUTE_NAMES.CREDENTIAL_INDEX]: "CREDENTIALS",
  [NODE_ROUTE_NAMES.NODE_INDEX]: "NODES",
  [NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX]: "NODE POOLS",
  [ROLE_ROUTE_NAMES.ROLE_INDEX]: "ROLES",
  [USER_ROUTES_NAMES.USER_INDEX]: "USERS",
  [SERVICE_ACCOUNTS_ROUTE_NAMES.SERVICE_ACCOUNTS_INDEX]: "SERVICE ACCOUNTS",
};

export default defineComponent({
  data() {
    return {
      appStore: useAppStore(),
    };
  },
  computed: {
    backPageName(): string {
      return this.$route.meta.backPageName as string;
    },
    btnText(): string {
      const backPage: string = this.backPageName;
      if (!backPage || !buttonRouteTexts[this.backPageName]) return "GO BACK TO A SAFE PLACE";
      return `GO TO ${buttonRouteTexts[this.backPageName]} GRID`;
    },
  },
  methods: {
    refresh(): void {
      window.location.reload();
    },
    async back(): Promise<void> {
      await this.$router.push({ name: this.backPageName });
    },
  },
});
</script>
<style lang="scss" scoped>
.runai-fallback-page {
  .refresh-anchor:hover {
    cursor: pointer;
  }
}
</style>
