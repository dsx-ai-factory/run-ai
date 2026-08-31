<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog
    :model-value="modelValue"
    size="dynamic"
    custom-class="user-edit-dialog"
    no-padding
    @close="onCancel"
  >
    <template #header>Edit user</template>
    <template #body>
      <div class="user-edit-body">
        <div class="field-section">
          <div class="disabled-field-wrapper">
            <q-input
              aid="user-edit-email"
              :model-value="user.username"
              label="Email address"
              stack-label
              disable
              no-error-icon
            />
            <q-tooltip anchor="top middle" self="bottom middle">This can't be modified when editing a user</q-tooltip>
          </div>
        </div>
        <div class="field-section">
          <div class="name-section-label text-b-subtitle2">
            Enter the user's first and last name
            <span class="optional-label">(optional)</span>
          </div>
          <div class="row q-col-gutter-md">
            <div class="col-6">
              <q-input aid="user-edit-first-name" v-model="editFirstName" label="First name" stack-label no-error-icon />
            </div>
            <div class="col-6">
              <q-input aid="user-edit-last-name" v-model="editLastName" label="Last name" stack-label no-error-icon />
            </div>
          </div>
        </div>
      </div>
    </template>
    <template #footer>
      <runai-btn flat color="primary" label="Cancel" aid="user-edit-cancel-btn" @click="onCancel" />
      <runai-btn
        color="primary"
        label="Save"
        aid="user-edit-save-btn"
        :loading="saving"
        :disable="!hasChanges || saving"
        @click="onSave"
      />
    </template>
  </runai-base-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useQuasar } from "quasar";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiBtn } from "@/components/common/runai-btn";
import type { User } from "@/swagger-models/identity-manager-client";
import userService from "@/services/control-plane/user.service/user.service";
import { alertUtil } from "@/utils/alert.util";
import { ErrorAlert } from "@/utils/error-alert.util";

export interface Props {
  modelValue: boolean;
  user: User;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "saved"): void;
}>();

const $q = useQuasar();

const editFirstName = ref(props.user.firstName || "");
const editLastName = ref(props.user.lastName || "");
const saving = ref(false);

const hasChanges = computed(
  () => editFirstName.value !== (props.user.firstName || "") || editLastName.value !== (props.user.lastName || ""),
);

function onCancel(): void {
  emit("close");
}

async function onSave(): Promise<void> {
  if (saving.value) return;
  saving.value = true;
  try {
    await userService.updateUser(props.user.id, {
      firstName: editFirstName.value.trim(),
      lastName: editLastName.value.trim(),
    });
    $q.notify(alertUtil.getSuccess(`User ${props.user.username} updated`));
    emit("saved");
  } catch (e: unknown) {
    const errorAlert = new ErrorAlert({
      generalMessage: ErrorAlert.failedUpdateMessage(`user ${props.user.username}`),
    });
    $q.notify(errorAlert.getNotification(e));
  } finally {
    saving.value = false;
  }
}

watch(
  () => props.user,
  (newUser: User) => {
    editFirstName.value = newUser.firstName || "";
    editLastName.value = newUser.lastName || "";
  },
);
</script>

<style lang="scss">
.user-edit-dialog {
  width: 600px;
}
</style>

<style lang="scss" scoped>
.user-edit-body {
  padding: 0 16px 16px;

  .field-section {
    margin-bottom: 16px;
  }

  .disabled-field-wrapper {
    position: relative;
    cursor: not-allowed;
  }

  .name-section-label {
    margin-bottom: 8px;
  }

  .optional-label {
    color: var(--b-text-sub);
    font-style: italic;
    margin-left: 4px;
  }
}
</style>
