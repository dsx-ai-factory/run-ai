<template>
  <runai-expansion-item
    class="form-workload-types-section"
    aid="workload-types-section"
    label="Workload architecture & type"
    :section-invalid="sectionInvalid"
    :default-opened="!isFolded"
    :subheader="summary"
  >
    <distributed-selection
      v-if="showDistributedSection"
      :read-only="readOnly"
      :disable-selection="lockedWorkloadTypes"
      :distributed-model="distributedModel"
      @update-distributed-model="updateDistributed"
      :form-type="EDistributedFormType.ENVIRONMENT"
      :cluster-id="clusterId"
    />
    <workload-types-selection
      :workload-types="workloadTypesModel"
      @update-workload-types="updateWorkloadTypes"
      :options="workloadTypeOptions"
    />
    <distributed-framework-selection-section
      v-if="workloadTypesSectionModel.distributed && workloadTypesSectionModel.training"
      class="q-mt-lg"
      :dist-framework="workloadTypesSectionModel.distFramework ?? null"
      :cluster-id="clusterId"
      :read-only="readOnly"
      :disable-selection="lockedWorkloadTypes"
      @update="updateDistFramework"
    />
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { DistributedSelection } from "../../section/multi-node-section/distributed-selection";
import { DistributedFrameworkSelectionSection } from "../../section/multi-node-section/distributed-selection/distributed-framework-selection-section";
import { WorkloadTypesSelection } from "./workload-types-selection";

// Model
import { EWorkloadName } from "@/models/workload.model";
import type { IWorkloadTypesOption } from "./workload-types-selection/workload-types-section.model";
import type {
  WorkloadSupportedTypes,
  WorkloadSupportedTypesDistFrameworkEnum,
} from "@/swagger-models/assets-service-client";
import type { IDistributedSelectionModel } from "@/components/section/multi-node-section/distributed-selection";

// stores
import { useAppStore } from "@/stores/app.store";

// Utils
import { deepCopy } from "@/utils/common.util";

import { EDistributedFormType } from "@/components/section/multi-node-section/distributed-selection/distributed-selection.vue";

