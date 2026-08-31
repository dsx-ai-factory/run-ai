<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

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
        <div class="template-card-image ellipsis">
          <span class="text-weight-medium q-mr-xs">Image:</span>{{ imageSummary }}
        </div>
        <div class="template-card-compute-resources ellipsis">
          <div class="text-weight-medium q-mr-xs">Total requested GPU resources:</div>
          <div class="q-mr-xs ellipsis">{{ totalRequestedGPUResourcesSummary }}</div>
        </div>
        <div class="template-card-data-sources ellipsis">
          <span class="text-weight-medium q-mr-xs">Data & storage:</span>
          {{ shortStorageSummary }}
        </div>

        <q-tooltip max-width="300px" class="bg-white text-body-color text-body2 shadow-2">
          <div class="template-card-env q-mb-sm">
            <div class="text-weight-medium q-mb-xs">Image:</div>
            <div>{{ imageSummary }}</div>
          </div>
          <div class="template-card-data-sources q-mb-sm">
            <div class="text-weight-medium q-mb-xs">Total requested GPU resources:</div>
            <div>{{ totalRequestedGPUResourcesSummary }}</div>
          </div>
          <div class="template-card-compute-resources">
            <div class="text-weight-medium q-mb-xs">Data & storage:</div>
            <div>{{ fullstorageSummary }}</div>
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
import type { DistributedFramework } from "@/swagger-models/assets-service-client";

// icons
import { distFrameworkIconsMap, toolTypeIconsMap } from "@/common/icons.constant";
import { getHuggingFaceIcon, getNimIconIfNimImage, getNvidiaImageBasedIcon } from "@/common/icon.util";

// utils
import { spaceToDash, toLowerCase } from "@/utils/string.util";
import type { SupersetTemplate } from "@/swagger-models/workloads-client";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { computeResourceUtil } from "@/utils/compute-resource.util";
import type { EWorkloadType } from "@/models/workload.model";

const STORAGE_TYPE_LABELS = [
  { key: "pvc", label: "PVC" },
  { key: "dataVolume", label: "Data volume" },
  { key: "s3", label: "S3" },
  { key: "nfs", label: "NFS" },
  { key: "configMapVolume", label: "Config map" },
  { key: "hostPath", label: "Host path" },
  { key: "secretVolume", label: "Secret" },
] as const;

export default defineComponent({
  name: "template-card-v2",
  components: {
    RunaiSvgIcon,
    RunaiSelectCard,
  },
  emits: ["card-clicked"],
  props: {
    data: {
      type: Object as PropType<SupersetTemplate>,
      required: true,
    },
    isSelected: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  computed: {
    imageSummary(): string {
      return this.data.spec?.image || this.data.combinedSpec?.image || "-";
    },
    totalRequestedGPUResourcesSummary(): string {
      return computeResourceUtil.totalGpuResources({
        ...(this.data.spec?.compute || this.data.combinedSpec?.compute),
        gpuDevicesRequest:
          this.data.spec?.compute?.gpuDevicesRequest || this.data.combinedSpec?.compute?.gpuDevicesRequest || 0,
      }).totalGpus;
    },
    shortStorageSummary(): string {
      // Get all storage types that have items
      const storageWithItems = STORAGE_TYPE_LABELS.filter(({ key }) => {
        const count = this.data.spec?.storage?.[key]?.length || 0;
        return count > 0;
      });

      if (storageWithItems.length === 0) {
        return "-";
      }

      // Get the first storage type's label and count
      const firstType = storageWithItems[0];
      const firstTypeCount = this.data.spec?.storage?.[firstType.key]?.length || 0;

      // Calculate total count of all items
      const totalItems = STORAGE_TYPE_LABELS.reduce((sum, { key }) => {
        return sum + (this.data.spec?.storage?.[key]?.length || 0);
      }, 0);

      // If there's only one item total, just return the type
      if (totalItems === 1) {
        return firstType.label;
      }

      // Otherwise return first type + count of remaining items
      const remainingItems = totalItems - firstTypeCount;
      return `${firstType.label} + ${remainingItems} more`;
    },
    fullstorageSummary(): string {
      const summaries = STORAGE_TYPE_LABELS.map(({ key, label }) => {
        const count = this.data.spec?.storage?.[key]?.length || 0;
        return count ? `${label}: ${count}` : "";
      }).filter(Boolean);

      return summaries.length ? summaries.join(", ") : "-";
    },
    icons(): Array<string> {
      const defaultIcon = "template-gray-new";

      // Check for image-based icons first (NVIDIA PyTorch, NVIDIA general, etc.)
      const image = this.data.spec?.image;
      const imageBasedIcon = getNvidiaImageBasedIcon(image);
      if (imageBasedIcon) {
        return [imageBasedIcon];
      }

      const nimIcon = getNimIconIfNimImage(image, this.data.spec?.labels);
      if (nimIcon) return [nimIcon];

      const hfIcon = getHuggingFaceIcon(this.data.meta.workloadType as EWorkloadType, image);
      if (hfIcon) return [hfIcon];

      const distFramework: DistributedFramework | undefined | null = this.data.spec?.distributedFramework;

      const connections = workloadHybridUtil.getConnectionsField(
        this.data.spec?.ports || [],
        this.data.spec?.exposedUrls || [],
        this.data.spec?.relatedUrls || [],
      );

      if (!connections[0]?.toolType || connections[0]?.toolType.length === 0) {
        if (distFramework) {
          return [distFrameworkIconsMap[distFramework]];
        }
        return [defaultIcon];
      }
      const toolIcons = connections.map(
        (connection) => (connection?.toolType && toolTypeIconsMap[connection.toolType]) || defaultIcon,
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
