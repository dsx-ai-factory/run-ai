<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="compute-resource-section"
    label="Compute resource"
    default-opened
    :section-invalid="sectionInvalid"
  >
    <template #subheader>
      <span>{{ computeResourceSummary }}</span>
    </template>
    <div v-if="computeResourceData.workers" class="workers-section">
      <div class="column q-pa-sm q-mb-sm">
        <div>Set the number of workers for your workload</div>
        <div class="row">
          <policy-number-input
            class="col-2 q-mb-xl"
            type="number"
            :model-value="computeResourceData.workers"
            @update:model-value="updateWorkers"
            stack-label
            :min-value="1"
            :disable="sectionDisabled"
            custom-ref="workersInput"
          />
        </div>
        <div class="dashed-seperator"></div>
      </div>
    </div>
    <runai-form-card-section
      :loading="loading"
      empty-message="Looks like you don't have any compute resources yet..."
      :main-message="cardsMainTitle"
      entity-name="compute resource"
      :cards-list="computeResourceCards"
      :selected-cards-ids="selectedCardsIds"
      icon-name="ph-bold ph-hard-drives"
      search-name="compute resources"
      @selected-card-changed="onSelectedComputeResource"
      @create-new="createNew"
      :disabled="sectionDisabled"
      :disable-create-new="disableCreateButton"
      :allow-actions="!sectionOptions.hideCardsActions"
      :sort-options="{ name: true, creation: true, recentUsage: true }"
      default-sort-option="recentUsage"
    />

    <div class="col-12 q-ml-sm" v-if="isRequired">
      <q-field
        class="hidden-field"
        :model-value="computeResourceData.computeResourceId"
        :rules="[isRequiredComputeResource]"
      ></q-field>
    </div>

    <compute-resource-describer
      v-if="computeResourceData.computeResourceId && !!computeResourceData.workers"
      label="workers"
      :compute-resource-name="selectedComputeResourceName"
      :amount="computeResourceData.workers"
    />

    <compute-resource-describer
      v-if="shouldShowReplicasDescriber && computeResourceData.autoScaleData"
      label="replica"
      :compute-resource-name="selectedComputeResourceName"
      :amount="computeResourceData.autoScaleData.minReplicas || 0"
      :max-amount="computeResourceData.autoScaleData.maxReplicas"
      max-label="replicas"
    />

    <div v-if="showNodeAffinitySelectSection">
      <node-affinity
        input-type="select"
        :node-type="selectedNodeType"
        :nodes-affinity="nodeAffinity"
        :disabled="sectionDisabled"
        @node-type-changed="updateNodeAffinity"
        :policy-rules="policyRules?.rules?.specificRunParams?.nodeType || undefined"
      />
    </div>

    <runai-sub-expansion-item
      v-if="sectionOptions.autoScale && computeResourceData.autoScaleData && !loading"
      label="Replica autoscaling"
      aid="compute-section-auto-scale-btn"
    >
      <auto-scale-section
        :disabled="sectionDisabled"
        :auto-scale-data="computeResourceData.autoScaleData"
        @on-auto-scale-changed="onAutoScaleChanged"
        :policy-rules="policyRules?.rules?.specificRunParams?.autoscaling"
        :auto-scale-options="autoScaleOptions"
        :policy-defaults="policyDefaults?.autoscaling"
      />
    </runai-sub-expansion-item>

    <runai-sub-expansion-item v-if="showMoreSettings && !loading" label="Nodes" aid="compute-section-more-settings-btn">
      <node-pools-resource-section
        v-if="nodePoolsData && !sectionOptions.hideNodePools"
        :selected-node-pools="nodePoolsData.defaultNodePools"
        :all-node-pools="nodePoolsData.allNodePoolsOptions"
        :node-pools-list-origin="nodePoolsData.nodePoolsListOrigin"
        @list-changed="onNodePoolsOrderChanged"
        :policy="nodePoolsPolicy"
        :apply-policy-defaults="applyPolicyDefaults"
        :disabled="sectionDisabled"
      />

      <div v-if="showNodeAffinityInputSection">
        <node-affinity
          class="q-mb-md"
          input-type="input"
          :node-type="selectedNodeType"
          :disabled="sectionDisabled"
          @node-type-changed="updateNodeAffinity"
          :policy-rules="policyRules?.rules?.specificRunParams?.nodeType || undefined"
        />
      </div>

      <tolerations-section
        v-if="supportTolerations"
        class="q-mt-md"
        :tolerations="computeResourceData.tolerations || undefined"
        @update-tolerations="updateTolerations"
        :read-only="sectionDisabled"
        :policy-rules="policyRules?.rules?.specificRunParams?.tolerations"
        :policy-defaults="itemizeDefaults?.tolerations"
        :allow-any-effect="sectionOptions.allowAnyEffect"
      />

      <topology-section
        v-if="shouldShowTopologySection"
        class="q-mt-md"
        :pod-affinity="computeResourceData.podAffinity"
        @update-topology="updateTopology"
        :read-only="sectionDisabled"
        :policy-rules="policyRules?.rules?.specificRunParams?.podAffinity"
      />
    </runai-sub-expansion-item>
  </runai-expansion-item>
