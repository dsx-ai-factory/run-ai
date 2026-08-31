<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="client-config q-pa-md">
    <div class="inner-text q-pb-md">
      To grant access to a specific cluster, add the following to each user's Kubernetes client configuration file
      <runai-tooltip
        aid="tooltip-client-config"
        tooltip-text="To grant access to multiple clusters, add the following to each context within the Kubernetes client configurations file (kubeconfig) "
        width="400px"
        tooltip-position="right"
      />
    </div>
    <runai-copyable-code-block :code-with-html="clientConfigHtml" :code="code" use-colors />
    <div class="inner-text q-pt-lg text-italic">
      For more information, see
      <a :href="DOCS.AUTHENTICATION.RESEARCHER.CLI_ACCESS" target="_blank">CLI access documentation</a>
    </div>
  </section>
</template>

<script lang="ts">
import { type PropType, defineComponent } from "vue";
// Models
import type { ITenantClientConfigInfo } from "@/models/tenant.model";
// Components
import RunaiCopyableCodeBlock from "@/components/common/runai-copyable-code-block/runai-copyable-code-block.vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { HTMLUtil } from "@/utils/html.util/html.util";
import { DOCS } from "@/constants/docs";

export default defineComponent({
  name: "client-config",
  components: {
    RunaiTooltip,
    RunaiCopyableCodeBlock,
  },
  props: {
    clientConfig: {
      type: Object as PropType<ITenantClientConfigInfo>,
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
      return HTMLUtil.HTMLToPlainText(this.clientConfigHtml);
    },
    clientConfigHtml() {
      let clientConfigHtml = `
<pre class="yaml-code" style="margin-top: -20px;overflow-x: scroll">
<span class="yaml-item">- </span><span class="yaml-key">name</span><span class="yaml-colon">: </span><span class="yaml-value">runai-authenticated-user</span>
  <span class="yaml-key">user</span><span class="yaml-colon">: </span>
    <span class="yaml-key">auth-provider</span><span class="yaml-colon">: </span>
      <span class="yaml-key">config</span><span class="yaml-colon">: </span>`;

      if (this.clientConfig["airgapped"]) {
        clientConfigHtml += `
        <span class="yaml-key">airgapped</span><span class="yaml-colon">: </span><span class="yaml-value">"true"</span>`;
      }

      clientConfigHtml += `
        <span class="yaml-key">auth-flow</span><span class="yaml-colon">: </span><span class="yaml-value">${this.clientConfig["authFlow"]}</span>
        <span class="yaml-key">realm<span class="yaml-colon">: </span></span><span class="yaml-value">${this.clientConfig["realm"]}</span>
        <span class="yaml-key">client-id<span class="yaml-colon">: </span></span><span class="yaml-value">${this.clientConfig["clientId"]}</span>`;

      clientConfigHtml += `
        <span class="yaml-key">idp-issuer-url<span class="yaml-colon">: </span></span><span class="yaml-value">${this.clientConfig["idpIssuerUrl"]}</span>`;

      if (this.clientConfig["redirectUri"]) {
        clientConfigHtml += `
        <span class="yaml-key">redirect-uri</span><span class="yaml-colon">: </span><span class="yaml-value">${this.clientConfig["redirectUri"]}</span>`;
      }

      clientConfigHtml += `
      <span class="yaml-key">name</span><span class="yaml-colon">: </span><span class="yaml-value">oidc</span></pre>`;
      return clientConfigHtml;
    },
  },
});
</script>
<style scoped lang="scss">
.client-config {
  border: $black-12 1px solid;
  .inner-text {
    color: $black-87;
    font-size: 14px;
  }
}
</style>
