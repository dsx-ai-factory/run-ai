<template>
  <wrapper-card
    @close="$emit('close')"
    header="Node type (Affinity)"
    subheader="Limit workloads to run on specific node types"
  >
    <template #body>
      <gray-card
        v-for="(workload, index) in workloads"
        :hide-close-button="isCloseButtonHidden(index)"
        @close="hideWorkloadSection(index)"
        :key="workload.value"
      >
        <div class="row items-center">
          <div class="col-4 q-mr-md">
            <runai-select
              label="Workload type"
              @update:model-value="onWorkloadSelect(workload, $event, index)"
              :model-value="getOptionValue(workload)"
              :options="workloadOptions"
              :disable="readOnly"
              :rules="[workloadTypeValidation]"
            />
          </div>
          <div class="col">
            <runai-addable-multi-select
              label="Node type (Affinity)"
              new-option-label="type"
              @options-select="updateSelectedNodeTypes(workload, $event, index)"
              @add-option="(name, doneCallback) => addNodeType(name, doneCallback, index)"
              :options="getNodeTypesOptionsForWorkload(workload)"
              :selected-options="workload.selectedTypes"
              :loading="workload.loading"
              :disable="readOnly"
              :support-adding-new-value="!hasParentNodeTypes"
              required
            />
          </div>
        </div>
      </gray-card>
    </template>
    <template #actions>
      <q-btn
        v-if="showNewWorkloadButton"
        @click="addWorkloadSection"
        :disable="readOnly"
        color="primary"
        flat
        class="btn-icon-sm"
        icon="ph-bold ph-plus"
        label="NODE TYPE"
      />
    </template>
  </wrapper-card>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
// cmps
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiAddableMultiSelect } from "@/components/common/runai-addable-multi-select/";
import { WrapperCard } from "@/components/project/project-edit-form/cards/";
import { GrayCard } from "@/components/project/project-edit-form/cards/";
//models
import type { ISelectOption } from "@/models/global.model";
import type { INodeAffinitySelectOption } from "@/models/project.model";
import { EWorkloadNodeAffinity } from "@/models/project.model";
//service
import { alertUtil } from "@/utils/alert.util";
//store
import { errorMessages } from "@/common/error-message.constant";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type {
  GetNodeTypes200ResponseNodeTypesInner,
  NodeTypesPerWorkload,
} from "@/swagger-models/org-unit-service-client";

const MAX_NUMBER_OF_NODE_AFFINITY_WORKLOADS = 2;
export default defineComponent({
  components: { GrayCard, WrapperCard, RunaiAddableMultiSelect, RunaiSelect },
  emits: [
    "close",
    "add-workload",
    "remove-workload",
    "workload-changed",
    "add-new-node-affinity-type",
    "update-workload-loading",
  ],
  props: {
    workloads: {
      type: Array as PropType<INodeAffinitySelectOption[]>,
      required: true,
    },
    clusterId: {
      type: String,
      required: true,
    },
    readOnly: {
      type: Boolean,
      default: false,
    },
    parentNodeTypes: {
      type: Object as PropType<NodeTypesPerWorkload>,
      required: false,
    },
  },
  data() {
    return {
      nodeTypes: [] as GetNodeTypes200ResponseNodeTypesInner[],
    };
  },
  created() {
    this.getNodeTypes();
  },
  computed: {
    workloadOptions(): ISelectOption[] {
      const isOptionDisabled = (value: EWorkloadNodeAffinity): boolean =>
        this.workloads.some((workload: INodeAffinitySelectOption) => workload.value === value);
      return [
        {
          value: EWorkloadNodeAffinity.Interactive,
          label: "Workspace",
          disable: isOptionDisabled(EWorkloadNodeAffinity.Interactive),
        },
        {
          value: EWorkloadNodeAffinity.Train,
          label: "Training",
          disable: isOptionDisabled(EWorkloadNodeAffinity.Train),
        },
      ] as ISelectOption[];
    },
    nodeTypesOptions(): ISelectOption[] {
      return this.nodeTypes.map(({ name, id }) => ({ label: name as string, value: id as string }));
    },
    hasParentNodeTypes(): boolean {
      return !!(this.parentNodeTypes?.training?.length || this.parentNodeTypes?.workspace?.length);
    },
    showNewWorkloadButton(): boolean {
      return this.workloads.length < MAX_NUMBER_OF_NODE_AFFINITY_WORKLOADS;
    },
  },
  methods: {
    getParentAllowedIds(workload: INodeAffinitySelectOption): string[] | undefined {
      if (!this.hasParentNodeTypes || !workload.value) return undefined;
      if (workload.value === EWorkloadNodeAffinity.Train) return this.parentNodeTypes?.training;
      if (workload.value === EWorkloadNodeAffinity.Interactive) return this.parentNodeTypes?.workspace;
      return undefined;
    },
    getNodeTypesOptionsForWorkload(workload: INodeAffinitySelectOption): ISelectOption[] {
      const allowedIds = this.getParentAllowedIds(workload);
      if (!allowedIds) return this.nodeTypesOptions;
      return this.nodeTypesOptions.filter((option) => allowedIds.includes(option.value as string));
    },
    isCloseButtonHidden(index: number): boolean {
      return index === 0 && this.workloads.length === 1;
    },
    workloadTypeValidation(option: ISelectOption | null): boolean | string {
      return !!option || errorMessages.SELECT_TYPE;
    },
    getOptionValue(workload: ISelectOption): null | ISelectOption {
      if (workload.value) {
        return workload;
      }
      return null;
    },
    async getNodeTypes(): Promise<void> {
      try {
        this.nodeTypes = await orgUnitService.getNodeTypes(this.clusterId);
      } catch (error: unknown) {
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to get node types"));
      }
    },
    async addNodeType(name: string, doneCallback: () => void, workloadIndex: number): Promise<void> {
      try {
        this.$emit("update-workload-loading", workloadIndex, true);
        const createdNodeType = await orgUnitService.createNodeType(name, this.clusterId);

        this.$emit("add-new-node-affinity-type", workloadIndex, { label: name, value: createdNodeType.id });
        this.nodeTypes.push(createdNodeType);
      } catch (error: unknown) {
        this.$q.notify(alertUtil.getError(`Failed to create ${name} node type`));
      } finally {
        this.$emit("update-workload-loading", workloadIndex, false);
        doneCallback();
      }
    },
    addWorkloadSection(): void {
      this.$emit("add-workload");
    },
    hideWorkloadSection(workloadIndex: number): void {
      this.$emit("remove-workload", workloadIndex);
    },
    updateSelectedNodeTypes(
      workload: INodeAffinitySelectOption,
      selectedTypes: ISelectOption[],
      workloadIndex: number,
    ): void {
      this.$emit("workload-changed", { ...workload, selectedTypes }, workloadIndex);
    },
    onWorkloadSelect(workload: INodeAffinitySelectOption, selectedWorkload: ISelectOption, workloadIndex: number): void {
      this.$emit("workload-changed", { ...selectedWorkload, selectedTypes: workload.selectedTypes }, workloadIndex);
    },
  },
});
</script>
