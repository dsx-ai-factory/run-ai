<template>
  <component :is="component" />
</template>

<script lang="ts">
import { useSettingStore } from "@/stores/setting.store";
import TemplateIndex from "@/views/workload-template/workload-template-hybrid/template-index.vue";
import WorkloadTemplateIndex from "@/views/workload-template/workload-template-index.vue";
import { SettingKeys } from "@/models/setting.model";
import { defineComponent } from "vue";

export default defineComponent({
  name: "template-index-wrapper",
  computed: {
    component() {
      return useSettingStore().isFeatureEnabled(SettingKeys.EnableTemplatesV2) ? TemplateIndex : WorkloadTemplateIndex;
    },
  },
});
</script>
