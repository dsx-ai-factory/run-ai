<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td>
    <q-btn
      flat
      :ripple="false"
      color="primary"
      class="btn-link no-padding"
      :label="displayText"
      @click.stop
      v-if="listPreview.length > 0 && isArray"
    >
      <tool-access-list-popover
        :search-placeholder="searchPlaceholderModal"
        :list-title="listTitleModal"
        :list="listToModal"
      />
    </q-btn>
    <div v-else>{{ displayText }}</div>
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type { ITableColumn } from "@/models/table.model";
import { ConnectionAuthorizationTypeEnum, type Connection } from "@/swagger-models/workloads-client";
import { InternalConnectionType } from "@/swagger-models/assets-service-client";

// services
import { previewStrings } from "@/utils/common.util";

// stores
import { useAppStore } from "@/stores/app.store";

// cmps
import { ToolAccessListPopover } from "@/components/tools/tool-access-list-popover";

const NON_EXTERNAL_URL_ACCESS_VALUE = "-";

export default defineComponent({
  components: { ToolAccessListPopover },
  props: {
    row: {
      type: Object as PropType<Connection>,
      required: true,
    },
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
  },
  data() {
    return {
      appStore: useAppStore(),
    };
  },
  computed: {
    displayText(): string {
      return Array.isArray(this.listPreview) ? previewStrings(this.listPreview) : this.listPreview;
    },
    isArray(): boolean {
      return Array.isArray(this.listPreview);
    },
    listTitleModal(): string {
      return this.row.authorizedGroups?.length ? "Groups" : "Users";
    },
    searchPlaceholderModal(): string {
      return this.row.authorizedGroups?.length ? "groups" : "users";
    },
    listToModal(): string[] {
      if (this.row.authorizedGroups) return this.row.authorizedGroups;
      if (this.row.authorizedUsers) return this.row.authorizedUsers;
      return [];
    },
    listPreview(): string | string[] {
      if (this.row.connectionType === InternalConnectionType.ServingPort) {
        return this.getListForServingPort;
      }

      if (this.row.connectionType === InternalConnectionType.ExternalUrl) {
        return this.getListForExternalUrl;
      }

      return NON_EXTERNAL_URL_ACCESS_VALUE;
    },
    getListForServingPort(): string | string[] {
      if (this.row.authorizationType === ConnectionAuthorizationTypeEnum.AuthorizedUsersOrGroups) {
        if (this.row.authorizedGroups?.length) return this.row.authorizedGroups;
        if (this.row.authorizedUsers?.length) return this.row.authorizedUsers;
      } else if (this.row.authorizationType === ConnectionAuthorizationTypeEnum.Public) {
        return "Public";
      }
      return "All authenticated users and service accounts";
    },
    getListForExternalUrl(): string | string[] {
      if (this.row.authorizedGroups?.length) return this.row.authorizedGroups;
      if (this.row.authorizedUsers?.length) return this.row.authorizedUsers;
      return "All authenticated users and service accounts";
    },
  },
});
</script>
