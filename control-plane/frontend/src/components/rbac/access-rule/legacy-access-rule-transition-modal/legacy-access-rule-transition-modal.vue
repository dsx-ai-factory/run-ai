<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="legacy-access-rule-transition-modal">
    <runai-base-dialog
      :model-value="true"
      @close="closeDialog"
      size="width-md"
      custom-class="transition-modal-width"
      no-padding
    >
      <template #header>Transition Access Rules To New Roles</template>

      <template #body>
        <p class="transition-modal-description q-px-lg q-pt-md q-mb-md">
          To complete the update, select a new role for each legacy role listed below.
        </p>

        <notification-alert-bar
          class="q-mb-sm"
          :variant="ENotificationBarVariant.Info"
          message="A new access rule with the same subject and scope will be created, and the previous rule will be deleted."
        />

        <legacy-access-rule-transition-table
          :rows="resolvedRows"
          :loading="loading"
          :available-roles="availableRoles"
          :selected="selectedRows"
          @update:selected="selectedRows = $event"
        />
      </template>

      <template #footer>
        <div class="row items-center justify-between full-width q-px-md q-py-sm">
          <p v-if="showLimitWarning" class="limit-warning q-mb-none">Up to 100 rules can be transitioned at a time.</p>
          <div v-else />
          <div class="row q-gutter-sm">
            <runai-btn
              aid="cancel-transition"
              flat
              label="cancel"
              color="primary"
              :disable="transitioning"
              @click="closeDialog"
            />
            <runai-btn
              aid="confirm-transition"
              :label="actionButtonLabel"
              color="primary"
              :disable="isActionDisabled"
              :loading="transitioning"
              @click="onTransition"
            />
          </div>
        </div>
      </template>
    </runai-base-dialog>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useQuasar } from "quasar";
import type { Role } from "@/swagger-models/authorization-client/api";
import type { ILegacyRoleTransitionRow } from "@/models/legacy-access-rule-transition.model";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiBtn } from "@/components/common/runai-btn";
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";
import { LegacyAccessRuleTransitionTable } from "./legacy-access-rule-transition-table";
import { useLegacyAccessRuleStore } from "@/stores/legacy-access-rule.store";
import { ROLE_NAME_API_MAP } from "@/models/roles.model";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { alertUtil } from "@/utils/alert.util";
import { ErrorAlert } from "@/utils/error-alert.util";

defineOptions({ name: "legacy-access-rule-transition-modal" });

const props = defineProps<{
  availableRoles: Role[];
}>();

const emit = defineEmits<{
  close: [];
}>();

const $q = useQuasar();
const legacyAccessRuleStore = useLegacyAccessRuleStore();
const { legacyRows, loading } = storeToRefs(legacyAccessRuleStore);

const selectedRows = ref<ILegacyRoleTransitionRow[]>([]);
const transitioning = ref(false);

const TRANSITION_LIMIT = 100;

const rawTransitionCount = computed(() => selectedRows.value.reduce((sum, r) => sum + r.accessRuleCount, 0));
const transitionCount = computed(() => Math.min(rawTransitionCount.value, TRANSITION_LIMIT));
const showLimitWarning = computed(() => rawTransitionCount.value > TRANSITION_LIMIT);
const actionButtonLabel = computed(() => `Transition ${transitionCount.value} access rules`);
const isActionDisabled = computed(
  () =>
    transitioning.value ||
    selectedRows.value.length === 0 ||
    transitionCount.value === 0 ||
    selectedRows.value.some((r) => !r.newRole),
);

// Stored in a ref (not computed) so each row is deep-reactive — the new-role
// cell mutates `row.newRole` directly, which would not trigger reactivity on
// the plain spread objects a computed would return.
const resolvedRows = ref<ILegacyRoleTransitionRow[]>([]);

watch(
  [legacyRows, () => props.availableRoles],
  () => {
    const availableNames = new Set(props.availableRoles.map((r) => r.name.toLowerCase()));
    resolvedRows.value = legacyRows.value.map((row) => {
      if (!row.newRole) return { ...row };
      const resolvedName = ROLE_NAME_API_MAP[row.newRole] ?? row.newRole;
      return {
        ...row,
        newRole: availableNames.has(resolvedName.toLowerCase()) ? resolvedName : undefined,
      };
    });
    selectedRows.value = [];
  },
  { immediate: true },
);

async function onTransition(): Promise<void> {
  transitioning.value = true;

  try {
    const row = selectedRows.value[0];
    if (!row.newRole) return;

    const newRole = row.newRole;
    const newRoleId = props.availableRoles.find((r) => r.name.toLowerCase() === newRole.toLowerCase())?.id;

    if (!newRoleId) {
      console.warn("[onTransition] newRoleId not found, aborting");
      return;
    }

    const response = await accessRuleService.migrateDeprecatedAccessRules(row.roleId, newRoleId);

    const deleted = response.succeeded.deleted.length;

    $q.notify(alertUtil.getSuccess(`${deleted}/${row.accessRuleCount} access rules transitioned`));

    await legacyAccessRuleStore.loadLegacyAccessRules();

    selectedRows.value = [];

    if (legacyRows.value.length === 0) {
      closeDialog();
    }
  } catch (error: unknown) {
    const errorAlert = new ErrorAlert({
      generalMessage: ErrorAlert.failedUpdateMessage("access rules"),
    });
    $q.notify(errorAlert.getNotification(error));
  } finally {
    transitioning.value = false;
  }
}

function closeDialog(): void {
  emit("close");
}
</script>

<style lang="scss" scoped>
.limit-warning {
  font-size: 12px;
  color: var(--black-54);
}

:deep(.transition-modal-width) {
  max-height: 700px;
  display: flex;
  flex-direction: column;

  .runai-base-dialog-body {
    flex: 1;
    overflow: hidden;
    min-height: 0;
  }
}
</style>
