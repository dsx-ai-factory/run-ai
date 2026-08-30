<template>
  <section class="form-security-section">
    <runai-section title="Set where the UID, GID, and supplementary groups for the container should be taken from">
      <runai-radio-options
        :options="securityOptions"
        :model-value="securityModel.uidGidSource || null"
        @update:model-value="updateUidGidSource"
        :policy-rules="uidGidSourcePolicyRules"
        :disable="disable"
        :disable-tooltip="disableUidGidSourceInfo"
      />
    </runai-section>

    <section v-if="securityModel.uidGidSource === 'custom'" class="custom-option-section">
      <section class="row items-center justify-between q-mt-md">
        <policy-number-input
          class="col-3"
          aid="uid-input"
          :min-value="0"
          is-integer
          :model-value="runAsUid"
          @update:model-value="updateUid"
          label="User ID (UID)"
          stack-label
          :policy-rules="policyRules?.security?.runAsUid"
          :disable="disable"
          :tooltip="disableUidGidSourceInfo"
        />

        <policy-number-input
          aid="gid-input"
          class="col-3"
          :min-value="0"
          is-integer
          :model-value="runAsGid"
          @update:model-value="updateGid"
          label="Group ID (GID)"
          stack-label
          :policy-rules="policyRules?.security?.runAsGid"
          :disable="disable"
          :tooltip="disableUidGidSourceInfo"
        />

        <div class="col-5">
          <policy-string-field
            aid="supplemental-groups-input"
            label="Supplementary groups"
            hint="Add multiple groups separated by commas"
            stack-label
            no-error-icon
            :rules="[isSupplementalGroupsValid]"
            :model-value="securityModel.supplementalGroups || null"
            @update:model-value="updateSupplementalGroups(String($event))"
            :disable="disable || disableSupplementalGroups"
            :policy-rules="policyRules?.security?.supplementalGroups"
            :tooltip="supplementalGroupsTooltip"
          />
        </div>
      </section>
      <section class="row items-center q-mt-lg" v-if="showOverrideToggle">
        <q-toggle
          aid="override-in-workspace-toggle"
          left-label
          label="Allow the values above to be modified within the workload"
          :model-value="securityModel.overrideUidGidInWorkspace"
          @update:model-value="updateOverrideInWorkspace"
          :disable="disable"
        ></q-toggle>
        <runai-tooltip
          tooltip-text="If enabled, the custom values set above can be modified within the workload using this environment. Otherwise, the custom values set above will be used by default."
          tooltip-position="right"
        />
      </section>
    </section>
    <section class="linux-capabilities-section q-my-lg">
      <div class="row items-center">
        <q-item-label>Select additional Linux capabilities for the container</q-item-label>
        <runai-tooltip
          tooltip-text="Grant certain privileges to a container without granting all the root user's privileges."
          tooltip-position="right"
        />
      </div>
      <runai-addable-multi-select
        custom-class="capabilities-select"
        label="Capabilities"
        :options="capabilitiesOptions"
        :selected-options="selectedCapabilitiesOptions || []"
        @options-select="updateSelectedCapabilities"
        :disable="disable"
        :policy-rules="policyRules?.security?.capabilities"
        :options-dense="false"
        :support-adding-new-value="false"
        :disabled-info="disableCapabilitiesInfo"
      />
    </section>
    <q-tooltip anchor="center middle" v-if="tooltipText">{{ tooltipText }}</q-tooltip>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { RunaiAddableMultiSelect } from "@/components/common/runai-addable-multi-select";
// Models
import { type ISecurityModel, type ISecurityOption } from "@/models/environment.model";
import { UidGidSource, Capability } from "@/swagger-models/assets-service-client";
import type { UnifiedPolicyInfoPerReplicaTypeRules } from "@/swagger-models/workloads-client";
import { errorMessages } from "@/common/error-message.constant";
// Utils
import { isCommaSemicolonSeperatedNumber } from "@/common/form.validators";
import { RunaiSection } from "@/components/common/runai-section";
import type { IGenericRadioOptionsPolicyRules } from "@/models/policy.model";
import type { ISelectOption } from "@/models/global.model";
const allLinuxCapabilities: Capability[] = Object.values(Capability);

