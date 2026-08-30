<template>
  <section class="scale-down-section">
    <div class="q-mt-md">
      Set when the replicas should be automatically scaled down to zero
      <runai-tooltip :tooltip-text="scaleDownTooltip" tooltip-position="right" width="315px" />
    </div>
    <div class="row items-center gap-15 q-mb-lg">
      <div>
        <runai-select
          aid="scale-down-select"
          class="scale-down-select"
          :model-value="scaleDownSelected"
          @update:model-value="onScaleDownChanged"
          :options="scaleDownOptions"
          :disable="isScaleDownDisabled"
          emit-value
        />
        <q-tooltip v-if="isScaleDownDisabled" anchor="bottom middle" self="top middle">
          {{ scaleDownDisabledTooltip }}
        </q-tooltip>
      </div>

      <div v-if="isCustomScaleDownToZeroSelected" class="custom-auto-scale row flex-1 gap-15">
        <span class="custom-auto-scale-label">After</span>
        <policy-number-input
          class="col-4"
          :model-value="customScaleDownSelectedAsMinutes"
          @update:model-value="updateCustomScaleToZero"
          stack-label
          :min-value="1"
          :max-value="60"
          :rules="[isValidScaleToZero]"
          :policy-rules="policyRules?.scaleToZeroRetentionSeconds"
          :disable="disabled"
          aid="scale-to-zero-retention-seconds-input"
        />
        <span class="custom-auto-scale-label custom-auto-scale-label-minutes">minutes of inactivity</span>
      </div>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiSelect } from "@/components/common/runai-select";
import { PolicyNumberInput } from "@/components/common/policy-number-input";

// models
import { type AutoScalingRules } from "@/swagger-models/assets-service-client";
import {
  CUSTOM_SCALE_DOWN_ID,
  EScaleDownSectionEvents,
  type IScaleDownData,
  MIN_REPLICAS_TO_ALLOW_SCALE_TO_ZERO_OPTIONS,
  NEVER_SCALE_DOWN_ID,
  SCALE_DOWN_OPTIONS,
  SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE,
} from "./scale-down-section.model";
import type { ISelectOption } from "@/models/global.model";

// utils
import { deepCopy } from "@/utils/common.util";
import { errorMessages } from "@/common/error-message.constant";
import { scaleDownUtil } from "./scale-down.util";
import type { AutoScaling } from "@/swagger-models/workloads-client";

function toMinutes(seconds: number): number {
  return seconds / 60;
}

function toSeconds(minutes: number): number {
  return minutes * 60;
}

