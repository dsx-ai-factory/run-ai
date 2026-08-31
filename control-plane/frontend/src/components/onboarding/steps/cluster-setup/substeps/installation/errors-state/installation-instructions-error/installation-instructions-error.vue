<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <onboarding-content-wrapper>
    <template #title>The installation instructions heard we were coming and made a run for it</template>
    <template #subtitle>
      Continue setting up your team and install your cluster later on the
      <a href="#" @click="close">Clusters page</a
      ><template v-if="DOCS.SUPPORT.URL"> or <a :href="DOCS.SUPPORT.URL" target="_blank">contact support</a></template
      >.
    </template>

    <template #content> <runai-svg-icon name="500" size="340" /> </template>
  </onboarding-content-wrapper>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import OnboardingContentWrapper from "@/components/onboarding/common/onboarding-content-wrapper";
import { DOCS } from "@/constants/docs";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";

export default defineComponent({
  name: "installation-instructions-error",
  components: {
    RunaiSvgIcon,
    OnboardingContentWrapper,
  },
  emits: {
    close: () => true,
  },
  setup(_, { emit }) {
    const close = () => {
      emit("close");
    };

    return {
      DOCS,
      close,
    };
  },
});
</script>

<style scoped lang="scss">
.error-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 500px;
  margin: 0 auto;
}

.error-illustration {
  margin-bottom: 48px;

  .error-icon-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 200px;
    height: 200px;
    opacity: 0.8;
    color: #f56565;
  }
}

.error-actions {
  display: flex;
  justify-content: center;
  margin-top: 32px;

  .continue-button {
    min-width: 120px;
  }
}
</style>