export default defineComponent({
  name: "multi-node-section",
  components: {
    RunaiExpansionItem,
    DistributedSelection,
    DistributedFrameworkSelectionSection,
    WorkloadTypesSelection,
  },
  emits: ["update-workload-types-section-model"],
  props: {
    workloadTypesSectionModel: {
      type: Object as PropType<WorkloadSupportedTypes>,
      required: true,
    },
    isFolded: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    lockedWorkloadTypes: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    showDistributedSection: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  data() {
    return {
      appStore: useAppStore(),
      workloadTypesOptionsMap: {} as Record<EWorkloadName, IWorkloadTypesOption>,
      EDistributedFormType: EDistributedFormType,
    };
  },
  created() {
    this.workloadTypesOptionsMap.workspace = { label: "Workspace", value: EWorkloadName.Workspace };
    if (!this.appStore.isWhiteLabel) {
      this.workloadTypesOptionsMap.training = { label: "Training", value: EWorkloadName.Training };
    }
    this.workloadTypesOptionsMap.inference = { label: "Inference", value: EWorkloadName.Inference };

    if (this.lockedWorkloadTypes) {
      const { distributed, inference, training, workspace } = this.workloadTypesSectionModel;
      const isDistributedInference = !!(distributed && inference && !training);
      const isDistributedTraining = !!(distributed && training && !inference);
      if (isDistributedInference || isDistributedTraining) {
        this.workloadTypesOptionsMap.workspace.disable = true;
        this.workloadTypesOptionsMap.workspace.toolTip =
          "This can't be modified when an environment is created via a distributed workload.";
        this.workloadTypesOptionsMap.training.disable = true;
        this.workloadTypesOptionsMap.training.toolTip =
          "This can't be modified when an environment is created via a distributed workload.";
        this.workloadTypesOptionsMap.inference.disable = true;
        this.workloadTypesOptionsMap.inference.toolTip =
          "This can't be modified when an environment is created via a distributed workload.";
      } else {
        if (workspace) {
          this.workloadTypesOptionsMap.workspace.disable = true;
          this.workloadTypesOptionsMap.workspace.toolTip =
            "This option must be checked when an environment is created via creation of workspace";
        }
        if (training) {
          this.workloadTypesOptionsMap.training.disable = true;
          this.workloadTypesOptionsMap.training.toolTip =
            "This option must be checked when an environment is created via creation of training";
        }
        if (inference) {
          this.workloadTypesOptionsMap.inference.disable = true;
          this.workloadTypesOptionsMap.inference.toolTip =
            "This option must be checked when an environment is created via creation of inference";
          this.workloadTypesOptionsMap.inference.informationText = "";
        }
      }
    } else if (this.readOnly) {
      Object.values(this.workloadTypesOptionsMap).forEach((option) => {
        option.disable = true;
        option.toolTip = "This can't be modified when editing an environment";
      });
    }
    if (this.workloadTypesOptionsMap.inference.disable) {
      this.workloadTypesOptionsMap.inference.toolTipWidth = "300px";
    }
  },
  computed: {
    workloadTypeOptions(): Array<IWorkloadTypesOption> {
      const workloadTypeOptions = deepCopy(this.workloadTypesOptionsMap);

      if (this.workloadTypesSectionModel.distributed && !this.lockedWorkloadTypes) {
        workloadTypeOptions.workspace.disable = true;
        workloadTypeOptions.workspace.toolTip = "Workspaces can only run as a standard workload";
      }

      return Object.values(workloadTypeOptions);
    },
    workloadTypesModel(): Record<Exclude<EWorkloadName, "distributed-inference">, boolean> {
      return {
        workspace: this.workloadTypesSectionModel.workspace || false,
        training: this.workloadTypesSectionModel.training || false,
        inference: this.workloadTypesSectionModel.inference || false,
        distributed: this.workloadTypesSectionModel.distributed || false,
      };
    },
    distributedModel(): IDistributedSelectionModel {
      return {
        isDistributed: !!this.workloadTypesSectionModel.distributed,
        distFramework: this.workloadTypesSectionModel.distFramework || null,
      };
    },
    sectionInvalid(): boolean {
      if (
        !this.workloadTypesSectionModel.distributed &&
        (this.workloadTypesSectionModel.workspace ||
          this.workloadTypesSectionModel.training ||
          this.workloadTypesSectionModel.inference)
      )
        return false;
      return (
        (this.workloadTypesSectionModel.distributed && !this.workloadTypesSectionModel.distFramework) ||
        (!this.workloadTypesSectionModel.workspace &&
          !this.workloadTypesSectionModel.training &&
          !this.workloadTypesSectionModel.inference)
      );
    },
    summary(): string {
      const { workspace, training, distributed, distFramework, inference } = this.workloadTypesSectionModel;
      const workloadTypes = [];
      if (workspace) workloadTypes.push("Workspace");
      if (training) workloadTypes.push("Training");
      if (inference) workloadTypes.push("Inference");
      const workloadArchType = distributed ? `Distributed${distFramework ? ` (${distFramework})` : ""}` : "Standard";
      return `${workloadArchType} - ${workloadTypes.join(" & ")}`;
    },
  },
  methods: {
    updateWorkloadTypes(selectedWorkloadTypes: Record<Exclude<EWorkloadName, "distributed-inference">, boolean>): void {
      let { workspace, training, inference } = selectedWorkloadTypes;
      if (this.workloadTypesSectionModel.distributed) {
        const selectedCount = [workspace, training, inference].filter(Boolean).length;
        if (selectedCount > 1) {
          const prev = this.workloadTypesSectionModel;
          const workloadKeys: Array<EWorkloadName> = [
            EWorkloadName.Workspace,
            EWorkloadName.Training,
            EWorkloadName.Inference,
          ];
          const newlySelected = workloadKeys.find(
            (k) =>
              selectedWorkloadTypes[k as Exclude<EWorkloadName, "distributed-inference">] &&
              !prev[k as Exclude<EWorkloadName, "distributed-inference">],
          );
          const only =
            newlySelected ??
            workloadKeys.find((k) => selectedWorkloadTypes[k as Exclude<EWorkloadName, "distributed-inference">]);
          workspace = only === EWorkloadName.Workspace;
          training = only === EWorkloadName.Training;
          inference = only === EWorkloadName.Inference;
        }
      }
      const payload = {
        ...this.workloadTypesSectionModel,
        workspace,
        training,
        inference,
      };
      if (this.workloadTypesSectionModel.distributed && inference && !training) {
        payload.distFramework = undefined;
      }
      this.$emit("update-workload-types-section-model", payload);
    },
    updateDistributed(distributedModel: IDistributedSelectionModel): void {
      const { isDistributed, distFramework } = distributedModel;
      const { training, inference } = this.workloadTypesSectionModel;
      const bothOrNeither = (training ?? false) === (inference ?? false);

      this.$emit("update-workload-types-section-model", {
        ...this.workloadTypesSectionModel,
        workspace: !isDistributed,
        training: !isDistributed || !bothOrNeither ? (training ?? false) : true,
        inference: !isDistributed || !bothOrNeither ? (inference ?? false) : false,
        distributed: isDistributed,
        distFramework: distFramework ?? null,
      });
    },
    updateDistFramework(distFramework: WorkloadSupportedTypesDistFrameworkEnum): void {
      this.$emit("update-workload-types-section-model", {
        ...this.workloadTypesSectionModel,
        distFramework,
      });
    },
  },
});
</script>
