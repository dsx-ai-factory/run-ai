<template>
  <section class="pvc-instance-section">
    <div class="q-mb-md">
      <div class="q-py-md">Set the data origin</div>
      <div class="row">
        <runai-tooltip-wrapper
          :display-tooltip="allowOnlyContainerPath"
          tooltip-text="This was defined when the setup was created and can't be modified"
          anchor="bottom middle"
          self="bottom middle"
          class="col-6"
        >
          <runai-select
            :disable="allowOnlyContainerPath || disable"
            class="claim-name-select"
            standard
            :model-value="modelValue.claimName"
            :options="pvcClaimNames"
            label="Claim name"
            stack-label
            placeholder="Select..."
            :rules="[validationRules.validateNameSelected]"
            @update:model-value="$emit('update:model-value', { ...modelValue, configMap: $event })"
            :loading="loadingPvcClaimNames"
            no-option-text="No names found"
          />
        </runai-tooltip-wrapper>
      </div>
    </div>

    <div>
      <div class="q-py-md">{{ containerPathTitle }}</div>
      <div class="row">
        <q-input
          :disable="disable"
          class="pvc-mount-path-input col-6"
          aid="mount-path-input"
          :model-value="modelValue.path"
          no-error-icon
          label="Container path"
          stack-label
          placeholder="e.g., /home/local/data"
          input-class="placeholder-italic"
          :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
          @update:model-value="$emit('update:model-value', { ...modelValue, path: $event })"
        />
        <q-tooltip class="explanatory-tooltip" target=".pvc-mount-path-input" v-if="disableInfo" max-width="350px">{{
          disableInfo
        }}</q-tooltip>
      </div>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
// Models
import type { IUIPvcInstance } from "@/models/workload.model";
// Constants
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { validationRules } from "@/common/validation-rules";

export default defineComponent({
  name: "pvc-instance-section",
  components: {
    RunaiSelect,
    RunaiTooltipWrapper,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<IUIPvcInstance>,
      required: true,
    },
    allowOnlyContainerPath: {
      type: Boolean,
      default: false,
    },
    pvcClaimNames: {
      type: Array as PropType<string[]>,
      required: true,
    },
    loadingPvcClaimNames: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disableInfo: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      validationRules,
    };
  },
  computed: {
    containerPathTitle(): string {
      return assetsUtil.containerPathTitle();
    },
  },
});
</script>
