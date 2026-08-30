<template>
  <q-td :aid="column.name" @click.stop="onCellClick" @keydown.enter="onCellClick">
    <div class="new-role-cell">
      <runai-select
        ref="selectRef"
        class="new-role-select select-italic-placeholder"
        :model-value="row.newRole ?? ''"
        :options="roleOptions"
        placeholder="Select..."
        dense
        outlined
        hide-bottom-space
        @update:model-value="onRoleSelect"
      />
      <runai-btn
        v-if="row.newRole"
        aid="view-new-role"
        flat
        round
        dense
        size="sm"
        tooltip="View role"
        tooltip-anchor="top middle"
        tooltip-self="bottom middle"
        @click="openReview"
      >
        <q-icon :name="detailsIcon" size="16px" color="black-54" />
      </runai-btn>
    </div>
  </q-td>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { ECustomCell, type ICustomCellNewRoleColConfig, type ITableColumn } from "@/models/table.model";
import type { ILegacyRoleTransitionRow } from "@/models/legacy-access-rule-transition.model";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiBtn } from "@/components/common/runai-btn";
import { ROLE_ROUTE_NAMES } from "@/router/role.routes/role.routes.names";
import { getCustomCellConfig } from "@/utils/table-format.util/table-format.util";
import { openNewWindow } from "@/utils/common.util";
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";

defineOptions({ name: "new-role-col" });

const props = defineProps<{
  row: ILegacyRoleTransitionRow;
  column: ITableColumn;
}>();

const emit = defineEmits<{ "custom-cell-event": [] }>();

const router = useRouter();
const detailsIcon = toQuasarIconName(runaiIcons.action.details);

const selectRef = ref();

function onCellClick(evt: MouseEvent): void {
  const btnEl = (evt.currentTarget as HTMLElement).querySelector(".q-btn");
  if (!btnEl?.contains(evt.target as Node)) {
    (selectRef.value?.$refs?.qSelectRef as { showPopup?: () => void } | undefined)?.showPopup?.();
  }
}

const config = computed(
  () => getCustomCellConfig(props.column, ECustomCell.NEW_ROLE_COL) as ICustomCellNewRoleColConfig | undefined,
);
const roleOptions = computed(() => config.value?.roleOptions ?? []);

function onRoleSelect(newRole: string): void {
  (props.row as ILegacyRoleTransitionRow).newRole = newRole;
  emit("custom-cell-event");
}

function openReview(): void {
  const roleId = config.value?.getReviewRoleId(props.row.newRole ?? "");
  if (!roleId) return;
  const route = router.resolve({ name: ROLE_ROUTE_NAMES.ACCESS_RULES_ROLE_VIEW, params: { id: roleId } });
  openNewWindow(route.href);
}
</script>

<style lang="scss" scoped>
.new-role-cell {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.new-role-select {
  width: 172px;
}
</style>
