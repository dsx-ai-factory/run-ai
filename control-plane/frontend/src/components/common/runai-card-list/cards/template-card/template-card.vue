<template>
  <runai-select-card :value="isSelected" @clicked="$emit('card-clicked')" :aid="getAid(data.meta.name)">
    <div class="template-card workload-card-content">
      <section class="top-content q-mb-md">
        <div class="template-card-icon">
          <runai-svg-icon v-for="(icon, idx) in icons.slice(0, 4)" :name="icon" :key="idx" size="47" />
        </div>
        <div class="template-card-title" :class="{ 'text-primary': isSelected }">
          <div class="col text-subtitle1 text-weight-medium ellipsis">
            {{ data.meta.name }}
          </div>
        </div>
      </section>

      <section class="bottom-content text-caption">
        <div class="template-card-env ellipsis">
          <span class="text-weight-medium q-mr-xs">Environment:</span>{{ environment }}
        </div>
        <div class="template-card-data-sources ellipsis">
          <span class="text-weight-medium q-mr-xs">Data source(s):</span>
          {{ shortDataSourceNames }}
        </div>
        <div class="template-card-compute-resources ellipsis">
          <span class="text-weight-medium q-mr-xs">Compute Resources:</span>{{ computeResource }}
        </div>

        <q-tooltip max-width="300px" class="bg-white text-body-color text-body2 shadow-2">
          <div class="template-card-env q-mb-sm">
            <div class="text-weight-medium q-mb-xs">Environment:</div>
            <div>{{ environment }}</div>
          </div>
          <div class="template-card-data-sources q-mb-sm">
            <div class="text-weight-medium q-mb-xs">Data source(s):</div>
            <div>{{ fullDataSourceNames }}</div>
          </div>
          <div class="template-card-compute-resources">
            <div class="text-weight-medium q-mb-xs">Compute Resources:</div>
            <div>{{ computeResource }}</div>
          </div>
        </q-tooltip>
      </section>
    </div>
  </runai-select-card>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmp
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { RunaiSelectCard } from "@/components/common/runai-select-card";

// models
import type { WorkloadTemplate, AssetRef, DistributedFramework } from "@/swagger-models/assets-service-client";

// icons
import { distFrameworkIconsMap, toolTypeIconsMap } from "@/common/icons.constant";

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
      type: Object as PropType<WorkloadTemplate>,
      required: true,
    },
    isSelected: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  computed: {
    environment(): string {
      return this.data.spec.assets.environment?.name || "";
    },
    computeResource(): string {
      return this.data.spec.assets.compute?.name || "";
    },
    shortDataSourceNames(): string {
      const datasources: Array<AssetRef> = this.data.spec.assets.datasources || [];
      if (datasources.length === 0) return "";
      if (datasources.length === 1) return datasources[0].name;

      return `${datasources[0].name} + ${datasources.length - 1} more`;
    },
    fullDataSourceNames(): string {
      const datasources: Array<AssetRef> | undefined = this.data.spec.assets.datasources;
      if (!datasources) return "";

      return datasources
        .reduce((acc: string, datasource: AssetRef) => {
          return `${acc} ${datasource.name}, `;
        }, "")
        .trim()
        .slice(0, -1);
    },
    icons(): Array<string> {
      const defaultIcon = "template-gray-new";
      const distFramework: DistributedFramework | undefined = this.data.meta.workloadSupportedTypes?.distFramework;
      if (!this.data.spec.assets.environment.toolTypes || this.data.spec.assets.environment.toolTypes.length === 0) {
        if (distFramework) {
          return [distFrameworkIconsMap[distFramework]];
        }
        return [defaultIcon];
      }
      const toolIcons = this.data.spec.assets.environment.toolTypes.map(
        (toolType: string) => toolTypeIconsMap[toolType] || defaultIcon,
      );
      if (!distFramework) return toolIcons;
      return [distFrameworkIconsMap[distFramework], ...toolIcons];
    },
  },
  methods: {
    getAid(name: string): string {
      return `${toLowerCase(spaceToDash(name))}-card`;
    },
  },
});
</script>