export default defineComponent({
  name: "scale-down-section",
  components: {
    RunaiTooltip,
    RunaiSelect,
    PolicyNumberInput,
  },
  emits: {
    [EScaleDownSectionEvents.UpdateModelValue]: (_data: IScaleDownData) => true,
  },
  props: {
    modelValue: {
      type: Object as PropType<IScaleDownData>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<AutoScalingRules | null>,
      required: false,
    },
    isAutoScaleEnabled: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    policyDefaults: {
      type: [Object, null] as PropType<AutoScaling | null>,
      required: false,
    },
  },
  data() {
    return {
      scaleDownSelected: null as ISelectOption | null,
      scaleDownTooltip:
        "Compute resources can be freed up when the model is inactive (i.e., there are no requests being sent)",
      scaleDownOptions: [] as ISelectOption[],
      neverOption: {} as ISelectOption,
      minMinutesForScaleToZero: 1 as number,
      maxMinutesForScaleToZero: 60 as number,
      customScaleToZeroMessageError: "" as string,
      customScaleDownSelectedAsMinutes: 0 as number,
    };
  },
  created() {
    this.initializeScaleDownOptions();
    this.neverOption = this.getNeverOption();
    this.customScaleToZeroMessageError = errorMessages.ENTER_A_NUMBER_FROM_MIN_TO_MAX.replace(
      "${1}",
      this.minMinutesForScaleToZero.toString(),
    ).replace("${2}", this.maxMinutesForScaleToZero.toString());

    this.initializeScaleDownSelectedOption(this.modelValue.scaleToZeroRetentionSeconds);
  },
  computed: {
    isCustomScaleDownToZeroSelected(): boolean {
      return this.scaleDownSelected?.value == CUSTOM_SCALE_DOWN_ID;
    },
    isScaleDownDisabled(): boolean {
      return (
        this.disabled || !this.isAutoScaleEnabled || this.policyRules?.scaleToZeroRetentionSeconds?.canEdit === false
      );
    },
    scaleDownDisabledTooltip(): string {
      if (this.policyRules?.scaleToZeroRetentionSeconds?.canEdit === false) {
        return errorMessages.CANT_BE_MODIFIED_POLICY;
      }

      if (!this.isAutoScaleEnabled) {
        return "Automatic scaling to zero is enabled only when the minimum number of replicas is 0";
      }

      return "";
    },
  },
  methods: {
    isValidScaleToZero(val: string): boolean | string {
      return (
        (Number(val) >= this.minMinutesForScaleToZero && Number(val) <= this.maxMinutesForScaleToZero) ||
        this.customScaleToZeroMessageError
      );
    },
    updateCustomScaleToZero(minutes: number | string | null): void {
      if (minutes === null || minutes === undefined || minutes === "") {
        minutes = 1;
      }

      const scaleToZeroRetentionMinutes = Math.round(Number(minutes));
      this.customScaleDownSelectedAsMinutes = scaleToZeroRetentionMinutes;
      this.updateModel({
        ...this.modelValue,
        scaleToZeroRetentionSeconds: toSeconds(scaleToZeroRetentionMinutes),
      });
    },
    onScaleDownChanged(scaleDownSelected: string): void {
      const scaleDownOption: ISelectOption | undefined = this.scaleDownOptions.find(
        (option: ISelectOption) => option.value === scaleDownSelected,
      );
      if (!scaleDownOption) return;

      this.scaleDownSelected = scaleDownOption;
      let scaleToZeroRetentionSeconds;
      this.customScaleDownSelectedAsMinutes = 0;
      if (scaleDownOption.value === this.neverOption.value) {
        scaleToZeroRetentionSeconds = undefined;
      } else if (scaleDownOption.value === CUSTOM_SCALE_DOWN_ID) {
        scaleToZeroRetentionSeconds = scaleDownUtil.getScaleToZeroRetentionSecondsDefaultValue(
          undefined,
          this.policyDefaults,
          this.policyRules,
        );
        this.customScaleDownSelectedAsMinutes = toMinutes(scaleToZeroRetentionSeconds);
      } else {
        scaleToZeroRetentionSeconds = Number(scaleDownOption.value);
      }
      this.updateModel({
        scaleToZeroRetentionSeconds,
      });
    },
    updateModel(data: IScaleDownData): void {
      this.$emit(EScaleDownSectionEvents.UpdateModelValue, data);
    },
    getNeverOption(): ISelectOption {
      return this.getScaleDownOptionById(NEVER_SCALE_DOWN_ID);
    },
    getScaleDownOptionById(id: string): ISelectOption {
      const option = this.scaleDownOptions.find((opt: ISelectOption) => opt.value === id);
      if (!option) {
        throw new Error(`Scale down option with id ${id} not found`);
      }
      return option;
    },
    findScaleDownOptionByRetentionSeconds(retentionSeconds: number | undefined): ISelectOption {
      // If undefined or zero, return never option
      if (retentionSeconds === undefined || retentionSeconds === 0) {
        return this.neverOption;
      }

      // First try to find a matching predefined option
      const matchingOption = this.scaleDownOptions.find(
        (option: ISelectOption) => option.value === String(retentionSeconds),
      );
      if (matchingOption) return matchingOption;

      // If retention seconds is greater than 0, return custom option
      if (retentionSeconds > MIN_REPLICAS_TO_ALLOW_SCALE_TO_ZERO_OPTIONS) {
        return this.getScaleDownOptionById(CUSTOM_SCALE_DOWN_ID);
      }

      // Default to never option for any other case
      return this.neverOption;
    },
    initializeScaleDownOptions(): void {
      this.scaleDownOptions = deepCopy(SCALE_DOWN_OPTIONS);
      if (this.policyRules?.scaleToZeroRetentionSeconds?.min || this.policyRules?.scaleToZeroRetentionSeconds?.max) {
        const min = this.policyRules?.scaleToZeroRetentionSeconds?.min || undefined;
        const max = this.policyRules?.scaleToZeroRetentionSeconds?.max || undefined;
        this.scaleDownOptions = this.scaleDownOptions.filter((option: ISelectOption) => {
          if (option.value === CUSTOM_SCALE_DOWN_ID || option.value === NEVER_SCALE_DOWN_ID) {
            return true;
          }

          const isMinAllowed = min !== undefined ? Number(option.value) >= min : true;
          const isMaxAllowed = max !== undefined ? Number(option.value) <= max : true;

          return isMinAllowed && isMaxAllowed;
        });
      }
    },
    initializeScaleDownSelectedOption(retentionSeconds: number | undefined): void {
      this.scaleDownSelected = this.findScaleDownOptionByRetentionSeconds(retentionSeconds);
      if (this.scaleDownSelected.value === CUSTOM_SCALE_DOWN_ID) {
        this.customScaleDownSelectedAsMinutes = retentionSeconds
          ? toMinutes(retentionSeconds)
          : SCALE_TO_ZERO_RETENTION_IN_MINUTES_DEFAULT_VALUE;
      }

      if (this.isAutoScaleEnabled) {
        this.neverOption.disable = true;
        this.neverOption.disabledTooltip = "To enable this, set the minimum number of replicas to at least 1";
      } else {
        this.neverOption.disable = false;
        this.neverOption.disabledTooltip = "";
      }
    },
  },
});
</script>
<style lang="scss" scoped>
.scale-down-section {
  .scale-down-select {
    width: 215px;
  }

  .custom-auto-scale {
    .custom-auto-scale-label {
      display: flex;
      align-items: center;
      padding-bottom: 22px;
    }
  }
}
</style>
