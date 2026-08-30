<template>
  <policy-non-compliance-card-wrapper
    :compliance-data="data.compliance || undefined"
    :field="data.meta.kind"
    :show-disabled-info="showDisabledInfo"
    :asset-name="data.meta.name"
  >
    <runai-select-card
      :value="isSelected"
      :disabled="disabled"
      @clicked="$emit('card-clicked')"
      :aid="getAid(data.meta.name)"
    >
      <q-tooltip v-if="data.meta.description" max-width="300px" anchor="top middle" self="bottom middle">{{
        data.meta.description
      }}</q-tooltip>
      <div class="data-source-card workload-card-content">
        <section class="top-card-section">
          <div class="data-source-logo">
            <runai-svg-icon :key="icon" :name="icon" size="42" />
          </div>
        </section>
        <div class="data-source-card-title q-mb-md" :class="{ 'text-primary': isSelected }">
          <div class="col text-subtitle1 text-weight-medium ellipsis">
            {{ data.meta.name }}
          </div>
        </div>
        <div class="data-source-card-details">
          <span class="text-weight-medium q-mr-xs">Type:</span>
          {{ type }}
        </div>
      </div>
    </runai-select-card>
  </policy-non-compliance-card-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmp
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { RunaiSelectCard } from "@/components/common/runai-select-card";

// models
import type { DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";

// constants
import { DATA_SOURCE_ICONS_NAMES } from "@/common/icons.constant";
import { DATA_SOURCE_TYPE, type TDataSourceKinds } from "@/models/data-source.model";

// utils
import { spaceToDash, toLowerCase } from "@/utils/string.util";
import { policyNonComplianceCardWrapper } from "../../policy-non-compliance-card-wrapper";

export default defineComponent({
  components: {
    RunaiSvgIcon,
    RunaiSelectCard,
    policyNonComplianceCardWrapper,
  },
  emits: ["card-clicked"],
  props: {
    data: {
      type: Object as PropType<DatasourceListResponseEntry>,
      required: true,
    },
    isSelected: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    showDisabledInfo: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  computed: {
    icon(): string {
      return (DATA_SOURCE_ICONS_NAMES[this.data.meta.kind as keyof object] as string) || DATA_SOURCE_ICONS_NAMES.default;
    },
    type(): string {
      const type: TDataSourceKinds = this.data.meta.kind as TDataSourceKinds;
      return DATA_SOURCE_TYPE[type] || this.data.meta.kind.toString();
    },
  },
  methods: {
    getAid(name: string): string {
      return `${toLowerCase(spaceToDash(name))}-card`;
    },
  },
});
</script>

<style lang="scss" scoped>
.data-source-card {
  .top-card-section {
    display: flex;
    justify-content: space-between;
  }
}
</style>
