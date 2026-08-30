<template>
  <runai-expansion-item
    class="inference-architecture-section"
    aid="inference-architecture-section"
    label="Workload architecture"
    :section-invalid="sectionInvalid"
    :default-opened="isSelectable"
    :disable-opening="!isSelectable"
    :disable-closing="isSelectable"
    hide-expend-icon
  >
    <template #subheader>
      <div class="row items-center justify-between q-pr-sm" aid="inference-architecture-section-summary">
        <span>{{ summary }}</span>
      </div>
    </template>

    <runai-radio-options
      title="Set whether the workload should run as a standard or distributed workload "
      :model-value="modelValue"
      :options="architectureOptions"
      @update:model-value="modelSelectionChanged"
    />
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";

// models
import type { ISelectOption } from "@/models/global.model";
import { EInferenceArchitecture, type IInferenceArchitectureClusterFeatures } from "@/models/inference.model";

const ARCHITECTURE_OPTION_TOOLTIP = {
  standardDisabledNoKnative:
    "Standard inference can't be selected because Knative is not installed on the selected cluster. Contact your administrator",
  distributedDisabledNoLws:
    "Distributed inference can't be selected because Leader Worker Set (LWS) is not installed on the selected cluster. Contact your administrator",
  standardInfo:
    "This workload will only be able to use environments that support standard workloads. A standard workload consists of a single process.",
  distributedInfo:
    "This workload will only be able to use environments that support distributed workloads. A distributed workload consists of multiple processes working together. These processes can run on different nodes.",
} as const;

const ARCHITECTURE_OPTION_TOOLTIP_WIDTH = {
  standardInfo: "407px",
  distributedInfo: "566px",
  disabledReason: "300px",
} as const;

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiRadioOptions,
  },
  emits: {
    "update-model-value": (inferenceArchitecture: EInferenceArchitecture) => inferenceArchitecture,
  },
  props: {
    modelValue: {
      type: String as PropType<EInferenceArchitecture>,
      required: true,
    },
    isSelectable: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    clusterArchitectureFeatures: {
      type: Object as PropType<IInferenceArchitectureClusterFeatures>,
      default: (): IInferenceArchitectureClusterFeatures => ({
        isKnativeInstalled: true,
        isLwsInstalled: true,
      }),
    },
  },
  computed: {
    architectureOptions(): Array<ISelectOption> {
      const standardDisabled = !this.clusterArchitectureFeatures.isKnativeInstalled;
      const distributedDisabled = !this.clusterArchitectureFeatures.isLwsInstalled;
      return [
        {
          value: EInferenceArchitecture.Standard,
          label: EInferenceArchitecture.Standard.toString(),
          toolTip: ARCHITECTURE_OPTION_TOOLTIP.standardInfo,
          toolTipWidth: ARCHITECTURE_OPTION_TOOLTIP_WIDTH.standardInfo,
          disable: standardDisabled,
          disabledTooltip: standardDisabled ? ARCHITECTURE_OPTION_TOOLTIP.standardDisabledNoKnative : undefined,
          disabledTooltipWidth: standardDisabled ? ARCHITECTURE_OPTION_TOOLTIP_WIDTH.disabledReason : undefined,
        },
        {
          value: EInferenceArchitecture.Distributed,
          label: EInferenceArchitecture.Distributed.toString(),
          toolTip: ARCHITECTURE_OPTION_TOOLTIP.distributedInfo,
          toolTipWidth: ARCHITECTURE_OPTION_TOOLTIP_WIDTH.distributedInfo,
          disable: distributedDisabled,
          disabledTooltip: distributedDisabled ? ARCHITECTURE_OPTION_TOOLTIP.distributedDisabledNoLws : undefined,
          disabledTooltipWidth: distributedDisabled ? ARCHITECTURE_OPTION_TOOLTIP_WIDTH.disabledReason : undefined,
        },
      ];
    },
    sectionInvalid(): boolean {
      return false;
    },
    summary(): string {
      switch (this.modelValue) {
        case EInferenceArchitecture.Standard:
          return EInferenceArchitecture.Standard;
        case EInferenceArchitecture.Distributed:
          return EInferenceArchitecture.Distributed;
        default:
          return EInferenceArchitecture.Standard;
      }
    },
  },
  methods: {
    modelSelectionChanged(modelValue: EInferenceArchitecture): void {
      this.$emit("update-model-value", modelValue);
    },
  },
});
</script>
<style lang="scss" scoped></style>
