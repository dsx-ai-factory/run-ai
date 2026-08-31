<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="image-pull-credential-layout row justify-between">
    <span class="col-6 flex flex-center">
      <runai-section :title="sectionTitle">
        <slot />
      </runai-section>
    </span>

    <transition name="fade">
      <span class="col-6" v-if="showInfoBanner">
        <q-banner class="info-box">
          <div class="row no-wrap items-center">
            <q-icon name="ph-bold ph-question" size="16px" color="icon-default" />
            <div class="q-ml-md info-message">
              NVIDIA NIM requires an image that is pulled from the
              <a :href="ngcCatalog" target="_blank" rel="noopener">NGC catalog</a>. Select a secret on the left or make
              sure your administrator configured a Docker registry credential in your project or cluster. For more
              information, see
              <a :href="deployNimInference" target="_blank" rel="noopener">Deploy inference workloads with NVIDIA NIM</a
              >.
            </div>
          </div>
        </q-banner>
      </span>
    </transition>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiSection } from "@/components/common/runai-section";

import DOCS from "@/constants/docs";

export default defineComponent({
  name: "image-pull-credential-layout",
  components: {
    RunaiSection,
  },
  props: {
    sectionTitle: {
      type: String as PropType<string>,
      required: true,
    },
    showInfoBanner: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
  },
  data() {
    return {
      ngcCatalog: DOCS.EXTERNAL.NIM.NGC_CATALOG,
      deployNimInference: DOCS.EXTERNAL.NIM.DEPLOY_INFERENCE,
    };
  },
});
</script>

<style scoped lang="scss">
.image-pull-credential-layout {
  height: 136px;

  .info-box {
    background: var(--bg-secondary-default, $body-background-color);

    .info-message {
      font-size: 12px;
      font-style: normal;
      font-weight: 400;
      line-height: 20px;
    }
  }
}
</style>
