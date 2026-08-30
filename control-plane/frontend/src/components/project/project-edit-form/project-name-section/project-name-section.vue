<template>
  <runai-expansion-item
    label="Project name & description"
    default-opened
    :section-invalid="sectionInvalid"
    disable-closing
  >
    <template #subheader> {{ expansionSubHeader }}</template>
    <!-- PROJECT NAME -->
    <div class="row items-center">
      <div class="col-7">
        <runai-name-validation
          aid="project-name-input"
          :rules="[isValidEntityName, isNameUnique]"
          :model-value="projectName"
          @update:model-value="$emit('update:projectName', $event)"
          :disable="!isNewProject"
        />
        <q-tooltip v-if="!isNewProject" anchor="center middle" self="center middle" :delay="300">
          {{ projectCannotBeModifiedTooltip }}
        </q-tooltip>
      </div>
      <div class="col-1">
        <runai-tooltip
          tooltip-position="right"
          width="320px"
          tooltip-text="Each project is associated with a namespace. It can be determine whether it will be automatically created with the
      same name as the project name, or will use an existing one. Each namespace will be associated with only one
      project"
        />
      </div>
    </div>
    <!-- PROJECT DESCRIPTION -->
    <div class="row q-mt-md">
      <div class="col-11">
        <q-input
          aid="project-description-input"
          :model-value="description"
          @update:model-value="$emit('update:description', $event)"
          :maxlength="250"
          counter
          label="Description"
          autogrow
        />
      </div>
    </div>
    <!-- PROJECT NAMESPACE -->
    <runai-sub-expansion-item aid="namespace-section-toggle" label="Namespace">
      <q-item-label class="q-my-md">Set the namespace associated to the project</q-item-label>
      <div>
        <div class="row items-center">
          <q-radio
            :disable="!isNewProject"
            v-model="selectedNamespaceCreationMethod"
            aid="project-namespace-radio"
            label="Create from the project name"
            :val="namespaceCreationMethod.PROJECT_NAME"
          />
          <runai-tooltip
            tooltip-position="right"
            tooltip-text="When a new project is created, a new namespace is created in the Kubernetes
          cluster using the project's name with the run:ai prefix."
          />
        </div>
        <div class="row items-center">
          <q-radio
            :disable="!isNewProject"
            v-model="selectedNamespaceCreationMethod"
            aid="existing-namespace-radio"
            label="Enter existing namespace from the cluster"
            :val="namespaceCreationMethod.EXISTING_NAMESPACE"
          />
          <runai-tooltip
            tooltip-position="right"
            tooltip-text="To use an existing Kubernetes namespace instead of the project name, enter that
        namespace here. The namespace can only be associated with one project and cannot be one of the Kubernetes global
        system default namespaces."
          />
        </div>
        <q-tooltip v-if="!isNewProject" anchor="center middle" self="center right">{{
          projectCannotBeModifiedTooltip
        }}</q-tooltip>
      </div>
      <div class="row items-center q-mt-lg">
        <div class="col-7">
          <q-input
            stack-label
            no-error-icon
            aid="existing-namespace-input"
            :rules="[isValidKubernetesNamespaceName]"
            label="Namespace"
            :disable="isNamespaceEditable"
            v-model="assignedNamespace"
            :ref="existingNamespaceInputRef"
          />
          <q-tooltip v-if="isProjectNameBasedNamespace" anchor="center right"
            >The namespace can only be modified if it is an existing namespace from the cluster.</q-tooltip
          >
          <q-tooltip v-if="!isNewProject" anchor="center middle" self="bottom start">{{
            projectCannotBeModifiedTooltip
          }}</q-tooltip>
        </div>
      </div>

      <template v-if="!isProjectNameBasedNamespace">
        <div class="q-pb-md q-pt-lg">
          Use the following to associate the namespace with the new project in the terminal interface
        </div>
        <runai-copyable-code-block :code="labelNamespaceCommand" />
      </template>
    </runai-sub-expansion-item>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

// cmps
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiNameValidation } from "@/components/common/runai-name-validation";
import { RunaiSubExpansionItem } from "@/components/common/runai-sub-expansion-item";
import RunaiCopyableCodeBlock from "@/components/common/runai-copyable-code-block/runai-copyable-code-block.vue";
// services

import { isValidEntityName, isValidKubernetesNamespaceName } from "@/common/form.validators";

