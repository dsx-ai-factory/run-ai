<template>
  <runai-select-card :value="isSelected" @clicked="$emit('card-clicked')" :aid="getAid(data.title)">
    <div class="k8s-dist workload-card-content">
      <div class="dist-logo">
        <runai-svg-icon :name="data.icon" size="42" />
      </div>
      <div class="k8s-dist-title q-mb-md" :class="{ 'text-primary': isSelected }">
        <div class="col text-subtitle1 text-weight-medium ellipsis">
          {{ data.title }}
        </div>
      </div>
      <div class="k8s-dist-details">
        {{ data.description }}
      </div>
    </div>
  </runai-select-card>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmp
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { RunaiSelectCard } from "@/components/common/runai-select-card";
// types
import type { IK8sDistributionCardData } from "@/models/cluster.model";
// utils
import { spaceToDash, toLowerCase } from "@/utils/string.util";

export default defineComponent({
  components: {
    RunaiSvgIcon,
    RunaiSelectCard,
  },
  emits: ["card-clicked"],
  props: {
    data: {
      type: Object as PropType<IK8sDistributionCardData>,
      required: true,
    },
    isSelected: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  methods: {
    getAid(name: string): string {
      return `${toLowerCase(spaceToDash(name))}-card`;
    },
  },
});
</script>
