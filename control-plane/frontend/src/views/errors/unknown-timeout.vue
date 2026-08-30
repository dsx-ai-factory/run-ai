<template>
  <section class="unknown-timeout row justify-center items-center">
    <div class="column">
      <span class="title">Our server tripped while running</span>
      <span class="title">your request</span>
      <span class="subtitle q-mt-xl">
        <a
          href="#"
          class="link-clr pointer underline-on-hover"
          aid="unknown-timeout-refresh-btn"
          @click.prevent="onGoBack"
        >
          Go back
        </a>
        and try again, or reach out to your administrator for help.
      </span>
    </div>
    <runai-svg-icon name="500" size="488px" class="q-ml-xl" />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";

export default defineComponent({
  name: "unknown-timeout",
  components: {
    RunaiSvgIcon,
  },
  methods: {
    onGoBack(): void {
      this.$router.replace("/");
    },
  },
});
</script>

<style lang="scss" scoped>
.unknown-timeout {
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
