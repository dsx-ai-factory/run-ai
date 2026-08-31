<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="internal-server-error row justify-center items-center">
    <div class="column">
      <span class="title">Our server tripped while running</span>
      <span class="title">your request</span>
      <span class="subtitle q-mt-xl">
        <a
          href="#"
          class="link-clr pointer underline-on-hover"
          aid="internal-server-refresh-btn"
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
  name: "internal-server-error",
  components: {
    RunaiSvgIcon,
  },
  methods: {
    onGoBack(): void {
      window.history.back();
    },
  },
});
</script>

<style lang="scss" scoped>
.internal-server-error {
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
