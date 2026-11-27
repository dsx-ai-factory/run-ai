<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <name-col class="template-name-col" :row="row" :column="column" :icon="icon">
    <template #name>
      <runai-tooltip-wrapper
        ref="nameRef"
        tag="span"
        class="template-name-text"
        max-width="300px"
        :tooltip-text="name"
        :display-tooltip="isTruncated"
      >
        {{ name }}
      </runai-tooltip-wrapper>
    </template>
    <runai-tooltip
      v-if="infoMessage"
      class="q-ml-sm info-icon"
      aid="template-info-icon"
      icon="ph-bold ph-info"
      color="warning"
      width="300px"
      tooltip-position="right"
      :tooltip-text="infoMessage"
      @click.stop
    />
  </name-col>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from "vue";

// cmps
import NameCol from "./name-col.vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";

// models
import type { ITableColumn } from "@/models/table.model";
import { Scope } from "@/swagger-models/assets-service-client";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import type { SupersetTemplate } from "@/swagger-models/workloads-client";

// stores / services / utils
import { useAuthStore } from "@/stores/auth.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { getFieldContent } from "@/utils/table-format.util";
import { templatesUtil } from "@/utils/templates.util/templates.util";

const props = defineProps<{
  row: SupersetTemplate;
  column: ITableColumn;
  icon?: string;
}>();

const authStore = useAuthStore();
const permissionStore = usePermissionStore();
const workloadTemplateStore = useWorkloadTemplateStore();

const nameRef = ref<InstanceType<typeof RunaiTooltipWrapper> | null>(null);
const isTruncated = ref(false);

const name = computed((): string => getFieldContent(props.column, props.row));

// Same rule as the template index edit action. System templates can't be edited from the UI.
// Expands the permitted scopes, so only evaluate it for templates that actually show a message.
const canEdit = computed((): boolean => {
  if (props.row.meta.scopeType === Scope.System) return false;
  if (!permissionStore.hasPermission(ResourceType.Templates, Action.Update)) return false;
  const permittedScopes = orgTreeService.getExpandedPermittedScopes(
    workloadTemplateStore.actionPermissionsByScopes[Action.Update] || {},
    authStore.orgUnitList,
  );
  return templatesUtil.isPermittedTemplate(props.row.meta, permittedScopes, String(authStore.currentUser.tenantId));
});

const infoMessage = computed((): string | undefined =>
  templatesUtil.isTgiTemplate(props.row) ? templatesUtil.getTgiInfoMessage(canEdit.value) : undefined,
);

// The name is cut with an ellipsis when it doesn't fit, show the full name in a tooltip only then.
const updateIsTruncated = async (): Promise<void> => {
  await nextTick();
  const el = nameRef.value?.$el as HTMLElement | undefined;
  isTruncated.value = !!el && el.scrollWidth > el.clientWidth;
};

onMounted(updateIsTruncated);
watch(name, updateIsTruncated);
</script>

<style lang="scss" scoped>
.template-name-col {
  max-width: 400px;

  .template-name-text {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  // The deprecation icon must stay visible however long the name is.
  .info-icon {
    flex-shrink: 0;
  }
}
</style>
