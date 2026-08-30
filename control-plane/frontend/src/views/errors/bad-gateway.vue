<template>
  <section class="bad-gateway row justify-center items-center">
    <div class="column">
      <span class="title">Hm, we can't seem to get in touch </span>
      <span class="title">with the server.</span>
      <span class="subtitle q-mt-xl">
        <a href="#" class="link-clr pointer underline-on-hover" aid="bad-gateway-refresh-btn" @click.prevent="onRefresh"
          >Refresh the page</a
        >
        or contact your administrator if the issue persists.
      </span>
    </div>
    <runai-svg-icon name="502" size="420px" class="q-pa-xl" />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";

export default defineComponent({
  name: "not-found",
  components: {
    RunaiSvgIcon,
  },
  methods: {
    onRefresh(): void {
      window.location.replace(window.location.origin);
    },
  },
});
</script>

<style lang="scss" scoped>
.bad-gateway {
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
