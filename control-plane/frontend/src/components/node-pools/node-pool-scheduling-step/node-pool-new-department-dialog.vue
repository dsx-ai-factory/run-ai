<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <runai-base-dialog :model-value="true" custom-class="scheduling-new-department" size="width-lg" @close="emit('close')">
    <template #header>New Department</template>
    <template #body>
      <q-form ref="form" class="new-department-form" @submit.prevent="reviewPreview">
        <p>The department will be created under the following cluster and can use all of its node pools</p>
        <q-input
          :model-value="clusterName || clusterId"
          label="Cluster"
          readonly
          stack-label
          class="department-cluster"
        />
        <department-name-section
          compact
          v-model:department-name="name"
          v-model:description="description"
          :cluster-id="clusterId"
          :read-only="false"
        />
        <runai-section-title
          title="Set the department administrator(s)"
          tooltip-text="Department administrators can manage the department and its projects. Adding them here enables an access rule scoped to the department. It can be edited later in the department."
        />
        <div v-if="administrator" class="administrator-rule">
          <q-btn
            flat
            round
            dense
            icon="ph ph-x"
            class="remove-administrator"
            aria-label="Remove administrator"
            @click="administrator = false"
          />
          <runai-select
            v-model="subject"
            :options="subjectTypeOptionsWithSSO"
            label="Subject"
            class="subject-select"
            @update:model-value="subjects = []"
          />
          <subject-addable-multi-select
            v-if="subject.value !== SubjectTypeExtended.AllAuthenticatedUsers"
            :subject-type="subject"
            :selected-subjects-options="subjects"
            @update:selected-subjects="subjects = $event"
            @add-subject="subjects = [...subjects, $event]"
          />
          <runai-select v-model="role" :options="roleOptions" label="Role" class="role-select" />
        </div>
        <runai-btn
          v-else
          flat
          color="primary"
          icon="ph ph-plus"
          label="Administrator"
          class="add-administrator"
          aid="add-department-administrator"
          @click="administrator = true"
        />
        <p class="department-preview-note" role="status">
          {{
            reviewed
              ? "Preview only. This department has not been created."
              : "Preview only — department creation is not available in this preview."
          }}
        </p>
      </q-form>
    </template>
    <template #footer
      ><runai-btn aid="cancel-new-department" flat color="primary" label="Cancel" @click="emit('close')" /><runai-btn
        color="primary"
        label="Create department"
        aid="create-department-preview"
        @click="reviewPreview"
    /></template>
  </runai-base-dialog>
</template>
<script setup lang="ts">
import { ref } from "vue";
import type { QForm } from "quasar";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import { RunaiSectionTitle } from "@/components/common/runai-section-title";
import { RunaiSelect } from "@/components/common/runai-select";
import DepartmentNameSection from "@/components/department/department-edit-form/department-name-section/department-name-section.vue";
import { SubjectAddableMultiSelect } from "@/components/rbac/common/subject-addable-multi-select";
import { subjectTypeOptionsWithSSO } from "@/models/access-rule.model";
import { SubjectTypeExtended } from "@/swagger-models/authorization-client";
import type { ISelectOption } from "@/models/global.model";
defineProps<{ clusterId: string; clusterName?: string }>();
const emit = defineEmits<{ close: [] }>();
const name = ref(""),
  description = ref(""),
  administrator = ref(false),
  reviewed = ref(false);
const subject = ref(subjectTypeOptionsWithSSO[0]);
const subjects = ref<ISelectOption[]>([]);
const roleOptions = [{ label: "Department administrator", value: "department-administrator" }];
const role = ref(roleOptions[0]);
const form = ref<QForm>();
async function reviewPreview() {
  if (await form.value?.validate()) reviewed.value = true;
}
</script>
<style lang="scss">
.scheduling-new-department {
  width: 800px !important;
  max-width: calc(100vw - 40px) !important;
}
.scheduling-new-department .runai-base-dialog-header {
  min-height: 64px;
  padding: 16px;
}
.scheduling-new-department .runai-base-dialog-body {
  padding: 32px 16px 24px;
}
.scheduling-new-department .runai-base-dialog-footer {
  min-height: 70px;
  padding: 16px;
}
.new-department-form .department-cluster {
  width: 85%;
  margin-top: 24px;
  margin-bottom: 44px;
}
.new-department-form .runai-expansion-item {
  border: 0;
  box-shadow: none;
  margin: 0 0 26px;
}
.new-department-form .runai-expansion-item .q-item {
  padding: 0;
  min-height: 30px;
}
.new-department-form .runai-expansion-item .q-item__section--side {
  display: none;
}
.new-department-form .runai-expansion-item .q-expansion-item__content > div {
  padding: 0;
}
.new-department-form .runai-expansion-item .col-7 {
  width: 66.6667%;
}
.new-department-form .runai-expansion-item .col-11 {
  width: 100%;
}
.new-department-form .runai-section-title {
  font-weight: 400;
}
.administrator-rule {
  position: relative;
  border: 1px solid var(--border-subtle);
  padding: 16px;
  margin-top: 26px;
}
.administrator-rule .subject-select,
.administrator-rule .role-select {
  width: 52%;
}
.administrator-rule .subject-addable-multi-select {
  margin: 26px 0 16px;
  width: 90%;
}
.administrator-rule .remove-administrator {
  position: absolute;
  top: 20px;
  right: 0;
}
.add-administrator {
  margin-top: 26px;
}
.department-preview-note {
  color: var(--b-text-sub);
  font-size: 12px;
  margin: 16px 0 0;
}
</style>
