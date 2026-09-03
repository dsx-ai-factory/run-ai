<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="host-path-create-form column items-center q-pt-md">
    <runai-form-wrapper ref="elHostPathForm" class="host-path-form" :form-state="hostPathModel" :enter-callback="save">
      <runai-expansion-wrapper>
        <data-source-type-section data-source-kind="hostPath" />

        <runai-scope-section
          kind-of-asset="data source"
          :read-only="!isScopeSelectable"
          :scope-model="scopeModel"
          @changed-scope="updateScope"
          :key="selectedCluster"
          :selected-cluster-id="selectedCluster"
          :allowed-scopes="allowedScopes"
        />
        <runai-name-section
          label="Data source"
          asset-kind="hostPath"
          v-model="hostPathModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="hostPathModel.meta.description"
          @update-description="updateDescription"
        />
        <transition name="fade">
          <host-path-data-mount-section v-if="isScopeSelected" v-model="hostPathModel.spec" />
        </transition>
      </runai-expansion-wrapper>

      <section class="row items-center q-mt-md">
        <q-field class="col-4 form-hint no-padding" :model-value="displayFormHint" :rules="[isFormIncomplete]"></q-field>
        <div class="buttons q-ml-auto">
          <runai-btn aid="cancel-btn" class="q-mr-md" @click="onCancel" flat color="primary" label="cancel" />
          <runai-btn
            color="primary"
            aid="submit-btn"
            @click="save"
            :loading="validating || submitting"
            label="create data source"
          />
        </div>
      </section>
    </runai-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { HostPathDataMountSection } from "@/components/data-source/host-path/host-path-data-mount-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiBtn } from "@/components/common/runai-btn";

// models
import type { IScopeModel } from "@/models/global.model";
import type { IUIHostPathCreationRequest } from "@/models/data-source.model";

// services
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { DataSourceTypeSection } from "../../sections/data-source-type-section";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiExpansionWrapper,
    RunaiNameSection,
    HostPathDataMountSection,
    DataSourceTypeSection,
    RunaiScopeSection,
    RunaiBtn,
  },
  emits: ["submit", "cancel"],
  props: {
    submitting: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    dataSourceModel: {
      type: Object as PropType<IUIHostPathCreationRequest>,
      required: true,
    },
    isScopeSelectable: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: true,
    },
  },
  data() {
    return {
      hostPathForm: null as HTMLFormElement | null,
      hostPathModel: deepCopy(this.dataSourceModel) as IUIHostPathCreationRequest,
      isFormInvalid: false as boolean,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      selectedCluster: "" as string,
      validating: false as boolean,
    };
  },
  mounted() {
    this.hostPathForm = this.$refs["elHostPathForm"] as HTMLFormElement;
  },
  computed: {
    isScopeSelected(): boolean {
      return !!this.scopeModel.scope;
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.hostPathModel.meta.scope,
        projectId: this.hostPathModel.meta.projectId,
        departmentId: this.hostPathModel.meta.departmentId,
        clusterId: this.hostPathModel.meta.clusterId,
      };
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.hostPathModel.meta.description = description;
    },
    updateScope(updatedScope: IScopeModel): void {
      this.hostPathModel.meta = scopeUtil.updateScope(this.hostPathModel.meta, updatedScope);
    },
    onCancel(): void {
      this.$emit("cancel");
    },
    async save(): Promise<void> {
      try {
        this.validating = true;
        this.displayFormHint = false;
        const success = await this.validate();
        if (!success) {
          this.showHint();
          return;
        }
        this.$emit("submit", this.hostPathModel);
      } catch (e: unknown) {
        console.error(e);
      } finally {
        this.validating = false;
      }
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.hostPathForm as HTMLFormElement).validate();
      return isValid;
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = setTimeout(() => (this.displayFormHint = false), 15000);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : errorMessages.FORM_INCOMPLETE;
    },
  },
  unmounted() {
    this.timeOutId && clearTimeout(this.timeOutId);
  },
});
</script>

<style lang="scss" scoped></style>
