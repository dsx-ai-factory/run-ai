<template>
  <section class="not-permitted row justify-center items-center" aid="missing-claim">
    <div class="column">
      <span class="title">We'd let you in, but it looks like your</span>
      <span class="title">account isn't associated with</span>
      <span class="title">NVIDIA Run:ai.</span>
      <span class="title">Want to try another one?</span>
      <span class="subtitle q-mt-xl">
        <span class="link-clr pointer underline-on-hover" @click="$router.back">Go back</span>
        and try again, or reach out to your administrator for help.
      </span>
    </div>
    <runai-svg-icon name="no-sso-email-mapper" size="500px" />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// cmps
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";

export default defineComponent({
  components: {
    RunaiSvgIcon,
  },
});
</script>

<style scoped lang="scss">
.not-permitted {
  .title {
    font-size: 40px;
    color: $black-87;
  }

  .subtitle {
    font-size: 18px;
    color: $black-87;
  }

  .link {
    color: $primary;
    cursor: pointer;
    &:hover {
      text-decoration: underline;
    }
  }
}
</style>