// models
import { errorMessages } from "@/common/error-message.constant";
import { useClusterStore } from "@/stores/cluster.store";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { ErrorAlert } from "@/utils/error-alert.util";
import { EOrgUnitFormTooltip } from "@/models/org-unit-forms.model";
export enum ENamespaceCreationMethod {
  PROJECT_NAME = "projectName",
  EXISTING_NAMESPACE = "existingNamespace",
}
export default defineComponent({
  components: {
    RunaiCopyableCodeBlock,
    RunaiSubExpansionItem,
    RunaiTooltip,
    RunaiNameValidation,
    RunaiExpansionItem,
  },
  emits: ["update:namespace", "update:projectName", "is-section-invalid", "update:description"],
  props: {
    projectName: {
      type: String as PropType<string>,
      required: true,
    },
    description: {
      type: String as PropType<string>,
      required: false,
    },
    namespace: {
      type: [String] as PropType<string>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    isNewProject: {
      type: Boolean,
      required: true,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      selectedNamespaceCreationMethod: ENamespaceCreationMethod.PROJECT_NAME as string,
      projectNameSectionForm: {} as HTMLFormElement,
      assignedNamespace: "" as string,
      isProjectNameUnique: true as boolean,
      existingNamespaceInputRef: "existing-namespace-input" as string,
    };
  },
  created() {
    this.initNamespaceSection();
  },
  computed: {
    labelNamespaceCommand(): string {
      if (this.projectName && this.namespace) {
        return `kubectl label ns ${this.namespace} runai/queue=${this.projectName}`;
      }
      return "kubectl label ns <NAMESPACE> runai/queue=<PROJECT_NAME>";
    },
    namespaceCreationMethod(): typeof ENamespaceCreationMethod {
      return ENamespaceCreationMethod;
    },
    isNamespaceEditable(): boolean {
      return this.isProjectNameBasedNamespace || !this.isNewProject;
    },
    sectionInvalid(): boolean {
      const isProjectNameInvalid = !isValidEntityName(this.projectName);
      const isNamespaceInvalid = !this.isProjectNameBasedNamespace && !this.namespace;

      return isProjectNameInvalid || isNamespaceInvalid || !this.isProjectNameUnique;
    },
    isProjectNameBasedNamespace(): boolean {
      return this.selectedNamespaceCreationMethod === ENamespaceCreationMethod.PROJECT_NAME;
    },
    expansionSubHeader(): string {
      if (isValidEntityName(this.projectName)) {
        return this.projectName;
      }
      return "None";
    },
    projectCannotBeModifiedTooltip(): string {
      return EOrgUnitFormTooltip.ProjectCannotBeModified;
    },
  },
  methods: {
    initNamespaceSection(): void {
      if (this.namespace) {
        this.assignedNamespace = this.namespace;
        this.selectedNamespaceCreationMethod = ENamespaceCreationMethod.EXISTING_NAMESPACE;
      } else if (isValidEntityName(this.projectName)) {
        this.assignedNamespace = `runai-${this.projectName}`;
        this.selectedNamespaceCreationMethod = ENamespaceCreationMethod.PROJECT_NAME;
      }
    },
    isValidEntityName(name: string): boolean | string {
      return isValidEntityName(name) || errorMessages.VALID_FORMAT;
    },
    isValidKubernetesNamespaceName(name: string): boolean | string {
      return isValidKubernetesNamespaceName(name) || errorMessages.VALID_NAMESPACE;
    },
    updateNamespaceByProjectName(): void {
      if (!this.isNewProject) return;
      if (this.isProjectNameBasedNamespace && isValidEntityName(this.projectName)) {
        this.assignedNamespace = `runai-${this.projectName}`;
        this.$emit("update:namespace", undefined);
      } else {
        this.assignedNamespace = "";
        this.$emit("update:namespace", this.assignedNamespace);
      }
    },
    async isNameUnique(name: string): Promise<boolean | string> {
      try {
        this.isProjectNameUnique = await orgUnitService.isProjectNameUnique(this.clusterId, name);
        if (!this.isProjectNameUnique) {
          return errorMessages.NAME_ALREADY_EXIST;
        }
        return true;
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedToValidateMessage("project name"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        return false;
      }
    },
  },
  watch: {
    projectName() {
      if (this.isProjectNameBasedNamespace && isValidEntityName(this.projectName)) {
        this.assignedNamespace = `runai-${this.projectName}`;
        this.$emit("update:namespace", undefined);
      } else {
        this.$emit("update:namespace", this.assignedNamespace);
      }
    },
    isProjectNameBasedNamespace() {
      this.updateNamespaceByProjectName();
    },
    assignedNamespace() {
      if (this.isProjectNameBasedNamespace && isValidEntityName(this.projectName)) {
        this.$emit("update:namespace", undefined);
      } else {
        this.$emit("update:namespace", this.assignedNamespace);
      }
    },
    selectedNamespaceCreationMethod(method: ENamespaceCreationMethod) {
      if (method === ENamespaceCreationMethod.EXISTING_NAMESPACE) {
        this.$nextTick(() => {
          //wait for disable attr to be removed
          const element = this.$refs[this.existingNamespaceInputRef] as HTMLInputElement | null;
          if (element) {
            element.focus();
          }
        });
      }
    },
    sectionInvalid: {
      handler(isSectionInvalid: boolean): void {
        this.$emit("is-section-invalid", isSectionInvalid);
      },
      immediate: true,
    },
  },
});
</script>
