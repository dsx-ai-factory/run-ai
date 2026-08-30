<template>
  <section class="item-details">
    <div
      v-for="(detail, idx) in item.details"
      class="item-details-item"
      :class="[showPadding ? 'q-pa-md' : '', showSeparators && !detail.noSeperator ? 'border-bottom' : '']"
      :key="`${detail.subtitle || detail.title || 'item'}-${idx}`"
    >
      <div class="row items-center q-mb-sm" v-if="detail.title || detail.icon">
        <runai-svg-icon v-if="detail.icon" :name="detail.icon" size="20px" class="q-mr-sm" />
        <span class="details-title">{{ detail.title }}</span>
      </div>
      <div v-if="detail.subtitle" class="details-title q-mb-sm">
        {{ detail.subtitle }}
      </div>
      <template v-for="(text, index) in detail.content" :key="text + index">
        <div v-if="renderSpace(text)" aid="spacing-div" class="q-mb-sm" />
        <div v-else class="content-text" :class="index && detail.spaceContent ? 'q-mt-sm' : ''">
          {{ text }}
        </div>
      </template>
    </div>
  </section>
</template>

<script lang="ts">
import type { IDrawerListItem, WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import { LINE_BREAK } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { defineComponent } from "vue";
import type { PropType } from "vue";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";

export default defineComponent({
  name: "item-details",
  components: {
    RunaiSvgIcon,
  },
  props: {
    item: {
      type: Object as PropType<IDrawerListItem>,
      required: true,
    },
    entity: {
      type: [String, null] as PropType<WorkloadCreationDrawerEntities | null>,
      required: false,
      default: null,
    },
    showSeparators: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    showPadding: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  methods: {
    renderSpace(text: string) {
      return text === LINE_BREAK;
    },
  },
});
</script>
<style lang="scss" scoped>
.item-details {
  .border-bottom {
    border-bottom: 1px solid $black-12;
  }
  .content-text {
    font-size: 12px;
    font-weight: 400;
  }
  .details-title {
    font-size: 14px;
    font-weight: 500;
  }
}
</style>
