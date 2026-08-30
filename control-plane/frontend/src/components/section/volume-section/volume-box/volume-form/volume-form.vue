<template>
  <section>
    <claim-info-section
      :model-value="claimInfo"
      @update:model-value="updateClaimInfo"
      :storage-classes="storageClasses"
      :loading="loadingStorageClass"
      :policy-rules="policyRules?.claimInfo || undefined"
      :disable="disable"
      :disable-info="disableInfo"
    />

    <section class="container-path-section">
      <div class="q-py-md q-mt-md">Set the volume target location</div>
      <policy-string-field
        aid="path-input"
        :model-value="volume.path"
        no-error-icon
        label="Container path"
        stack-label
        placeholder="e.g., /home/local/data"
        input-class="placeholder-italic"
        :rules="[notEmpty, isDirectoryPath]"
        @update:model-value="pathChanged"
        :disable="disable"
        :tooltip="disableInfo"
        :policy-rules="policyRules?.path"
      />
    </section>
    <section class="q-pt-lg q-mb-xs" v-if="showVolumePersistencySection">
      <div>Volume persistency</div>
      <runai-radio-options
        :model-value="!!volume.ephemeral"
        @update:model-value="volumePesistencyChanged"
        :options="volumePersistencyOptions"
        class="q-pt-sm column"
        :disable="disable"
      />
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { ClaimInfoSection } from "@/components/data-source/pvc/claim-info-section";
// Models
import type { ClaimInfo } from "@/swagger-models/assets-service-client";
import type { ISelectOption } from "@/models/global.model";
import { NONE_STORAGE_CLASS, NONE_STORAGE_CLASS_LABEL } from "@/models/data-source.model";
import type { StorageClassV2 } from "@/swagger-models/k8s-objects-tracker-client";
// Utils
import { isNotEmpty, isValidDirectoryPath } from "@/common/form.validators";
// Constants
import { errorMessages } from "@/common/error-message.constant";
import type { PvcRules } from "@/swagger-models/policy-service-client";
import type { IUIPvcInstance } from "@/models/workload.model";

export default defineComponent({
  name: "volume-form",
  components: {
    RunaiRadioOptions,
    PolicyStringField,
    ClaimInfoSection,
  },
  emits: ["update-volume", "remove-volume"],
  props: {
    volume: {
      type: Object as PropType<IUIPvcInstance>,
      required: true,
    },
    storageClasses: {
      type: Array as PropType<Array<StorageClassV2>>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<PvcRules | null>,
      required: false,
      default: () => ({}),
    },
    showVolumePersistencySection: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    loadingStorageClass: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disableInfo: {
      type: String as PropType<string>,
      required: false,
    },
  },
  computed: {
    claimInfo(): ClaimInfo {
      return {
        storageClass: this.volume.claimInfo?.storageClass,
        size: this.volume.claimInfo?.size,
        accessModes: this.volume.claimInfo?.accessModes,
        volumeMode: this.volume.claimInfo?.volumeMode,
      };
    },
    storageClass(): string | null {
      return this.volume.claimInfo?.storageClass === NONE_STORAGE_CLASS
        ? NONE_STORAGE_CLASS_LABEL
        : this.volume.claimInfo?.storageClass || null;
    },
    selectedStorageClass(): StorageClassV2 | undefined {
      return this.storageClasses.find((sc) => this.volume.claimInfo?.storageClass === sc.storageClassName);
    },
    volumePersistencyOptions(): ISelectOption[] {
      return [
        {
          value: false,
          label: "Persistent",
          toolTip: "The volume and its data will be deleted only when the workload is deleted.",
          disable: this.selectedStorageClass?.permissions?.allowedForPersistentVolumes === false || false,
          disabledTooltip: "This can't be modified based on the storage class setup",
        },
        {
          value: true,
          label: "Ephemeral",
          toolTip: `The volume and its data will be deleted every time the workload's status changes to "Stopped".`,
          disable: this.selectedStorageClass?.permissions?.allowedForEphemeralVolumes === false || false,
          disabledTooltip: "This can't be modified based on the storage class setup",
        },
      ];
    },
  },
  methods: {
    updateClaimInfo(claimInfo: ClaimInfo): void {
      const selectedStorageClass = this.storageClasses.find((sc) => sc.storageClassName === claimInfo.storageClass);
      let ephemeral = this.volume.ephemeral;

      // AllowedForPersistentVolumes and AllowedForEphemeralVolumes can not be both false, therefore we can be sure that only one of the conditions will be true
      if (selectedStorageClass?.permissions?.allowedForPersistentVolumes === false) {
        ephemeral = true;
      }
      if (selectedStorageClass?.permissions?.allowedForEphemeralVolumes === false) {
        ephemeral = false;
      }
      this.volumeChanged({ ...this.volume, claimInfo, ephemeral });
    },
    volumePesistencyChanged(ephemeral: boolean): void {
      this.volumeChanged({ ...this.volume, ephemeral });
    },
    pathChanged(path: string | number | null): void {
      this.volumeChanged({ ...this.volume, path: path as string });
    },
    sizeChanged(size: string): void {
      this.volumeChanged({ ...this.volume, claimInfo: { ...this.volume.claimInfo, size } });
    },
    volumeChanged(volume: IUIPvcInstance): void {
      this.$emit("update-volume", volume);
    },
    // Rules
    isDirectoryPath(val: string): boolean | string {
      return isValidDirectoryPath(val) || errorMessages.INVALID_DIRECTORY_PATH;
    },
    notEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.PATH_NOT_EMPTY;
    },
  },
});
</script>
