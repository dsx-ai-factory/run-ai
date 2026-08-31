<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="yaml-workload-section"
    label="YAML"
    aid="yaml-workload-section"
    default-opened
    body-no-padding
    :section-invalid="!isYamlValid"
  >
    <template #subheader>
      <span>{{ subheaderText }}</span>
    </template>

    <div class="yaml-section-content q-pa-md">
      <div class="q-mb-sm">Upload or paste a workload YAML</div>

      <runai-btn
        class="upload-btn q-mb-md q-mt-lg btn-icon-sm"
        icon="ph-bold ph-upload"
        label="Upload YAML file"
        aid="upload-yaml-btn"
        color="grey-4"
        text-color="b-main"
        @click="triggerFileInput"
      />
      <input
        ref="fileInput"
        type="file"
        accept=".yaml,.yml"
        aria-label="Upload YAML file"
        style="display: none"
        @change="onFileSelected"
      />

      <div class="editor-wrapper" :style="{ height: editorHeight }">
        <runai-editor
          ref="yamlEditorRef"
          :code="yamlContent"
          :height="editorHeight"
          show-expand-button
          :is-expanded="isExpanded"
          show-placeholder
          placeholder="Or paste your YAML content here"
          @changed="onYamlChanged"
          @validation-status="onValidationStatusChanged"
          @toggle-expand="toggleExpand"
          @pasted="onYamlPasted"
        />
      </div>

      <div class="validation-results row justify-between items-center q-mt-md">
        <div class="validation-hint text-negative">
          <span v-if="!isYamlValid && yamlContent">Add a valid Kubernetes YAML</span>
        </div>
        <div
          class="supported-types-wrapper"
          role="button"
          tabindex="0"
          aria-haspopup="true"
          :aria-expanded="showTypesMenu"
          @mouseenter="showTypesMenu = true"
          @mouseleave="showTypesMenu = false"
          @focusin="showTypesMenu = true"
          @focusout="showTypesMenu = false"
        >
          <span class="supported-types-link text-black-54" aid="supported-workload-types-link">
            Supported workload types
          </span>
          <q-menu
            v-model="showTypesMenu"
            anchor="top right"
            self="bottom right"
            :offset="[0, 4]"
            no-parent-event
            @mouseenter="showTypesMenu = true"
            @mouseleave="showTypesMenu = false"
            @focusin="showTypesMenu = true"
            @focusout="showTypesMenu = false"
          >
            <q-list dense style="min-width: 180px; max-height: 300px; overflow-y: auto">
              <template v-if="loadingWorkloadTypes">
                <q-item>
                  <q-item-section>
                    <q-spinner size="sm" />
                  </q-item-section>
                </q-item>
              </template>
              <template v-else-if="sortedWorkloadTypes.length > 0">
                <q-item v-for="type in sortedWorkloadTypes" :key="type.id" dense>
                  <q-item-section>{{ type.name }}</q-item-section>
                </q-item>
              </template>
              <template v-else>
                <q-item>
                  <q-item-section class="text-grey">No workload types available</q-item-section>
                </q-item>
              </template>
            </q-list>
          </q-menu>
        </div>
      </div>
    </div>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { debounce } from "quasar";

// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiEditor } from "@/components/common/runai-editor";
import { RunaiBtn } from "@/components/common/runai-btn";

// Utils
import { yamlUtil } from "@/utils/yaml.util";
import { fileUtil } from "@/utils/file.util";
import { alertUtil } from "@/utils/alert.util";
import { extractErrorMessages } from "@/utils/error-alert.util";

// Services
import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

// Models
import type { IEditorValidationStatus } from "@/components/common/runai-editor/runai-editor.model";
import type { WorkloadTypeConfig } from "@/swagger-models/workloads-client";
import type { IYamlDryRunFields } from "@/utils/yaml.util";
import type { WorkloadV2CreateRequest, Preemptibility } from "@/swagger-models/workloads-manager-client";

// Interface for selected form values from parent component
export interface ISelectedFormValues {
  projectId: number | null;
  priority: string | null;
  preemptibility: string | null;
}

const MIN_HEIGHT = "300px";
const MAX_HEIGHT = "700px";

