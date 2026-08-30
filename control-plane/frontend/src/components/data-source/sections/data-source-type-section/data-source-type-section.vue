<template>
  <runai-expansion-item class="data-source-type-section" label="Type" hide-expend-icon>
    <template #subheader>
      <div class="row items-center">
        <runai-svg-icon :name="icon" size="32px" /><span class="q-pl-sm">{{ title }}</span>
      </div>
    </template>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { DATA_SOURCE_TYPE, type TDataSourceKinds } from "@/models/data-source.model";
import { DATA_SOURCE_ICONS_NAMES } from "@/common/icons.constant";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiSvgIcon,
  },
  props: {
    dataSourceKind: {
      type: String as PropType<TDataSourceKinds>,
      required: true,
    },
  },
  computed: {
    icon(): DATA_SOURCE_ICONS_NAMES {
      return DATA_SOURCE_ICONS_NAMES[this.dataSourceKind];
    },
    title(): string {
      return DATA_SOURCE_TYPE[this.dataSourceKind] || "";
    },
  },
});
</script>
