<template>
  <hybrid-section-wrapper
    default-opened
    aid="hybrid-compute-resource-section"
    class="hybrid-compute-resource-section"
    label="Compute resources"
    :selected-name="selectedName"
    :section-disabled="sectionDisabled"
    :support-linking="supportLinking"
    :show-content-header="showContentHeader"
    :is-linked="isLinked"
    @load-section-templates="$emit('load-section-templates')"
    @clear="$emit('clear')"
    @reset="$emit('reset')"
    @toggle-linked="$emit('toggle-linked')"
  >
    <template #subheader>
      <span>{{ computeSummary }}</span>
    </template>

    <compute-resource-section-layout
      :model-value="resourcesModel"
      @update:model-value="updateResourcesModel"
      :support-gpu-limit="supportGpuLimit"
      :hide-gpu-resources="hideGpuResources"
      :policy-defaults="policy?.defaults?.compute"
      :policy-rules="policy?.rules?.compute"
      :support-multi-fraction-gpu="supportMultiFractionGpu"
      :section-disabled="sectionDisabled || isLinked"
      :tooltip-text="isLinked ? linkedDisableText : undefined"
      :workers="workloadSpec.numWorkers || undefined"
      :replicas="computedReplicas"
      :section-title="sectionSubTitle"
    >
      <template #top-left>
        <distributed-workers-input
          class="q-mb-md"
          v-if="showWorkersInfo && !showWorkerElasticity"
          :workers="workloadSpec.numWorkers || 1"
          :section-disabled="sectionDisabled"
          :policy-rules="policy?.rules?.numWorkers"
          @update="updateWorkers"
          :title="distributedWorkersInputTitle"
        />
      </template>

      <template #below-gpu-resources-section>
        <template v-if="showWorkerElasticity">
          <distributed-workers-input
            class="q-mb-md"
            :workers="workloadSpec.numWorkers || 1"
            :section-disabled="sectionDisabled"
            :policy-rules="policy?.rules?.numWorkers"
            @update="updateWorkers"
          />
          <distributed-segment-size-input
            v-if="showSegmentSizeInput"
            class="q-mb-md"
            :segment-size="workloadSpec.segmentSize"
            :section-disabled="sectionDisabled || isLinked"
            @update="updateSegmentSize"
          />
          <worker-elasticity-section
            class="q-mb-sm"
            :min-workers="workloadSpec.minReplicas"
            :max-workers="workloadSpec.maxReplicas"
            :num-workers="workloadSpec.numWorkers || undefined"
            :section-disabled="sectionDisabled || isLinked"
            :policy-rules-min="policy?.rules?.minReplicas"
            :policy-rules-max="policy?.rules?.maxReplicas"
            @update:min-workers="updateMinWorkers"
            @update:max-workers="updateMaxWorkers"
          />
        </template>
      </template>
    </compute-resource-section-layout>

    <section class="q-py-md">
      <runai-folding-sub-section label="Extended resources" aid="extended-resources-sub-section">
        <large-shm-toggle
          :model-value="workloadSpec.compute?.largeShmRequest || false"
          @update:model-value="updateLargeShmRequest"
          :policy-rules="policy?.rules?.compute?.largeShmRequest"
          :disable="sectionDisabled || isLinked"
          :tooltip-text="isLinked ? linkedDisableText : undefined"
        />

        <extended-resources
          :key="`extended-resources-${renderKey}`"
          :extended-resources="workloadSpec.compute?.extendedResources || undefined"
          @update:model-value="updateExtendedResources"
          :policy-defaults="policy?.defaults?.compute?.extendedResources"
          :policy-rules="policy?.rules?.compute?.extendedResources"
          :disable="sectionDisabled || isLinked"
          :tooltip-text="isLinked ? linkedDisableText : undefined"
        />
      </runai-folding-sub-section>

      <runai-folding-sub-section
        v-if="workloadSpec.autoscaling"
        label="Replica autoscaling"
        aid="compute-section-auto-scale-btn"
      >
        <auto-scale-section
          v-if="autoscaleData && autoScaleOptions"
          :disabled="sectionDisabled"
          :auto-scale-data="autoscaleData"
          @on-auto-scale-changed="onAutoScaleChanged"
          :policy-rules="policy?.rules?.autoscaling"
          :policy-defaults="policy?.defaults?.autoscaling"
          :auto-scale-options="autoScaleOptions"
        />
      </runai-folding-sub-section>

      <runai-folding-sub-section label="Nodes" aid="nodes-sub-section" v-if="showNodesSubSection">
        <node-pools-resource-section
          v-if="!hideSpecificSections?.nodes"
          :node-pools-list-origin="ENodePoolsListOrigin.Project"
          :apply-policy-defaults="false"
          :selected-node-pools="defaultNodePools"
          :all-node-pools="allNodePoolsNames"
          :policy="nodePoolsPolicy"
          :disabled="sectionDisabled || disableFields?.nodePools"
          :disabled-tooltip="disableFields?.nodePoolsTooltip"
          @list-changed="updateNodePools"
        />

        <node-affinity
          v-if="!hideSpecificSections?.affinity"
          class="q-ml-none"
          :input-type="nodeAffinities.length ? 'select' : 'input'"
          :node-type="workloadSpec.nodeType || null"
          :disabled="sectionDisabled"
          @node-type-changed="updateNodeType"
          :nodes-affinity="nodeAffinities"
          :policy-rules="policy?.rules?.nodeType || undefined"
        />

        <tolerations-section
          v-if="supportTolerations"
          class="q-mt-md"
          :tolerations="workloadSpec.tolerations || undefined"
          @update-tolerations="updateTolerations"
          :read-only="sectionDisabled || !!disableFields?.tolerations"
          :policy-rules="policy?.rules?.tolerations"
          :policy-defaults="policy?.defaults?.tolerations"
          :allow-any-effect="isAnyTolerationAllowed"
          :disabled-info="disableFields?.tolerations"
        />

        <topology-section
          v-if="isTopologyEnabled && showTopology"
          class="q-mt-md"
          :pod-affinity="workloadSpec.podAffinity"
          @update-topology="updatePodAffinity"
          :read-only="sectionDisabled"
          :policy-rules="policy?.rules?.podAffinity"
        />
      </runai-folding-sub-section>
    </section>
  </hybrid-section-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { HybridSectionWrapper } from "../hybrid-section-wrapper";
