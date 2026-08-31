<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="not-permitted row justify-center items-center" aid="no-sso-email-mapper">
    <div class="column">
      <span class="title">We're having trouble identifying your</span>
      <span class="title">account because your email is</span>
      <span class="title">incorrect or can't be found.</span>
      <span class="subtitle q-mt-xl"> Please reach out to your administrator for help. </span>
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
