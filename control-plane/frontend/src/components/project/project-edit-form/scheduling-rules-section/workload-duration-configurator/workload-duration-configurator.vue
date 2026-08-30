<template>
  <wrapper-card
    :aid="wrapperAid"
    @close="$emit('close')"
    :header="header"
    :subheader="subheader"
    :hide-close-button="hideCloseConfiguratorButton"
  >
    <template #body>
      <gray-card
        :hide-close-button="index === 0 && workloadsDuration.length == 1"
        @close="hideWorkloadSection(index)"
        @reset="resetWorkloadDuration(workload, index)"
        v-for="(workload, index) in workloadsDuration"
        :key="workload.value"
        :ph-icon-name="departmentIcon"
        :title="getCardTitle(workload)"
        :reset-tooltip-text="getResetTooltipText(workload)"
        :aid="workloadDurationAid(workload)"
      >
        <div class="row items-center q-pt-md">
          <div class="col-4 q-mr-md">
            <runai-select
              label="Workload type"
              @update:model-value="onWorkloadSelect(workload, $event, index)"
              :model-value="getOptionValue(workload)"
              :rules="[workloadTypeValidation]"
              :options="workloadOptions"
              aid="gpu-workload-type"
              :disable="!!workload.inheritedDuration || readOnly"
            />
          </div>
          <div class="col">
            <runai-duration-input
              hide-seconds
              @update:model-value="updateDuration(workload, $event, index)"
              :model-value="getWorkloadDuration(workload)"
              :rules="[durationValidation, ($event) => inheritedDurationValidation($event, workload)]"
              :disable="readOnly"
            />
          </div>
        </div>
      </gray-card>
    </template>

    <template #actions>
      <q-btn
        aid="add-worklaod-duration-button"
        v-if="showAddDurationButton"
        @click="addWorkloadSection"
        :disable="readOnly"
        color="primary"
        flat
        class="btn-icon-sm"
        icon="ph-bold ph-plus"
        :label="addDurationButtonLabel"
      />
    </template>
  </wrapper-card>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
// cmps
import { RunaiDurationInput } from "@/components/common/runai-duration-input";
import { WrapperCard } from "@/components/project/project-edit-form/cards/";
import { GrayCard } from "@/components/project/project-edit-form/cards/";
import { RunaiSelect } from "@/components/common/runai-select";
// models
import type { ISelectOption } from "@/models/global.model";
import type { IWorkloadDurationOption } from "@/models/workload.model";
import { errorMessages } from "@/common/error-message.constant";
import { dateUtil } from "@/utils/date.util";
import { stringUtil } from "@/utils/string.util";
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";

export default defineComponent({
  name: "workload-duration-configurator",
  components: { GrayCard, WrapperCard, RunaiDurationInput, RunaiSelect },
  emits: ["close", "add-duration", "remove-duration", "duration-changed"],
  props: {
    header: {
      type: String,
      required: true,
    },
    subheader: {
      type: String,
      required: true,
    },
    workloadsDuration: {
      type: Array as PropType<IWorkloadDurationOption[]>,
      required: true,
    },
    workloadOptions: {
      type: Array as PropType<ISelectOption[]>,
      required: true,
    },
    addDurationButtonLabel: {
      type: String,
      required: true,
    },
    readOnly: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      departmentIcon: toQuasarIconName(runaiIcons.entity.department),
    };
  },
  computed: {
    wrapperAid(): string {
      return stringUtil.slug(this.header);
    },
    showAddDurationButton() {
      return this.workloadsDuration.length < this.workloadOptions.length;
    },
    hideCloseConfiguratorButton(): boolean {
      return this.workloadsDuration.some(
        (workload) => workload.inheritedDuration !== null && workload.inheritedDuration !== undefined,
      );
    },
  },
  methods: {
    getCardTitle(workload: IWorkloadDurationOption): string | undefined {
      return workload.inheritedDuration
        ? `This rule is inherited from the parent department's configurations and can't exceed ${dateUtil.formatTimestamp(
            workload.inheritedDuration,
          )}`
        : undefined;
    },
    workloadDurationAid(workload: IWorkloadDurationOption): string {
      return `${stringUtil.slug(workload.value)}-workload-duration`;
    },
    getResetTooltipText(workload: IWorkloadDurationOption): string | undefined {
      if (workload.inheritedDuration) {
        return "Reset to department's default";
      }
    },
    durationValidation(val: number): boolean | string {
      return val <= 0 ? errorMessages.DURATION_GREATER_THAN_ZERO : true;
    },
    inheritedDurationValidation(val: number, workload: IWorkloadDurationOption): boolean | string {
      if (workload.inheritedDuration === null || workload.inheritedDuration === undefined) return true;
      return val > workload.inheritedDuration
        ? `Enter a limit lower than ${dateUtil.formatTimestamp(workload.inheritedDuration)}`
        : true;
    },
    workloadTypeValidation(option: ISelectOption | null): boolean | string {
      return !!option || errorMessages.SELECT_TYPE;
    },
    getWorkloadDuration(workload: IWorkloadDurationOption): number {
      if (workload.duration === null || workload.duration === undefined) {
        return 0;
      }
      return workload.duration;
    },
    getOptionValue(workload: ISelectOption): null | ISelectOption {
      if (workload.value) {
        return workload;
      }
      return null;
    },
    addWorkloadSection(): void {
      this.$emit("add-duration");
    },
    hideWorkloadSection(workloadIndex: number): void {
      this.$emit("remove-duration", workloadIndex);
    },
    onWorkloadSelect(workload: IWorkloadDurationOption, selectedWorkload: ISelectOption, workloadIndex: number): void {
      this.$emit("duration-changed", { ...selectedWorkload, duration: workload.duration }, workloadIndex);
    },
    updateDuration(workload: IWorkloadDurationOption, duration: number, workloadIndex: number): void {
      this.$emit("duration-changed", { ...workload, duration }, workloadIndex);
    },
    resetWorkloadDuration(workload: IWorkloadDurationOption, workloadIndex: number): void {
      this.$emit("duration-changed", { ...workload, duration: workload.inheritedDuration }, workloadIndex);
    },
  },
});
</script>
