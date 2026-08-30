<template>
  <section class="claim-info-section">
    <section class="row">
      <runai-select
        aid="storage-class-select"
        class="storage-class-select col-6"
        standard
        unclearable
        :model-value="modelValue.storageClass || null"
        :options="storageClassOptions"
        label="Storage class"
        @update:model-value="storageClassChanged"
        :rules="[isStorageClassSelected]"
        :disable="disable"
        emit-value
        no-option-text="No storage classes found"
        :policy-rules="storageClassRules"
      />
      <q-tooltip
        class="explanatory-tooltip"
        target=".storage-class-select"
        v-if="disableInfo || policyRules?.storageClass?.canEdit === false"
        max-width="350px"
        >{{ disableInfo || errorMessages.CANT_BE_MODIFIED_POLICY }}</q-tooltip
      >
    </section>
    <section class="row">
      <runai-select
        aid="access-mode-select"
        class="access-mode-select col-6"
        standard
        counter
        max-values="3"
        label="Access mode"
        multiple
        :model-value="accessModesModel"
        option-label="label"
        map-options
        lazy-rules
        :options="accessModeOptions"
        no-error-icon
        emit-value
        :rules="[isAccessModeSelected]"
        @update:model-value="accessModeChanged"
        :disable="disable"
        :loading="loading"
      >
        <template #option="scope">
          <q-item v-bind="scope.itemProps" class="row justify-between items-center">
            <q-item-label>{{ scope.opt.label }}</q-item-label>
            <runai-tooltip class="q-ml-md" tooltip-position="right" :tooltip-text="scope.opt.toolTip" />
          </q-item>
        </template>
      </runai-select>
      <q-tooltip class="explanatory-tooltip" target=".access-mode-select" v-if="disableInfo" max-width="350px">{{
        disableInfo
      }}</q-tooltip>
    </section>
    <section class="row">
      <volume-size-section
        class="volume-size-section col-6"
        :model-value="modelValue.size || ''"
        @update:model-value="sizeChanged"
        :disable="disable"
        :policy-rules="policyRules?.size"
        :loading="loading"
        :claim-size-customization="selectedStorageClass?.customization?.claimSize || undefined"
      />
      <q-tooltip class="explanatory-tooltip" target=".volume-size-section" v-if="disableInfo" max-width="350px">{{
        disableInfo
      }}</q-tooltip>
    </section>
    <div class="row q-mb-sm">
      <runai-select
        aid="volume-mode-select"
        class="volume-mode-select col-6"
        label="Volume mode"
        :options="volumeModeOptions"
        :rules="[isVolumeModeSelected]"
        :model-value="modelValue.volumeMode || null"
        @update:model-value="volumeModeChanged"
        :disable="disable"
        emit-value
        :loading="loading"
      />
      <q-tooltip class="explanatory-tooltip" target=".volume-mode-select" v-if="disableInfo" max-width="350px">{{
        disableInfo
      }}</q-tooltip>
    </div>
    <transition name="fade">
      <div class="column" v-if="supportCustomAttributes && !loading">
        <runai-custom-field
          v-for="(field, index) in attributesToCustomFields"
          :key="field.key"
          :field="field"
          @update="updateCustomAttribute(field.key, $event)"
          :class="index !== attributesToCustomFields.length - 1 ? 'q-mb-sm' : ''"
        />
      </div>
    </transition>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

// Models
import {
  PvcAddedAttributeFormat,
  type PvcAddedAttribute,
  type StorageClassV2,
} from "@/swagger-models/k8s-objects-tracker-client";
import {
  type PvcAccessModes,
  PvcVolumeMode,
  type ClaimInfo,
  type ClaimInfoRules,
} from "@/swagger-models/assets-service-client";
import type { ISelectOption } from "@/models/global.model";
import { NONE_STORAGE_CLASS, NONE_STORAGE_CLASS_LABEL } from "@/models/data-source.model";
// Utils
import { isNotEmpty, isNotEmptyArray } from "@/common/form.validators";
// Constants
import { errorMessages } from "@/common/error-message.constant";
// Components
import type { ICustomField } from "@/components/common/runai-custom-field";
import { RunaiCustomField } from "@/components/common/runai-custom-field";
import { RunaiSelect } from "@/components/common/runai-select";
import { VolumeSizeSection } from "@/components/section/volume-section/volume-box/volume-size-section";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";

export type TAccessMode = "readWriteOnce" | "readOnlyMany" | "readWriteMany";

