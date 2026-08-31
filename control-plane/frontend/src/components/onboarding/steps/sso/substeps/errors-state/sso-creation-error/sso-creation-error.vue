<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <onboarding-content-wrapper>
    <template #title>We ran into an issue. Literally.</template>
    <template #subtitle>
      To give SSO setup another try,
      <a aid="sso-refresh-page-link" href="#" @click="handleRefreshPage">refresh the page</a
      ><template v-if="DOCS.SUPPORT.URL">
        or
        <a aid="sso-contact-support-link" :href="DOCS.SUPPORT.URL" target="_blank" rel="noopener noreferrer"
          >contact support</a
        ></template
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
  name: "sso-creation-error",
  components: {
    RunaiSvgIcon,
    OnboardingContentWrapper,
  },
  emits: {
    retry: () => true,
  },
  setup() {
    const handleRefreshPage = () => {
      window.location.reload();
    };

    return {
      DOCS,
      handleRefreshPage,
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
  gap: 16px;
  margin-top: 32px;

  .retry-button {
    min-width: 120px;
  }

  .support-button {
    min-width: 120px;
  }
}
</style>
