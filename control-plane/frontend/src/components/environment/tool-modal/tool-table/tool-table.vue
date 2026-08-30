<template>
  <q-table
    :="tableDefaultProps"
    :rows="tools"
    :columns="columns"
    rows-per-page-label="Rows per page"
    :filter="filterBy.searchTerm"
    :pagination="filterBy"
  >
    <template v-slot:body-cell-tool-name="props">
      <name-col :row="props.row" :column="props.col" :icon="getIcon(props.row)" />
    </template>

    <template v-slot:body-cell-connection-type="props">
      <q-td>
        {{ getUIConnectionType(props.row) }}
      </q-td>
    </template>
    <template v-slot:no-data="{}">
      <runai-table-no-data
        :filter-by="filterBy"
        entity-name="tool"
        @clear-filters="$emit('clear-filters')"
        :show-filter-icon-and-button="!modalView"
        :show-create-btn="!modalView"
      />
    </template>
  </q-table>
</template>

<script lang="ts">
import { toolTypeIconsMap } from "@/common/icons.constant";
import type { Connection } from "@/swagger-models/assets-service-client";
import { defineComponent, type PropType } from "vue";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";

import { NameCol } from "@/components/common/runai-table/runai-custom-table-columns";

// constants
import { tableDefaultProps, type TTableDefaultProps } from "@/common/table.constant";
import type { IFilterBy } from "@/models/filter.model";

import type { ITableColumn } from "@/models/table.model";

const connectionTypeMap = {
  ExternalUrl: "External URL",
  NodePort: "NodePort",
  LoadBalancer: "Load balancer",
} as Record<string, string>;

export default defineComponent({
  components: {
    RunaiTableNoData,
    NameCol,
  },
  props: {
    tools: {
      type: Array as PropType<Array<Connection>>,
      required: true,
    },
    columns: {
      type: Array as PropType<Array<ITableColumn>>,
      required: true,
    },
    filterBy: {
      type: Object as PropType<IFilterBy>,
      required: false,
      default: {} as IFilterBy,
    },
    modalView: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  data() {
    return {
      tableDefaultProps: tableDefaultProps as TTableDefaultProps,
    };
  },
  methods: {
    getIcon(tool: Connection): string {
      return (
        (tool.externalToolInfo && toolTypeIconsMap[tool.externalToolInfo.toolType]) ||
        (tool.internalToolInfo && toolTypeIconsMap[tool.internalToolInfo.toolType]) ||
        "environment"
      );
    },
    getUIConnectionType(tool: Connection): string {
      return (
        (tool.internalToolInfo && connectionTypeMap[tool.internalToolInfo.connectionType]) ||
        connectionTypeMap.ExternalUrl
      );
    },
  },
});
</script>
