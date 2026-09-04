<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-menu fit>
    <q-list>
      <q-item
        clickable
        v-close-popup
        v-for="opt in options"
        :key="opt.title"
        @click="$emit('data-source-selected', opt.id)"
        :disable="opt.disable"
        :aid="opt.title + '-btn'"
      >
        <q-item-section class="option">
          <runai-svg-icon v-if="opt.icon" :name="opt.icon" size="25px" />{{ opt.title }}
        </q-item-section>
      </q-item>
    </q-list>
  </q-menu>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import type { TUIDataSourceInstancesKinds } from "@/models/data-source.model";
import { DATA_SOURCES_TYPES_TO_NAMES, UI_EMPTY_DIR_KIND, UI_VOLUME_KIND } from "@/models/data-source.model";
import { DATA_SOURCE_ICONS_NAMES } from "@/common/icons.constant";
import { AssetKind } from "@/swagger-models/assets-service-client";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";

export interface IDataSourceOption {
  id: TUIDataSourceInstancesKinds;
  title: string;
  icon: string;
  disable?: boolean;
}

export default defineComponent({
  components: {
    RunaiSvgIcon,
  },
  props: {
    allowedOptions: {
      type: Object as PropType<Set<TUIDataSourceInstancesKinds> | undefined>,
      required: false,
      default: undefined,
    },
    showWorkloadVolumeOption: {
      type: Boolean,
      required: false,
      default: false,
    },
    includeDataVolume: {
      type: Boolean,
      required: false,
      default: false,
    },
    avoidPermissionsCheck: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
    };
  },
  computed: {
    allOptions(): Array<IDataSourceOption> {
      const options: Array<IDataSourceOption> = [
        {
          id: AssetKind.Nfs,
          title: DATA_SOURCES_TYPES_TO_NAMES.NFS,
          icon: DATA_SOURCE_ICONS_NAMES.nfs,
          disable: this.isOptionDisabled(ResourceType.NfsAssets),
        },
        {
          id: AssetKind.Pvc,
          title: DATA_SOURCES_TYPES_TO_NAMES.PVC,
          icon: DATA_SOURCE_ICONS_NAMES.pvc,
          disable: this.isOptionDisabled(ResourceType.PvcAssets),
        },
        {
          id: AssetKind.S3,
          title: DATA_SOURCES_TYPES_TO_NAMES.S3,
          icon: DATA_SOURCE_ICONS_NAMES.s3,
          disable: this.isOptionDisabled(ResourceType.S3Assets),
        },
        {
          id: AssetKind.Git,
          title: DATA_SOURCES_TYPES_TO_NAMES.GIT,
          icon: DATA_SOURCE_ICONS_NAMES.git,
          disable: this.isOptionDisabled(ResourceType.GitAssets),
        },
        {
          id: AssetKind.HostPath,
          title: DATA_SOURCES_TYPES_TO_NAMES.HOST_PATH,
          icon: DATA_SOURCE_ICONS_NAMES.hostPath,
          disable: this.isOptionDisabled(ResourceType.HostPathAssets),
        },
        {
          id: AssetKind.ConfigMap,
          title: DATA_SOURCES_TYPES_TO_NAMES.CONFIG_MAP,
          icon: DATA_SOURCE_ICONS_NAMES["config-map"],
          disable: this.isOptionDisabled(ResourceType.CmVolumeAssets),
        },
        {
          id: AssetKind.SecretVolume,
          title: DATA_SOURCES_TYPES_TO_NAMES.SECRET_VOLUME,
          icon: DATA_SOURCE_ICONS_NAMES["secret-volume"],
          disable: this.isOptionDisabled(ResourceType.SecretVolumeAssets),
        },
        {
          id: UI_EMPTY_DIR_KIND,
          title: "EmptyDir",
          icon: DATA_SOURCE_ICONS_NAMES[UI_EMPTY_DIR_KIND],
        },
      ];
      if (this.includeDataVolume) {
        options.push({
          id: AssetKind.DataVolume,
          title: DATA_SOURCES_TYPES_TO_NAMES.DATA_VOLUME,
          icon: DATA_SOURCE_ICONS_NAMES[AssetKind.DataVolume],
          disable: this.isOptionDisabled(ResourceType.PvcAssets),
        });
      }

      if (this.showWorkloadVolumeOption) {
        options.push({
          id: UI_VOLUME_KIND,
          title: "Volume",
          icon: DATA_SOURCE_ICONS_NAMES[UI_VOLUME_KIND],
          disable: this.isOptionDisabled(ResourceType.PvcAssets),
        });
      }

      return options;
    },
    options(): Array<IDataSourceOption> {
      const allowedOptions = this.allowedOptions;
      if (!allowedOptions) return this.allOptions;

      return this.allOptions.filter((dataSourceOption: IDataSourceOption) => allowedOptions.has(dataSourceOption.id));
    },
  },
  methods: {
    isOptionDisabled(resourceType: ResourceType): boolean {
      return this.avoidPermissionsCheck ? false : !this.permissionStore.hasPermission(resourceType, Action.Create);
    },
  },
});
</script>

<style lang="scss" scoped>
.option {
  flex-direction: row;
  gap: 10px;
  justify-content: flex-start;
  align-items: center;
}
</style>
