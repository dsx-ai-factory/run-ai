<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <policy-non-compliance-card-wrapper
    :compliance-data="data.compliance || undefined"
    :field="data.meta.kind"
    :show-disabled-info="showDisabledInfo"
    :asset-name="data.meta.name"
  >
    <runai-select-card
      class="environment-card"
      :value="isSelected"
      :disabled="disabled"
      @clicked="$emit('card-clicked')"
      :aid="getAid(data.meta.name)"
    >
      <q-tooltip v-if="data.meta.description" max-width="300px" anchor="top middle" self="bottom middle">{{
        data.meta.description
      }}</q-tooltip>
      <div class="workload-card-content">
        <section class="top-card-section">
          <div class="environment-logo">
            <runai-svg-icon v-for="item in icons.slice(0, 4)" :name="item.icon" :key="item.name" size="50" />
          </div>
        </section>
        <div class="environment-card-title q-mb-md" :class="{ 'text-primary': isSelected }">
          <div class="col text-subtitle1 text-weight-medium ellipsis">
            {{ data.meta.name }}
          </div>
        </div>
        <div class="environment-card-details">
          <span class="text-weight-medium q-mr-xs">Image:</span>
          {{ data.spec.image }}
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
import type { EnvironmentAsset, Connection } from "@/swagger-models/assets-service-client";

// icons
import { toolTypeIconsMap, distFrameworkIconsMap } from "@/common/icons.constant";
import { getNvidiaImageBasedIcon } from "@/common/icon.util";

// utils
import { spaceToDash, toLowerCase } from "@/utils/string.util";
import { policyNonComplianceCardWrapper } from "../../policy-non-compliance-card-wrapper";

export interface Icon {
  name: string;
  icon: string;
}

export default defineComponent({
  components: {
    RunaiSvgIcon,
    RunaiSelectCard,
    policyNonComplianceCardWrapper,
  },
  emits: ["card-clicked"],
  props: {
    data: {
      type: Object as PropType<EnvironmentAsset>,
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
    icons(): Array<Icon> {
      const defaultIcon: Icon = { name: "default", icon: "environment-gray-new" };

      if (!this.data.spec.connections?.length && !this.data.meta.workloadSupportedTypes?.distFramework) {
        return [defaultIcon];
      }
      const icons: Array<Icon> = [];

      const image = this.data.spec.image;
      const imageBasedIcon = getNvidiaImageBasedIcon(image);
      if (imageBasedIcon) {
        icons.push({ name: imageBasedIcon, icon: imageBasedIcon });
      }

      const { distFramework } = this.data.meta.workloadSupportedTypes || {};
      if (distFramework) {
        // DistFramwork icon should be the first icon.
        icons.push({ name: distFramework, icon: distFrameworkIconsMap[distFramework] });
      }

      this.data.spec.connections?.forEach((connection: Connection) => {
        const toolType: string | undefined =
          connection.internalToolInfo?.toolType || connection.externalToolInfo?.toolType;
        icons.push({
          name: connection.name,
          icon: (toolType && toolTypeIconsMap[toolType]) || defaultIcon.icon,
        });
      });

      return icons;
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
.environment-card {
  .top-card-section {
    display: flex;
    justify-content: space-between;
  }

  .environment-card-details {
    line-break: anywhere;
  }
}
</style>
