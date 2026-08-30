<template>
  <runai-radio-options
    title="Set how the model profile should be selected"
    title-tooltip="A NIM model profile sets compatible model engines and criteria for engine selection, such as precision, latency, throughput optimization, and GPU requirements. Profiles are optimized to balance either latency or throughput, with quantized profiles (e.g., fp8) preferred to reduce memory usage and enhance performance."
    title-tooltip-width="535px"
    tooltip-width="400px"
    :model-value="profileMethod"
    :options="profilesOptions"
    @update:model-value="onProfileSelectionChanged"
  />

  <div class="row q-mt-sm">
    <policy-string-field
      v-if="isProfileManually"
      class="col-9"
      aid="profile-input-input"
      @update:model-value="onProfileInputChanged"
      :model-value="profileNameOrHash"
      label="Profile name or hash"
      placeholder="e.g., tensorrt_llm-h100-fp8-tp2-latency"
      stack-label
      input-class="placeholder-italic"
      :rules="profileValidationRules"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { PolicyStringField } from "@/components/common/policy-string-field";

// models
import { EProfileNameOrHashOptions } from "./nim-profile.model";
import type { ISelectOption } from "@/models/global.model";
import type { ValidationRule } from "quasar";

// others
import { errorMessages } from "@/common/error-message.constant";
import { isNotEmpty, nimProfileValidator } from "@/common/form.validators";

export default defineComponent({
  components: {
    RunaiRadioOptions,
    PolicyStringField,
  },
  emits: ["nim-profile-changed", "is-valid"],
  props: {
    profileNameOrHash: {
      type: String as PropType<string>,
      required: true,
      default: "",
    },
    isProfileRequired: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
  },
  data() {
    return {
      profileMethod: EProfileNameOrHashOptions.Automatically as EProfileNameOrHashOptions,
      profilesOptions: [
        {
          value: EProfileNameOrHashOptions.Automatically,
          label: "Automatically (recommended) ",
          toolTip:
            "NIM is designed to automatically select the most suitable profile from the list of compatible profiles based on the detected hardware. Each profile consists of different parameters that influence the selection process.",
        },
        {
          value: EProfileNameOrHashOptions.Manually,
          label: "Manually",
          toolTip: "Enter the profile name or unique hash identifier",
        },
      ] as Array<ISelectOption>,
      localProfileNameOrHash: "" as string,
      profileValidationRules: [] as Array<ValidationRule>,
    };
  },
  created() {
    if (this.profileNameOrHash) {
      this.profileMethod = EProfileNameOrHashOptions.Manually;
      this.localProfileNameOrHash = this.profileNameOrHash;
    }

    if (this.isProfileRequired) {
      this.profileValidationRules = [this.profileInputNotEmpty, this.profileIsValid];
    } else {
      this.profileValidationRules = [this.profileInputNotRequired];
    }
  },
  computed: {
    isProfileManually(): boolean {
      return this.profileMethod === EProfileNameOrHashOptions.Manually;
    },
  },
  methods: {
    profileInputNotEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.PROFILE_NAME_OR_HASH_NOT_EMPTY;
    },
    profileIsValid(val: string): boolean | string {
      return nimProfileValidator(val) || errorMessages.ENTER_A_VALID_PROFILE_NAME_OR_HASH;
    },
    profileInputNotRequired(val: string): boolean | string {
      return val ? this.profileIsValid(val) : true;
    },
    onProfileSelectionChanged(profileMethod: EProfileNameOrHashOptions): void {
      this.profileMethod = profileMethod;
      if (profileMethod === EProfileNameOrHashOptions.Manually) {
        this.emitChanges(this.localProfileNameOrHash);
      } else {
        this.emitChanges("");
        this.$emit("is-valid", true);
      }
    },
    onProfileInputChanged(profileNameOrHash: string): void {
      this.localProfileNameOrHash = profileNameOrHash;
      this.emitChanges(profileNameOrHash);
    },
    emitChanges(profileNameOrHash: string): void {
      this.$emit("nim-profile-changed", profileNameOrHash);
      const isValid = isNotEmpty(profileNameOrHash) && nimProfileValidator(profileNameOrHash);
      this.$emit("is-valid", isValid);
    },
  },
});
</script>