export default defineComponent({
  name: "form-security-settings",
  components: {
    RunaiTooltip,
    RunaiRadioOptions,
    RunaiSection,
    PolicyNumberInput,
    PolicyStringField,
    RunaiAddableMultiSelect,
  },
  emits: ["update"],
  props: {
    securityModel: {
      type: Object as PropType<ISecurityModel>,
      required: true,
    },
    isSSO: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    showOverrideToggle: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<UnifiedPolicyInfoPerReplicaTypeRules | null>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    disableCapabilitiesInfo: {
      type: String as PropType<string>,
      required: false,
    },
    disableUidGidSourceInfo: {
      type: String as PropType<string>,
      required: false,
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      allLinuxCapabilities,
    };
  },
  computed: {
    runAsGid(): number | null {
      return this.securityModel.runAsGid === undefined ? null : this.securityModel.runAsGid;
    },
    runAsUid(): number | null {
      return this.securityModel.runAsUid === undefined ? null : this.securityModel.runAsUid;
    },
    uidGidSourcePolicyRules(): IGenericRadioOptionsPolicyRules {
      return this.policyRules?.security?.uidGidSource as IGenericRadioOptionsPolicyRules;
    },
    disableSupplementalGroups(): boolean {
      if (this.securityModel.runAsGid === 0) return false;
      return !this.securityModel.runAsGid;
    },
    supplementalGroupsTooltip(): string | undefined {
      if (this.disableSupplementalGroups) return "To modify this value, enter a GID first.";
      return this.disableUidGidSourceInfo;
    },
    securityOptions(): Array<ISecurityOption> {
      const options: Array<ISecurityOption> = [
        {
          label: "From the image",
          value: UidGidSource.FromTheImage,
          toolTip:
            "The UID, GID, and supplementary groups applied to the workload using this environment will be taken from the image.",
          display: true,
        },
        {
          label: "Custom",
          value: UidGidSource.Custom,
          toolTip:
            "The UID, GID, and supplementary groups will be applied to the workload using this environment. if left empty, they will be automatically taken from the image.",
          display: true,
        },
      ];

      if (this.isSSO) {
        options.unshift({
          label: "From the IdP token",
          value: UidGidSource.FromIdpToken,
          toolTip:
            "The workload creator's UID, GID, and supplementary groups will be taken automatically from the IdP token. If they don't exist, the workload creator will be asked to fill them in.",
          display: this.isSSO,
        });
      }

      return options;
    },
    showLinuxCapabilities(): boolean {
      const hide: boolean = this.securityModel.runAsGid === 0 && !this.securityModel.overrideUidGidInWorkspace;
      return !hide;
    },
    capabilitiesOptions(): ISelectOption[] {
      return allLinuxCapabilities.map((c) => ({ label: c, value: c }));
    },
    selectedCapabilitiesOptions(): ISelectOption[] {
      return this.securityModel.capabilities?.map((c) => ({ label: c, value: c })) || [];
    },
  },
  methods: {
    updateSelectedCapabilities(capabilities: ISelectOption[]) {
      const newModel: ISecurityModel = {
        ...this.securityModel,
        capabilities: capabilities.map((c) => c.value as Capability),
      };
      this.updateModel(newModel);
    },
    isSupplementalGroupsValid(val: string): boolean | string {
      return isCommaSemicolonSeperatedNumber(val) || errorMessages.SUPPLEMENTAL_GROUPS_NOT_VALID;
    },
    updateUidGidSource(uidGidSource: UidGidSource): void {
      const isCustom: boolean = uidGidSource === UidGidSource.Custom;
      const isFromToken: boolean = uidGidSource === UidGidSource.FromIdpToken;
      // When changing from custom to other options, we want to reset the values to the default values

      let overrideUidGidInWorkspace: boolean | undefined = undefined;
      if ((isCustom || isFromToken) && this.showOverrideToggle) overrideUidGidInWorkspace = true;

      const runAsUid = isCustom ? this.securityModel.runAsUid : undefined;
      const runAsGid = isCustom ? this.securityModel.runAsGid : undefined;
      const supplementalGroups = isCustom ? this.securityModel.supplementalGroups : undefined;
      const newModel: ISecurityModel = {
        ...this.securityModel,
        uidGidSource,
        overrideUidGidInWorkspace,
        runAsGid,
        runAsUid,
        supplementalGroups,
      };
      this.updateModel(newModel);
    },
    updateUid(runAsUid: number | string | null): void {
      if (runAsUid === "" || runAsUid === null) {
        runAsUid = null;
      } else {
        runAsUid = +runAsUid;
      }
      const newModel: ISecurityModel = { ...this.securityModel, runAsUid };
      this.updateModel(newModel);
    },
    updateGid(runAsGid: number | string | null): void {
      if (runAsGid === "" || runAsGid === null) {
        runAsGid = null;
      } else {
        runAsGid = +runAsGid;
      }
      // When gid is deleted we reset the supplementalGroups
      const supplementalGroups = runAsGid === null ? null : this.securityModel.supplementalGroups;
      const newModel: ISecurityModel = { ...this.securityModel, runAsGid, supplementalGroups };
      this.updateModel(newModel);
    },
    updateOverrideInWorkspace(overrideUidGidInWorkspace: boolean): void {
      const newModel: ISecurityModel = { ...this.securityModel, overrideUidGidInWorkspace };
      this.updateModel(newModel);
    },
    updateSupplementalGroups(supplementalGroups: string | null): void {
      supplementalGroups = supplementalGroups === "" ? null : supplementalGroups;
      const newModel: ISecurityModel = { ...this.securityModel, supplementalGroups };
      this.updateModel(newModel);
    },
    updateModel(newModel: ISecurityModel): void {
      this.$emit("update", newModel);
    },
  },
});
</script>

<style lang="scss">
.form-security-section {
  .q-radio {
    margin-right: 20px;
  }
  .linux-capabilities-section {
    .q-chip__content {
      padding-left: 5px;
      padding-right: 5px;
    }
  }
}
</style>
