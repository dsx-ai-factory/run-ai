<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-system-page class="oidc-index">
    <section class="content column items-center justify-center flex-1 q-mx-xl b-gap-l">
      <h1 v-if="code" class="title">Login Successful!</h1>
      <p class="subtitle">{{ message }}</p>
      <runai-copyable-code-block v-if="code" class="code-block" :code="code" />
    </section>
  </runai-system-page>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// cmps
import { RunaiSystemPage } from "@/components/common/runai-system-page";
import { RunaiCopyableCodeBlock } from "@/components/common/runai-copyable-code-block";

export default defineComponent({
  components: {
    RunaiSystemPage,
    RunaiCopyableCodeBlock,
  },
  data() {
    return {
      message: "Copy this code and paste it into your terminal to complete the login." as string,
      code: "" as string,
    };
  },
  created() {
    // RFC 6749 §4.1.2 defines code = 1*VSCHAR (%x21-7E: printable ASCII, no space).
    const isValidCode = (v: string) => /^[\x21-\x7E]+$/.test(v);

    // Prefer the hash fragment (response_mode=fragment, used by the browser login
    // flow via pkce.util). Fall back to the query string for the CLI remote-browser
    // flow, which uses the identity-manager-built URL that does not yet set fragment
    // mode. Requiring typeof === "string" rejects Vue Router arrays (duplicate params).
    const hashParams = new URLSearchParams(this.$route.hash.slice(1));
    const rawCode =
      hashParams.get("code") ?? (typeof this.$route.query.code === "string" ? this.$route.query.code : null);
    const code: string | undefined = typeof rawCode === "string" && isValidCode(rawCode) ? rawCode : undefined;

    if (code) {
      this.code = code;
    } else {
      this.message =
        "No OIDC auth code detected, either you reached this page by mistake or you should redo the auth flow.";
    }
  },
});
</script>

<style lang="scss" scoped>
.content {
  .title {
    font-size: 40px;
    line-height: normal;
    font-weight: 400;
    color: var(--b-text-main);
  }

  .subtitle {
    font-size: 18px;
    color: var(--b-text-main);
  }

  .code-block {
    width: 100%;
    max-width: 500px;
  }
}
</style>
