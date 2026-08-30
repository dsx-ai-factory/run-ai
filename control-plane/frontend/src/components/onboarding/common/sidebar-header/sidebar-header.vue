<template>
  <div class="sidebar-header">
    <runai-distribution-logo v-bind="distributionLogoPreset" class="onboarding-distribution-logo" />
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { DISTRIBUTION_LOGO_PRESETS, RunaiDistributionLogo } from "@/components/common/runai-distribution-logo";

export default defineComponent({
  name: "sidebar-header",
  components: {
    RunaiDistributionLogo,
  },
  setup() {
    return { distributionLogoPreset: DISTRIBUTION_LOGO_PRESETS.onboardingSidebar };
  },
});
</script>

<style scoped lang="scss">
.onboarding-distribution-logo[data-profile="ce"] {
  overflow: visible;
}
</style>
