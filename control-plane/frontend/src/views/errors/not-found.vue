<template>
  <section class="not-found row justify-center items-center">
    <div class="column">
      <span class="title">Looks like the page you were</span>
      <span class="title">looking for ran away or doesn't exist...</span>
      <span class="subtitle q-mt-xl">
        <a href="#" class="link-clr pointer underline-on-hover" @click.prevent="$router.back()">Go back</a>
        or try navigating to
        <a href="#" class="link-clr pointer underline-on-hover" @click.prevent="redirectByPermission">{{
          landingPageTitle
        }}</a>
        instead
      </span>
    </div>
    <runai-svg-icon name="404" size="520px" />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { useAppStore } from "@/stores/app.store";
import { DASHBOARD_ROUTE_NAMES } from "@/router/dashboard.routes/dashboard.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

export default defineComponent({
  name: "not-found",
  components: {
    RunaiSvgIcon,
  },
  data() {
    return {
      appStore: useAppStore(),
    };
  },
  computed: {
    landingPageTitle(): string {
      const landingPageMap: Record<string, string> = {
        [DASHBOARD_ROUTE_NAMES.OVERVIEW_INDEX]: "Dashboards",
        [WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX]: "Workloads",
      };
      return this.redirectPathName ? landingPageMap[this.redirectPathName] : "Dashboards";
    },
    redirectPathName(): string | undefined {
      return this.appStore.getUserLandingPage();
    },
  },
  methods: {
    redirectByPermission(): void {
      const landingPage = this.appStore.getUserLandingPage();
      this.$router.push({ name: landingPage });
    },
  },
});
</script>

<style lang="scss" scoped>
.not-found {
  .title {
    font-size: 40px;
    color: $black-87;
  }

  .subtitle {
    font-size: 18px;
    color: $black-87;
  }
}
</style>