export default defineComponent({
  name: "claim-info-section",
  components: {
    RunaiSelect,
    VolumeSizeSection,
    RunaiTooltip,
    RunaiCustomField,
  },
  emits: {
    "update:model-value": (payload: ClaimInfo) => payload,
  },
  props: {
    modelValue: {
      type: Object as PropType<ClaimInfo>,
      required: true,
    },
    storageClasses: {
      type: Array as PropType<StorageClassV2[]>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    policyRules: {
      type: Object as PropType<ClaimInfoRules>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    disableInfo: {
      type: String as PropType<string>,
      required: false,
    },
    dataSharing: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    supportCustomAttributes: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  data() {
    return {
      errorMessages,
    };
  },
  computed: {
    selectedStorageClassCustomAttributes(): PvcAddedAttribute[] | undefined {
      if (!this.supportCustomAttributes) return undefined;
      return (
        this.storageClasses.find((sc) => sc.storageClassName === this.modelValue.storageClass)?.customization
          ?.attributes || undefined
      );
    },
    attributesToCustomFields(): ICustomField[] {
      if (!this.selectedStorageClassCustomAttributes) return [];

      return this.selectedStorageClassCustomAttributes.map((attr) => {
        const key = attr.key;
        let value: ICustomField["value"] =
          this.modelValue?.addedAttrValues?.find((a) => a.key === attr.key)?.value || null;
        if (!value) {
          if (attr.format === PvcAddedAttributeFormat.Boolean) value = false;
        } else if (typeof value === "string" && attr.format !== PvcAddedAttributeFormat.Text) {
          try {
            value = JSON.parse(value);
          } catch (e) {
            console.error("unable to parse value", e);
          }
        }

        return {
          key,
          value,
          label: attr.display || "",
          format: attr.format || "text",
          tooltip: attr.description || "",
          required: attr.required || undefined,
          min: typeof attr.minValue === "number" ? attr.minValue : undefined,
          max: typeof attr.maxValue === "number" ? attr.maxValue : undefined,
          step: typeof attr.step === "number" ? attr.step : undefined,
          regex: attr.validationRegexp || undefined,
          placeholder: attr.placeholder || undefined,
          invalidMessage: attr.validationError || undefined,
        };
      });
    },
    accessModesModel(): TAccessMode[] {
      const options: TAccessMode[] = [];
      for (const key in this.modelValue.accessModes) {
        if (this.modelValue.accessModes[key as TAccessMode]) options.push(key as TAccessMode);
      }
      return options;
    },
    selectedStorageClass(): StorageClassV2 | undefined {
      return this.storageClasses.find((sc) => this.modelValue.storageClass === sc.storageClassName);
    },
    storageClassOptions(): ISelectOption[] {
      return this.storageClasses.map((storageClass) => {
        const option: ISelectOption = {
          label:
            storageClass.storageClassName === NONE_STORAGE_CLASS
              ? NONE_STORAGE_CLASS_LABEL
              : storageClass.storageClassName,
          value: storageClass.storageClassName,
        };

        if (this.dataSharing) {
          option.disable = !storageClass.permissions?.allowDataSharing;
          option.disabledTooltip = "This can't be selected when the volume is shared";
        }

        return option;
      });
    },
    accessModeOptions(): ISelectOption[] {
      return [
        {
          label: "Read-write by one node",
          toolTip: "The volume can be mounted as read-write by a single node.",
          value: "readWriteOnce",
          disable:
            this.selectedStorageClass?.customization?.accessMode?.supportedValues?.readWriteOnce === false ||
            this.policyRules?.accessModes?.readWriteOnce?.canEdit === false,
          disabledTooltip:
            this.policyRules?.accessModes?.readWriteOnce?.canEdit === false
              ? errorMessages.CANT_BE_MODIFIED_POLICY
              : errorMessages.CANT_BE_SELECTED_STORAGE_CONFIG,
        },
        {
          label: "Read-only by many nodes",
          toolTip: "The volume can be mounted as read-only by many nodes.",
          value: "readOnlyMany",
          disable:
            this.selectedStorageClass?.customization?.accessMode?.supportedValues?.readOnlyMany === false ||
            this.policyRules?.accessModes?.readOnlyMany?.canEdit === false,
          disabledTooltip:
            this.policyRules?.accessModes?.readOnlyMany?.canEdit === false
              ? errorMessages.CANT_BE_MODIFIED_POLICY
              : errorMessages.CANT_BE_SELECTED_STORAGE_CONFIG,
        },
        {
          label: "Read-write by many nodes",
          toolTip: "The volume can be mounted as read-write by many nodes.",
          value: "readWriteMany",
          disable:
            this.selectedStorageClass?.customization?.accessMode?.supportedValues?.readWriteMany === false ||
            this.policyRules?.accessModes?.readWriteMany?.canEdit === false,
          disabledTooltip:
            this.policyRules?.accessModes?.readWriteMany?.canEdit === false
              ? errorMessages.CANT_BE_MODIFIED_POLICY
              : errorMessages.CANT_BE_SELECTED_STORAGE_CONFIG,
        },
      ];
    },
    volumePersistencyOptions(): ISelectOption[] {
      return [
        {
          value: false,
          label: "Persistent",
          toolTip: "The volume and its data will be deleted only when the workload is deleted.",
          disable: this.selectedStorageClass?.permissions?.allowedForPersistentVolumes === false || false,
          disabledTooltip: errorMessages.CANT_BE_MODIFIED_STORAGE_CONFIG,
        },
        {
          value: true,
          label: "Ephemeral",
          toolTip: `The volume and its data will be deleted every time the workload's status changes to "Stopped".`,
          disable: this.selectedStorageClass?.permissions?.allowedForEphemeralVolumes === false || false,
          disabledTooltip: errorMessages.CANT_BE_MODIFIED_STORAGE_CONFIG,
        },
      ];
    },
    volumeModeOptions(): ISelectOption[] {
      return [
        {
          value: PvcVolumeMode.Filesystem,
          label: PvcVolumeMode.Filesystem,
          disable: this.selectedStorageClass?.customization?.volumeMode?.supportedValues?.filesystem === false,
          disabledTooltip: errorMessages.CANT_BE_SELECTED_STORAGE_CONFIG,
        },
        {
          value: PvcVolumeMode.Block,
          label: PvcVolumeMode.Block,
          disable: this.selectedStorageClass?.customization?.volumeMode?.supportedValues?.block === false,
          disabledTooltip: errorMessages.CANT_BE_SELECTED_STORAGE_CONFIG,
        },
      ];
    },
    storageClassRules(): IGenericSelectPolicyRules | undefined {
      return this.policyRules?.storageClass ? { ...this.policyRules?.storageClass, isClosedList: true } : undefined;
    },
  },
  methods: {
    storageClassChanged(storageClass: string): void {
      if (storageClass === NONE_STORAGE_CLASS_LABEL) storageClass = NONE_STORAGE_CLASS;
      this.claimInfoChanged({ ...this.modelValue, storageClass });
    },
    volumeModeChanged(volumeMode: PvcVolumeMode): void {
      this.claimInfoChanged({ ...this.modelValue, volumeMode });
    },
    sizeChanged(size: string): void {
      this.claimInfoChanged({ ...this.modelValue, size });
    },
    accessModeChanged(accessModel: TAccessMode[]): void {
      const claimInfo: ClaimInfo = { ...this.modelValue, accessModes: this.readyAccessModesData(accessModel) };
      this.claimInfoChanged(claimInfo);
    },
    readyAccessModesData(val: TAccessMode[]): PvcAccessModes {
      return val.reduce((acc: PvcAccessModes, curr: TAccessMode) => {
        acc[curr] = true;
        return acc;
      }, {});
    },
    claimInfoChanged(claimInfo: ClaimInfo): void {
      this.$emit("update:model-value", claimInfo);
    },
    // Validation Rules
    isStorageClassSelected(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.STORAGE_CLASS_NOT_EMPTY;
    },
    isAccessModeSelected(val: TAccessMode[]): boolean | string {
      if (!this.selectedStorageClass?.customization?.accessMode?.required) return true;
      return isNotEmptyArray(val) || errorMessages.SELECT_A_MODE;
    },
    isVolumeModeSelected(val: string): boolean | string {
      if (!this.selectedStorageClass?.customization?.volumeMode?.required) return true;
      return isNotEmpty(val) || errorMessages.SELECT_A_MODE;
    },
    updateCustomAttribute(key: string, value: ICustomField["value"]): void {
      if (!this.supportCustomAttributes) return;
      this.claimInfoChanged({
        ...this.modelValue,
        addedAttrValues: this.modelValue.addedAttrValues?.map((attr) => {
          if (attr.key === key) {
            if (!value && value !== 0) return { ...attr, value: undefined };
            if (value && typeof value === "string") return { ...attr, value };
            return { ...attr, value: JSON.stringify(value) };
          }
          return attr;
        }),
      });
    },
  },
});
</script>

<style lang="scss" scoped>
.claim-info-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
</style>
