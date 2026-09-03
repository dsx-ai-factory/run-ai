<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="pvc-data-mount" default-opened label="Data mount" should-validate>
    <template #subheader>
      <section class="summary row">
        <template v-if="!!pvcDataModel.existingPvc">
          Origin:
          <div class="ellipsis-l">&lrm;{{ claimNameSummary }}</div>
          <span> > </span>
        </template>
        Target:
        <div class="ellipsis-l">&lrm;{{ containerPathSummary }}</div>
      </section>
    </template>

    <q-linear-progress animation-speed="100" indeterminate v-if="loading" />

    <template v-else>
      <section class="pvc-options">
        <div class="q-mb-lg">
          <div class="q-py-md">Select PVC</div>
          <runai-radio-options
            :model-value="!!pvcDataModel.existingPvc"
            :options="pvcOptions"
            @update:model-value="pvcTypeChanged"
            horizontal
          />
        </div>
      </section>
      <section class="claim-info">
        <pvc-data-mount-existing
          v-if="pvcDataModel.existingPvc"
          :claim-name="pvcDataModel.claimName || ''"
          :pvc-list="pvcList"
          @claim-name-changed="claimNameChanged"
          :loading="loading"
        />

        <claim-info-section
          v-else
          :storage-classes="storageClasses"
          :model-value="claimInfoModel"
          @update:model-value="updateClaimInfo"
          :loading="loadingStorageClass"
          :data-sharing="pvcDataModel.dataSharing || undefined"
          :support-custom-attributes="supportCustomPvcAttributes"
        />
      </section>
      <section class="container-path-section">
        <div class="q-py-md q-mt-md">
          {{ containerPathTitle }}
        </div>
        <q-input
          aid="container-path-input"
          :model-value="pvcDataModel.path"
          no-error-icon
          label="Container path"
          stack-label
          placeholder="e.g., /home/local/data"
          input-class="placeholder-italic"
          :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
          @update:model-value="pathChanged"
        />
      </section>
      <section class="restrictions-section">
        <runai-sub-expansion-item label="Restrictions">
          <section class="row items-center q-mb-md">
            <q-toggle
              :model-value="pvcDataModel.readOnly"
              label="Prevent data modification"
              left-label
              class="text-no-wrap"
              @update:model-value="readOnlyChanged"
            />
            <runai-tooltip
              class="q-ml-md"
              tooltip-position="right"
              tooltip-text="When enabled, the data will be mounted with read-only permissions."
            />
          </section>
        </runai-sub-expansion-item>
      </section>
    </template>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSubExpansionItem } from "@/components/common/runai-sub-expansion-item";
import { PvcDataMountExisting } from "./pvc-data-mount-existing";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { ClaimInfoSection } from "../claim-info-section";

// Types
import type { ClaimInfo, PvcAccessModes, PVCAssetSpec } from "@/swagger-models/assets-service-client";
import type { ISelectOption } from "@/models/global.model";
import type { StorageClassV2 } from "@/swagger-models/k8s-objects-tracker-client";

// services
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { validationRules } from "@/common/validation-rules";

// Personal Types
export interface INewPvcCmpData extends ClaimInfo {
  claimName: PVCAssetSpec["claimName"];
}

export default defineComponent({
  components: {
    RunaiTooltip,
    RunaiExpansionItem,
    RunaiSubExpansionItem,
    PvcDataMountExisting,
    ClaimInfoSection,
    RunaiRadioOptions,
  },
  emits: ["update:pvc-data-model"],
  props: {
    pvcDataModel: {
      type: Object as PropType<PVCAssetSpec>,
      required: true,
    },
    pvcList: {
      type: Array as PropType<string[]>,
      required: true,
    },
    storageClasses: {
      type: Array as PropType<StorageClassV2[]>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    loadingStorageClass: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    initClaimInfoFromScratch: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    isProjectScoped: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    supportCustomPvcAttributes: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      validationRules,
    };
  },
  computed: {
    claimInfoModel(): ClaimInfo {
      return {
        storageClass: this.pvcDataModel.claimInfo?.storageClass || "",
        accessModes: this.pvcDataModel.claimInfo?.accessModes || {},
        size: this.pvcDataModel.claimInfo?.size || "1G",
        volumeMode: this.pvcDataModel.claimInfo?.volumeMode || null,
        addedAttrValues: this.pvcDataModel.claimInfo?.addedAttrValues || undefined,
      };
    },
    claimNameSummary(): string {
      return `${this.pvcDataModel.claimName ? this.pvcDataModel.claimName : "None"}`;
    },
    containerPathSummary(): string {
      return `${this.pvcDataModel.path ? this.pvcDataModel.path : "None"}`;
    },
    pvcOptions(): Array<ISelectOption> {
      return [
        this.existingPvcOption,
        {
          label: "New PVC",
          value: false,
        },
      ];
    },
    existingPvcOption(): ISelectOption {
      const option: ISelectOption = {
        label: "Existing PVC",
        value: true,
      };

      if (this.pvcList.length === 0) {
        option.disable = true;
        option.disabledTooltip = `No existing PVCs ${this.isProjectScoped ? "in the selected project" : ""}`;
        return option;
      }

      if (this.pvcDataModel.dataSharing) {
        option.disable = true;
        option.disabledTooltip = "This can't be selected when the volume is shared";
        return option;
      }

      return option;
    },
    containerPathTitle(): string {
      return assetsUtil.containerPathTitle(true);
    },
  },
  methods: {
    isAccessModeSelected(): boolean {
      if (!this.pvcDataModel.claimInfo || !this.pvcDataModel.claimInfo.accessModes) return false;
      const accessModes: PvcAccessModes = this.pvcDataModel.claimInfo.accessModes;

      return !!(accessModes.readOnlyMany || accessModes.readWriteMany || accessModes.readWriteOnce);
    },
    modelChanged(updatedModel: PVCAssetSpec): void {
      this.$emit("update:pvc-data-model", updatedModel);
    },
    claimNameChanged(claimName: string): void {
      this.modelChanged({ ...this.pvcDataModel, claimName });
    },
    pathChanged(path: string | number | null): void {
      this.modelChanged({ ...this.pvcDataModel, path: path as string });
    },
    readOnlyChanged(readOnly: boolean): void {
      this.modelChanged({ ...this.pvcDataModel, readOnly });
    },
    pvcTypeChanged(isExists: boolean): void {
      if (!isExists) {
        this.modelChanged({ ...this.pvcDataModel, claimName: "", existingPvc: isExists });
      } else {
        this.modelChanged({ claimName: "", path: this.pvcDataModel.path, existingPvc: isExists });
      }
    },
    updateClaimInfo(claimInfo: ClaimInfo): void {
      this.modelChanged({ ...this.pvcDataModel, claimInfo });
    },
  },
});
</script>

<style lang="scss">
.pvc-data-mount {
  .claim-name-input {
    width: 500px;
  }

  .claim-select {
    width: 400px;
  }

  .type-options {
    margin-left: -18px;
    gap: 40px;
  }
  .summary {
    white-space: pre;
  }
  .ellipsis-l {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 100px;
    direction: rtl;
    text-align: left;
  }
}
</style>
