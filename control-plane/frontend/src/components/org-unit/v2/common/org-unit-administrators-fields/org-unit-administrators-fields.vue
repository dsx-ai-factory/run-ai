<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="org-unit-administrators-fields">
    <runai-section-title flat class="administrator-title" :title="title" :tooltip-text="tooltip" />
    <div v-if="administrator" class="administrator-rule">
      <runai-btn
        icon-button
        aid="remove-administrator"
        icon="ph-bold ph-x"
        size="sm"
        tooltip="Remove administrator"
        class="remove-administrator"
        aria-label="Remove administrator"
        :disable="disable"
        @click="administrator = false"
      />
      <runai-select
        v-model="subject"
        :options="subjectOptions"
        :rules="[requiredSubject]"
        hide-bottom-space
        label="Subject"
        class="subject-select"
        :disable="disable"
        @update:model-value="subjects = []"
      />
      <subject-addable-multi-select
        v-if="subject && subject.value !== SubjectTypeExtended.AllAuthenticatedUsers"
        :subject-type="subject"
        :disable="disable"
        :selected-subjects-options="subjects"
        @update:selected-subjects="subjects = $event"
        @add-subject="subjects = [...subjects, $event]"
      />
      <runai-select
        v-model="role"
        :options="roleOptions"
        :rules="[validateRole]"
        :loading="rolesLoading"
        :disable="disable || !canAssignRoles || rolesLoading || !!rolesError"
        label="Role"
        class="role-select"
      />
      <p v-if="rolesError" role="alert" class="text-negative">
        {{ rolesError }} <runai-btn aid="retry-administrator-roles" flat label="Retry" @click="loadRoleOptions" />
      </p>
    </div>
    <runai-btn
      v-else
      flat
      color="primary"
      icon="ph-bold ph-plus"
      label="Administrator"
      class="add-administrator"
      aid="add-administrator"
      :disable="disable || !canAssignRoles"
      @click="administrator = true"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useAccessRuleRoles } from "@/composables/node-pool/use-access-rule-roles.composable";
import { useAuthStore } from "@/stores/auth.store/index";
import { errorMessages } from "@/common/error-message.constant";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import type { ScopeType, SubjectType } from "@/swagger-models/authorization-client/index";
import { SubjectTypeExtended } from "@/swagger-models/authorization-client/index";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import { RunaiSectionTitle } from "@/components/common/runai-section-title/index";
import { RunaiSelect } from "@/components/common/runai-select/index";
import { SubjectAddableMultiSelect } from "@/components/rbac/common/subject-addable-multi-select/index";
import { subjectTypeOptions, subjectTypeOptionsWithSSO } from "@/models/access-rule.model/index";
import type { ISelectOption } from "@/models/global.model";

export interface IAdministratorsScope {
  scopeId: string;
  scopeType: ScopeType;
  clusterId: string;
}

export interface IAdministratorsValidation {
  valid: boolean;
  message?: string;
}

export interface IProps {
  title: string;
  tooltip: string;
  defaultRole: string;
  disable?: boolean;
}

const props = withDefaults(defineProps<IProps>(), { disable: false });

const emit = defineEmits<{ "update:count": [count: number] }>();

const administrator = ref(false);
const authStore = useAuthStore();
const subjectOptions = computed(() => (authStore.isSSO ? subjectTypeOptionsWithSSO : subjectTypeOptions));
const subject = ref<ISelectOption | null>(
  subjectOptions.value.find((option) => option.value === SubjectTypeExtended.User) ?? null,
);
const subjects = ref<ISelectOption[]>([]);
const { roleOptions, canAssignRoles, loadRoles, validateRole, loading: rolesLoading } = useAccessRuleRoles();
const role = ref<ISelectOption | null>(null);
const rolesError = ref("");
// A retried save skips the subjects whose access rule already exists.
const completedSubjects = new Set<string>();

const isAllAuthenticatedUsers = computed(() => subject.value?.value === SubjectTypeExtended.AllAuthenticatedUsers);

const count = computed(() => {
  if (!administrator.value) return 0;
  return isAllAuthenticatedUsers.value ? 1 : subjects.value.length;
});

watch(count, (value) => emit("update:count", value), { immediate: true });

watch(
  roleOptions,
  (options) => {
    if (!role.value) role.value = options.find((option) => option.label === props.defaultRole) ?? null;
  },
  { immediate: true },
);

const requiredSubject = (value: ISelectOption | null) => !!value || errorMessages.SUBJECT_NOT_EMPTY;

async function loadRoleOptions(): Promise<void> {
  rolesError.value = "";
  try {
    await loadRoles();
  } catch {
    rolesError.value = "Could not load roles.";
  }
}

onMounted(loadRoleOptions);

function validate(): IAdministratorsValidation {
  if (!administrator.value) return { valid: true };
  if (!canAssignRoles.value || rolesLoading.value || rolesError.value || validateRole(role.value) !== true) {
    return { valid: false };
  }
  if (!isAllAuthenticatedUsers.value && !subjects.value.length) {
    return { valid: false, message: "Select at least one subject." };
  }
  return { valid: true };
}

async function save(scope: IAdministratorsScope): Promise<void> {
  if (!administrator.value || !role.value || !subject.value) return;
  const rule = { ...scope, roleId: Number(role.value.value) };
  if (isAllAuthenticatedUsers.value) {
    await accessRuleService.createAuthenticatedUsersAccessRule(rule);
    return;
  }
  const failedSubjects: string[] = [];
  for (const selected of subjects.value) {
    const subjectId = String(selected.value);
    if (completedSubjects.has(subjectId)) continue;
    try {
      await accessRuleService.createAccessRule({ ...rule, subjectId, subjectType: subject.value.value as SubjectType });
      completedSubjects.add(subjectId);
    } catch {
      failedSubjects.push(subjectId);
    }
  }
  if (failedSubjects.length) {
    throw new Error(`Couldn't add ${failedSubjects.join(", ")} as administrator`);
  }
}

defineExpose({ validate, save, hasAdministrator: computed(() => administrator.value) });
</script>

<style scoped lang="scss">
.administrator-rule {
  position: relative;
  box-shadow: inset 0 0 0 1px var(--border-subtle);
  padding: var(--space-m);
  margin-top: var(--space-xl);

  .subject-select,
  .role-select {
    width: 380px;
    max-width: 100%;
  }

  .subject-addable-multi-select {
    margin: var(--space-xl) 0;
    width: 90%;
  }

  // runai-btn renders more than one root node, so scoped classes don't reach its button.
  :deep(.remove-administrator) {
    position: absolute;
    top: var(--space-m);
    right: var(--space-m);
    width: 32px;
    height: 32px;
    min-width: 32px;
    min-height: 32px;
    padding: 0;
  }
}

.org-unit-administrators-fields :deep(.add-administrator) {
  margin-top: var(--space-xl);
  height: 38px;
  min-height: 38px;
  padding: 0 var(--space-s);

  .q-icon {
    width: 20px;
    height: 20px;
    font-size: 20px;
    margin-right: var(--space-s);
  }
}
</style>
