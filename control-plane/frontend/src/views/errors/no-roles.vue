<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="not-permitted row justify-center items-center">
    <div class="column" aid="no-roles">
      <span class="title">To start using Run:ai you need</span>
      <span class="title">the magic key: permissions</span>
      <span class="subtitle q-mt-xl"> Ask your administrator for them </span>
    </div>
    <runai-svg-icon name="403" size="500px" />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// cmps
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";

// stores
import { usePermissionStore } from "@/stores/permissions.store";
import { useAuthStore } from "@/stores/auth.store";
import { useAppStore } from "@/stores/app.store";

export default defineComponent({
  components: {
    RunaiSvgIcon,
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
      authStore: useAuthStore(),
      appStore: useAppStore(),
    };
  },
  created() {
    this.appStore.setPageLoading(false);
  },
});
</script>

<style scoped lang="scss">
.not-permitted {
  background-color: $white;
  min-height: calc(100vh - 60px);
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
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
