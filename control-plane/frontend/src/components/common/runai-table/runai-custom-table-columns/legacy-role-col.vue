<template>
  <q-td :aid="column.name" :style="column.style">
    <div class="legacy-role-cell">
      <span class="legacy-role-text">{{ row.legacyRole }}</span>
      <runai-btn
        aid="view-legacy-role"
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
import { useRouter } from "vue-router";
import type { ITableColumn } from "@/models/table.model";
import type { ILegacyRoleTransitionRow } from "@/models/legacy-access-rule-transition.model";
import { RunaiBtn } from "@/components/common/runai-btn";
import { ROLE_ROUTE_NAMES } from "@/router/role.routes/role.routes.names";
import { openNewWindow } from "@/utils/common.util";
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";

defineOptions({ name: "legacy-role-col" });

const props = defineProps<{
  row: ILegacyRoleTransitionRow;
  column: ITableColumn;
}>();

defineEmits<{ "custom-cell-event": [] }>();

const router = useRouter();
const detailsIcon = toQuasarIconName(runaiIcons.action.details);

function openReview(): void {
  const route = router.resolve({ name: ROLE_ROUTE_NAMES.ACCESS_RULES_ROLE_VIEW, params: { id: props.row.roleId } });
  openNewWindow(route.href);
}
</script>

<style lang="scss" scoped>
.legacy-role-cell {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-width: 0;
}

.legacy-role-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
