<template>
  <section class="secret-volume-instance-section">
    <div class="q-py-md">Set the data origin</div>
    <div class="credentials-section q-mb-md row">
      <runai-tooltip-wrapper
        :display-tooltip="allowOnlyContainerPath"
        tooltip-text="This was defined when the setup was created and can't be modified"
        anchor="bottom middle"
        self="bottom middle"
        class="col-6"
      >
        <runai-select
          :disable="allowOnlyContainerPath || disable"
          aid="secret-select"
          :model-value="modelValue.secret"
          @update:model-value="$emit('update:model-value', { ...modelValue, credentialAssetId: $event })"
          :loading="loadingSecrets"
          :options="secretOptions"
          map-options
          label="Secret"
          emit-value
          no-option-text="No secrets found"
          :rules="[isCredSelected]"
        />
      </runai-tooltip-wrapper>
    </div>

    <div class="q-py-md">{{ containerPathTitle }}</div>
    <div class="row">
      <q-input
        :disable="disable"
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
      />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { IUISecretVolumeInstance } from "@/models/workload.model";
// Components
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { errorMessages } from "@/common/error-message.constant";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import type { ISelectOption } from "@/models/global.model";
import { validationRules } from "@/common/validation-rules";

export default defineComponent({
  name: "secret-volume-instance-section",
  components: {
    RunaiSelect,
    RunaiTooltipWrapper,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<IUISecretVolumeInstance>,
      required: true,
    },
    secrets: {
      type: Array as PropType<string[]>,
      required: true,
    },
    loadingSecrets: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    allowOnlyContainerPath: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    disable: {
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
    containerPathTitle(): string {
      return assetsUtil.containerPathTitle();
    },
    secretOptions(): ISelectOption[] {
      const options: ISelectOption[] = this.secrets.map((secret) => ({ label: secret, value: secret }));
      return options;
    },
  },
  methods: {
    isCredSelected(val: string): boolean | string {
      return this.isCredValueValid(val) || errorMessages.SELECT_CREDENTIAL;
    },
    isCredValueValid(val: string | null | undefined) {
      return val === null || !!val?.length;
    },
  },
});
</script>
