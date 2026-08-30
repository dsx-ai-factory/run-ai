<template>
  <q-td :aid="column.name" class="application-department-col">
    {{ departmentName }}
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { ITableColumn } from "@/models/table.model";
import { useAuthStore } from "@/stores/auth.store";
import { ScopeType } from "@/swagger-models/org-unit-service-client";
import { getFieldContent } from "@/utils/table-format.util";

export default defineComponent({
  props: {
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
    row: {
      type: Object as PropType<any>, // eslint-disable-line @typescript-eslint/no-explicit-any
      required: true,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
    };
  },
  computed: {
    projectId(): string {
      return getFieldContent(this.column, this.row);
    },
    departmentName(): string {
      if (!this.projectId || this.projectId === "-") return "-";

      const projectUnit = this.authStore.orgUnitList.find(
        (ou) => ou.type === ScopeType.Project && ou.id === this.projectId,
      );
      if (!projectUnit?.parentId) return "-";

      const departmentUnit = this.authStore.orgUnitList.find(
        (ou) => ou.type === ScopeType.Department && ou.id === projectUnit.parentId,
      );
      return departmentUnit?.name || "-";
    },
  },
});
</script>