</template>

<script lang="ts">
import { computed, defineComponent, type PropType } from "vue";

// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSubExpansionItem } from "@/components/common/runai-sub-expansion-item";
import { RunaiFormCardSection } from "@/components/common";
import { NodePoolsResourceSection } from "./node-pools-resource-section";
import { ComputeResourceDescriber } from "./compute-resource-describer";
import { NodeAffinity } from "./node-affinity";
import { AutoScaleSection } from "./auto-scale-section";
import { TolerationsSection } from "./tolerations-section";
import { TopologySection } from "./topology-section";
import { PolicyNumberInput } from "@/components/common/policy-number-input";

// Models
import { EWorkloadFormType } from "@/models/workload.model";
import type { ICardListItem } from "@/components/common/runai-card-list";
import { TCardCmpName } from "@/components/common/runai-card-list";
import type {
  ComputeAsset,
  ComplianceInfoReason,
  SpecificRunAutoScalingAutoScaling,
  PodAffinity,
  PolicyInfo,
} from "@/swagger-models/assets-service-client";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import { errorMessages } from "@/common/error-message.constant";
import type {
  IComputeSectionData,
  IComputeSectionPolicy,
  IWorkloadComputeSectionOptions,
  IComputeSectionNodePoolData,
} from "./compute-resource-section.models";
import type { IComputeSectionNodePoolsPolicy } from "./node-pools-resource-section";
import type { Toleration } from "@/swagger-models/assets-service-client";
import type { IAutoScaleOptions } from "./auto-scale-section/auto-scale-section.model";

// stores
import { usePermissionStore } from "@/stores/permissions.store";
import { useSettingStore } from "@/stores/setting.store";

// utils
import { fallbackDefaultIfNullOrUndefined } from "@/utils/common.util";
import { convertToBytes, parseSizeString } from "@/utils/format.util";
import type { ILoadedItemizeDefaults } from "@/models/policy.model";
import type { InferencePolicyDefaultsAndRulesV2Defaults } from "@/swagger-models/policy-service-client";

