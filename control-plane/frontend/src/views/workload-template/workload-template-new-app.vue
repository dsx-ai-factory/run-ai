<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <router-view />
</template>

<script lang="ts">
import { defineComponent } from "vue";

// store
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";

import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import { useSettingStore } from "@/stores/setting.store";
import { SettingKeys } from "@/models/setting.model";

export default defineComponent({
  data() {
    return {
      workloadTemplateStore: useWorkloadTemplateStore(),
    };
  },
  created() {
    if (!this.$router.options.history.state.back?.toString().includes("/workload-templates")) return;
    this.workloadTemplateStore.resetCreateWorkloadTemplateCreationData();
    if (
      ![
        WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW,
        WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_NEW,
        WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_NEW,
      ].includes(this.$route.name as WORKLOAD_TEMPLATE_ROUTE_NAMES) &&
      this.$route.query.fromCopyId === undefined
    ) {
      if (useSettingStore().isFeatureEnabled(SettingKeys.EnableTemplatesV2)) {
        this.$router.replace({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX });
      } else {
        this.$router.replace({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW });
      }
    }
  },
});
</script>