import { RunaiFoldingSubSection } from "@/components/common/runai-folding-sub-section";
import { LargeShmToggle } from "@/components/compute-resource/large-shm-toggle";
import { ExtendedResources } from "@/components/compute-resource/compute-resource-resources-section/extended-resources";
import { TolerationsSection } from "@/components/section/compute-resource-section/tolerations-section";
import { NodeAffinity } from "@/components/section/compute-resource-section/node-affinity";
import { TopologySection } from "@/components/section/compute-resource-section/topology-section";
import { ComputeResourceSectionLayout } from "@/components/compute-resource/compute-resource-resources-section/compute-resource-section-layout";
import { WorkerElasticitySection } from "@/components/compute-resource/compute-resource-resources-section/worker-elasticity-section";
import DistributedWorkersInput from "@/components/workload-creation/hybrid-compute-resource-section/distributed-workers-input/distributed-workers-input.vue";
import { DistributedSegmentSizeInput } from "@/components/workload-creation/hybrid-compute-resource-section/distributed-segment-size-input";
import { MIN_CLUSTER_VERSION_FOR_ANY_EFFECT, MIN_CLUSTER_VERSION_FOR_TOPOLOGY } from "@/common/version.constant";
import type { IComputeSectionNodePoolsPolicy } from "@/components/section/compute-resource-section/node-pools-resource-section";
import {
  ENodePoolsListOrigin,
  NodePoolsResourceSection,
} from "@/components/section/compute-resource-section/node-pools-resource-section";
import { AutoScaleSection } from "@/components/section/compute-resource-section/auto-scale-section";

// Models
import type { IUISuperWorkloadSpec } from "@/models/workload.model";
import type { IResourcesSectionModel } from "@/models/compute-resource.model";
import type { IHideSpecificSectionsComputeResource } from "@/models/workload-hybrid.model";
import type { PodAffinity, UnifiedPolicyInfoPerReplicaType } from "@/swagger-models/workloads-client";
import type { AutoScaling, ExtendedResource, StringOption, Toleration } from "@/swagger-models/assets-service-client";
import { DistributedFramework } from "@/swagger-models/assets-service-client";
import type { IAutoScaleData } from "@/components/section/compute-resource-section/auto-scale-section";
import type { IAutoScaleOptions } from "@/components/section/compute-resource-section/auto-scale-section/auto-scale-section.model";
import type { IHybridComputeSectionDisableFields } from "./hybrid-compute-resource-section.model";

// Stores
import { useSettingStore } from "@/stores/setting.store";
import { useClusterStore } from "@/stores/cluster.store";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { computeResourceUtil } from "@/utils/compute-resource.util";

// Constants

