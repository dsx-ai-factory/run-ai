<template>
  <section class="host-path-instance-section">
    <section>
      <div class="q-mb-md">
        <div class="q-py-md">Set the data origin</div>
        <runai-tooltip-wrapper
          :display-tooltip="allowOnlyContainerPath"
          tooltip-text="This was defined when the setup was created and can't be modified"
          anchor="bottom middle"
          self="bottom middle"
        >
          <policy-string-field
            :disable="allowOnlyContainerPath || disable"
            :tooltip="disableInfo"
            aid="host-path-input"
            :model-value="modelValue.path"
            no-error-icon
            label="Host path"
            stack-label
            placeholder="e.g., /usr/data"
            input-class="placeholder-italic"
            :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
            @update:model-value="$emit('update:model-value', { ...modelValue, path: $event })"
            :policy-rules="policyRules?.path"
          />
        </runai-tooltip-wrapper>
      </div>

      <div>
        <div class="q-py-md">{{ containerPathTitle }}</div>
        <div class="row">
          <policy-string-field
            :disable="disable"
            :tooltip="disableInfo"
            class="col-6"
            aid="mount-path-input"
            :model-value="modelValue.mountPath"
            no-error-icon
            label="Container path"
            stack-label
            placeholder="e.g., /home/local/data"
            input-class="placeholder-italic"
            :rules="[validationRules.validatePathNotEmpty, validationRules.validatePathFormat]"
            @update:model-value="$emit('update:model-value', { ...modelValue, mountPath: $event })"
            :policy-rules="policyRules?.mountPath"
          />
        </div>
      </div>
    </section>
    <section v-if="showRestrictions">
      <runai-sub-expansion-item label="Restrictions">
        <section class="row items-center q-mb-md">
          <runai-toggle
            :disable="disable"
            aid="read-only-toggle"
            :model-value="modelValue.readOnly || false"
            label="Prevent data modification"
            left-label
            class="text-no-wrap"
            @update:model-value="$emit('update:model-value', { ...modelValue, readOnly: $event })"
            :policy-rules="policyRules?.readOnly"
          />
          <runai-tooltip
            class="q-ml-md"
            tooltip-position="right"
            tooltip-text="When enabled, the data will be mounted with read-only permissions."
          />
        </section>
      </runai-sub-expansion-item>
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSubExpansionItem } from "@/components/common/runai-sub-expansion-item";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { RunaiToggle } from "@/components/common";
// Models
import { assetsUtil } from "@/utils/assets.util/assets.util";
import type { IUIHostPathInstance } from "@/models/workload.model";
import type { HostPathRules } from "@/swagger-models/assets-service-client";
import { validationRules } from "@/common/validation-rules";

export default defineComponent({
  name: "host-path-instance-section",
  components: {
    RunaiSubExpansionItem,
    RunaiTooltip,
    RunaiTooltipWrapper,
    PolicyStringField,
    RunaiToggle,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<IUIHostPathInstance>,
      required: true,
    },
    showRestrictions: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    allowOnlyContainerPath: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<HostPathRules | null>,
      required: false,
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