export default defineComponent({
  name: "yaml-workload-section",
  components: {
    RunaiExpansionItem,
    RunaiEditor,
    RunaiBtn,
  },
  emits: [
    "yaml-input", // Fires on every keystroke
    "yaml-loaded", // Fires on file upload or paste (bulk content load)
    "validation-changed", // Fires when YAML syntax validation status changes
    "dryrun-success", // Fires when dry-run validation succeeds (errors handled internally)
  ],
  props: {
    yaml: {
      type: String as PropType<string>,
      default: "",
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    // Selected values from other sections to include in validation
    selectedFormValues: {
      type: Object as PropType<ISelectedFormValues>,
      default: () => ({ projectId: null, priority: null, preemptibility: null }),
    },
  },
  data() {
    return {
      yamlContent: this.yaml,
      isYamlValid: false,
      validationError: "" as string,
      isExpanded: false,
      workloadTypes: [] as WorkloadTypeConfig[],
      loadingWorkloadTypes: false,
      showTypesMenu: false,
      // Dry-run validation state
      dryRunError: "" as string,
      isDryRunValid: true,
      previousDryRunFields: { name: null, apiVersion: null, kind: null, namespace: null } as IYamlDryRunFields,
      debouncedRunDryRunValidation: (() => {}) as (content: string) => void,
      // Flag to prevent initialYaml watcher from re-triggering when change originated internally
      isInternalUpdate: false,
    };
  },
  async created() {
    // Debounce the dry-run validation to avoid excessive API calls while typing
    this.debouncedRunDryRunValidation = debounce(this.runDryRunValidation, 500);
    await this.loadWorkloadTypes();

    // Validate initial content if yaml prop was provided (e.g., from restored state)
    if (this.yaml) {
      this.validateYaml(this.yaml);
      // Trigger dry-run validation for restored content
      this.previousDryRunFields = { name: null, apiVersion: null, kind: null, namespace: null };
      this.checkDryRunFieldsChanged(this.yaml);
    }
  },
  computed: {
    editorHeight(): string {
      return this.isExpanded ? MAX_HEIGHT : MIN_HEIGHT;
    },
    subheaderText(): string {
      if (!this.yamlContent) return "Name: None";
      const name = yamlUtil.extractWorkloadName(this.yamlContent);
      return `Name: ${name || "None"}`;
    },
    sortedWorkloadTypes(): WorkloadTypeConfig[] {
      return [...this.workloadTypes].sort((a, b) => (a.name || "").localeCompare(b.name || ""));
    },
  },
  methods: {
    setEditorContent(content: string): void {
      // Set Monaco content imperatively without remounting (no blink).
      // `runai-editor` exposes `setContent` for this purpose.
      (this.$refs.yamlEditorRef as InstanceType<typeof RunaiEditor>)?.setContent(content);
    },
    triggerFileInput(): void {
      (this.$refs.fileInput as HTMLInputElement).click();
    },
    async onFileSelected(event: Event): Promise<void> {
      const input = event.target as HTMLInputElement;
      const file = input.files?.[0];
      if (!file) return;

      // Validate file extension
      const validExtensions = [".yaml", ".yml"];
      const fileExtension = file.name.toLowerCase().slice(file.name.lastIndexOf("."));
      if (!validExtensions.includes(fileExtension)) {
        console.error("Invalid file type. Only .yaml and .yml files are allowed.");
        input.value = "";
        return;
      }

      try {
        const content = await fileUtil.readAsText(file);
        // Mark as internal update to prevent watcher from re-triggering setEditorContent
        this.isInternalUpdate = true;
        this.yamlContent = content;
        this.setEditorContent(content);
        this.validateYaml(content);
        this.$emit("yaml-input", content);
        this.$emit("yaml-loaded", content);
        // Reset and trigger dry-run validation for uploaded YAML
        this.previousDryRunFields = { name: null, apiVersion: null, kind: null, namespace: null };
        this.checkDryRunFieldsChanged(content);
        // Reset flag after Vue processes the prop update
        this.$nextTick(() => {
          this.isInternalUpdate = false;
        });
      } catch (error) {
        console.error("Error reading file:", error);
      }

      // Reset input so the same file can be selected again
      input.value = "";
    },
    onYamlChanged(content: string): void {
      // Mark as internal update to prevent watcher from re-triggering
      this.isInternalUpdate = true;
      this.yamlContent = content;
      this.$emit("yaml-input", content);
      this.checkDryRunFieldsChanged(content);
      // Reset flag after Vue processes the prop update
      this.$nextTick(() => {
        this.isInternalUpdate = false;
      });
    },
    onYamlPasted(content: string): void {
      this.$emit("yaml-loaded", content);
      // Reset and trigger dry-run validation for new YAML
      this.previousDryRunFields = { name: null, apiVersion: null, kind: null, namespace: null };
      this.checkDryRunFieldsChanged(content);
    },
    onValidationStatusChanged(status: IEditorValidationStatus): void {
      this.isYamlValid = status.isValid;
      this.validationError = status.error || "";
      this.$emit("validation-changed", status.isValid);
    },
    validateYaml(content: string): void {
      const error = yamlUtil.validateYaml(content);
      this.isYamlValid = !error;
      this.validationError = error || "";
      this.$emit("validation-changed", !error);
    },
    toggleExpand(): void {
      this.isExpanded = !this.isExpanded;
    },
    async loadWorkloadTypes(): Promise<void> {
      this.loadingWorkloadTypes = true;
      try {
        // externalTypesOnly=true filters out Run:ai types, showing only external workload types
        const response = await workloadServiceApi.workloadPropertiesApi.listWorkloadTypes(true, undefined, false);
        this.workloadTypes = response.data.types || [];
      } catch (error) {
        console.error("Failed to load workload types:", error);
        this.workloadTypes = [];
      } finally {
        this.loadingWorkloadTypes = false;
      }
    },
    checkDryRunFieldsChanged(content: string): void {
      if (!this.isYamlValid || !content) {
        this.clearDryRunValidation();
        return;
      }

      const currentFields = yamlUtil.extractDryRunFields(content);
      const fieldsToCompare = ["name", "apiVersion", "kind", "namespace"] as const;
      const hasFieldsChanged = fieldsToCompare.some(
        (field) => currentFields[field] !== this.previousDryRunFields[field],
      );

      if (hasFieldsChanged) {
        this.previousDryRunFields = { ...currentFields };
        this.debouncedRunDryRunValidation(content);
      }
    },
    async runDryRunValidation(content: string): Promise<void> {
      // Need clusterId and valid YAML to run validation
      if (!this.clusterId || !this.isYamlValid || !content) {
        this.clearDryRunValidation();
        return;
      }

      const manifest = yamlUtil.yamlToObject(content);
      if (!manifest) {
        return;
      }

      const workloadName = yamlUtil.extractWorkloadName(content);
      if (!workloadName) {
        return;
      }

      this.dryRunError = "";

      try {
        const { projectId, priority, preemptibility } = this.selectedFormValues;
        const hasValidProject = projectId && projectId > 0;

        const request: WorkloadV2CreateRequest = {
          metadata: {
            name: workloadName,
            projectId: hasValidProject ? String(projectId) : undefined,
            clusterId: hasValidProject ? undefined : this.clusterId,
            priority: priority || undefined,
            preemptibility: (preemptibility as Preemptibility) || undefined,
          },
          manifest: manifest as Record<string, unknown>,
        };

        const workload = await workloadService.createWorkloadFromYaml(request, true, true);
        this.isDryRunValid = true;
        this.dryRunError = "";
        this.$emit("dryrun-success", workload);
      } catch (error) {
        console.error("Dry-run validation error:", error);
        this.isDryRunValid = false;

        const errorMessages = extractErrorMessages(error, "Validation failed");
        this.dryRunError = errorMessages.join("\n");

        errorMessages.forEach((msg: string) => {
          this.$q.notify({ ...alertUtil.getError(msg), timeout: 4500 });
        });
      }
    },
    clearDryRunValidation(): void {
      this.dryRunError = "";
      this.isDryRunValid = true;
    },
  },
  watch: {
    yaml: {
      handler(newVal: string): void {
        // Skip if change originated from this component (e.g., file upload)
        if (this.isInternalUpdate) return;

        if (newVal !== this.yamlContent) {
          this.yamlContent = newVal;
          this.setEditorContent(newVal);
          if (newVal) {
            this.validateYaml(newVal);
          }
        }
      },
      immediate: true,
    },
    clusterId: {
      handler(): void {
        if (this.yamlContent && this.isYamlValid) {
          this.runDryRunValidation(this.yamlContent);
        }
      },
    },
    // Trigger validation when project or general section values change
    selectedFormValues: {
      handler(): void {
        if (this.yamlContent && this.isYamlValid) {
          this.debouncedRunDryRunValidation(this.yamlContent);
        }
      },
      deep: true,
    },
  },
});
</script>

<style lang="scss" scoped>
.yaml-workload-section {
  .yaml-section-content {
    .editor-wrapper {
      position: relative;
      border-radius: 4px;
      overflow: hidden;
      transition: height 0.3s ease;
    }

    .validation-hint {
      font-size: 12px;
    }

    .upload-btn {
      :deep(.q-icon) {
        font-size: 20px;
      }
    }

    .supported-types-wrapper {
      position: relative;
      padding: 7px;
      border-radius: 4px;
      transition: background-color 0.2s ease;
      color: $black-87;
      min-width: fit-content;
      &:hover {
        background-color: #eee;
      }
    }

    .supported-types-link {
      font-size: 14px;
      cursor: default;
    }
  }
  .validation-results {
    flex-wrap: nowrap;
    gap: 24px;
  }
}
</style>