export default defineComponent({
  name: "hybrid-compute-resource-section",
  components: {
    NodePoolsResourceSection,
    HybridSectionWrapper,
    RunaiFoldingSubSection,
    LargeShmToggle,
    ExtendedResources,
    TolerationsSection,
    NodeAffinity,
    TopologySection,
    DistributedWorkersInput,
    DistributedSegmentSizeInput,
    WorkerElasticitySection,
    AutoScaleSection,
    ComputeResourceSectionLayout,
  },
  emits: ["load-section-templates", "reset", "clear", "update", "toggle-linked"],
  props: {
    selectedName: {
      type: String as PropType<string>,
      required: false,
    },
    workloadSpec: {
      type: Object as PropType<IUISuperWorkloadSpec>,
      required: true,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    policy: {
      type: [Object, null] as PropType<UnifiedPolicyInfoPerReplicaType | null>,
      required: false,
    },
    nodePoolsResources: {
      type: Array as PropType<StringOption[]>,
      required: false,
      default: () => [],
    },
    supportMultiFractionGpu: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    supportTolerations: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: false,
    },
    autoScaleOptions: {
      type: Object as PropType<IAutoScaleOptions>,
      required: false,
    },
    showContentHeader: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    disableFields: {
      type: Object as PropType<IHybridComputeSectionDisableFields>,
      required: false,
    },
    hideSpecificSections: {
      type: Object as PropType<IHideSpecificSectionsComputeResource>,
      required: false,
    },
    sectionSubTitle: {
      type: String as PropType<string>,
      required: false,
    },
    supportLinking: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    isLinked: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    renderKey: {
      type: String as PropType<string>,
      required: false,
      default: "",
    },
    distributedWorkersInputTitle: {
      type: String as PropType<string>,
      required: false,
    },
    showTopology: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    hideGpuResources: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  data() {
    return {
      settingStore: useSettingStore(),
      clusterStore: useClusterStore(),
      ENodePoolsListOrigin,
      linkedDisableText: "This can't be modified when the Link button is active",
    };
  },
  computed: {
    supportGpuLimit(): boolean {
      return this.settingStore.isGPUOverProvisioningEnabled;
    },
    computeSummary(): string {
      return computeResourceUtil.computeResourceUISummary(this.resourcesModel);
    },
    nodeAffinities(): string[] {
      return this.policy?.rules?.nodeType?.options?.map((nt) => nt.value) || [];
    },
    nodePoolsPolicy(): IComputeSectionNodePoolsPolicy | null {
      return {
        rules: this.policy?.rules?.nodePools || undefined,
        defaults: this.policy?.defaults?.nodePools || undefined,
      };
    },
    isAnyTolerationAllowed(): boolean {
      if (!this.clusterId) return false;
      return this.clusterStore.isClusterVersionSufficient(this.clusterId, MIN_CLUSTER_VERSION_FOR_ANY_EFFECT);
    },
    resourcesModel(): IResourcesSectionModel {
      return {
        gpuDevicesRequest: this.workloadSpec.compute?.gpuDevicesRequest || 0,
        gpuRequestType: this.workloadSpec.compute?.gpuRequestType,
        gpuPortionRequest: this.workloadSpec.compute?.gpuPortionRequest,
        gpuPortionLimit: this.workloadSpec.compute?.gpuPortionLimit,
        gpuMemoryRequest: this.workloadSpec.compute?.gpuMemoryRequest,
        gpuMemoryLimit: this.workloadSpec.compute?.gpuMemoryLimit,
        cpuCore: {
          request: this.workloadSpec.compute?.cpuCoreRequest || 0,
          limit: this.workloadSpec.compute?.cpuCoreLimit || null,
        },
        cpuMemory: {
          request: this.workloadSpec.compute?.cpuMemoryRequest || "0M",
          limit: this.workloadSpec.compute?.cpuMemoryLimit || null,
        },
        minWorkers: this.workloadSpec.minReplicas,
        maxWorkers: this.workloadSpec.maxReplicas,
      };
    },
    defaultNodePools(): string[] {
      return this.workloadSpec?.nodePools || [];
    },
    allNodePoolsNames(): string[] {
      return this.nodePoolsResources.map((nodePoolResource: StringOption) => nodePoolResource.value || "");
    },
    showWorkersInfo(): boolean {
      return this.workloadSpec.numWorkers !== undefined;
    },
    showSegmentSizeInput(): boolean {
      return this.showWorkersInfo && this.workloadSpec.distributedFramework === DistributedFramework.PyTorch;
    },
    replicasInfo(): boolean {
      return !!this.workloadSpec.autoscaling;
    },
    isTopologyEnabled(): boolean {
      if (!this.clusterId) return false;
      return this.clusterStore.isClusterVersionSufficient(this.clusterId, MIN_CLUSTER_VERSION_FOR_TOPOLOGY);
    },
    autoscaleData(): IAutoScaleData | undefined {
      if (!this.workloadSpec.autoscaling) return undefined;
      return inferenceWorkloadUtil.convertAutoScaleFlatToUI(this.workloadSpec.autoscaling);
    },
    showNodesSubSection(): boolean {
      if (!this.hideSpecificSections) return true;
      return this.supportTolerations || (this.isTopologyEnabled && this.showWorkersInfo);
    },
    showWorkerElasticity(): boolean {
      return (
        this.showWorkersInfo &&
        this.workloadSpec.numWorkers !== null &&
        this.workloadSpec.distributedFramework === DistributedFramework.PyTorch
      );
    },
    computedReplicas(): { min: number | null | undefined; max: number | null | undefined } | undefined {
      // If worker elasticity is enabled
      if (this.showWorkerElasticity) {
        const hasMin = this.workloadSpec.minReplicas !== null && this.workloadSpec.minReplicas !== undefined;
        const hasMax = this.workloadSpec.maxReplicas !== null && this.workloadSpec.maxReplicas !== undefined;

        // Return replicas if at least one of min or max is set
        if (hasMin || hasMax) {
          return {
            min: this.workloadSpec.minReplicas,
            max: this.workloadSpec.maxReplicas,
          };
        }
      }
      // Otherwise, use autoscaling replicas for inference
      if (this.workloadSpec.autoscaling) {
        return {
          min: this.workloadSpec.autoscaling.minReplicas,
          max: this.workloadSpec.autoscaling.maxReplicas,
        };
      }
      return undefined;
    },
  },
  methods: {
    updateTolerations(tolerations: Toleration[]) {
      this.$emit("update", { ...this.workloadSpec, tolerations });
    },
    updatePodAffinity(podAffinity: PodAffinity) {
      this.$emit("update", { ...this.workloadSpec, podAffinity });
    },
    updateLargeShmRequest(largeShmRequest: boolean) {
      this.$emit("update", { ...this.workloadSpec, compute: { ...this.workloadSpec.compute, largeShmRequest } });
    },
    updateWorkers(numWorkers: number) {
      if (typeof numWorkers === "string") {
        numWorkers = parseInt(numWorkers);
      }
      this.$emit("update", { ...this.workloadSpec, numWorkers: numWorkers || 1 });
    },
    updateSegmentSize(segmentSize: number | null) {
      this.$emit("update", { ...this.workloadSpec, segmentSize });
    },
    updateMinWorkers(minWorkers: number | null) {
      this.$emit("update", { ...this.workloadSpec, minReplicas: minWorkers });
    },
    updateMaxWorkers(maxWorkers: number | null) {
      this.$emit("update", { ...this.workloadSpec, maxReplicas: maxWorkers });
    },
    updateResourcesModel(resourcesModel: IResourcesSectionModel) {
      const computeFields = {
        gpuDevicesRequest: resourcesModel.gpuDevicesRequest,
        gpuRequestType: resourcesModel.gpuRequestType,
        gpuPortionRequest: resourcesModel.gpuPortionRequest,
        gpuPortionLimit: resourcesModel.gpuPortionLimit,
        gpuMemoryRequest: resourcesModel.gpuMemoryRequest,
        gpuMemoryLimit: resourcesModel.gpuMemoryLimit,
        cpuCoreRequest: resourcesModel.cpuCore.request,
        cpuCoreLimit: resourcesModel.cpuCore.limit,
        cpuMemoryRequest: resourcesModel.cpuMemory.request,
        cpuMemoryLimit: resourcesModel.cpuMemory.limit,
      };

      this.$emit("update", {
        ...this.workloadSpec,
        compute: {
          ...this.workloadSpec.compute,
          ...computeFields,
        },
      });
    },
    updateExtendedResources(extendedResources: ExtendedResource[]) {
      this.$emit("update", { ...this.workloadSpec, compute: { ...this.workloadSpec.compute, extendedResources } });
    },
    updateNodeType(nodeType: string) {
      this.$emit("update", { ...this.workloadSpec, nodeType });
    },
    updateNodePools(nodePools: Array<string>): void {
      this.$emit("update", { ...this.workloadSpec, nodePools });
    },
    onAutoScaleChanged(autoScaleData: IAutoScaleData): void {
      const autoscaling: AutoScaling = inferenceWorkloadUtil.convertAutoScaleUIToFlat(autoScaleData);
      this.$emit("update", { ...this.workloadSpec, autoscaling });
    },
  },
});
</script>
