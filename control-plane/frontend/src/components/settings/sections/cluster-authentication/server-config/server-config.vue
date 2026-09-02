<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="server-config q-pa-md">
    <div class="inner-text q-pb-sm">
      To enable cluster authentication, add the following flags to the Kubernetes API server of each cluster
    </div>
    <runai-copyable-code-block :code-with-html="serverConfigHtml" :code="code" use-colors />
    <div class="inner-text q-pt-lg text-italic">
      For more information, see
      <a :href="DOCS.AUTHENTICATION.RESEARCHER.K8S_CONFIG" target="_blank">Cluster authentication guide</a>
    </div>
  </section>
</template>

<script lang="ts">
import { type PropType, defineComponent } from "vue";
// Models
import type { ITenantServerConfigInfo } from "@/models/tenant.model";
// Components
import RunaiCopyableCodeBlock from "@/components/common/runai-copyable-code-block/runai-copyable-code-block.vue";
import { HTMLUtil } from "@/utils/html.util/html.util";
import { DOCS } from "@/constants/docs";

export default defineComponent({
  name: "server-config",
  components: {
    RunaiCopyableCodeBlock,
  },
  props: {
    serverConfig: {
      type: Object as PropType<ITenantServerConfigInfo>,
      required: true,
    },
  },
  data() {
    return {
      DOCS,
    };
  },
  computed: {
    code() {
      return HTMLUtil.HTMLToPlainText(this.serverConfigHtml);
    },
    serverConfigHtml() {
      let serverConfigHtml = `
<pre class="yaml-code" style="margin-top: -20px;overflow-x: scroll">
<span class="yaml-key">spec</span><span class="yaml-colon">:</span>
  <span class="yaml-key">containers</span><span class="yaml-colon">:</span>
  <span class="yaml-item">- </span><span class="yaml-key">command</span><span class="yaml-colon">:</span>
    <span class="yaml-value">...</span>`;

      serverConfigHtml += `
    <span class="yaml-item">- </span><span class="yaml-value">--oidc-client-id=${this.serverConfig["clientId"]}</span>
    <span class="yaml-item">- </span><span class="yaml-value">--oidc-issuer-url=${this.serverConfig["issuerUrl"]}</span>
    <span class="yaml-item">- </span><span class="yaml-value">--oidc-username-prefix=-</span>`;

      if (this.serverConfig["groupsClaim"]) {
        serverConfigHtml += `
    <span class="yaml-item">- </span><span class="yaml-value">--oidc-groups-claim=${this.serverConfig["groupsClaim"]}</span>`;
      }

      serverConfigHtml += `</pre>`;
      return serverConfigHtml;
    },
  },
});
</script>
<style scoped lang="scss">
.server-config {
  border: $black-12 1px solid;
  .inner-text {
    color: $black-87;
    font-size: 14px;
  }
}
</style>