export interface ICardVisibilityState {
  disabled: boolean;
  tooltip: string;
  showDisabledInfo: boolean;
}

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiFormCardSection,
    RunaiSubExpansionItem,
    NodePoolsResourceSection,
    ComputeResourceDescriber,
    NodeAffinity,
    AutoScaleSection,
    TolerationsSection,
    TopologySection,
    PolicyNumberInput,
  },
  emits: ["compute-resource-data-changed", "create-new"],
  props: {
    entityType: {
      type: String as PropType<EWorkloadFormType>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isRequired: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    computeResourceData: {
      type: Object as PropType<IComputeSectionData>,
      required: true,
    },
    computeResources: {
      type: Array as PropType<Array<ComputeAsset>>,
      required: true,
    },
    nodeAffinity: {
      type: Array as PropType<string[]>,
      required: false,
    },
    policy: {
      type: [Object, null] as PropType<IComputeSectionPolicy | null>,
      required: false,
    },
    sectionOptions: {
      type: Object as PropType<IWorkloadComputeSectionOptions>,
      required: false,
      default: () => ({}),
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<PolicyInfo | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<InferencePolicyDefaultsAndRulesV2Defaults | null>,
      required: false,
    },
    itemizeDefaults: {
      type: [Object, null] as PropType<ILoadedItemizeDefaults | null>,
      required: false,
    },
    supportTolerations: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  provide() {
    return {
      isGPUOverProvisioningEnabled: computed(() => this.settingStore.isGPUOverProvisioningEnabled),
    };
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
      settingStore: useSettingStore(),
    };
  },
  mounted() {
    // Quasar input is incrementing its value on scroll if the window reaches the bottom of the page, therefore we blur the element whenever the scroll event is fired in order to prevent that
    window.addEventListener("scroll", this.blurWorkersInput);
  },
  computed: {
    shouldShowTopologySection(): boolean {
      return !!this.sectionOptions.showTopologySection;
    },
    shouldShowReplicasDescriber(): boolean {
      return (
        !!this.computeResourceData.computeResourceId &&
        this.entityType === EWorkloadFormType.Inference &&
        !!this.sectionOptions.autoScale
      );
    },
    disableCreateButton(): boolean {
      return !this.permissionStore.hasPermission(ResourceType.ComputeResources, Action.Create);
    },
    entityText(): string {
      if (!this.entityType || this.entityType === EWorkloadFormType.Template) return "workload";
      return this.entityType;
    },
    computeResourceSummary(): string {
      const cardData: ComputeAsset = (this.selectedComputeResource?.data as ComputeAsset) || undefined;
      if (!cardData || !cardData.meta) return "None";
      return this.computeResourceData.workers
        ? `${cardData.meta.name} X ${this.computeResourceData.workers}`
        : cardData.meta.name;
    },
    // minGpuMemoryRequirementsString(): string {
    //   if (!this.minGpuMemoryRequirementsMb) return "";
    //   const minGpuMemoryRequirementsBytes = convertToBytes(this.minGpuMemoryRequirementsMb, "MB");
    //   return memoryFormat(minGpuMemoryRequirementsBytes);
    // },

    computeResourceCards(): Array<ICardListItem> {
      // let minGpuMemoryRequirementsBytes: number | undefined = undefined;
      // if (this.minGpuMemoryRequirementsMb) {
      //   minGpuMemoryRequirementsBytes = convertToBytes(this.minGpuMemoryRequirementsMb, "MB");
      // }

      return this.computeResources.map((computeResource: ComputeAsset) => {
        const cardVisibilityState: ICardVisibilityState = this.getCardVisibilityState(computeResource);
        const name = computeResource.meta.name;
        const cpu = (computeResource.spec.cpuCoreRequest || "").toString();
        const memory = computeResource.spec.cpuMemoryRequest || "";
        const gpu = (computeResource.spec.gpuPortionRequest || computeResource.spec.gpuMemoryRequest || "").toString();

        return {
          id: computeResource.meta.id,
          cardName: TCardCmpName.COMPUTE_RESOURCE,
          data: computeResource,
          searchValues: [name, gpu, memory, cpu].filter((val) => !!val),
          ...cardVisibilityState,
          sortInfo: {
            name: computeResource.meta.name,
            createdAt: computeResource.meta.createdAt,
            recentUsage: computeResource.usageTimes?.lastUsedByWorkload,
          },
        };
      });
    },
    selectedComputeResource(): ICardListItem | null {
      if (!this.computeResourceId) return null;
      return this.computeResourceCards.find((item: ICardListItem) => item.id === this.computeResourceId) || null;
    },
    sectionInvalid(): boolean {
      return (
        (this.isRequired && !this.computeResourceId) ||
        (!!this.nodeAffinity && !!this.nodeAffinity.length && !this.selectedNodeType)
      );
    },
    selectedCardsIds(): Array<string> {
      return this.computeResourceId ? [this.computeResourceId] : [];
    },
    nodePoolsPolicy(): IComputeSectionNodePoolsPolicy | null {
      return this.policy?.nodePools || null;
    },
    selectedComputeResourceName(): string {
      return (
        this.computeResources.find((compute: ComputeAsset) => compute.meta.id === this.computeResourceId)?.meta.name ||
        ""
      );
    },
    cardsMainTitle(): string {
      return this.sectionOptions.cardsTitle || `Select the node resources needed to run your workload`;
    },
    applyPolicyDefaults(): boolean {
      return fallbackDefaultIfNullOrUndefined(this.sectionOptions.applyPolicyDefaults, false);
    },
    showMoreSettings(): boolean {
      return (
        !!(this.nodePoolsData && !this.sectionOptions.hideNodePools) ||
        this.showNodeAffinityInputSection ||
        this.supportTolerations
      );
    },
    hideNodeAffinitySection(): boolean {
      return this.nodeAffinity === undefined || this.loading;
    },
    showNodeAffinityInputSection(): boolean {
      if (this.hideNodeAffinitySection) return false;
      return this.nodeAffinity !== undefined && !this.nodeAffinity.length;
    },
    showNodeAffinitySelectSection(): boolean {
      if (this.hideNodeAffinitySection) return false;
      return this.nodeAffinity !== undefined && !!this.nodeAffinity.length;
    },
    nodePoolsData(): IComputeSectionNodePoolData | null {
      return this.computeResourceData.nodePools || null;
    },
    computeResourceId(): string | null {
      return this.computeResourceData.computeResourceId;
    },
    selectedNodeType(): string | null {
      return this.computeResourceData.nodeType;
    },
    autoScaleOptions(): IAutoScaleOptions {
      return {
        latencyEnabled: this.sectionOptions.latencyEnabled || false,
        isAutoScaleEnabled: this.sectionOptions.autoScaleEnabled || false,
      };
    },
  },
  methods: {
    getCardVisibilityState(computeResource: ComputeAsset): ICardVisibilityState {
      const cardVisibilityState: ICardVisibilityState = {
        disabled: false,
        tooltip: "",
        showDisabledInfo: false,
      };

      if (computeResource.compliance && !computeResource.compliance?.compliance) {
        return this.getCardVisibilityStateByCompliance(computeResource);
      }

      return cardVisibilityState;
    },
    getCardVisibilityStateByCompliance(computeResource: ComputeAsset): ICardVisibilityState {
      const cardVisibilityState: ICardVisibilityState = {
        disabled: true,
        tooltip: "",
        showDisabledInfo: false,
      };

      // if it does not comply due to policy vs if it does not comply for some other reason
      if (computeResource.compliance?.reason?.some((reason: ComplianceInfoReason) => reason.field)) {
        cardVisibilityState.showDisabledInfo = true;
        cardVisibilityState.tooltip =
          "This compute resource can't be used because it doesn't comply with the policy your administrator set.";
      } else if (computeResource.compliance?.reason?.length) {
        cardVisibilityState.tooltip = computeResource.compliance.reason[0].details;
      }

      return cardVisibilityState;
    },
    parseStringSizeToBytes(sizeString: string): number | null {
      const sizeAndString: [number, string] | null = parseSizeString(sizeString);
      if (!sizeAndString) return null;
      return convertToBytes(sizeAndString[0], sizeAndString[1]);
    },
    updateTolerations(tolerations: Array<Toleration>): void {
      this.$emit("compute-resource-data-changed", {
        ...this.computeResourceData,
        tolerations,
      });
    },
    updateTopology(podAffinity: PodAffinity | null): void {
      this.$emit("compute-resource-data-changed", {
        ...this.computeResourceData,
        podAffinity,
      });
    },
    blurWorkersInput(): void {
      const workersInput = this.$refs.workersInput as HTMLElement | null;
      if (workersInput && workersInput.blur) workersInput.blur();
    },
    isRequiredComputeResource(val: string | null): string | boolean {
      return val ? true : errorMessages.SELECT_COMPUTE;
    },
    onSelectedComputeResource(selectedItems: Array<string>): void {
      const selectedItem: string | null = selectedItems.length === 0 ? null : selectedItems[0];
      this.$emit("compute-resource-data-changed", {
        ...this.computeResourceData,
        computeResourceId: selectedItem,
      });
    },
    createNew(): void {
      this.$emit("create-new");
    },
    updateNodeAffinity(nodeAffinity: string | number | null): void {
      this.$emit("compute-resource-data-changed", {
        ...this.computeResourceData,
        nodeType: nodeAffinity,
      });
    },
    isSelected(selectedOption: string | null): boolean | string {
      return !!selectedOption || errorMessages.SELECT_NODE_TYPE;
    },
    onNodePoolsOrderChanged(nodePools: Array<string>): void {
      this.$emit("compute-resource-data-changed", {
        ...this.computeResourceData,
        nodePools: {
          ...this.computeResourceData.nodePools,
          defaultNodePools: nodePools,
        },
      });
    },
    updateWorkers(workers: number | string | null): void {
      let newValue: number;
      if (!workers) {
        newValue = 1;
        return;
      }
      if (typeof workers === "string") {
        workers = parseInt(workers);
      }
      newValue = workers < 1 ? 1 : workers;

      this.$emit("compute-resource-data-changed", {
        ...this.computeResourceData,
        workers: newValue,
      });
    },
    onAutoScaleChanged(autoScaleData: SpecificRunAutoScalingAutoScaling): void {
      this.$emit("compute-resource-data-changed", {
        ...this.computeResourceData,
        autoScaleData,
      });
    },
  },
  unmounted() {
    // Quasar input is incrementing its value on scroll if the window reaches the bottom of the page, therefore we blur the element whenever the scroll event is fired in order to prevent that
    window.removeEventListener("scroll", this.blurWorkersInput);
  },
});
</script>
<style lang="scss" scoped>
.compute-resource-section {
  .workers-display-section {
    color: $black-54;
    border: 1px $black-12 solid;
  }
  .dashed-seperator {
    border: 1px dashed $black-12;
  }
}
</style>
