<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="data-volume-create-form column items-center q-pt-md">
    <runai-form-wrapper ref="elDataVolumeForm" :form-state="dataVolumeModel" :enter-callback="save">
      <data-source-type-section data-source-kind="data-volume" />
      <data-origin-section
        :model-value="originSectionModel"
        :pvcs="availablePVCs"
        :loading-pvcs="loadingPvcs"
        :allowed-scopes="allowedScopes"
        @update:model-value="updateOriginSectionModel"
        :disable="editMode"
      />

      <runai-name-section
        label="Data volume"
        asset-kind="data-volume"
        v-model="dataVolumeModel.meta.name"
        :description="dataVolumeModel.meta.description"
        @update-description="dataVolumeModel.meta.description = $event"
        :scope-model="{ scope: 'project' }"
        support-description
        :disable-name-input="editMode"
        :name-input-tooltip="editMode ? 'this can’t be modified when editing an existing data volume' : undefined"
        :custom-existing-name-validation="isUniqueName"
      />

      <transition name="fade">
        <scopes-section
          v-if="parentClusterId"
          v-model="dataVolumeModel.spec.sharedScopes"
          :allowed-scopes="allowedScopes"
          :parent-cluster-id="parentClusterId"
          allow-empty-selection
        />
      </transition>

      <section class="row items-center q-mt-md">
        <q-field class="col-4 form-hint no-padding" :model-value="displayFormHint" :rules="[isFormIncomplete]"></q-field>
        <div class="buttons q-ml-auto">
          <runai-btn aid="cancel-btn" class="q-mr-md" @click="onCancel" flat color="primary" label="cancel" />
          <runai-btn
            color="primary"
            aid="data-volume-submit-btn"
            @click="save"
            :loading="submitting"
            :label="`${editMode ? 'save' : 'create'} data volume`"
          />
        </div>
      </section>
    </runai-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { DataSourceTypeSection } from "@/components/data-source/sections/data-source-type-section";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { DataOriginSection } from "../data-origin-section";
import { ScopesSection } from "../scopes-section";
import { RunaiBtn } from "@/components/common/runai-btn";
// Services
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
// Models
import { ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";
import type { PVCListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import type { IDataVolumeCreationRequest } from "@/models/data-source.model";
import type { IDataOriginModel } from "../data-origin-section/data-origin-section.vue";
// Utils
import { alertUtil } from "@/utils/alert.util";
import { deepCopy } from "@/utils/common.util";
// Constants
import { errorMessages } from "@/common/error-message.constant";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { useAuthStore } from "@/stores/auth.store";
import { dataVolumeService } from "@/services/control-plane/data-volume.service/data-volume.service";
import type { ISelectOption, IValidationRule } from "@/models/global.model";

export default defineComponent({
  name: "data-volume-create-form",
  components: {
    RunaiFormWrapper,
    DataSourceTypeSection,
    RunaiNameSection,
    DataOriginSection,
    ScopesSection,
    RunaiBtn,
  },
  emits: {
    submit: (dataVolumeModel: IDataVolumeCreationRequest) => dataVolumeModel,
    cancel: () => true,
  },
  props: {
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: true,
    },
    submitting: {
      type: Boolean,
      required: true,
    },
    model: {
      type: Object as PropType<IDataVolumeCreationRequest>,
      required: true,
    },
    editMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      dataVolumeModel: deepCopy(this.model),
      existingPVCs: [] as Array<PVCListInfo>,
      usedPVCs: [] as Array<string>,
      loadingPvcs: false as boolean,
      displayFormHint: false as boolean,
      timeOutId: 0 as number,
      authStore: useAuthStore(),
      isUniqueName: (() => true) as IValidationRule,
    };
  },
  created() {
    if (this.dataVolumeModel.meta.projectId) {
      this.loadPvcs();
    }
  },
  computed: {
    originSectionModel(): IDataOriginModel {
      return {
        projectId: this.dataVolumeModel.meta.projectId,
        originPvcName: this.dataVolumeModel.spec.originPvcName,
      };
    },
    parentClusterId(): string | undefined {
      return (
        orgTreeService.getParentClusterUuid(
          this.authStore.orgUnitList,
          this.dataVolumeModel.meta.projectId,
          "project",
        ) || undefined
      );
    },
    availablePVCs(): ISelectOption[] {
      const usedPVCsMap = new Set(this.usedPVCs);

      return this.existingPVCs.reduce((acc, pvc) => {
        const option: ISelectOption = {
          label: pvc.name,
          value: pvc.name,
        };

        if (usedPVCsMap.has(pvc.name)) {
          option.disable = true;
          option.disabledTooltip = "This can't be selected because it is already in use by another data volume";
        }

        acc.push(option);
        return acc;
      }, [] as ISelectOption[]);
    },
  },
  methods: {
    updateOriginSectionModel(originSectionModel: IDataOriginModel): void {
      this.dataVolumeModel.meta.projectId = originSectionModel.projectId;
      this.dataVolumeModel.spec.originPvcName = originSectionModel.originPvcName;
    },
    async loadPvcs(): Promise<void> {
      try {
        this.loadingPvcs = true;
        await Promise.all([this.loadExistingPvcs(), this.loadUsedPvcs()]);
      } catch (err) {
        console.error(err);
        this.$q.notify(alertUtil.getError("Failed to load pvcs"));
        throw err;
      } finally {
        this.loadingPvcs = false;
      }
    },
    async loadExistingPvcs(): Promise<void> {
      if (!this.dataVolumeModel.meta.projectId) return;
      try {
        const existingPvcs = await dataSourceService.listExistingPVCs(
          ScopeType.Project,
          this.dataVolumeModel.meta.projectId,
          "all",
          true,
        );
        this.existingPVCs = existingPvcs;
      } catch (err) {
        console.error(err);
        this.$q.notify(alertUtil.getError("Failed to load existing pvcs"));
        throw err;
      }
    },
    async loadUsedPvcs(): Promise<void> {
      if (!this.dataVolumeModel.meta.projectId) return;
      try {
        const usedPvcs = await dataVolumeService.listUsedPvcs(this.dataVolumeModel.meta.projectId);
        this.usedPVCs = usedPvcs;
      } catch (err) {
        console.error(err);
        this.$q.notify(alertUtil.getError("Failed to load used PVCs"));
        throw err;
      }
    },
    onCancel(): void {
      this.$emit("cancel");
    },
    async save(): Promise<void> {
      this.displayFormHint = false;
      const success = await this.validate();
      if (!success) {
        this.showHint();
        return;
      }
      this.$emit("submit", this.dataVolumeModel);
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.$refs.elDataVolumeForm as HTMLFormElement).validate();
      return isValid;
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = window.setTimeout(() => (this.displayFormHint = false), 15000);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : errorMessages.FORM_INCOMPLETE;
    },
  },
  watch: {
    "dataVolumeModel.meta.projectId": {
      handler(): void {
        this.loadPvcs();
        this.dataVolumeModel.spec.sharedScopes = [];
        this.dataVolumeModel.spec.originPvcName = "";
        if (this.parentClusterId) {
          this.isUniqueName = dataVolumeService.getNameValidationFunc(this.parentClusterId);
        }
      },
    },
  },
});
</script>
