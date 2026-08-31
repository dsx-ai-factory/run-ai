<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="server-error row justify-center items-center">
    <div class="column">
      <span class="title">Our server got a cluster headache</span>
      <span class="title">and can't be reached right now</span>
      <span class="subtitle q-mt-xl">
        Try again in a few seconds<span v-if="DOCS.SUPPORT.URL">
          or
          <a aid="server-error-contact-support-link" :href="DOCS.SUPPORT.URL" target="_blank" rel="noopener noreferrer"
            >contact support</a
          > </span
        >.
      </span>
    </div>
    <runai-svg-icon name="503" size="500px" />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { DOCS } from "@/constants/docs";

export default defineComponent({
  name: "server-error",
  components: {
    RunaiSvgIcon,
  },
  data() {
    return {
      DOCS,
    };
  },
});
</script>

<style lang="scss" scoped>
.server-error {
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
